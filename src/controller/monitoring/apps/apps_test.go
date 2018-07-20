/*******************************************************************************
 * Copyright 2018 Samsung Electronics All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *******************************************************************************/
package apps

import (
	"commons/errors"
	"controller/dockercontroller"
	dockermocks "controller/dockercontroller/mocks"
	dbmocks "db/bolt/service/mocks"
	"github.com/golang/mock/gomock"
	"reflect"
	"testing"
)

const (
	serviceName             = "test_service"
	serviceName2            = "test_service2"
	oldTag                  = "1.0"
	repositoryWithPortImage = "test_url:5000/test"
	descriptionJson         = "{\"services\":{\"" + serviceName + "\":{\"image\":\"" + repositoryWithPortImage + ":" + oldTag + "\"},\"" +
		serviceName2 + "\":{\"image\":\"" + repositoryWithPortImage + ":" + oldTag + "\"}},\"version\":\"2\"}"

	appId = "test_app_id"
	path  = "test_path"
)

var (
	psWithUpObj = []map[string]string{
		{
			"State": "Up",
		},
	}
	psWithForcefullyExitedObj = []map[string]string{
		{
			"State": "Exited (137)",
		},
	}
	psWithGracefullyExitedObj = []map[string]string{
		{
			"State": "Exited (0)",
		},
	}

	dbGetAppObj = map[string]interface{}{
		"id":          appId,
		"state":       RUNNING_STATE,
		"description": descriptionJson,
		"images": []map[string]interface{}{
			{
				"name": repositoryWithPortImage,
				"changes": map[string]interface{}{
					"tag":   oldTag,
					"state": "update",
				},
			},
		},
	}
	unknownError = errors.Unknown{}
)

func TestEnableEventMonitoring_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Events(appId, path, gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	err := Executor{}.EnableEventMonitoring(appId, path)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestEnableEventMonitoringWhenFailedToSetEventChannel_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Events(appId, path, gomock.Any()).Return(unknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	err := Executor{}.EnableEventMonitoring(appId, path)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", err.Error())
	case errors.Unknown:
	}
}

func TestDisableEventMonitoring_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Events(appId, path, nil).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	err := Executor{}.DisableEventMonitoring(appId, path)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestDisableEventMonitoringWhenFailedToSetEventChannel_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Events(appId, path, nil).Return(unknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	err := Executor{}.DisableEventMonitoring(appId, path)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", err.Error())
	case errors.Unknown:
	}
}

func TestLockUpdateState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Executor{}.LockUpdateAppState()
}

func TestUnLockUpdateState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Executor{}.UnlockUpdateAppState()
}

func TestGetEventChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testChan := make(chan dockercontroller.Event)
	ret := Executor{}.GetEventChannel()

	if reflect.TypeOf(ret) != reflect.TypeOf(testChan) {
		t.Errorf("Expected type of ret : chan dockercontroller.Event, actual type of ret: %v", reflect.TypeOf(ret))
	}
}

func TestExtractStringInParenthesis(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedRet := "123"
	testStr := "(" + expectedRet + ") test string"
	ret := extractStringInParenthesis(testStr)

	if ret != expectedRet {
		t.Errorf("Expected result : %s, actual result : %s", expectedRet, ret)
	}
}

func TestUpdateAppStateWhenGetAppFailed_ExpectReturnAfterGetApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(appId).Return(nil, unknownError),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	testEvent := dockercontroller.Event{
		AppID: appId,
	}
	updateAppState(testEvent)
}

func TestUpdateAppstate_ExpectUpdateAppStateToPartiallyExited(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(appId).Return(dbGetAppObj, nil),
		dockerExecutorMockObj.EXPECT().Ps(appId, "docker-compose.yml", gomock.Any()).Return(psWithForcefullyExitedObj, nil),
		dockerExecutorMockObj.EXPECT().Ps(appId, "docker-compose.yml", gomock.Any()).Return(psWithUpObj, nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(appId, PARTIALLY_EXITED_STATE),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	testEvent := dockercontroller.Event{
		AppID: appId,
	}
	updateAppState(testEvent)
}

func TestUpdateAppstate_ExpectUpdateAppStateToExited(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(appId).Return(dbGetAppObj, nil),
		dockerExecutorMockObj.EXPECT().Ps(appId, "docker-compose.yml", gomock.Any()).Return(psWithForcefullyExitedObj, nil),
		dockerExecutorMockObj.EXPECT().Ps(appId, "docker-compose.yml", gomock.Any()).Return(psWithForcefullyExitedObj, nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(appId, EXITED_STATE),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	testEvent := dockercontroller.Event{
		AppID: appId,
	}
	updateAppState(testEvent)
}

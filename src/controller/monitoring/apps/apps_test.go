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
	dockermocks "controller/dockercontroller/mocks"
	"github.com/golang/mock/gomock"
	"testing"
)

const (
	appId = "test_app_id"
	path  = "test_path"
)

var (
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

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
	dbmocks "db/bolt/event/mocks"
	servicedbmocks "db/bolt/service/mocks"
	"github.com/golang/mock/gomock"
	"testing"
)

const (
	appId          = "test_app_id"
	path           = "compose_path"
	cid            = "container_id"
	status         = "event_status"
	serviceName    = "test_service_name"
	imageName      = "test_image_name"
	tag            = "latest"
	testBodyString = `{"eventid":"test_event_id","appid":"test_app_id","imagename":"test_image_name"}`
	description    = "{\"services\":{\"" + serviceName + "\":{\"image\":\"" + imageName + ":" + tag + "\"}},\"version\":\"2\"}"
)

var (
	testEvent = dockercontroller.Event{
		ID:          "",
		Type:        "container",
		AppID:       appId,
		ServiceName: serviceName,
		Status:      status,
		ContainerEvent: dockercontroller.ContainerEvent{
			CID: cid,
		},
	}
	testBody = map[string]interface{}{
		"eventid":   "test_event_id",
		"appid":     "test_app_id",
		"imagename": "test_image_name",
	}
	app = map[string]interface{}{
		"id":          "test_app_id",
		"state":       "running",
		"description": description,
		"images": []map[string]interface{}{
			{
				"name": "test_image_name",
			},
		},
	}
	unknownError = errors.Unknown{}
)

func TestSubscribeEvent_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().InsertEvent(testBody["eventid"], testBody["appid"], testBody["imagename"]).Return(testBody, nil),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	_, err := Executor{}.SubscribeEvent(testBodyString)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestSubscribeEventWhenInsertEventFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().InsertEvent(testBody["eventid"], testBody["appid"], testBody["imagename"]).Return(nil, unknownError),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	_, err := Executor{}.SubscribeEvent(testBodyString)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", err.Error())
	case errors.Unknown:
	}
}

func TestSendNotification_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	evt := map[string]interface{}{
		"id":        "test_event_id",
		"appid":     "test_app_id",
		"imagename": "test_image_name",
	}

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	serviceDbExecutor := servicedbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		serviceDbExecutor.EXPECT().GetApp(appId).Return(app, nil),
		dbExecutorMockObj.EXPECT().GetEvents(appId, imageName).Return([]map[string]interface{}{evt}, nil),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj
	serviceExecutor = serviceDbExecutor

	Executor{}.SendNotification(testEvent)
}

func TestUnsubscribeEvent_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().DeleteEvent(testBody["eventid"]).Return(nil),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	err := Executor{}.UnsubscribeEvent(testBodyString)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestUnsubscribeEventWhenDeleteEventFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().DeleteEvent(testBody["eventid"]).Return(unknownError),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	err := Executor{}.UnsubscribeEvent(testBodyString)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", err.Error())
	case errors.Unknown:
	}
}

/*******************************************************************************
 * Copyright 2017 Samsung Electronics All Rights Reserved.
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
package event

import (
	"commons/errors"
	dbmocks "db/bolt/wrapper/mocks"
	gomock "github.com/golang/mock/gomock"
	"reflect"
	"testing"
)

const (
	EVENTID         = "test_event_id"
	APPID           = "test_app_id"
	IMAGENAME       = "test_image_name"
	EVENT_JSON      = "{\"id\":\"test_event_id\",\"appid\":\"test_app_id\",\"imagename\":\"test_image_name\"}"
	DUMMY_ERROR_MSG = "dummy_errors"
)

var (
	event = map[string]interface{}{
		"id":        EVENTID,
		"appid":     APPID,
		"imagename": IMAGENAME,
	}
	dummy_error = errors.NotFound{DUMMY_ERROR_MSG}
)

func TestCalledInsertEvent_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := dbmocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(EVENTID)).Return(nil, dummy_error),
		dbMockObj.EXPECT().Put([]byte(EVENTID), gomock.Any()).Return(nil),
	)

	db = dbMockObj
	executor := Executor{}

	_, err := executor.InsertEvent(EVENTID, APPID, IMAGENAME)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledInsertEventWhenAlreadyExistsInDB_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := dbmocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(EVENTID)).Return([]byte(EVENT_JSON), nil),
	)

	db = dbMockObj
	executor := Executor{}

	_, err := executor.InsertEvent(EVENTID, APPID, IMAGENAME)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "AlreadyReported", err.Error())
	case errors.AlreadyReported:
	}
}

func TestCalledInsertEventWhenDBReturnsError_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := dbmocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(EVENTID)).Return(nil, dummy_error),
		dbMockObj.EXPECT().Put([]byte(EVENTID), gomock.Any()).Return(dummy_error),
	)

	db = dbMockObj
	executor := Executor{}

	_, err := executor.InsertEvent(EVENTID, APPID, IMAGENAME)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}

func TestCalledGetEvents_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	returnedEvents := map[string]interface{}{
		EVENTID: EVENT_JSON,
	}
	expectedRes := []map[string]interface{}{event}

	dbMockObj := dbmocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().List().Return(returnedEvents, nil),
	)

	db = dbMockObj
	executor := Executor{}
	res, err := executor.GetEvents(APPID, IMAGENAME)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(expectedRes, res) {
		t.Errorf("Expected res: %s, actual res: %s", expectedRes, res)
	}
}

func TestCalledGetEventsWhenDBReturnsError_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := dbmocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().List().Return(nil, dummy_error),
	)

	db = dbMockObj
	executor := Executor{}
	_, err := executor.GetEvents(APPID, IMAGENAME)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}

func TestCalledDeleteEvent_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := dbmocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Delete([]byte(EVENTID)).Return(nil),
	)

	db = dbMockObj
	executor := Executor{}
	err := executor.DeleteEvent(EVENTID)

	if err != nil {
		t.Error()
	}
}

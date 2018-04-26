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
package service

import (
	"commons/errors"
	"db/bolt/wrapper/mocks"
	"encoding/json"
	gomock "github.com/golang/mock/gomock"
	"reflect"
	"testing"
)

const (
	DUMMY_ERROR_MSG     = "dummy_errors"
	INVALID_APPID       = ""
	INVALID_DESCRIPTION = ""
	VALID_APPID         = "e1f63701c26b8bbf6e41fd7c2bdf12e075b768b5"
	VALID_STATE         = "STATE"
	REPO                = "test_image_name"
	TAG                 = ""
	EVENT               = "update"
	VALID_DESCRIPTION   = `{
	  "services": {
	    "test_service_name": {
	      "image": "test_image_name"
	    }
	  }
	}`
)

var (
	dummy_error = errors.NotFound{DUMMY_ERROR_MSG}
	image       = map[string]interface{}{
		"name": "test_image_name",
	}
	service = map[string]interface{}{
		"id":          VALID_APPID,
		"description": VALID_DESCRIPTION,
		"images":      []map[string]interface{}{image},
		"state":       VALID_STATE,
	}
)

func TestCalled_InsertComposeFile_WithEmptyDescription_ExpectErrorReturn(t *testing.T) {
	dbExecutor := Executor{}

	_, err := dbExecutor.InsertComposeFile(INVALID_DESCRIPTION, VALID_STATE)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYamlError or UnknownError", "nil")
	}

	switch err.(type) {
	default:
		t.Error()
	case errors.InvalidYaml:
	case errors.Unknown:
	}
}

func TestCalled_InsertComposeFile_WithInvalidDescription_Service_ExpectErrorReturn(t *testing.T) {
	dbExecutor := Executor{}

	inVALID_DESCRIPTION_without_service := `{"services":}`

	_, err := dbExecutor.InsertComposeFile(inVALID_DESCRIPTION_without_service, VALID_STATE)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYamlError", "nil")
	}

	switch err.(type) {
	default:
		t.Error()
	case errors.InvalidYaml:
	}
}

func TestCalled_InsertComposeFile_WithInvalidDescription_Image_ExpectErrorReturn(t *testing.T) {
	dbExecutor := Executor{}

	inVALID_DESCRIPTION_without_image := `{
  "services": {
    "test_service_name": {}
  }
}`
	_, err := dbExecutor.InsertComposeFile(inVALID_DESCRIPTION_without_image, VALID_STATE)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYamlError", "nil")
	}
	switch err.(type) {
	default:
		t.Error()
	case errors.InvalidYaml:
	}
}

func TestCalled_InsertComposeFile_WithValidDescription_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(VALID_APPID)).Return(nil, dummy_error),
		dbMockObj.EXPECT().Put([]byte(VALID_APPID), gomock.Any()).Return(nil),
	)
	db = dbMockObj
	dbExecutor := Executor{}

	expectedRes := map[string]interface{}{
		"id":          VALID_APPID,
		"description": VALID_DESCRIPTION,
		"images":      []map[string]interface{}{image},
		"state":       VALID_STATE,
	}

	res, err := dbExecutor.InsertComposeFile(VALID_DESCRIPTION, VALID_STATE)

	if err != nil {
		t.Error()
	}

	if !reflect.DeepEqual(res, expectedRes) {
		t.Errorf("Expected res: %s, actual res: %s", expectedRes, res)
	}
}

func TestCalled_InsertComposeFile_WhenAlreadyExistsInDB_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	returnedService, _ := json.Marshal(service)

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(VALID_APPID)).Return([]byte(returnedService), nil),
	)
	db = dbMockObj
	dbExecutor := Executor{}

	_, err := dbExecutor.InsertComposeFile(VALID_DESCRIPTION, VALID_STATE)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "AlreadyREPOrted", err.Error())
	case errors.AlreadyReported:
	}
}

func TestCalled_GetAppList_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	encodedService, _ := json.Marshal(service)
	returnedServices := map[string]interface{}{
		VALID_APPID: string(encodedService),
	}
	expectedRes := []map[string]interface{}{service}

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().List().Return(returnedServices, nil),
	)
	db = dbMockObj
	dbExecutor := Executor{}

	res, err := dbExecutor.GetAppList()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(expectedRes, res) {
		t.Errorf("Expected res: %s, actual res: %s", expectedRes, res)
	}
}

func TestCalled_GetAppListWhenDBReturnsError_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().List().Return(nil, dummy_error),
	)
	db = dbMockObj
	dbExecutor := Executor{}

	_, err := dbExecutor.GetAppList()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}

func TestCalled_GetApp_WithInvaliddAppID_ExpectErrorReturn(t *testing.T) {
	dbExecutor := Executor{}

	_, err := dbExecutor.GetApp(INVALID_APPID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", err.Error())
	case errors.InvalidParam:
	}
}

func TestCalled_GetApp_WhenDBHasMatchedApp_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	returnedService, _ := json.Marshal(service)
	expectedRes := map[string]interface{}{
		"id":          VALID_APPID,
		"description": VALID_DESCRIPTION,
		"state":       VALID_STATE,
		"images":      []map[string]interface{}{image},
	}

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(VALID_APPID)).Return([]byte(returnedService), nil),
	)

	db = dbMockObj
	dbExecutor := Executor{}
	res, err := dbExecutor.GetApp(VALID_APPID)

	if err != nil {
		t.Error()
	}

	if !reflect.DeepEqual(res, expectedRes) {
		t.Errorf("Expected res: %s, actual res: %s", expectedRes, res)
	}
}

func TestCalled_GetApp_WhenDBHasNotMatchedApp_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(VALID_APPID)).Return(nil, dummy_error),
	)

	db = dbMockObj
	dbExecutor := Executor{}
	_, err := dbExecutor.GetApp(VALID_APPID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}

func TestCalled_UpdateAppInfo_WithInvaliddAppID_ExpectErrorReturn(t *testing.T) {
	dbExecutor := Executor{}

	err := dbExecutor.UpdateAppInfo(INVALID_APPID, VALID_DESCRIPTION)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", err.Error())
	case errors.InvalidParam:
	}
}

func TestCalled_UpdateAppInfo_WhenDBHasNotMatchedApp_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(VALID_APPID)).Return(nil, dummy_error),
	)

	db = dbMockObj
	dbExecutor := Executor{}
	err := dbExecutor.UpdateAppInfo(VALID_APPID, VALID_DESCRIPTION)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", err.Error())
	case errors.NotFound:
	case errors.Unknown:
	}
}

func TestCalled_UpdateAppInfo_WhenDBHasMatchedApp_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	returnedService, _ := json.Marshal(service)

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(VALID_APPID)).Return([]byte(returnedService), nil),
		dbMockObj.EXPECT().Put([]byte(VALID_APPID), gomock.Any()).Return(nil),
	)

	db = dbMockObj
	dbExecutor := Executor{}
	err := dbExecutor.UpdateAppInfo(VALID_APPID, VALID_DESCRIPTION)

	if err != nil {
		t.Error()
	}
}

func TestCalled_UpdateAppState_WithInvalidAppID_ExpectErrorReturn(t *testing.T) {
	dbExecutor := Executor{}

	err := dbExecutor.UpdateAppInfo(INVALID_APPID, VALID_DESCRIPTION)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", err.Error())
	case errors.InvalidParam:
	}
}

func TestCalled_UpdateAppState_WhenDBHasNotMatchedApp_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(VALID_APPID)).Return(nil, dummy_error),
	)

	db = dbMockObj
	dbExecutor := Executor{}
	err := dbExecutor.UpdateAppState(VALID_APPID, VALID_STATE)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", err.Error())
	case errors.NotFound:
	case errors.Unknown:
	}
}

func TestCalled_UpdateAppState_WhenDBHasMatchedApp_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	returnedService, _ := json.Marshal(service)

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(VALID_APPID)).Return([]byte(returnedService), nil),
		dbMockObj.EXPECT().Put([]byte(VALID_APPID), gomock.Any()).Return(nil),
	)

	db = dbMockObj
	dbExecutor := Executor{}
	err := dbExecutor.UpdateAppState(VALID_APPID, VALID_STATE)

	if err != nil {
		t.Error()
	}
}

func TestCalled_UpdateAppEVENT_WithInvalidAppID_ExpectErrorReturn(t *testing.T) {
	dbExecutor := Executor{}

	err := dbExecutor.UpdateAppEvent(INVALID_APPID, REPO, TAG, EVENT)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", err.Error())
	case errors.InvalidParam:
	}
}

func TestCalled_UpdateAppEVENT_WhenDBHasNotMatchedApp_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(VALID_APPID)).Return(nil, dummy_error),
	)

	db = dbMockObj
	dbExecutor := Executor{}
	err := dbExecutor.UpdateAppEvent(VALID_APPID, REPO, TAG, EVENT)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", err.Error())
	case errors.NotFound:
	case errors.Unknown:
	}
}

func TestCalled_UpdateAppEVENT_WhenDBHasMatchedApp_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	returnedService, _ := json.Marshal(service)

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(VALID_APPID)).Return([]byte(returnedService), nil),
		dbMockObj.EXPECT().Put([]byte(VALID_APPID), gomock.Any()).Return(nil),
	)

	db = dbMockObj
	dbExecutor := Executor{}
	err := dbExecutor.UpdateAppEvent(VALID_APPID, REPO, TAG, EVENT)

	if err != nil {
		t.Error()
	}
}

func TestCalled_DeleteApp_WithInvlaidAppID_ExpectErrorReturn(t *testing.T) {
	dbExecutor := Executor{}
	err := dbExecutor.DeleteApp(INVALID_APPID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", err.Error())
	case errors.InvalidParam:
	}
}

func TestCalled_DeleteApp_WhenDBHasNotMatchedApp_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Delete([]byte(VALID_APPID)).Return(nil),
	)

	db = dbMockObj
	dbExecutor := Executor{}
	err := dbExecutor.DeleteApp(VALID_APPID)

	if err != nil {
		t.Error()
	}
}

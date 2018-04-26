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
package configuration

import (
	"commons/errors"
	dbmocks "db/bolt/wrapper/mocks"
	gomock "github.com/golang/mock/gomock"
	"reflect"
	"testing"
)

const (
	PROP_NAME       = "name"
	PROP_VALUE      = "value"
	PROP_JSON       = "{\"name\":\"name\",\"value\":\"value\",\"readonly\":false}"
	DUMMY_ERROR_MSG = "dummy_errors"
)

var (
	dummy_error = errors.NotFound{DUMMY_ERROR_MSG}
	property    = map[string]interface{}{
		"name":     PROP_NAME,
		"value":    PROP_VALUE,
		"readOnly": false,
	}
)

func TestCalledSetProperty_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := dbmocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(PROP_NAME)).Return(nil, dummy_error),
		dbMockObj.EXPECT().Put([]byte(PROP_NAME), gomock.Any()).Return(nil),
	)

	db = dbMockObj
	executor := Executor{}

	err := executor.SetProperty(property)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledSetPropertyWhenAlreadyExistsInDB_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := dbmocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(PROP_NAME)).Return([]byte(PROP_JSON), nil),
		dbMockObj.EXPECT().Put([]byte(PROP_NAME), gomock.Any()).Return(nil),
	)

	db = dbMockObj
	executor := Executor{}

	err := executor.SetProperty(property)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledGetProperty_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expectedRes := property

	dbMockObj := dbmocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(PROP_NAME)).Return([]byte(PROP_JSON), nil),
	)

	db = dbMockObj
	executor := Executor{}
	res, err := executor.GetProperty(PROP_NAME)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(expectedRes, res) {
		t.Errorf("Expected res: %s, actual res: %s", expectedRes, res)
	}
}

func TestCalledGetPropertyWhenDBReturnsError_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := dbmocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().Get([]byte(PROP_NAME)).Return(nil, dummy_error),
	)

	db = dbMockObj
	executor := Executor{}
	_, err := executor.GetProperty(PROP_NAME)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}

func TestCalledGetProperties_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	returnedProps := map[string]interface{}{
		PROP_NAME: PROP_JSON,
	}
	expectedRes := []map[string]interface{}{property}

	dbMockObj := dbmocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().List().Return(returnedProps, nil),
	)

	db = dbMockObj
	executor := Executor{}
	res, err := executor.GetProperties()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(expectedRes, res) {
		t.Errorf("Expected res: %s, actual res: %s", expectedRes, res)
	}
}

func TestCalledGetPropertiesWhenDBReturnsError_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dbMockObj := dbmocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		dbMockObj.EXPECT().List().Return(nil, dummy_error),
	)

	db = dbMockObj
	executor := Executor{}
	_, err := executor.GetProperties()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}

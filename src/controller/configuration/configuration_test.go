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
	dbmocks "db/bolt/configuration/mocks"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"os"
	"reflect"
	"testing"
)

var (
	properties = map[string]interface{}{
		"properties": []map[string]interface{}{{
			"name":     "name",
			"value":    "value",
			"readOnly": false,
		}},
	}
	newProperties = map[string]interface{}{
		"properties": []map[string]interface{}{{
			"name": "value",
		}},
	}
	notFoundError = errors.NotFound{}
)

func TestGetAnchorEndPointWithNoAnchorRPEnv_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedRet := "http://127.0.0.1:48099/api/v1"

	os.Setenv("ANCHOR_ADDRESS", "127.0.0.1")
	ret, err := getAnchorEndPoint()
	os.Unsetenv("ANCHOR_ADDRESS")

	if err != nil {
		t.Errorf("Expected error : nil, actual error : %s", err.Error())
	}

	if ret != expectedRet {
		t.Errorf("Expected result : %v, actual result : %v", expectedRet, ret)
	}
}

func TestGetAnchorEndPointWithAnchorRPEnvTrue_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedRet := "http://127.0.0.1:80/pharos-anchor/api/v1"

	os.Setenv("ANCHOR_ADDRESS", "127.0.0.1")
	os.Setenv("ANCHOR_REVERSE_PROXY", "true")
	ret, err := getAnchorEndPoint()
	os.Unsetenv("ANCHOR_ADDRESS")
	os.Unsetenv("ANCHOR_REVERSE_PROXY")

	if err != nil {
		t.Errorf("Expected error : nil, actual error : %s", err.Error())
	}

	if ret != expectedRet {
		t.Errorf("Expected result : %v, actual result : %v", expectedRet, ret)
	}
}

func TestGetProxyInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedRet := map[string]interface{}{
		"enabled": true,
	}

	os.Setenv("REVERSE_PROXY", "true")
	ret, err := getProxyInfo()
	os.Unsetenv("REVERSE_PROXY")

	if err != nil {
		t.Errorf("Expected err : nil, actual err : %s", err.Error())
	}

	if !reflect.DeepEqual(expectedRet, ret) {
		t.Errorf("Expected result : %v, Actual Result : %v", expectedRet, ret)
	}
}

func TestGetProxyInfoWithNoEnvironment_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedRet := map[string]interface{}{
		"enabled": false,
	}

	ret, err := getProxyInfo()

	if err != nil {
		t.Errorf("Expected err : nil, actual err : %s", err.Error())
	}

	if !reflect.DeepEqual(expectedRet, ret) {
		t.Errorf("Expected result : %v, Actual Result : %v", expectedRet, ret)
	}
}

func TestGetProxyInfoWithInvalidEnvValue_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	os.Setenv("REVERSE_PROXY", "test")
	_, err := getProxyInfo()
	os.Unsetenv("REVERSER_PROXY")

	if err == nil {
		t.Errorf("Expected err : InvalidParam, actual err : %s", err.Error())
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParam", err.Error())
	case errors.InvalidParam:
	}
}

func TestGetOSInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result1, result2, err := getOSInfo()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if result1 == "" || result2 == "" {
		t.Errorf("Unexpected err : os info is empty")
	}
}

func TestGetProcessorInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result, err := getProcessorInfo()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if result == nil || len(result) == 0 {
		t.Errorf("Unexpected err : processor info array is empty")
	}
}

func TestGetConfiguration_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetProperties().Return(properties["properties"], nil),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	_, err := Executor{}.GetConfiguration()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestGetConfigurationWhenDBReturnsError_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetProperties().Return(nil, notFoundError),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	_, err := Executor{}.GetConfiguration()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}

func TestSetConfiguration_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	prop := properties["properties"].([]map[string]interface{})[0]
	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetProperty(prop["name"].(string)).Return(prop, nil),
		dbExecutorMockObj.EXPECT().SetProperty(prop).Return(nil),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	jsonString, _ := json.Marshal(newProperties)
	err := Executor{}.SetConfiguration(string(jsonString))

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestSetConfigurationWhenGetPropertyReturnsError_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	prop := properties["properties"].([]map[string]interface{})[0]
	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetProperty(prop["name"].(string)).Return(nil, errors.InvalidJSON{}),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	jsonString, _ := json.Marshal(newProperties)
	err := Executor{}.SetConfiguration(string(jsonString))

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "InvalidJSON", err.Error())
	case errors.InvalidJSON:
	}
}

func TestSetConfigurationWhenSetPropertyReturnsError_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	prop := properties["properties"].([]map[string]interface{})[0]
	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetProperty(prop["name"].(string)).Return(prop, nil),
		dbExecutorMockObj.EXPECT().SetProperty(prop).Return(notFoundError),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	jsonString, _ := json.Marshal(newProperties)
	err := Executor{}.SetConfiguration(string(jsonString))

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}

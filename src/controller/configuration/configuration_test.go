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
	dbmocks "db/mongo/configuration/mocks"
	"encoding/json"
	"github.com/golang/mock/gomock"
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

func TestGetOSInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result, err := getOSInfo()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if result == "" {
		t.Errorf("Unexpected err : os info is empty")

	}
}

func TestGetPlatformInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result, err := getPlatformInfo()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if _, ok := result["platform"]; !ok {
		t.Errorf("Unexpected err: platform key does not exist")
	}

	if _, ok := result["family"]; !ok {
		t.Errorf("Unexpected err: family key does not exist")
	}

	if _, ok := result["version"]; !ok {
		t.Errorf("Unexpected err: version key does not exist")
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

func TestSetConfigurationWhenDBReturnsError_ExpectErrorReturn(t *testing.T) {
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

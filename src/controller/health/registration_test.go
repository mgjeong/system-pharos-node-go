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
package health

import (
	configmocks "controller/configuration/mocks"
	dbmocks "db/bolt/configuration/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	msgmocks "messenger/mocks"
	"os"
	"testing"
)

var (
	ANCHOR_IP      = "192.168.0.1"
	ANCHOR_ADDRESS = map[string]interface{}{
		"anchoraddress": ANCHOR_IP,
		"policy":        []string{"readable"},
	}
	REVERSE_PROXY = map[string]interface{}{
		"reverseproxy": map[string]interface{}{
			"enabled": false,
		},
	}
	CONFIGURATION = map[string]interface{}{
		"properties": []map[string]interface{}{ANCHOR_ADDRESS, REVERSE_PROXY},
	}
	PROPERTY = map[string]interface{}{
		"name":     "deviceid",
		"value":    "test_device_id",
		"readonly": "true",
	}
)

var healthExecutor Command

func init() {
	healthExecutor = Executor{}
}

func TestCalledRegisterWhenFailedToGetConfiguration_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMockObj := configmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		configMockObj.EXPECT().GetConfiguration().Return(CONFIGURATION, errors.New("Error")),
	)
	configurator = configMockObj

	err := register(false)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledRegisterWhenFailedToSetConfiguration_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMockObj := configmocks.NewMockCommand(ctrl)
	msgMockObj := msgmocks.NewMockCommand(ctrl)
	dbMockObj := dbmocks.NewMockCommand(ctrl)

	url := "http://192.168.0.1:48099/api/v1/management/nodes/register"
	expectedResp := `{"id":"deviceid"}`

	gomock.InOrder(
		configMockObj.EXPECT().GetConfiguration().Return(CONFIGURATION, nil),
		msgMockObj.EXPECT().SendHttpRequest("POST", url, gomock.Any()).Return(200, expectedResp, nil),
		dbMockObj.EXPECT().GetProperty("deviceid").Return(PROPERTY, nil),
		dbMockObj.EXPECT().SetProperty(gomock.Any()).Return(errors.New("Error")),
	)
	configurator = configMockObj
	httpExecutor = msgMockObj
	configDbExecutor = dbMockObj

	os.Setenv("ANCHOR_ADDRESS", ANCHOR_IP)
	os.Setenv("ANCHOR_REVERSE_PROXY", "false")
	err := register(false)
	os.Unsetenv("ANCHOR_ADDRESS")
	os.Unsetenv("ANCHOR_REVERSE_PROXY")

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledUnregister_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMockObj := configmocks.NewMockCommand(ctrl)
	dbMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbMockObj.EXPECT().GetProperty("deviceid").Return(PROPERTY, nil),
		dbMockObj.EXPECT().SetProperty(gomock.Any()).Return(nil),
	)
	configurator = configMockObj
	configDbExecutor = dbMockObj

	err := healthExecutor.Unregister()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledSendRegisterRequestWhenFailedToSendHttpRequest_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockCommand(ctrl)

	url := "http://192.168.0.1:48099/api/v1/management/nodes/register"

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", url, gomock.Any()).Return(500, "", errors.New("Error")),
	)
	httpExecutor = msgMockObj

	os.Setenv("ANCHOR_ADDRESS", ANCHOR_IP)
	os.Setenv("ANCHOR_REVERSE_PROXY", "false")
	_, _, err := sendRegisterRequest(CONFIGURATION)
	os.Unsetenv("ANCHOR_ADDRESS")
	os.Unsetenv("ANCHOR_REVERSE_PROXY")

	if err == nil {
		t.Errorf("Expected err: %s", err.Error())
	}
}

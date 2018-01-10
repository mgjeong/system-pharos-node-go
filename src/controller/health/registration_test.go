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
	"errors"
	"github.com/golang/mock/gomock"
	msgmocks "messenger/mocks"
	"testing"
)

var (
	CONFIGURATION = map[string]interface{}{
		"serveraddress": "192.168.0.1",
		"devicename":    "Edge Device #1",
		"deviceid":      "54919CA5-4101-4AE4-595B-353C51AA983C",
		"manufacturer":  "Manufacturer Name",
		"modelnumber":   "Model number as designated by the manufacturer",
		"serialnumber":  "Serial number",
		"platform":      "Platform name and version",
		"os":            "Operationg system name and version",
		"location":      "Human readable location",
		"pinginterval":  "10",
		"deviceaddress": "192.168.0.1",
		"nodeid":       "Pharos Node ID",
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

	err := register()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledRegisterWhenFailedToSetConfiguration_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMockObj := configmocks.NewMockCommand(ctrl)
	msgMockObj := msgmocks.NewMockCommand(ctrl)

	url := "http://192.168.0.1:48099/api/v1/management/nodes/register"
	expectedResp := `{"id":"nodeid"}`
	expectedNewConfig := map[string]interface{}{
		"nodeid": "nodeid",
	}

	gomock.InOrder(
		configMockObj.EXPECT().GetConfiguration().Return(CONFIGURATION, nil),
		msgMockObj.EXPECT().SendHttpRequest("POST", url, gomock.Any()).Return(200, expectedResp, nil),
		configMockObj.EXPECT().SetConfiguration(expectedNewConfig).Return(errors.New("Error")),
	)
	configurator = configMockObj
	httpExecutor = msgMockObj

	err := register()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledUnregister_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configMockObj := configmocks.NewMockCommand(ctrl)

	expectedNewConfig := map[string]interface{}{
		"nodeid": "",
	}

	gomock.InOrder(
		configMockObj.EXPECT().SetConfiguration(expectedNewConfig).Return(nil),
	)
	configurator = configMockObj

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

	_, _, err := sendRegisterRequest(CONFIGURATION)

	if err == nil {
		t.Errorf("Expected err: %s", err.Error())
	}
}
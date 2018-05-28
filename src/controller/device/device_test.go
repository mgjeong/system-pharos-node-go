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
package device

import (
	"commons/errors"
	"github.com/golang/mock/gomock"
	msgmocks "messenger/mocks"
	"testing"
)

var (
	testSystemContainerIP     = "0.0.0.0"
	testResponse = `{"response":"response"}`
)

var deviceExecutor Command

func init() {
	deviceExecutor = Executor{}
}

func TestReboot_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockCommand(ctrl)

	url := "http://" + testSystemContainerIP + "/api/v1/device/management/reboot"

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest(POST, url, gomock.Any()).Return(200, testResponse, nil),
	)

	httpExecutor = msgMockObj

	systemContainerIP = testSystemContainerIP
	err := deviceExecutor.Reboot()

	if err != nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestRebootWithNosystemContainerIP_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	systemContainerIP = ""
	err := deviceExecutor.Reboot()

	switch err.(type) {
	default:
		t.Errorf("Expected err: NotFound, actual err: %s", err.Error())
	case errors.NotFound:
	}
}

func TestRebootWhenSendHttpRequestFailed_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockCommand(ctrl)

	url := "http://" + testSystemContainerIP + "/api/v1/device/management/reboot"

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest(POST, url, gomock.Any()).Return(500, "", errors.Unknown{}),
	)

	httpExecutor = msgMockObj

	systemContainerIP = testSystemContainerIP
	err := deviceExecutor.Reboot()

	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestRestore_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockCommand(ctrl)

	url := "http://" + testSystemContainerIP + "/api/v1/device/management/restore"

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest(POST, url, gomock.Any()).Return(200, testResponse, nil),
	)

	httpExecutor = msgMockObj

	systemContainerIP = testSystemContainerIP
	err := deviceExecutor.Restore()

	if err != nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestRestoreWithNosystemContainerIP_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	systemContainerIP = ""
	err := deviceExecutor.Restore()

	switch err.(type) {
	default:
		t.Errorf("Expected err: NotFound, actual err: %s", err.Error())
	case errors.NotFound:
	}
}

func TestRestoreWhenSendHttpRequestFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockCommand(ctrl)

	url := "http://" + testSystemContainerIP + "/api/v1/device/management/restore"

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest(POST, url, gomock.Any()).Return(500, "", errors.Unknown{}),
	)

	httpExecutor = msgMockObj

	systemContainerIP = testSystemContainerIP
	err := deviceExecutor.Restore()

	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

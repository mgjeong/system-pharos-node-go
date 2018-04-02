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
	urls "commons/url"
	devicemocks "controller/device/mocks"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	invalidOperationList = map[string][]string{
		"/api/v1/management/device/reboot":  []string{GET, PUT, DELETE},
		"/api/v1/management/device/restore": []string{GET, PUT, DELETE},
	}
	testMap = map[string]interface{}{
		"test": "test",
	}
	testList = []testObj{
		{"InvalidYamlError", errors.InvalidYaml{}, http.StatusBadRequest},
		{"InvalidAppId", errors.InvalidAppId{}, http.StatusBadRequest},
		{"InvalidParamError", errors.InvalidParam{}, http.StatusBadRequest},
		{"NotFoundImage", errors.NotFoundImage{}, http.StatusBadRequest},
		{"AlreadyAllocatedPort", errors.AlreadyAllocatedPort{}, http.StatusBadRequest},
		{"AlreadyUsedName", errors.AlreadyUsedName{}, http.StatusBadRequest},
		{"InvalidContainerName", errors.InvalidContainerName{}, http.StatusBadRequest},
		{"IOError", errors.IOError{}, http.StatusInternalServerError},
		{"UnknownError", errors.Unknown{}, http.StatusInternalServerError},
		{"NotFoundError", errors.NotFound{}, http.StatusServiceUnavailable},
		{"AlreadyReported", errors.AlreadyReported{}, http.StatusAlreadyReported},
	}
)

type testObj struct {
	name       string
	err        error
	expectCode int
}

var deviceAPIExecutor Command

func init() {
	deviceAPIExecutor = Executor{}
}

func TestDeviceAPIInvalidOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for api, invalidMethodList := range invalidOperationList {
		for _, method := range invalidMethodList {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, api, nil)

			deviceAPIExecutor.Handle(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected error : %d, Actual Error : %d", http.StatusMethodNotAllowed, w.Code)
			}
		}
	}
}

func TestRebootAPI_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deviceExecutorMockObj := devicemocks.NewMockCommand(ctrl)

	gomock.InOrder(
		deviceExecutorMockObj.EXPECT().Reboot().Return(nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Device()+urls.Reboot(), nil)

	deviceExecutor = deviceExecutorMockObj

	deviceAPIExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Unexpected error code : %d", w.Code)
	}
}

func TestRebootAPIWhenControllerFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deviceExecutorMockObj := devicemocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			deviceExecutorMockObj.EXPECT().Reboot().Return(test.err),
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Device()+urls.Reboot(), nil)

		deviceExecutor = deviceExecutorMockObj

		deviceAPIExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Unexpected error code : %d\n", w.Code)
		}
	}
}

func TestRestoreAPI_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deviceExecutorMockObj := devicemocks.NewMockCommand(ctrl)

	gomock.InOrder(
		deviceExecutorMockObj.EXPECT().Restore().Return(nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Device()+urls.Restore(), nil)

	deviceExecutor = deviceExecutorMockObj

	deviceAPIExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Unexpected error code : %d", w.Code)
	}
}

func TestRestoreAPIWhenControllerFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deviceExecutorMockObj := devicemocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			deviceExecutorMockObj.EXPECT().Restore().Return(test.err),
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Device()+urls.Restore(), nil)

		deviceExecutor = deviceExecutorMockObj

		deviceAPIExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Unexpected error code : %d\n", w.Code)
		}
	}
}
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
package resource

import (
	"commons/errors"
	urls "commons/url"
	resourcemocks "controller/monitoring/resource/mocks"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	invalidOperationList = map[string][]string{
		"/api/v1/monitoring/apps/appId/resource": []string{POST, PUT, DELETE},
		"/api/v1/monitoring/resource":            []string{POST, PUT, DELETE},
	}
	testAppId = "testAppId"
	testMap   = map[string]interface{}{
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

var resourceAPIExecutor Command

func init() {
	resourceAPIExecutor = Executor{}
}

func TestResourceAPIInvalidOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for api, invalidMethodList := range invalidOperationList {
		for _, method := range invalidMethodList {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, api, nil)

			resourceAPIExecutor.Handle(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected error : %d, Actual Error : %d", http.StatusMethodNotAllowed, w.Code)
			}
		}
	}
}

func TestResourceAPIInvalidUrl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	InvalidUrl := "http://0.0.0.0:48098/api/v1/monitoring/resource/resource"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(GET, InvalidUrl, nil)

	resourceAPIExecutor.Handle(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected error : %d, Actual Error : %d", http.StatusNotFound, w.Code)
	}
}

func TestHostResourceAPI_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceExecutorMockObj := resourcemocks.NewMockCommand(ctrl)

	gomock.InOrder(
		resourceExecutorMockObj.EXPECT().GetHostResourceInfo().Return(testMap, nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(GET, urls.Base()+urls.Monitoring()+urls.Resource(), nil)

	resourceExecutor = resourceExecutorMockObj

	resourceAPIExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Unexpected error code : %d", w.Code)
	}
}

func TestHostResourceAPIWhenControllerFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceExecutorMockObj := resourcemocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			resourceExecutorMockObj.EXPECT().GetHostResourceInfo().Return(nil, test.err),
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(GET, urls.Base()+urls.Monitoring()+urls.Resource(), nil)

		resourceExecutor = resourceExecutorMockObj

		resourceAPIExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Unexpected error code : %d\n", w.Code)
		}
	}
}

func TestAppResourceAPI_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceExecutorMockObj := resourcemocks.NewMockCommand(ctrl)

	gomock.InOrder(
		resourceExecutorMockObj.EXPECT().GetAppResourceInfo(testAppId).Return(testMap, nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(GET, urls.Base()+urls.Monitoring()+urls.Apps()+"/"+testAppId+urls.Resource(), nil)

	resourceExecutor = resourceExecutorMockObj

	resourceAPIExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Unexpected error code : %d", w.Code)
	}
}

func TestAppResourceAPIWhenControllerFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceExecutorMockObj := resourcemocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			resourceExecutorMockObj.EXPECT().GetAppResourceInfo(testAppId).Return(nil, test.err),
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(GET, urls.Base()+urls.Monitoring()+urls.Apps()+"/"+testAppId+urls.Resource(), nil)

		resourceExecutor = resourceExecutorMockObj

		resourceAPIExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Unexpected error code : %d\n", w.Code)
		}
	}
}
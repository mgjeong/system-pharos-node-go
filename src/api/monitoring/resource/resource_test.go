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
		"/api/v1/monitoring/resource": []string{POST, PUT, DELETE},
	}
	testMap = map[string]interface{}{
		"cpu":  "test",
		"mem":  "test",
		"disk": "test",
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

var resourceApiExecutor Command

func init() {
	resourceApiExecutor = Executor{}
}

func TestResourceApiInvalidOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for api, invalidMethodList := range invalidOperationList {
		for _, method := range invalidMethodList {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, api, nil)

			resourceApiExecutor.Handle(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected error : %d, Actual Error : %d", http.StatusMethodNotAllowed, w.Code)
			}
		}
	}
}

func TestResourceApi_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceExecutorMockObj := resourcemocks.NewMockCommand(ctrl)

	gomock.InOrder(
		resourceExecutorMockObj.EXPECT().GetResourceInfo().Return(testMap, nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(GET, urls.Base()+urls.Monitoring()+urls.Resource(), nil)

	resourceExecutor = resourceExecutorMockObj

	resourceApiExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Unexpected error code : %d", w.Code)
	}
}

func TestResourceApiWhenControllerFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceExecutorMockObj := resourcemocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			resourceExecutorMockObj.EXPECT().GetResourceInfo().Return(nil, test.err),
		)

		w := httptest.NewRecorder()
		print(urls.Base()+urls.Monitoring()+urls.Resource())
		req, _ := http.NewRequest(GET, urls.Base()+urls.Monitoring()+urls.Resource(), nil)

		resourceExecutor = resourceExecutorMockObj

		resourceApiExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Unexpected error code : %d\n", w.Code)
		}
	}
}

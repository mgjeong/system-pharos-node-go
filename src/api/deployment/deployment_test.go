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
package deployment

import (
	"bytes"
	"commons/errors"
	urls "commons/url"
	deploymentmocks "controller/deployment/mocks"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const (
	ID string = "id"
)

var (
	appId                = "0000000000001"
	invalidOperationList = map[string][]string{
		"/api/v1/management/apps":           []string{PUT, POST, DELETE},
		"/api/v1/management/apps/deploy":    []string{GET, PUT, DELETE},
		"/api/v1/management/apps/11":        []string{PUT},
		"/api/v1/management/apps/11/update": []string{GET, PUT, DELETE},
		"/api/v1/management/apps/11/stop":   []string{GET, PUT, DELETE},
		"/api/v1/management/apps/11/start":  []string{GET, PUT, DELETE},
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
	testMap = map[string]interface{}{
		"id": appId,
	}
)

type testObj struct {
	name       string
	err        error
	expectCode int
}

var deploymentApiExecutor Command

func init() {
	deploymentApiExecutor = Executor{}
}

func TestDeploymentApiWithInvalidOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for api, invalidMethodList := range invalidOperationList {
		for _, method := range invalidMethodList {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, api, nil)

			deploymentApiExecutor.Handle(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Error("Expected error : %d, Actual Error : %d", http.StatusMethodNotAllowed, w.Code)
			}
		}
	}
}

func TestDeployApi_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	data := url.Values{}
	data.Set("name", "test")
	body := bytes.NewBufferString(data.Encode())

	deployedApp := make(map[string]interface{})
	deployedApp[ID] = appId

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().DeployApp(gomock.Any()).Return(deployedApp, nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+urls.Deploy(), body)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentApiExecutor.Handle(w, req)

	if w.Header().Get("Location") != urls.Base()+urls.Management()+urls.Apps()+"/"+appId ||
		w.Code != http.StatusOK {
		t.Error()
	}
}

func TestDeployApiWhenControllerFailed_ExpecReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			deploymentExecutorMockObj.EXPECT().DeployApp(gomock.Any()).Return(nil, test.err),
		)

		data := url.Values{}
		data.Set("name", "test")
		body := bytes.NewBufferString(data.Encode())

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+urls.Deploy(), body)

		deploymentExecutor = deploymentExecutorMockObj

		deploymentApiExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Error("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestDeployApiWithEmptyBodyStr_ExpecReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	deployedApp := make(map[string]interface{})
	deployedApp[ID] = appId

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+urls.Deploy(), nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentApiExecutor.Handle(w, req)

	if w.Code == http.StatusOK {
		t.Error("Expected return error but return http.StatusOK")
	}
}

func TestAppsApi_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	apps := make(map[string]interface{})
	apps["apps"] = "test"

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().Apps().Return(apps, nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(GET, urls.Base()+urls.Management()+urls.Apps(), nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentApiExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Error("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestAppsApiWhenControllerFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			deploymentExecutorMockObj.EXPECT().Apps().Return(nil, test.err),
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(GET, urls.Base()+urls.Management()+urls.Apps(), nil)

		deploymentExecutor = deploymentExecutorMockObj

		deploymentApiExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Error("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestGETAppApi_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().App(appId).Return(testMap, nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(GET, urls.Base()+urls.Management()+urls.Apps()+"/"+appId, nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentApiExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Error("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestGETAppApiWhenControllerFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			deploymentExecutorMockObj.EXPECT().App(appId).Return(nil, test.err),
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(GET, urls.Base()+urls.Management()+urls.Apps()+"/"+appId, nil)

		deploymentExecutor = deploymentExecutorMockObj

		deploymentApiExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Error("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestPOSTAppApi_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	data := url.Values{}
	data.Set("name", "test")
	body := bytes.NewBufferString(data.Encode())

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().UpdateAppInfo(appId, gomock.Any()).Return(nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId, body)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentApiExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Error("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestPOSTAppApiWithEmptyBody_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId, nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentApiExecutor.Handle(w, req)

	if w.Code == http.StatusOK {
		t.Error("Expected return error but return http.StatusOK")
	}
}

func TestPOSTAppApiWhenControllerFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	data := url.Values{}
	data.Set("name", "test")
	body := bytes.NewBufferString(data.Encode())

	for _, test := range testList {
		gomock.InOrder(
			deploymentExecutorMockObj.EXPECT().UpdateAppInfo(appId, body.String()).Return(test.err),
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId, body)

		deploymentExecutor = deploymentExecutorMockObj

		deploymentApiExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Error("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestDELETEAppApi_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().DeleteApp(appId).Return(nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(DELETE, urls.Base()+urls.Management()+urls.Apps()+"/"+appId, nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentApiExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Error("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestDELETEAppApiWhenControllerFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			deploymentExecutorMockObj.EXPECT().DeleteApp(appId).Return(test.err),
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(DELETE, urls.Base()+urls.Management()+urls.Apps()+"/"+appId, nil)

		deploymentExecutor = deploymentExecutorMockObj

		deploymentApiExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Error("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestEventsApi_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	data := url.Values{}
	data.Set("name", "test")
	body := bytes.NewBufferString(data.Encode())

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().HandleEvents(appId, body.String()).Return(nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId+urls.Events(), body)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentApiExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Error("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestEventsApiWithEmptyBodyStr_ExpecReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	deployedApp := make(map[string]interface{})
	deployedApp[ID] = appId

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId+urls.Events(), nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentApiExecutor.Handle(w, req)

	if w.Code == http.StatusOK {
		t.Error("Expected return error but return http.StatusOK")
	}
}

func TestStartApi_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().StartApp(appId).Return(nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId+urls.Start(), nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentApiExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Error("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestStartApiWhenControllerFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			deploymentExecutorMockObj.EXPECT().StartApp(appId).Return(test.err),
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId+urls.Start(), nil)

		deploymentExecutor = deploymentExecutorMockObj

		deploymentApiExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Error("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestStopApi_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().StopApp(appId).Return(nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId+urls.Stop(), nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentApiExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Error("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestStopApiWhenControllerFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			deploymentExecutorMockObj.EXPECT().StopApp(appId).Return(test.err),
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId+urls.Stop(), nil)

		deploymentExecutor = deploymentExecutorMockObj

		deploymentApiExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Error("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestUpdateApi_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().UpdateApp(appId, nil).Return(nil),
	)

	w := httptest.NewRecorder()
	print(urls.Base() + urls.Management() + urls.Apps() + appId + urls.Update())
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId+urls.Update(), nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentApiExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Error("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestUpdateApiWhenControllerFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			deploymentExecutorMockObj.EXPECT().UpdateApp(appId, nil).Return(test.err),
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId+urls.Update(), nil)

		deploymentExecutor = deploymentExecutorMockObj

		deploymentApiExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Error("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

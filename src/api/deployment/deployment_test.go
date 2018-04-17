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

var deploymentAPIExecutor Command

func init() {
	deploymentAPIExecutor = Executor{}
}

func TestDeploymentAPIWithInvalidOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for api, invalidMethodList := range invalidOperationList {
		for _, method := range invalidMethodList {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, api, nil)

			deploymentAPIExecutor.Handle(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected error : %d, Actual Error : %d", http.StatusMethodNotAllowed, w.Code)
			}
		}
	}
}

func TestDeployAPI_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	data := url.Values{}
	data.Set("name", "test")
	body := bytes.NewBufferString(data.Encode())

	deployedApp := make(map[string]interface{})
	deployedApp[ID] = appId

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().DeployApp(gomock.Any(), nil).Return(deployedApp, nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+urls.Deploy(), body)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentAPIExecutor.Handle(w, req)

	if w.Header().Get("Location") != urls.Base()+urls.Management()+urls.Apps()+"/"+appId ||
		w.Code != http.StatusOK {
		t.Error()
	}
}

func TestDeployAPIWhenControllerFailed_ExpecReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	for _, test := range testList {
		gomock.InOrder(
			deploymentExecutorMockObj.EXPECT().DeployApp(gomock.Any(), nil).Return(nil, test.err),
		)

		data := url.Values{}
		data.Set("name", "test")
		body := bytes.NewBufferString(data.Encode())

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+urls.Deploy(), body)

		deploymentExecutor = deploymentExecutorMockObj

		deploymentAPIExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestDeployAPIWithEmptyBodyStr_ExpecReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	deployedApp := make(map[string]interface{})
	deployedApp[ID] = appId

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+urls.Deploy(), nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentAPIExecutor.Handle(w, req)

	if w.Code == http.StatusOK {
		t.Errorf("Expected return error but return http.StatusOK")
	}
}

func TestAppsAPI_ExpectSuccess(t *testing.T) {
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

	deploymentAPIExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestAppsAPIWhenControllerFailed_ExpectReturnError(t *testing.T) {
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

		deploymentAPIExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestGETAppAPI_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().App(appId).Return(testMap, nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(GET, urls.Base()+urls.Management()+urls.Apps()+"/"+appId, nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentAPIExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestGETAppAPIWhenControllerFailed_ExpectReturnError(t *testing.T) {
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

		deploymentAPIExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestPOSTAppAPI_ExpectSuccess(t *testing.T) {
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

	deploymentAPIExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestPOSTAppAPIWithEmptyBody_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId, nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentAPIExecutor.Handle(w, req)

	if w.Code == http.StatusOK {
		t.Errorf("Expected return error but return http.StatusOK")
	}
}

func TestPOSTAppAPIWhenControllerFailed_ExpectReturnError(t *testing.T) {
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

		deploymentAPIExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestDELETEAppAPI_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().DeleteApp(appId).Return(nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(DELETE, urls.Base()+urls.Management()+urls.Apps()+"/"+appId, nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentAPIExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestDELETEAppAPIWhenControllerFailed_ExpectReturnError(t *testing.T) {
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

		deploymentAPIExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestEventsAPI_ExpectSuccess(t *testing.T) {
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

	deploymentAPIExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestEventsAPIWithEmptyBodyStr_ExpecReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	deployedApp := make(map[string]interface{})
	deployedApp[ID] = appId

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId+urls.Events(), nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentAPIExecutor.Handle(w, req)

	if w.Code == http.StatusOK {
		t.Errorf("Expected return error but return http.StatusOK")
	}
}

func TestStartAPI_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().StartApp(appId).Return(nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId+urls.Start(), nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentAPIExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestStartAPIWhenControllerFailed_ExpectReturnError(t *testing.T) {
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

		deploymentAPIExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestStopAPI_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().StopApp(appId).Return(nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId+urls.Stop(), nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentAPIExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestStopAPIWhenControllerFailed_ExpectReturnError(t *testing.T) {
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

		deploymentAPIExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

func TestUpdateAPI_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentExecutorMockObj := deploymentmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		deploymentExecutorMockObj.EXPECT().UpdateApp(appId, nil).Return(nil),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(POST, urls.Base()+urls.Management()+urls.Apps()+"/"+appId+urls.Update(), nil)

	deploymentExecutor = deploymentExecutorMockObj

	deploymentAPIExecutor.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected return OK, Actual Return : %d", w.Code)
	}
}

func TestUpdateAPIWhenControllerFailed_ExpectReturnError(t *testing.T) {
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

		deploymentAPIExecutor.Handle(w, req)

		if w.Code != test.expectCode {
			t.Errorf("Expected error code : %d, Actual error code : %d\n", test.expectCode, w.Code)
		}
	}
}

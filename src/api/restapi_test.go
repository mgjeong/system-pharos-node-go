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
package api

import (
	"encoding/json"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	configurationapi "api/configuration/mocks"
	deploymentapi "api/deployment/mocks"
	deviceapi "api/device/mocks"
	healthapi "api/health/mocks"
	resourceapi "api/monitoring/resource/mocks"
	notificationapi "api/notification/mocks"
)

const (
	GET    string = "GET"
	PUT    string = "PUT"
	POST   string = "POST"
	DELETE string = "DELETE"

	appId1 = "000000000000000000000000"
)

func TestInvalidUrlList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlList := make(map[string][]string)
	urlList["/test"] = []string{GET, PUT, POST, DELETE}
	urlList["/api/v1/test"] = []string{GET, PUT, POST, DELETE}
	urlList["/api/v1/apps/11/test"] = []string{GET, PUT, POST, DELETE}
	urlList["/api/v1/apps/11/test/"] = []string{GET, PUT, POST, DELETE}

	for key, vals := range urlList {
		for _, method := range vals {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, key, nil)

			NodeAPIs.ServeHTTP(w, req)

			msg := make(map[string]interface{})
			err := json.Unmarshal(w.Body.Bytes(), &msg)
			if err != nil {
				t.Error("Expected results : invalid method msg, Actual err : json unmarshal failed.")
			}

			if !strings.Contains(msg["message"].(string), "unsupported url") {
				t.Errorf("Expected results : invalid method msg, Actual err : %s.", msg["message"])
			}
		}
	}
}

func TestServeHTTPsendUnregisterAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	healthAPIExecutorMockObj := healthapi.NewMockCommand(ctrl)

	gomock.InOrder(
		healthAPIExecutorMockObj.EXPECT().Handle(gomock.Any(), gomock.Any()),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/management/unregister", nil)

	healthAPIExecutor = healthAPIExecutorMockObj
	NodeAPIs.ServeHTTP(w, req)
}

func TestServeHTTPsendDeploymentAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentAPIExecutorMockObj := deploymentapi.NewMockCommand(ctrl)

	urlList := make(map[string][]string)
	urlList["/api/v1/management/apps"] = []string{GET}
	urlList["/api/v1/management/apps/deploy"] = []string{POST}
	urlList["/api/v1/management/apps/"+appId1] = []string{GET, POST, DELETE}
	urlList["/api/v1/management/apps/"+appId1+"/update"] = []string{POST}
	urlList["/api/v1/management/apps/"+appId1+"/stop"] = []string{POST}
	urlList["/api/v1/management/apps/"+appId1+"/start"] = []string{POST}

	for key, vals := range urlList {
		for _, method := range vals {
			gomock.InOrder(
				deploymentAPIExecutorMockObj.EXPECT().Handle(gomock.Any(), gomock.Any()),
			)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, key, nil)

			deploymentAPIExecutor = deploymentAPIExecutorMockObj
			NodeAPIs.ServeHTTP(w, req)
		}
	}
}

func TestServeHTTPsendResourceAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceAPIExecutorMockObj := resourceapi.NewMockCommand(ctrl)

	urlList := make(map[string][]string)
	urlList["/api/v1/monitoring/resource"] = []string{GET}
	urlList["/api/v1/monitoring/apps/"+appId1+"/resource"] = []string{GET}

	for key, vals := range urlList {
		for _, method := range vals {
			gomock.InOrder(
				resourceAPIExecutorMockObj.EXPECT().Handle(gomock.Any(), gomock.Any()),
			)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, key, nil)

			resourceAPIExecutor = resourceAPIExecutorMockObj
			NodeAPIs.ServeHTTP(w, req)
		}
	}
}

func TestServeHTTPsendDeviceAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deviceAPIExecutorMockObj := deviceapi.NewMockCommand(ctrl)

	urlList := make(map[string][]string)
	urlList["/api/v1/management/device/reboot"] = []string{POST}
	urlList["/api/v1/management/device/restore"] = []string{POST}

	for key, vals := range urlList {
		for _, method := range vals {
			gomock.InOrder(
				deviceAPIExecutorMockObj.EXPECT().Handle(gomock.Any(), gomock.Any()),
			)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, key, nil)

			deviceAPIExecutor = deviceAPIExecutorMockObj
			NodeAPIs.ServeHTTP(w, req)
		}
	}
}

func TestServeHTTPsendConfigurationApi(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configApiExecutorMockObj := configurationapi.NewMockCommand(ctrl)

	urlList := make(map[string][]string)
	urlList["/api/v1/management/device/configuration"] = []string{GET, POST}

	for key, vals := range urlList {
		for _, method := range vals {
			gomock.InOrder(
				configApiExecutorMockObj.EXPECT().Handle(gomock.Any(), gomock.Any()),
			)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, key, nil)

			configurationAPIExecutor = configApiExecutorMockObj
			NodeAPIs.ServeHTTP(w, req)
		}
	}
}

func TestServeHTTPsendNotificationApi(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notiApiExecutorMockObj := notificationapi.NewMockCommand(ctrl)

	urlList := make(map[string][]string)
	urlList["/api/v1/notification/apps/watch"] = []string{POST, DELETE}

	for key, vals := range urlList {
		for _, method := range vals {
			gomock.InOrder(
				notiApiExecutorMockObj.EXPECT().Handle(gomock.Any(), gomock.Any()),
			)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, key, nil)

			notificationAPIExecutor = notiApiExecutorMockObj
			NodeAPIs.ServeHTTP(w, req)
		}
	}
}

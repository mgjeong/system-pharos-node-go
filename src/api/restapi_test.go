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

	deploymentapi "api/deployment/mocks"
	healthapi "api/health/mocks"
	resourceapi "api/monitoring/resource/mocks"
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

			NodeApis.ServeHTTP(w, req)

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

func TestServeHTTPsendUnregisterApi(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	healthApiExecutorMockObj := healthapi.NewMockCommand(ctrl)

	gomock.InOrder(
		healthApiExecutorMockObj.EXPECT().Handle(gomock.Any(), gomock.Any()),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/management/unregister", nil)

	healthApiExecutor = healthApiExecutorMockObj
	NodeApis.ServeHTTP(w, req)
}

func TestServeHTTPsendDeploymentApi(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deploymentApiExecutorMockObj := deploymentapi.NewMockCommand(ctrl)

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
				deploymentApiExecutorMockObj.EXPECT().Handle(gomock.Any(), gomock.Any()),
			)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, key, nil)

			deploymentApiExecutor = deploymentApiExecutorMockObj
			NodeApis.ServeHTTP(w, req)
		}
	}
}

func TestServeHTTPsendResourceApi(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceApiExecutorMockObj := resourceapi.NewMockCommand(ctrl)

	urlList := make(map[string][]string)
	urlList["/api/v1/monitoring/resource"] = []string{GET}
	urlList["/api/v1/monitoring/apps/"+appId1+"/resource"] = []string{GET}

	for key, vals := range urlList {
		for _, method := range vals {
			gomock.InOrder(
				resourceApiExecutorMockObj.EXPECT().Handle(gomock.Any(), gomock.Any()),
			)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, key, nil)

			resourceApiExecutor = resourceApiExecutorMockObj
			NodeApis.ServeHTTP(w, req)
		}
	}
}

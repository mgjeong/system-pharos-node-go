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

package configuration

import (
	"bytes"
	configurationmocks "controller/configuration/mocks"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testBodyString = `{"test":"body"}`
)

var testBody = map[string]interface{}{
	"test": "body",
}

var Handler Command

func init() {
	Handler = Executor{}
}

func TestCalledHandleWithInvalidURL_UnExpectCalledAnyHandle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurationMockObj := configurationmocks.NewMockCommand(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/management/invalid", nil)

	// pass mockObj to a real object.
	configurationExecutor = configurationMockObj

	Handler.Handle(w, req)
}

func TestCalledHandleWithExcludedBaseURL_UnExpectCalledAnyHandle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurationMockObj := configurationmocks.NewMockCommand(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/management/device/configuration", nil)

	// pass mockObj to a real object.
	configurationExecutor = configurationMockObj

	Handler.Handle(w, req)
}

func TestCalledHandleWithGetConfigurationRequest_ExpectCalledGetResourceInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurationMockObj := configurationmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		configurationMockObj.EXPECT().GetConfiguration(),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/management/device/configuration", nil)

	// pass mockObj to a real object.
	configurationExecutor = configurationMockObj

	Handler.Handle(w, req)
}

func TestCalledHandleWithSetResourceRequest_ExpectCalledGetPerformanceInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurationMockObj := configurationmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		configurationMockObj.EXPECT().SetConfiguration(testBodyString),
	)

	w := httptest.NewRecorder()
	body, _ := json.Marshal(testBody)
	req, _ := http.NewRequest("POST", "/api/v1/management/device/configuration", bytes.NewReader(body))

	// pass mockObj to a real object.
	configurationExecutor = configurationMockObj

	Handler.Handle(w, req)
}

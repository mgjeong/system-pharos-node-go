/*******************************************************************************
 * Copyright 2018 Samsung Electronics All Rights Reserved.
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

package notification

import (
	appmocks "api/notification/apps/mocks"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

var executor Command

func init() {
	executor = Executor{}
}

func TestCalledHandleWithInvalidURL_UnExpectCalledAnyHandle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/invalid", nil)

	executor.Handle(w, req)
}

func TestCalledHandleWithExcludedBaseURL_UnExpectCalledAnyHandle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	appHandlerMockObj := appmocks.NewMockCommand(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/apps/watch", nil)

	// pass mockObj to a real object.
	appsNotificationHandler = appHandlerMockObj

	executor.Handle(w, req)
}

func TestCalledHandleWithSubscribeAppEventRequest_ExpectCalledAppHandle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	appHandlerMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		appHandlerMockObj.EXPECT().Handle(gomock.Any(), gomock.Any()),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/notification/apps/watch", nil)

	// pass mockObj to a real object.
	appsNotificationHandler = appHandlerMockObj

	executor.Handle(w, req)
}

func TestCalledHandleWithUnsubscribeAppEventRequest_ExpectCalledAppHandle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	appHandlerMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		appHandlerMockObj.EXPECT().Handle(gomock.Any(), gomock.Any()),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/notification/apps/watch", nil)

	// pass mockObj to a real object.
	appsNotificationHandler = appHandlerMockObj

	executor.Handle(w, req)
}

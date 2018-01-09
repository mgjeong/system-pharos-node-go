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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	GET    string = "GET"
	PUT    string = "PUT"
	POST   string = "POST"
	DELETE string = "DELETE"
)

// Test
var status int
var head http.Header

type testResponseWriter struct {
}

func init() {
	NodeApis = Executor{}
}

func (w testResponseWriter) Header() http.Header {
	return head
}

func (w testResponseWriter) Write(b []byte) (int, error) {
	if string(b) == http.StatusText(http.StatusOK) {
		w.WriteHeader(http.StatusOK)
	}
	return 0, nil
}

func (w testResponseWriter) WriteHeader(code int) {
	status = code
}

func newRequest(method string, url string, body io.Reader) *http.Request {
	status = 0
	head = make(map[string][]string)

	r, _ := http.NewRequest(method, url, body)
	r.URL.Path = url
	return r
}

func invalidOperation(t *testing.T, method string, url string, code int) {
	w, req := testResponseWriter{}, newRequest(method, url, nil)
	NodeApis.ServeHTTP(w, req)

	t.Log(status)
	if status != code {
		t.Error()
	}
}

func getInvalidUrlList() map[string][]string {
	urlList := make(map[string][]string)
	urlList["/test"] = []string{GET, PUT, POST, DELETE}
	urlList["/api/v1/test"] = []string{GET, PUT, POST, DELETE}
	urlList["/api/v1/apps/11/test"] = []string{GET, PUT, POST, DELETE}
	urlList["/api/v1/apps/11/test/"] = []string{GET, PUT, POST, DELETE}

	return urlList
}

func TestInvalidUrl(t *testing.T) {
	urlList := getInvalidUrlList()

	for key, vals := range urlList {
		for _, tc := range vals {
			t.Run(key+"="+tc, func(t *testing.T) {
				invalidOperation(t, tc, key, http.StatusNotFound)
			})
		}
	}
}

type deploymentApiExecutorMock struct {
	handlerCall bool
}

type healthApiExecutorMock struct {
	handlerCall bool
}

type resourceApiExecutorMock struct {
	handlerCall bool
}

var dm deploymentApiExecutorMock
var hm healthApiExecutorMock
var rm resourceApiExecutorMock

func setUp() func() {
	dm.handlerCall = false
	hm.handlerCall = false
	rm.handlerCall = false
	defaultDeploymentApiExecutor := deploymentApiExecutor
	defaultHealthApiExecutor := healthApiExecutor
	defaultResourceApiExecutor := resourceApiExecutor
	deploymentApiExecutor = &dm
	healthApiExecutor = &hm
	resourceApiExecutor = &rm

	return func() {
		deploymentApiExecutor = defaultDeploymentApiExecutor
		healthApiExecutor = defaultHealthApiExecutor
		resourceApiExecutor = defaultResourceApiExecutor
	}
}

func TestServeHTTPsendDeploymentApi(t *testing.T) {
	tearDown := setUp()
	defer tearDown()
	w := httptest.NewRecorder()
	req := newRequest("POST", "localhost:48098/api/v1/management/apps/deploy", nil)
	NodeApis.ServeHTTP(w, req)

	if rm.handlerCall && hm.handlerCall && !dm.handlerCall {
		t.Error("TestServeHTTPsendDeploymentApi is invalid")
	}
}

func TestServeHTTPsendUnregisterApi(t *testing.T) {
	tearDown := setUp()
	defer tearDown()
	w := httptest.NewRecorder()
	req := newRequest("POST", "localhost:48098/api/v1/management/unregister", nil)
	NodeApis.ServeHTTP(w, req)

	if rm.handlerCall && !hm.handlerCall && dm.handlerCall {
		t.Error("TestServeHTTPsendUnregisterApi is invalid")
	}
}

func TestServeHTTPsendResourceApi(t *testing.T) {
	tearDown := setUp()
	defer tearDown()
	w := httptest.NewRecorder()
	req := newRequest("POST", "localhost:48098/api/v1/monitoring/resource", nil)
	NodeApis.ServeHTTP(w, req)

	if !rm.handlerCall && hm.handlerCall && dm.handlerCall {
		t.Error("TestServeHTTPsendResourceApi is invalid")
	}
}

func (dm *deploymentApiExecutorMock) Handle(w http.ResponseWriter, req *http.Request) {
	dm.handlerCall = true
}

func (hm *healthApiExecutorMock) Handle(w http.ResponseWriter, req *http.Request) {
	hm.handlerCall = true
}

func (rm *resourceApiExecutorMock) Handle(w http.ResponseWriter, req *http.Request) {
	rm.handlerCall = true
}

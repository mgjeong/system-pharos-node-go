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
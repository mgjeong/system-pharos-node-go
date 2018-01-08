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
package health

import (
	"bytes"
	"commons/errors"
	urls "commons/url"
	"io"
	"net/http"
	"net/url"
	"testing"
)

// Test
var status int
var head http.Header
var healthApiExecutor Command

type testResponseWriter struct {
}

func init() {
	healthApiExecutor = Executor{}
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
	healthApiExecutor.Handle(w, req)

	t.Log(status)
	if status != code {
		t.Error()
	}
}

func getInvalidMethodList() map[string][]string {
	urlList := make(map[string][]string)
	urlList["/api/v1/management/unregister"] = []string{GET, PUT, DELETE}

	return urlList
}

func TestInvalidMethod(t *testing.T) {
	urlList := getInvalidMethodList()

	for key, vals := range urlList {
		for _, tc := range vals {
			t.Run(key+"="+tc, func(t *testing.T) {
				invalidOperation(t, tc, key, http.StatusMethodNotAllowed)
			})
		}
	}
}

func setup(t *testing.T, mock mockinghci) func(*testing.T) {
	healthExecutor = mock
	return func(*testing.T) {}
}

// Test using mocking for healthcheckinterface
var doSomethingFunc func(*mockinghci)

type mockinghci struct {
	err error
}

func makeMockingHci() mockinghci {
	mock := mockinghci{}
	mock.err = nil
	return mock
}

func (m mockinghci) Unregister() error {
	doSomethingFunc(&m)
	return m.err
}

func getBody() io.Reader {
	data := url.Values{}
	data.Set("name", "test")
	return bytes.NewBufferString(data.Encode())
}

type returnValue struct {
	id   string
	err  error
	path string
}

func executeFuncImpl(t *testing.T, method string, url string, isBody bool) {
	var body io.Reader
	body = nil
	if isBody {
		body = getBody()
	}
	w, req := testResponseWriter{}, newRequest(method, url, body)
	healthApiExecutor.Handle(w, req)

	t.Log(status)
	t.Log(head)
}

func executeFunc(t *testing.T, method string, url string, r returnValue, isBody bool) {
	doSomethingFunc = func(m *mockinghci) {
		m.err = r.err
	}
	executeFuncImpl(t, method, url, isBody)
}

type testObj struct {
	name       string
	err        error
	expectCode int
}

func getErrorTestList() []testObj {
	testList := []testObj{
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
	return testList
}

func TestUnregister(t *testing.T) {
	mock := makeMockingHci()
	tearDown := setup(t, mock)
	defer tearDown(t)

	t.Run("Success", func(t *testing.T) {
		r := returnValue{id: "", err: nil, path: ""}
		executeFunc(t, POST, urls.Base()+urls.Management()+urls.Unregister(), r, true)

		if status != http.StatusOK {
			t.Error()
		}
	})

	testList := getErrorTestList()
	for _, test := range testList {
		t.Run("Error/"+test.name, func(t *testing.T) {
			r := returnValue{id: "", err: test.err, path: ""}
			executeFunc(t, POST, urls.Base()+urls.Management()+urls.Unregister(), r, true)
			if status != test.expectCode {
				t.Error()
			}
		})
	}
}
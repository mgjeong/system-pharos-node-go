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
var resourceApiExecutor Command

type testResponseWriter struct {
}

func init() {
	resourceApiExecutor = Executor{}
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
	resourceApiExecutor.Handle(w, req)

	t.Log(status)
	if status != code {
		t.Error()
	}
}

func getInvalidMethodList() map[string][]string {
	urlList := make(map[string][]string)
	urlList["/api/v1/monitoring/resource"] = []string{POST, PUT, DELETE}
	urlList["/api/v1/monitoring/resource/performance"] = []string{POST, PUT, DELETE}

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

func setup(t *testing.T, mock mockingci) func(*testing.T) {
	resourceExecutor = mock
	return func(*testing.T) {}
}

var getResourceInfoCalled bool
var getPerformanceInfoCalled bool

// Test using mocking for resourceinterface
var doSomethingFunc func(*mockingci)

type mockingci struct {
	data map[string]interface{}
	err  error
}

func makeMockingci() mockingci {
	mock := mockingci{}
	mock.data = make(map[string]interface{})
	mock.err = nil
	return mock
}

func (m mockingci) GetResourceInfo() (map[string]interface{}, error) {
	doSomethingFunc(&m)
	getResourceInfoCalled = true
	return m.data, m.err
}

func (m mockingci) GetPerformanceInfo() (map[string]interface{}, error) {
	doSomethingFunc(&m)
	getPerformanceInfoCalled = true
	return m.data, m.err
}

func getBody() io.Reader {
	data := url.Values{}
	data.Set("name", "test")
	return bytes.NewBufferString(data.Encode())
}

type returnValue struct {
	err error
}

func executeFuncImpl(t *testing.T, method string, url string, isBody bool) {
	var body io.Reader
	body = nil
	if isBody {
		body = getBody()
	}
	w, req := testResponseWriter{}, newRequest(method, url, body)
	resourceApiExecutor.Handle(w, req)

	t.Log(status)
	t.Log(head)
}

func executeFunc(t *testing.T, method string, url string, r returnValue, isBody bool) {
	doSomethingFunc = func(m *mockingci) {
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
		{"UnknownError", errors.Unknown{}, http.StatusInternalServerError},
	}
	return testList
}

func TestResource(t *testing.T) {
	mock := makeMockingci()
	tearDown := setup(t, mock)
	defer tearDown(t)

	t.Run("Success", func(t *testing.T) {
		getResourceInfoCalled = false
		getPerformanceInfoCalled = false

		r := returnValue{err: nil}
		executeFunc(t, GET, urls.Base()+urls.Monitoring()+urls.Resource(), r, true)

		if status != http.StatusOK {
			t.Error()
		}
		if getResourceInfoCalled == false || getPerformanceInfoCalled == true {
			t.Error()
		}
	})

	testList := getErrorTestList()
	for _, test := range testList {
		t.Run("Error/"+test.name, func(t *testing.T) {
			getResourceInfoCalled = false
			getPerformanceInfoCalled = false

			r := returnValue{err: test.err}
			executeFunc(t, GET, urls.Base()+urls.Monitoring()+urls.Resource(), r, true)

			if status != test.expectCode {
				t.Error()
			}
			if getResourceInfoCalled == false || getPerformanceInfoCalled == true {
				t.Error()
			}
		})
	}
}

func TestPerformance(t *testing.T) {
	mock := makeMockingci()
	tearDown := setup(t, mock)
	defer tearDown(t)

	t.Run("Success", func(t *testing.T) {
		r := returnValue{err: nil}
		executeFunc(t, GET, urls.Base()+urls.Monitoring()+urls.Resource()+urls.Performance(), r, true)

		if status != http.StatusOK {
			t.Error()
		}
	})

	testList := getErrorTestList()
	for _, test := range testList {
		t.Run("Error/"+test.name, func(t *testing.T) {
			r := returnValue{err: test.err}
			executeFunc(t, GET, urls.Base()+urls.Monitoring()+urls.Performance(), r, true)
			if status != test.expectCode {
				t.Error()
			}
		})
	}
}
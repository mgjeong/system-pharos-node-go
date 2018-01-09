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
	"io"
	"net/http"
	"net/url"
	"testing"
)

// Test
var status int
var head http.Header
var deploymentApiExecutor Command

type testResponseWriter struct {
}

func init() {
	deploymentApiExecutor = Executor{}
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
	deploymentApiExecutor.Handle(w, req)

	t.Log(status)
	if status != code {
		t.Error()
	}
}

func getInvalidMethodList() map[string][]string {
	urlList := make(map[string][]string)
	urlList["/api/v1/management/apps"] = []string{PUT, POST, DELETE}
	urlList["/api/v1/management/apps/deploy"] = []string{GET, PUT, DELETE}
	urlList["/api/v1/management/apps/11"] = []string{PUT}
	urlList["/api/v1/management/apps/11/update"] = []string{GET, PUT, DELETE}
	urlList["/api/v1/management/apps/11/stop"] = []string{GET, PUT, DELETE}
	urlList["/api/v1/management/apps/11/start"] = []string{GET, PUT, DELETE}

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

// Test using mocking for controllerinterface
var doSomethingFunc func(*mockingci)
var doSomethingFuncWithAppId func(*mockingci, string)

type mockingci struct {
	data map[string]interface{}
	path string
	err  error
}

func makeMockingCi() mockingci {
	mock := mockingci{}
	mock.data = make(map[string]interface{})
	mock.err = nil
	return mock
}

func (m mockingci) DeployApp(body string) (map[string]interface{}, error) {
	doSomethingFunc(&m)
	return m.data, m.err
}
func (m mockingci) App(appId string) (map[string]interface{}, error) {
	doSomethingFuncWithAppId(&m, appId)
	return m.data, m.err
}
func (m mockingci) Apps() (map[string]interface{}, error) {
	doSomethingFunc(&m)
	return m.data, m.err
}
func (m mockingci) UpdateAppInfo(appId string, body string) error {
	doSomethingFuncWithAppId(&m, appId)
	return m.err
}
func (m mockingci) DeleteApp(appId string) error {
	doSomethingFuncWithAppId(&m, appId)
	return m.err
}
func (m mockingci) UpdateApp(appId string) error {
	doSomethingFuncWithAppId(&m, appId)
	return m.err
}
func (m mockingci) StopApp(appId string) error {
	doSomethingFuncWithAppId(&m, appId)
	return m.err
}
func (m mockingci) StartApp(appId string) error {
	doSomethingFuncWithAppId(&m, appId)
	return m.err
}

func getBody() io.Reader {
	data := url.Values{}
	data.Set("name", "test")
	return bytes.NewBufferString(data.Encode())
}

func setup(t *testing.T, mock mockingci) func(*testing.T) {
	deploymentExecutor = mock
	return func(*testing.T) {}
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
	deploymentApiExecutor.Handle(w, req)

	t.Log(status)
	t.Log(head)
}

func executeFunc(t *testing.T, method string, url string, r returnValue, isBody bool) {
	doSomethingFunc = func(m *mockingci) {
		m.data["id"] = r.id
		m.err = r.err
		m.path = r.path
	}
	executeFuncImpl(t, method, url, isBody)
}

func executeFuncWithAppId(t *testing.T, appId string, method string, url string, r returnValue, isBody bool) {
	doSomethingFuncWithAppId = func(m *mockingci, id string) {

		m.data["id"] = r.id
		m.err = r.err
		m.path = r.path
		if appId != id {
			t.Error()
		}
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

func TestDeploy(t *testing.T) {
	mock := makeMockingCi()
	tearDown := setup(t, mock)
	defer tearDown(t)

	id := "12345"

	t.Run("Success", func(t *testing.T) {
		r := returnValue{id: id, err: nil, path: ""}
		executeFunc(t, POST, urls.Base()+urls.Management()+urls.Apps()+urls.Deploy(), r, true)
		if status != http.StatusOK ||
			head.Get("Location") != urls.Base()+urls.Management()+urls.Apps()+"/"+id {
			t.Error()
		}
	})

	testList := getErrorTestList()

	for _, test := range testList {
		t.Run("Error/"+test.name, func(t *testing.T) {
			r := returnValue{id: id, err: test.err, path: ""}
			executeFunc(t, POST, urls.Base()+urls.Management()+urls.Apps()+urls.Deploy(), r, true)

			if status != test.expectCode {
				t.Error()
			}
		})
	}

	t.Run("ErrorEmptyBody", func(t *testing.T) {
		r := returnValue{id: id, err: errors.Unknown{}, path: ""}
		executeFunc(t, POST, urls.Base()+urls.Management()+urls.Apps()+urls.Deploy(), r, false)

		if status == http.StatusOK {
			t.Error()
		}
	})

}

func TestApps(t *testing.T) {
	mock := makeMockingCi()
	tearDown := setup(t, mock)
	defer tearDown(t)

	t.Run("Success", func(t *testing.T) {
		r := returnValue{id: "", err: nil, path: ""}
		executeFunc(t, GET, urls.Base()+urls.Management()+urls.Apps(), r, false)

		if status != http.StatusOK {
			t.Error()
		}
	})

	testList := getErrorTestList()
	for _, test := range testList {
		t.Run("Error/"+test.name, func(t *testing.T) {
			r := returnValue{id: "", err: test.err, path: ""}
			executeFunc(t, GET, urls.Base()+urls.Management()+urls.Apps(), r, false)
			if status != test.expectCode {
				t.Error()
			}
		})
	}
}

func TestApp(t *testing.T) {
	mock := makeMockingCi()
	tearDown := setup(t, mock)
	defer tearDown(t)

	var list []string = []string{GET, POST, DELETE}

	id := "111"
	url := urls.Base() + urls.Management() + urls.Apps() + "/" + id

	for _, method := range list {
		t.Run(method+"/Success", func(t *testing.T) {
			r := returnValue{id: id, err: nil, path: ""}
			executeFuncWithAppId(t, id, method, url, r, true)

			if status != http.StatusOK {
				t.Error()
			}
		})

		testList := getErrorTestList()
		for _, test := range testList {
			t.Run(method+"/Error/"+test.name, func(t *testing.T) {
				r := returnValue{id: id, err: test.err, path: ""}
				executeFuncWithAppId(t, id, method, url, r, true)

				if status != test.expectCode {
					t.Error()
				}
			})
		}
	}

	t.Run(POST+"/ErrorEmptyBody", func(t *testing.T) {
		r := returnValue{id: id, err: errors.Unknown{}, path: ""}
		executeFuncWithAppId(t, id, POST, url, r, false)

		if status == http.StatusOK {
			t.Error()
		}
	})
}

func TestFunc(t *testing.T) {
	mock := makeMockingCi()
	tearDown := setup(t, mock)
	defer tearDown(t)

	id := "111"
	url := urls.Base() + urls.Management() + urls.Apps() + "/" + id

	urls := []string{
		urls.Update(), urls.Start(), urls.Stop(),
	}

	for _, u := range urls {
		t.Run(u+"/Success", func(t *testing.T) {
			r := returnValue{id: "", err: nil, path: ""}
			executeFuncWithAppId(t, id, POST, url+u, r, false)

			if status != http.StatusOK {
				t.Error()
			}
		})

		testList := getErrorTestList()
		for _, test := range testList {
			t.Run(u+"/Error/"+test.name, func(t *testing.T) {
				r := returnValue{id: "", err: test.err, path: ""}
				executeFuncWithAppId(t, id, POST, url+u, r, false)

				if status != test.expectCode {
					t.Error()
				}
			})
		}
	}
}

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
package common

import (
	"bytes"
	"commons/errors"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testBodyString = `{"test":"body"}`
)

var testBody = map[string]interface{}{
	"test": "body",
}

func TestMakeErrorResponseWithInternalDefinedError_ExpectConversionToMatchHttpErrorCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := httptest.NewRecorder()

	MakeErrorResponse(w, errors.NotFoundURL{})
	if w.Code != http.StatusNotFound {
		t.Errorf("Unexpected Error code : %d", w.Code)
	}

	w = httptest.NewRecorder()
	MakeErrorResponse(w, errors.InvalidMethod{})
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Unexpected Error code : %d", w.Code)
	}

	w = httptest.NewRecorder()
	MakeErrorResponse(w, errors.InvalidYaml{})
	if w.Code != http.StatusBadRequest {
		t.Errorf("Unexpected Error code : %d", w.Code)
	}

	w = httptest.NewRecorder()
	MakeErrorResponse(w, errors.IOError{})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Unexpected Error code : %d", w.Code)
	}

	w = httptest.NewRecorder()
	MakeErrorResponse(w, errors.ConnectionError{})
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Unexpected Error code : %d", w.Code)
	}

	w = httptest.NewRecorder()
	MakeErrorResponse(w, errors.AlreadyReported{})
	if w.Code != http.StatusAlreadyReported {
		t.Errorf("Unexpected Error code : %d", w.Code)
	}

	w = httptest.NewRecorder()
	MakeErrorResponse(w, errors.Unknown{})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Unexpected Error code : %d", w.Code)
	}
}

func TestMakeResponseWithEmptyData_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var data []byte
	w := httptest.NewRecorder()

	MakeResponse(w, data)
	if w.Code != http.StatusOK {
		t.Errorf("Unexpected Error code : %d", w.Code)
	}
}

func TestMakeResponse_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	data := []byte{'1', '2', '3'}
	w := httptest.NewRecorder()

	MakeResponse(w, data)
	if w.Code != http.StatusOK {
		t.Errorf("Unexpected Error code : %d", w.Code)
	}
}

func TestCheckSupportedMethodWithValidMethod_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := httptest.NewRecorder()

	if !CheckSupportedMethod(w, POST, GET, POST, DELETE) {
		t.Error("Not Supported Method")
	}
}

func TestCheckSupportedMethodWithInvalidMethod_ExpectResponseWithMethodNotAllowedCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := httptest.NewRecorder()

	if !CheckSupportedMethod(w, PUT, GET, POST, DELETE) {
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Unexpected Error code : %d", w.Code)
		}
	}
}

func TestGetBodyFromReqWithValidBody_ExpectReturnBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	body, _ := json.Marshal(testBody)
	req, _ := http.NewRequest(POST, "test/url", bytes.NewReader(body))

	stringBody, err := GetBodyFromReq(req)
	if err != nil {
		t.Error("invalid body")
	}
	if strings.Compare(stringBody, testBodyString) != 0 {
		t.Error("The value of the parsed body is incorrect.")
	}
}

func TestGetBodyFromReqWithEmptyBody_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, _ := http.NewRequest(POST, "test/url", nil)

	_, err := GetBodyFromReq(req)
	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParam", "nil")
	}
}

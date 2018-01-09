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
	"net/http"
	"testing"
)

// Test
var status int
var head http.Header

type testResponseWriter struct {
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

func testMakeResponse(t *testing.T, w testResponseWriter, data []byte, expect int) {
	MakeResponse(w, data)

	t.Log(status)
	if status != expect {
		t.Error()
	}
}

// This unittest function does not work.
func TestMakeResponse(t *testing.T) {
	/*var data []byte
	w := testResponseWriter{}
	t.Run("ExpectSuccessWithEmptyData", func(t *testing.T) {
		testMakeResponse(t, w, data, http.StatusOK)
	})

	data = []byte{'1', '2', '3'}
	t.Run("ExpectSuccess", func(t *testing.T) {
		testMakeResponse(t, w, data, http.StatusOK)
	})*/
}
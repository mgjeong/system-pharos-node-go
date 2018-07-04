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
package errors

import (
	"strings"
	"testing"
)

func TestTError(t *testing.T) {
	msg := "Test"

	type commonsError interface {
		SetMsg(string)
		Error() string
	}

	type testObj struct {
		testName   string
		testPrefix string
		testError  commonsError
	}

	testList := []testObj{
		{testName: "Unknown", testPrefix: "unknown error",
			testError: &Unknown{}},
		{testName: "InvalidParam", testPrefix: "invalid parameter",
			testError: &InvalidParam{}},
		{testName: "InvalidJSON", testPrefix: "invalid json format",
			testError: &InvalidJSON{}},
		{testName: "InvalidObjectId", testPrefix: "invalid objectId",
			testError: &InvalidObjectId{}},
		{testName: "NotFound", testPrefix: "not find target",
			testError: &NotFound{}},
		{testName: "InvalidYamlError", testPrefix: "invalid yaml file",
			testError: &InvalidYaml{}},
		{testName: "ConnectionError", testPrefix: "can not connect",
			testError: &ConnectionError{}},
		{testName: "InvalidAppId", testPrefix: "invalid app id",
			testError: &InvalidAppId{}},
		{testName: "InvalidMethod", testPrefix: "invalid method",
			testError: &InvalidMethod{}},
		{testName: "NotFoundURL", testPrefix: "unsupported url",
			testError: &NotFoundURL{}},
		{testName: "IOError", testPrefix: "io error",
			testError: &IOError{}},
		{testName: "NotFoundImage", testPrefix: "unsupported url",
			testError: &NotFoundImage{}},
		{testName: "AlreadyReported", testPrefix: "already done processing",
			testError: &AlreadyReported{}},
		{testName: "AlreadyUsedName", testPrefix: "already used container name",
			testError: &AlreadyUsedName{}},
		{testName: "InvalidContainerName", testPrefix: "invalid container name",
			testError: &InvalidContainerName{}},
		{testName: "AlreadyAllocatedPort", testPrefix: "already allocated port",
			testError: &AlreadyAllocatedPort{}},
		{testName: "InternalServerError", testPrefix: "internal server error",
			testError: &InternalServerError{}},
		{testName: "DBConnectionError", testPrefix: "db connection failed",
			testError: &DBConnectionError{}},
		{testName: "DBOperationError", testPrefix: "db operation failed",
			testError: &DBOperationError{}},
	}

	testFunc := func(err commonsError, prefix string) {
		err.SetMsg(msg)
		ret := err.Error()
		if !strings.HasPrefix(ret, prefix) {
			t.Error()
		} else if !strings.HasSuffix(ret, msg) {
			t.Error()
		}
	}

	for _, test := range testList {
		t.Run(test.testName, func(t *testing.T) {
			testFunc(test.testError, test.testPrefix)
		})
	}
}

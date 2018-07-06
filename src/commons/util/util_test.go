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
package util

import (
	"commons/errors"
	"github.com/golang/mock/gomock"
	"os"
	"testing"
)

var (
	ip                    = "127.0.0.1"
	anchorAddressEnv      = "ANCHOR_ADDRESS"
	anchorReverseProxyEnv = "ANCHOR_REVERSE_PROXY"
	properties            = map[string]interface{}{
		"properties": []map[string]interface{}{{
			"name":     "name",
			"value":    "value",
			"readOnly": false,
		}},
	}
	newProperties = map[string]interface{}{
		"properties": []map[string]interface{}{{
			"name": "value",
		}},
	}
	notFoundError = errors.NotFound{}
)

func TestMakeAnchorRequestUrlWhenRPEnvTrue_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testUrlPart := "/testurl"
	expectedUrl := "http://" + ip + ":" + UNSECURED_ANCHOR_PORT_WITH_REVERSE_PROXY + "/pharos-anchor/api/v1/management" + testUrlPart

	os.Setenv(anchorAddressEnv, ip)
	os.Setenv(anchorReverseProxyEnv, "true")
	ret, err := MakeAnchorRequestUrl(testUrlPart)
	os.Unsetenv(anchorAddressEnv)
	os.Unsetenv(anchorReverseProxyEnv)

	if err != nil {
		t.Errorf("Expected error : nil, actual error : %s", err.Error())
	}

	if ret != expectedUrl {
		t.Errorf("Expected result : %s, actual result : %s", expectedUrl, ret)
	}
}

func TestMakeAnchorRequestUrlWhenRPEnvFalse_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testUrlPart := "/testurl"
	expectedUrl := "http://" + ip + ":" + DEFAULT_ANCHOR_PORT + "/api/v1/management" + testUrlPart

	os.Setenv(anchorAddressEnv, ip)
	os.Setenv(anchorReverseProxyEnv, "false")
	ret, err := MakeAnchorRequestUrl(testUrlPart)
	os.Unsetenv(anchorAddressEnv)
	os.Unsetenv(anchorReverseProxyEnv)

	if err != nil {
		t.Errorf("Expected error : nil, actual error : %s", err.Error())
	}

	if ret != expectedUrl {
		t.Errorf("Expected result : %s, actual result : %s", expectedUrl, ret)
	}
}

func TestMakeAnchorRequestUrlWithInvalidAnchorIPEnv_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	os.Setenv("ANCHOR_ADDRESS", "192.2")
	_, err := MakeAnchorRequestUrl("")
	os.Unsetenv("ANCHOR_ADDRESS")

	if err == nil {
		t.Errorf("Expected error : InvalidParam, actual error : nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected error : %s, actual error : %s", "InvalidParam", err.Error())
	case errors.InvalidParam:
	}
}

func TestMakeAnchorRequestUrlWithNoAnchorIPEnv_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testUrlPart := "/testurl"
	_, err := MakeAnchorRequestUrl(testUrlPart)

	if err == nil {
		t.Errorf("Expected error : NotFound, actual error : nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected error : %s, actual error : %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}

func TestMakeAnchorRequestUrlWithNoAnchorRPEnv_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testUrlPart := "/testurl"
	expectedRet := "http://" + ip + ":" + DEFAULT_ANCHOR_PORT + "/api/v1/management" + testUrlPart
	os.Setenv(anchorAddressEnv, ip)
	ret, err := MakeAnchorRequestUrl(testUrlPart)
	os.Unsetenv(anchorAddressEnv)

	if err != nil {
		t.Errorf("Expected error : nil, actual error : %s", err.Error())
	}

	if expectedRet != ret {
		t.Errorf("Expected return : %s, actual return : %s", expectedRet, ret)
	}
}

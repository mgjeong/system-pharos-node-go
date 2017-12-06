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
package registration

import (
	"commons/errors"
	"github.com/golang/mock/gomock"
	msgmocks "messenger/mocks"
	"testing"
)

var regObj RegistrationInterface

func init() {
	regObj = Registration{}
}

func TestCalledRegisterWhenAlreadyRegistered_ExpectErrorReturn(t *testing.T) {
	agentId = "id"
	err := regObj.Register("")

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledRegisterWithoutBody_ExpectErrorReturn(t *testing.T) {
	agentId = ""
	err := regObj.Register("")

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParam", "nil")
	}
}

func TestCalledRegisterWithInvalidBodyNotIncludingIPField_ExpectErrorReturn(t *testing.T) {
	agentId = ""
	invalidBody := `{"key":"value"}`
	err := regObj.Register(invalidBody)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParam", "nil")
	}
}

func TestCalledRegisterWithInvalidBodyNotIncludingHealthCheckField_ExpectErrorReturn(t *testing.T) {
	agentId = ""
	invalidBody := `{"ip":"value", "key":"value"}`
	err := regObj.Register(invalidBody)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParam", "nil")
	}
}

func TestCalledRegisterWithInvalidBodyNotIncludingIntervalField_ExpectErrorReturn(t *testing.T) {
	agentId = ""
	invalidBody := `{"ip":"value", "healthCheck":{"key":"value"}}`
	err := regObj.Register(invalidBody)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParam", "nil")
	}
}

func TestCalledRegisterWhenFailedToSendRegisterRequest_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockMessengerInterface(ctrl)

	agentId = ""
	expectedBody := `{"ip":"2"}`
	url := "http://1:48099/api/v1/agents/register"
	unknownErr := errors.Unknown{"error"}

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", url, []byte(expectedBody)).Return(500, "", unknownErr),
	)
	
	httpRequester = msgMockObj
	
	body := `{"ip":{"manager":"1", "agent":"2"}, "healthCheck":{"interval":"1"}}`
	err := regObj.Register(body)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledRegisterWhenReceiveInvalidResponseBodyFromSDAM_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockMessengerInterface(ctrl)

	agentId = ""
	expectedBody := `{"ip":"2"}`
	url := "http://1:48099/api/v1/agents/register"
	invalidResp := `{"response"}`

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", url, []byte(expectedBody)).Return(200, invalidResp, nil),
	)

	httpRequester = msgMockObj
	
	body := `{"ip":{"manager":"1", "agent":"2"}, "healthCheck":{"interval":"1"}}`
	err := regObj.Register(body)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledRegisterWhenReceiveErrorResponseFromSDAM_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockMessengerInterface(ctrl)

	agentId = ""
	expectedBody := `{"ip":"2"}`
	url := "http://1:48099/api/v1/agents/register"
	resp := `{"message":"error"}`

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", url, []byte(expectedBody)).Return(500, resp, nil),
	)
	
	httpRequester = msgMockObj
	
	body := `{"ip":{"manager":"1", "agent":"2"}, "healthCheck":{"interval":"1"}}`
	err := regObj.Register(body)
	
	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledUnregisterWhenAlreadyUnregistered_ExpectErrorReturn(t *testing.T) {
	agentId = ""
	err := regObj.Unregister()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledUnregisterWhenFailedToSendUnregisterRequest_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockMessengerInterface(ctrl)
	
	agentId = "id"
	url := "http://1:48099/api/v1/agents/id/unregister"
	unknownErr := errors.Unknown{"error"}

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", url).Return(500, "", unknownErr),
	)
	
	httpRequester = msgMockObj
	
	err := regObj.Unregister()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledUnregisterWhenReceiveErrorResponseFromSDAM_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockMessengerInterface(ctrl)

	agentId = "id"
	url := "http://1:48099/api/v1/agents/id/unregister"
	resp := `{"message":"error"}`

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", url).Return(500, resp, nil),
	)
	
	httpRequester = msgMockObj
	
	err := regObj.Unregister()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledUnregisterWhenReceiveInvalidErrorResponseFromSDAM_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockMessengerInterface(ctrl)
	
	agentId = "id"
	url := "http://1:48099/api/v1/agents/id/unregister"
	invalidResp := `{"response"}`

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", url).Return(500, invalidResp, nil),
	)
	
	httpRequester = msgMockObj
	
	err := regObj.Unregister()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledUnregister_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockMessengerInterface(ctrl)
	
	agentId = "id"
	interval = "1"
	url := "http://1:48099/api/v1/agents/id/unregister"

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", url).Return(200, "", nil),
	)

	httpRequester = msgMockObj
	
	err := regObj.Unregister()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledSendPingRequestWhenFailedToSendHttpRequest_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockMessengerInterface(ctrl)
	
	agentId = "id"
	interval := "1"
	url := "http://1:48099/api/v1/agents/id/ping"
	expectedBody := `{"interval":"1"}`
	unknownErr := errors.Unknown{"error"}

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", url, []byte(expectedBody)).Return(500, "", unknownErr),
	)

	httpRequester = msgMockObj
	
	_, err := sendPingRequest(interval)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

func TestCalledSendPingRequest_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockMessengerInterface(ctrl)
	
	agentId = "id"
	interval := "1"
	url := "http://1:48099/api/v1/agents/id/ping"
	expectedBody := `{"interval":"1"}`

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", url, []byte(expectedBody)).Return(200, "", nil),
	)

	httpRequester = msgMockObj
	
	_, err := sendPingRequest(interval)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledSendRegisterRequestWhenFailedToSendHttpRequest_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockMessengerInterface(ctrl)
	
	agentId = ""
	url := "http://1:48099/api/v1/agents/register"
	expectedBody := `{"ip":"2"}`
	unknownErr := errors.Unknown{"error"}

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", url, []byte(expectedBody)).Return(500, "", unknownErr),
	)
	
	httpRequester = msgMockObj
	
	_, _, err := sendRegisterRequest()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
}

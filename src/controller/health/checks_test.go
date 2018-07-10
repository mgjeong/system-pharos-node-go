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
	dbmocks "db/bolt/configuration/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	msgmocks "messenger/mocks"
	"os"
	"testing"
)

func TestCalledSendPingRequestWhenFailedToSendHttpRequest_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockCommand(ctrl)
	dbMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbMockObj.EXPECT().GetProperty("deviceid").Return(PROPERTY, nil),
		msgMockObj.EXPECT().SendHttpRequest("POST", gomock.Any(), gomock.Any()).Return(500, "", errors.New("Error")),
	)
	configDbExecutor = dbMockObj
	httpExecutor = msgMockObj

	interval := "1"
	os.Setenv("ANCHOR_ADDRESS", "127.0.0.1")
	_, err := sendPingRequest(interval)
	os.Unsetenv("ANCHOR_ADDRESS")

	if err == nil {
		t.Errorf("Expected err: %s", err.Error())
	}
}

func TestCalledSendPingRequest_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockCommand(ctrl)
	dbMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbMockObj.EXPECT().GetProperty("deviceid").Return(PROPERTY, nil),
		msgMockObj.EXPECT().SendHttpRequest("POST", gomock.Any(), gomock.Any()).Return(200, "", nil),
	)
	configDbExecutor = dbMockObj
	httpExecutor = msgMockObj

	interval := "1"
	os.Setenv("ANCHOR_ADDRESS", "127.0.0.1")
	os.Setenv("ANCHOR_REVERSE_PROXY", "false")
	_, err := sendPingRequest(interval)
	os.Unsetenv("ANCHOR_ADDRESS")
	os.Unsetenv("ANCHOR_REVERSE_PROXY")

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

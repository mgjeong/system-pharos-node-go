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
	"errors"
	"github.com/golang/mock/gomock"
	msgmocks "messenger/mocks"
	"testing"
)

func TestCalledSendPingRequestWhenFailedToSendHttpRequest_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", gomock.Any(), gomock.Any()).Return(500, "", errors.New("Error")),
	)

	httpExecutor = msgMockObj

	interval := "1"
	_, err := sendPingRequest("id", interval)

	if err == nil {
		t.Errorf("Expected err: %s", err.Error())
	}
}

func TestCalledSendPingRequest_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msgMockObj := msgmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		msgMockObj.EXPECT().SendHttpRequest("POST", gomock.Any(), gomock.Any()).Return(200, "", nil),
	)
	httpExecutor = msgMockObj

	interval := "1"
	_, err := sendPingRequest("id", interval)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

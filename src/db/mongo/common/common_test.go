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
	"commons/errors"
	"db/mongo/wrapper/mocks"
	gomock "github.com/golang/mock/gomock"
	"testing"
)

const valid_URL = "localhost:27017"
const dummy_error_msg = "dummy_errors"
const emptyString = ""
const dbName = "DeploymentAgentDB"
const collectionName = "APP"

var dummy_error = errors.Unknown{dummy_error_msg}
var dummy_session mocks.MockSession

/*
	Unit-test for Connect
*/

func TestCalled_Connect_WithEmptyURL_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	connectionMockObj := mocks.NewMockConnection(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(emptyString).Return(&dummy_session, dummy_error),
	)

	mgoDial = connectionMockObj

	builder := MongoBuilder{}
	_ = builder.Connect(emptyString)
}

func TestCalled_Connect_WithValidURL_ExpectToSuccessWithoutPanic(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer func() {
		if recover() != nil {
			t.Fatalf("panic occured")
		}
		mockCtrl.Finish()
	}()

	connectionMockObj := mocks.NewMockConnection(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(valid_URL).Return(&dummy_session, nil),
	)

	mgoDial = connectionMockObj

	builder := MongoBuilder{}
	_ = builder.Connect(valid_URL)
}

/*
	Unit-test for CreateDB
*/

func TestCalled_CreateDB_WithInvalidSession_ExpectToErrorReturn(t *testing.T) {
	builder := MongoBuilder{}
	_, err := builder.CreateDB()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", "nil")
	}

	switch err.(type) {
	default:
		t.Error()
	case errors.InvalidParam:
	}
}

func TestCalled_CreateDB_WithValidSession_ExpectSuccessWithoutError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	connectionMockObj := mocks.NewMockConnection(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(valid_URL).Return(&dummy_session, nil),
	)

	mgoDial = connectionMockObj

	builder := MongoBuilder{}
	_ = builder.Connect(valid_URL)

	_, err := builder.CreateDB()

	if err != nil {
		t.Error()
	}
}

/*
	Unit-test for Close
*/

func TestCalled_Close_ExpectToSessionClosed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sessionMockObj := mocks.NewMockSession(mockCtrl)

	dbManager := MongoDBManager{
		mgoSession: sessionMockObj,
	}

	gomock.InOrder(
		sessionMockObj.EXPECT().Close(),
	)

	dbManager.Close()
}

/*
	Unit-test for GetCollcetion
*/

func TestCalled_GetCollcetion_ExpectToCCalled(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sessionMockObj := mocks.NewMockSession(mockCtrl)
	collectionMockObj := mocks.NewMockCollection(mockCtrl)
	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	dbManager := MongoDBManager{
		mgoSession: sessionMockObj,
	}

	gomock.InOrder(
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
	)

	dbManager.GetCollection(dbName, collectionName)
}


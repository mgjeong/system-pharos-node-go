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
package service

import (
	"commons/errors"
	"db/mongo/wrapper/mocks"
	gomock "github.com/golang/mock/gomock"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"testing"
)

const valid_URL = "localhost:27017"
const dummy_error_msg = "dummy_errors"
const emptyString = ""
const invalidID = ""
const invalidDescription = ""
const dbName = "DeploymentAgentDB"
const collectionName = "APP"
const validID = "e1f63701c26b8bbf6e41fd7c2bdf12e075b768b5"
const valid_description = `{
  "services": {
    "test_service_name": {
      "image": "test_image_name"
    }
  }
}`
const valid_state = "STATE"

var dummy_error = errors.Unknown{dummy_error_msg}
var dummy_session mocks.MockSession
var appsArgs = []App{{ID: validID, Description: valid_description, State: valid_state}}
var appArgs = App{ID: validID, Description: valid_description, State: valid_state}

/*
	Unit-test for connect
*/
func TestCalled_connect_WithEmptyURL_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	connectionMockObj := mocks.NewMockConnection(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(emptyString).Return(&dummy_session, dummy_error),
	)

	mgoDial = connectionMockObj

	_, err := connect(emptyString)
	
	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound or Unknown", "nil")
	}
	
	switch err.(type) {
	default:
		t.Error()
	case errors.NotFound:
	case errors.Unknown:
	}
}

func TestCalled_connect_WithValidURL_ExpectToSuccessWithoutPanic(t *testing.T) {
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
	_, _ = connect(valid_URL)
}

/*
	Unit-test for getCollection
*/

func TestCalled_getCollection_WithInvalidSession_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sessionMockObj := mocks.NewMockSession(mockCtrl)
	dbMockObj := mocks.NewMockDatabase(mockCtrl)
	collectionMockObj := mocks.NewMockCollection(mockCtrl)
	
	gomock.InOrder(
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
	)

	_ = getCollection(sessionMockObj, dbName, collectionName)
}

/*
	Unit-test for InsertComposeFile
*/

func TestCalled_InsertComposeFile_WithEmptyDescription_ExpectErrorReturn(t *testing.T) {
	dbManager := DBManager{}

	_, err := dbManager.InsertComposeFile(invalidDescription)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYamlError or UnknownError", "nil")
	}

	switch err.(type) {
	default:
		t.Error()
	case errors.InvalidYaml:
	case errors.Unknown:
	}
}

func TestCalled_InsertComposeFile_WithInvalidDescription_Service_ExpectErrorReturn(t *testing.T) {
	dbManager := DBManager{}

	invalid_description_without_service := `{"services":}`

	_, err := dbManager.InsertComposeFile(invalid_description_without_service)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYamlError", "nil")
	}

	switch err.(type) {
	default:
		t.Error()
	case errors.InvalidYaml:
	}
}

func TestCalled_InsertComposeFile_WithInvalidDescription_Image_ExpectErrorReturn(t *testing.T) {
	dbManager := DBManager{}

	invalid_description_without_image := `{
  "services": {
    "test_service_name": {}
  }
}`
	_, err := dbManager.InsertComposeFile(invalid_description_without_image)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYamlError", "nil")
	}
	switch err.(type) {
	default:
		t.Error()
	case errors.InvalidYaml:
	}
}

func TestCalled_InsertComposeFile_WithValidDescription_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	
	connectionMockObj := mocks.NewMockConnection(mockCtrl)
	sessionMockObj := mocks.NewMockSession(mockCtrl)
	dbMockObj := mocks.NewMockDatabase(mockCtrl)
	collectionMockObj := mocks.NewMockCollection(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(valid_URL).Return(sessionMockObj, nil),
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
		collectionMockObj.EXPECT().Insert(gomock.Any()).Return(nil),
		sessionMockObj.EXPECT().Close(),
	)
	mgoDial = connectionMockObj
	dbManager := DBManager{}
	
	expectedRes := map[string]interface{}{
		"id":          validID,
		"description": valid_description,
		"state":       "DEPLOY",
	}

	res, err := dbManager.InsertComposeFile(valid_description)

	if err != nil {
		t.Error()
	}

	if !reflect.DeepEqual(res, expectedRes) {
		t.Errorf("Expected res: %s, actual res: %s", expectedRes, res)
	}
}

/*
	Unit-test for GetAppList
*/

func TestCalled_GetAppList_WhenDBHasNotAppsData_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	connectionMockObj := mocks.NewMockConnection(mockCtrl)
	sessionMockObj := mocks.NewMockSession(mockCtrl)
	collectionMockObj := mocks.NewMockCollection(mockCtrl)
	dbMockObj := mocks.NewMockDatabase(mockCtrl)
	queryMockObj := mocks.NewMockQuery(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(valid_URL).Return(sessionMockObj, nil),
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
		collectionMockObj.EXPECT().Find(nil).Return(queryMockObj),
		queryMockObj.EXPECT().All(gomock.Any()).Return(mgo.ErrNotFound),
		sessionMockObj.EXPECT().Close(),
	)
	mgoDial = connectionMockObj
	dbManager := DBManager{}
	
	_, err := dbManager.GetAppList()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", err.Error())
	case errors.NotFound:
	case errors.Unknown:
	}
}

func TestCalled_GetAppList_WhenDBHasAppsData_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	appsArgs := []App{{ID: validID, Description: valid_description, State: valid_state}}
	expectedRes := []map[string]interface{}{{
		"id":          validID,
		"description": valid_description,
		"state":       valid_state,
	}}
	
	connectionMockObj := mocks.NewMockConnection(mockCtrl)
	sessionMockObj := mocks.NewMockSession(mockCtrl)
	collectionMockObj := mocks.NewMockCollection(mockCtrl)
	dbMockObj := mocks.NewMockDatabase(mockCtrl)
	queryMockObj := mocks.NewMockQuery(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(valid_URL).Return(sessionMockObj, nil),	
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
		collectionMockObj.EXPECT().Find(nil).Return(queryMockObj),
		queryMockObj.EXPECT().All(gomock.Any()).SetArg(0, appsArgs).Return(nil),
		sessionMockObj.EXPECT().Close(),
	)
	
	mgoDial = connectionMockObj
	dbManager := DBManager{}
	res, err := dbManager.GetAppList()

	if err != nil {
		t.Error()
	}

	if !reflect.DeepEqual(res, expectedRes) {
		t.Errorf("Expected res: %s, actual res: %s", expectedRes, res)
	}
}

/*
	Unit-test for GetApp
*/

func TestCalled_GetApp_WhenDBHasNotMatchedApp_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	query := bson.M{"_id": validID}

	connectionMockObj := mocks.NewMockConnection(mockCtrl)
	sessionMockObj := mocks.NewMockSession(mockCtrl)
	collectionMockObj := mocks.NewMockCollection(mockCtrl)
	dbMockObj := mocks.NewMockDatabase(mockCtrl)
	queryMockObj := mocks.NewMockQuery(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(valid_URL).Return(sessionMockObj, nil),
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
		collectionMockObj.EXPECT().Find(query).Return(queryMockObj),
		queryMockObj.EXPECT().One(gomock.Any()).Return(mgo.ErrNotFound),
		sessionMockObj.EXPECT().Close(),
	)

	mgoDial = connectionMockObj
	dbManager := DBManager{}
	_, err := dbManager.GetApp(validID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", err.Error())
	case errors.NotFound:
	case errors.Unknown:
	}
}

func TestCalled_GetAppList_WhenDBHasMatchedApp_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	query := bson.M{"_id": validID}

	expectedRes := map[string]interface{}{
		"id":          validID,
		"description": valid_description,
		"state":       valid_state,
	}
	
	connectionMockObj := mocks.NewMockConnection(mockCtrl)	
	sessionMockObj := mocks.NewMockSession(mockCtrl)
	collectionMockObj := mocks.NewMockCollection(mockCtrl)
	dbMockObj := mocks.NewMockDatabase(mockCtrl)
	queryMockObj := mocks.NewMockQuery(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(valid_URL).Return(sessionMockObj, nil),	
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
		collectionMockObj.EXPECT().Find(query).Return(queryMockObj),
		queryMockObj.EXPECT().One(gomock.Any()).SetArg(0, appArgs).Return(nil),
		sessionMockObj.EXPECT().Close(),	
	)

	mgoDial = connectionMockObj
	dbManager := DBManager{}
	res, err := dbManager.GetApp(validID)

	if err != nil {
		t.Error()
	}

	if !reflect.DeepEqual(res, expectedRes) {
		t.Errorf("Expected res: %s, actual res: %s", expectedRes, res)
	}
}

/*
	Unit-test for UpdateAppInfo
*/

func TestCalled_UpdateAppInfo_WithInvlaidAppID_ExpectErrorReturn(t *testing.T) {
	dbManager := DBManager{}

	err := dbManager.UpdateAppInfo(invalidID, valid_description)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", err.Error())
	case errors.InvalidParam:
	}
}

func TestCalled_UpdateAppInfo_WhenDBHasNotMatchedApp_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	query := bson.M{"_id": validID}
	update := bson.M{"$set": bson.M{"description": valid_description}}

	connectionMockObj := mocks.NewMockConnection(mockCtrl)	
	sessionMockObj := mocks.NewMockSession(mockCtrl)
	collectionMockObj := mocks.NewMockCollection(mockCtrl)
	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(valid_URL).Return(sessionMockObj, nil),	
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
		collectionMockObj.EXPECT().Update(query, update).Return(mgo.ErrNotFound),
		sessionMockObj.EXPECT().Close(),	
	)
	
	mgoDial = connectionMockObj
	dbManager := DBManager{}
	err := dbManager.UpdateAppInfo(validID, valid_description)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", err.Error())
	case errors.NotFound:
	case errors.Unknown:
	}
}

func TestCalled_UpdateAppInfo_WhenDBHasMatchedApp_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	query := bson.M{"_id": validID}
	update := bson.M{"$set": bson.M{"description": valid_description}}

	connectionMockObj := mocks.NewMockConnection(mockCtrl)	
	sessionMockObj := mocks.NewMockSession(mockCtrl)
	collectionMockObj := mocks.NewMockCollection(mockCtrl)
	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(valid_URL).Return(sessionMockObj, nil),	
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
		collectionMockObj.EXPECT().Update(query, update).Return(nil),
		sessionMockObj.EXPECT().Close(),	
	)

	mgoDial = connectionMockObj
	dbManager := DBManager{}
	err := dbManager.UpdateAppInfo(validID, valid_description)

	if err != nil {
		t.Error()
	}
}

/*
	Unit-test for DeleteApp
*/

func TestCalled_DeleteApp_WithInvlaidAppID_ExpectErrorReturn(t *testing.T) {
	dbManager := DBManager{}
	err := dbManager.DeleteApp(invalidID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "InvalidParamError", err.Error())
	case errors.InvalidParam:
	}
}

func TestCalled_DeleteApp_WhenDBHasNotMatchedApp_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	query := bson.M{"_id": validID}

	connectionMockObj := mocks.NewMockConnection(mockCtrl)	
	sessionMockObj := mocks.NewMockSession(mockCtrl)
	collectionMockObj := mocks.NewMockCollection(mockCtrl)
	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(valid_URL).Return(sessionMockObj, nil),	
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
		collectionMockObj.EXPECT().Remove(query).Return(mgo.ErrNotFound),
		sessionMockObj.EXPECT().Close(),	
	)

	mgoDial = connectionMockObj
	dbManager := DBManager{}
	err := dbManager.DeleteApp(validID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFoundError", err.Error())
	case errors.NotFound:
	case errors.Unknown:
	}
}

func TestCalled_DeleteApp_WhenDBHasMatchedApp_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	query := bson.M{"_id": validID}
	
	connectionMockObj := mocks.NewMockConnection(mockCtrl)	
	sessionMockObj := mocks.NewMockSession(mockCtrl)
	collectionMockObj := mocks.NewMockCollection(mockCtrl)
	dbMockObj := mocks.NewMockDatabase(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(valid_URL).Return(sessionMockObj, nil),	
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
		collectionMockObj.EXPECT().Remove(query).Return(nil),
		sessionMockObj.EXPECT().Close(),	
	)

	mgoDial = connectionMockObj
	dbManager := DBManager{}
	err := dbManager.DeleteApp(validID)

	if err != nil {
		t.Error()
	}
}
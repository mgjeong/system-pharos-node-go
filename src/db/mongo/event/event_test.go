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
package event

import (
	"commons/errors"
	mgomocks "db/mongo/wrapper/mocks"
	gomock "github.com/golang/mock/gomock"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"testing"
)

const (
	validUrl        = "127.0.0.1:27017"
	dbName          = "DeploymentNodeDB"
	collectionName  = "EVENT"
	eventId         = "test_event_id"
	appId           = "test_app_id"
	imageName       = "test_image_name"
	dummy_error_msg = "dummy_errors"
	emptyString     = ""
)

var (
	dummy_error   = errors.Unknown{dummy_error_msg}
	dummy_session mgomocks.MockSession
	property      = map[string]interface{}{
		"name":     "name",
		"value":    "value",
		"readOnly": false,
	}
)

func TestCalled_connect_WithEmptyURL_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	connectionMockObj := mgomocks.NewMockConnection(mockCtrl)

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

	connectionMockObj := mgomocks.NewMockConnection(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(validUrl).Return(&dummy_session, nil),
	)

	mgoDial = connectionMockObj
	_, _ = connect(validUrl)
}

func TestCalled_getCollection_WithInvalidSession_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sessionMockObj := mgomocks.NewMockSession(mockCtrl)
	dbMockObj := mgomocks.NewMockDatabase(mockCtrl)
	collectionMockObj := mgomocks.NewMockCollection(mockCtrl)

	gomock.InOrder(
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
	)

	_ = getCollection(sessionMockObj, dbName, collectionName)
}

func TestCalledInsertEvent_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	query := bson.M{"_id": eventId}

	connectionMockObj := mgomocks.NewMockConnection(mockCtrl)
	sessionMockObj := mgomocks.NewMockSession(mockCtrl)
	dbMockObj := mgomocks.NewMockDatabase(mockCtrl)
	queryMockObj := mgomocks.NewMockQuery(mockCtrl)
	collectionMockObj := mgomocks.NewMockCollection(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(validUrl).Return(sessionMockObj, nil),
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(gomock.Any()).Return(collectionMockObj),
		collectionMockObj.EXPECT().Find(query).Return(queryMockObj),
		queryMockObj.EXPECT().One(gomock.Any()).Return(mgo.ErrNotFound),
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(gomock.Any()).Return(collectionMockObj),
		collectionMockObj.EXPECT().Insert(gomock.Any()).Return(nil),
		sessionMockObj.EXPECT().Close(),
	)

	mgoDial = connectionMockObj
	executor := Executor{}

	_, err := executor.InsertEvent(eventId, appId, imageName)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledInsertEventWhenAlreadyExistsInDB_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	query := bson.M{"_id": eventId}
	arg := Event{ID: eventId, AppID: appId, ImageName: imageName}

	connectionMockObj := mgomocks.NewMockConnection(mockCtrl)
	sessionMockObj := mgomocks.NewMockSession(mockCtrl)
	dbMockObj := mgomocks.NewMockDatabase(mockCtrl)
	queryMockObj := mgomocks.NewMockQuery(mockCtrl)
	collectionMockObj := mgomocks.NewMockCollection(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(validUrl).Return(sessionMockObj, nil),
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(gomock.Any()).Return(collectionMockObj),
		collectionMockObj.EXPECT().Find(query).Return(queryMockObj),
		queryMockObj.EXPECT().One(gomock.Any()).SetArg(0, arg).Return(nil),
		sessionMockObj.EXPECT().Close(),
	)

	mgoDial = connectionMockObj
	executor := Executor{}

	_, err := executor.InsertEvent(eventId, appId, imageName)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "AlreadyReported", err.Error())
	case errors.AlreadyReported:
	}
}

func TestCalledGetEvents_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	args := []Event{{ID: eventId, AppID: appId, ImageName: imageName}}
	expectedRes := []map[string]interface{}{{
		"id":        eventId,
		"appId":     appId,
		"imageName": imageName,
	}}

	query := bson.M{"appid": bson.M{"$in": []string{"", appId}}, "imagename": bson.M{"$in": []string{"", imageName}}}

	connectionMockObj := mgomocks.NewMockConnection(mockCtrl)
	sessionMockObj := mgomocks.NewMockSession(mockCtrl)
	dbMockObj := mgomocks.NewMockDatabase(mockCtrl)
	collectionMockObj := mgomocks.NewMockCollection(mockCtrl)
	queryMockObj := mgomocks.NewMockQuery(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(validUrl).Return(sessionMockObj, nil),
		sessionMockObj.EXPECT().DB(gomock.Any()).Return(dbMockObj),
		dbMockObj.EXPECT().C(gomock.Any()).Return(collectionMockObj),
		collectionMockObj.EXPECT().Find(query).Return(queryMockObj),
		queryMockObj.EXPECT().All(gomock.Any()).SetArg(0, args).Return(nil),
		sessionMockObj.EXPECT().Close(),
	)

	mgoDial = connectionMockObj
	executor := Executor{}
	res, err := executor.GetEvents(appId, imageName)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(expectedRes, res) {
		t.Errorf("Expected res: %s, actual res: %s", expectedRes, res)
	}
}

func TestCalledGetEventsWhenDBReturnsError_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	query := bson.M{"appid": bson.M{"$in": []string{"", appId}}, "imagename": bson.M{"$in": []string{"", imageName}}}

	connectionMockObj := mgomocks.NewMockConnection(mockCtrl)
	sessionMockObj := mgomocks.NewMockSession(mockCtrl)
	dbMockObj := mgomocks.NewMockDatabase(mockCtrl)
	collectionMockObj := mgomocks.NewMockCollection(mockCtrl)
	queryMockObj := mgomocks.NewMockQuery(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(validUrl).Return(sessionMockObj, nil),
		sessionMockObj.EXPECT().DB(gomock.Any()).Return(dbMockObj),
		dbMockObj.EXPECT().C(gomock.Any()).Return(collectionMockObj),
		collectionMockObj.EXPECT().Find(query).Return(queryMockObj),
		queryMockObj.EXPECT().All(gomock.Any()).Return(mgo.ErrNotFound),
		sessionMockObj.EXPECT().Close(),
	)

	mgoDial = connectionMockObj
	executor := Executor{}
	_, err := executor.GetEvents(appId, imageName)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}

func TestCalledDeleteEvent_ExpectSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	query := bson.M{"_id": eventId}

	connectionMockObj := mgomocks.NewMockConnection(mockCtrl)
	sessionMockObj := mgomocks.NewMockSession(mockCtrl)
	dbMockObj := mgomocks.NewMockDatabase(mockCtrl)
	collectionMockObj := mgomocks.NewMockCollection(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(validUrl).Return(sessionMockObj, nil),
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
		collectionMockObj.EXPECT().Remove(query).Return(nil),
		sessionMockObj.EXPECT().Close(),
	)

	mgoDial = connectionMockObj
	executor := Executor{}
	err := executor.DeleteEvent(eventId)

	if err != nil {
		t.Error()
	}
}

func TestCalledDeleteEventWhenDBHasNotMatchedEvent_ExpectErrorReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	query := bson.M{"_id": eventId}

	connectionMockObj := mgomocks.NewMockConnection(mockCtrl)
	sessionMockObj := mgomocks.NewMockSession(mockCtrl)
	dbMockObj := mgomocks.NewMockDatabase(mockCtrl)
	collectionMockObj := mgomocks.NewMockCollection(mockCtrl)

	gomock.InOrder(
		connectionMockObj.EXPECT().Dial(validUrl).Return(sessionMockObj, nil),
		sessionMockObj.EXPECT().DB(dbName).Return(dbMockObj),
		dbMockObj.EXPECT().C(collectionName).Return(collectionMockObj),
		collectionMockObj.EXPECT().Remove(query).Return(mgo.ErrNotFound),
		sessionMockObj.EXPECT().Close(),
	)

	mgoDial = connectionMockObj
	executor := Executor{}
	err := executor.DeleteEvent(eventId)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "NotFound", err.Error())
	case errors.NotFound:
	}
}

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

package event

import (
	"commons/errors"
	. "db/mongo/wrapper"
	"gopkg.in/mgo.v2/bson"
)

// Interface of Event model's operations.
type Command interface {
	InsertEvent(eventId, appId, imageName string) (map[string]interface{}, error)
	GetEvents(appId, imageName string) ([]map[string]interface{}, error)
	DeleteEvent(eventId string) error
}

const (
	DB_NAME          = "DeploymentNodeDB"
	EVENT_COLLECTION = "EVENT"
	DB_URL           = "127.0.0.1:27017"
)

type Event struct {
	ID        string `bson:"_id,omitempty"`
	AppID     string
	ImageName string
}

type Executor struct {
}

var mgoDial Connection

func init() {
	mgoDial = MongoDial{}
}

// Try to connect with mongo db server.
// if succeed to connect with mongo db server, return error as nil,
// otherwise, return error.
func connect(url string) (Session, error) {
	// Create a MongoDB Session
	session, err := mgoDial.Dial(url)

	if err != nil {
		return nil, ConvertMongoError(err, "")
	}

	return session, err
}

// close of mongodb session.
func close(mgoSession Session) {
	mgoSession.Close()
}

// Getting collection by name.
// return mongodb Collection
func getCollection(mgoSession Session, dbname string, collectionName string) Collection {
	return mgoSession.DB(dbname).C(collectionName)
}

// Convert to map by object of struct Configuration.
// will return App information as map.
func (event Event) convertToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":        event.ID,
		"appId":     event.AppID,
		"imageName": event.ImageName,
	}
}

func (Executor) InsertEvent(eventId, appId, imageName string) (map[string]interface{}, error) {
	session, err := connect(DB_URL)
	if err != nil {
		return nil, err
	}
	defer close(session)

	event := Event{}
	err = getCollection(session, DB_NAME, EVENT_COLLECTION).Find(bson.M{"_id": eventId}).One(&event)
	if err == nil {
		return event.convertToMap(), errors.AlreadyReported{Msg: eventId}
	}

	event = Event{
		ID:        eventId,
		AppID:     appId,
		ImageName: imageName,
	}

	err = getCollection(session, DB_NAME, EVENT_COLLECTION).Insert(event)
	if err != nil {
		return nil, ConvertMongoError(err, "")
	}

	result := event.convertToMap()
	return result, err
}

func (Executor) GetEvents(appId, imageName string) ([]map[string]interface{}, error) {
	session, err := connect(DB_URL)
	if err != nil {
		return nil, err
	}
	defer close(session)

	events := []Event{}
	query := bson.M{"appid": bson.M{"$in": []string{"", appId}}, "imagename": bson.M{"$in": []string{"", imageName}}}
	err = getCollection(session, DB_NAME, EVENT_COLLECTION).Find(query).All(&events)
	if err != nil {
		return nil, ConvertMongoError(err, "Failed to find events")
	}

	result := make([]map[string]interface{}, len(events))
	for i, event := range events {
		result[i] = event.convertToMap()
	}
	return result, err
}

func (Executor) DeleteEvent(eventId string) error {
	session, err := connect(DB_URL)
	if err != nil {
		return err
	}
	defer close(session)

	err = getCollection(session, DB_NAME, EVENT_COLLECTION).Remove(bson.M{"_id": eventId})
	if err != nil {
		errMsg := "Failed to remove an event by " + eventId
		err = ConvertMongoError(err, errMsg)
		return err
	}

	return err
}

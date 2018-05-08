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
	. "db/bolt/wrapper"
	"encoding/json"
)

// Interface of Event model's operations.
type Command interface {
	InsertEvent(eventId, appId, imageName string) (map[string]interface{}, error)
	GetEvents(appId, imageName string) ([]map[string]interface{}, error)
	DeleteEvent(eventId string) error
}

const (
	BUCKET_NAME = "event"
)

type Event struct {
	ID        string `json:"id"`
	AppID     string `json:"appid"`
	ImageName string `json:"imagename"`
}

type Executor struct {
}

var db Database

func init() {
	db = NewBoltDB(BUCKET_NAME)
}

// Convert to map by object of struct Configuration.
// will return App information as map.
func (event Event) convertToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":        event.ID,
		"appid":     event.AppID,
		"imagename": event.ImageName,
	}
}

func (event Event) encode() ([]byte, error) {
	encoded, err := json.Marshal(event)
	if err != nil {
		return nil, errors.InvalidJSON{Msg: err.Error()}
	}
	return encoded, nil
}

func decode(data []byte) (*Event, error) {
	var event *Event
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, errors.InvalidJSON{Msg: err.Error()}
	}
	return event, nil
}

func (e Executor) InsertEvent(eventId, appId, imageName string) (map[string]interface{}, error) {
	value, err := db.Get([]byte(eventId))
	if err == nil {
		event, err := decode(value)
		if err == nil {
			return event.convertToMap(), errors.AlreadyReported{Msg: eventId}
		}
	}

	event := Event{
		ID:        eventId,
		AppID:     appId,
		ImageName: imageName,
	}

	encoded, err := event.encode()
	if err != nil {
		return nil, err
	}

	err = db.Put([]byte(eventId), encoded)
	if err != nil {
		return nil, err
	}

	return event.convertToMap(), nil
}

func (Executor) GetEvents(appId, imageName string) ([]map[string]interface{}, error) {
	events, err := db.List()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, value := range events {
		event, err := decode([]byte(value.(string)))
		if err != nil {
			continue
		}
		if ((event.AppID == "") || (event.AppID == appId)) && ((event.ImageName == "") || (event.ImageName == imageName)) {
			result = append(result, event.convertToMap())
		}
	}
	return result, nil
}

func (Executor) DeleteEvent(eventId string) error {
	return db.Delete([]byte(eventId))
}

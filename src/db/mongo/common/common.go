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
	. "db/mongo/wrapper"
)

var mgoDial Connection

func init() {
	mgoDial = MongoDial{}
}

// MongoDBClient provides persistence logic for "app" collection.
type (
	Builder interface {
		Connect(url string) error
		CreateDB() (*MongoDBManager, error)
	}

	MongoBuilder struct {
		session Session
	}

	MongoDBManager struct {
		mgoSession Session
	}
)

// Try to connect with mongo db server.
// if succeed to connect with mongo db server, return error as nil,
// otherwise, return error.
func (builder *MongoBuilder) Connect(url string) error {
	// Create a MongoDB Session
	session, err := mgoDial.Dial(url)

	if err != nil {
		return ConvertMongoError(err, "")
	}

	builder.session = session
	return nil
}

// Create mongodb manager.
// if session is valid, will return mongoDBManager
// otherwise, return error.
func (builder *MongoBuilder) CreateDB() (*MongoDBManager, error) {
	if builder.session == nil {
		return nil, errors.InvalidParam{"Invaild session of mongo."}
	}

	return &MongoDBManager{
		mgoSession: builder.session,
	}, nil
}

// Closer of mongodb manager.
func (client *MongoDBManager) Close() {
	client.mgoSession.Close()
}

// Getting collection by name.
// return mongodb Collection
func (client *MongoDBManager) GetCollection(dbname string, collectionName string) Collection {
	return client.mgoSession.DB(dbname).C(collectionName)
}

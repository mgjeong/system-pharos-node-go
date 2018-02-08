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

package configuration

import (
	"commons/errors"
	. "db/mongo/wrapper"
	"gopkg.in/mgo.v2/bson"
)

// Interface of Service model's operations.
type Command interface {
	// SetProperty updates configuration sets.
	SetProperty(property map[string]interface{}) error

	// GetProperty returns a single configuration property specified by name parameter.
	GetProperty(name string) (map[string]interface{}, error)

	// GetProperties returns a list of configurations stored in database.
	GetProperties() ([]map[string]interface{}, error)
}

const (
	DB_NAME                  = "DeploymentNodeDB"
	CONFIGURATION_COLLECTION = "CONFIGURATION"
	DB_URL                   = "127.0.0.1:27017"
)

type Property struct {
	Name   string `bson:"_id,omitempty"`
	Value  interface{}
	Policy []string
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
func (prop Property) convertToMap() map[string]interface{} {
	return map[string]interface{}{
		"name":   prop.Name,
		"value":  prop.Value,
		"policy": prop.Policy,
	}
}

// SetProperty inserts a map of configuration into the database.
// if succeed to add new configuration sets, returns an error as nil.
// otherwise, return error.
func (Executor) SetProperty(property map[string]interface{}) error {
	session, err := connect(DB_URL)
	if err != nil {
		return err
	}
	defer close(session)

	prop := Property{}
	query := bson.M{"_id": property["name"].(string)}
	err = getCollection(session, DB_NAME, CONFIGURATION_COLLECTION).Find(query).One(&prop)
	if err != nil {
		err = ConvertMongoError(err, "")
		switch err.(type) {
		default:
			return err
		case errors.NotFound:
			prop := Property{
				Name:   property["name"].(string),
				Value:  property["value"].(interface{}),
				Policy: property["policy"].([]string),
			}

			err = getCollection(session, DB_NAME, CONFIGURATION_COLLECTION).Insert(prop)
			if err != nil {
				return ConvertMongoError(err, "")
			}
			return err
		}
	}

	query = bson.M{"_id": property["name"].(string)}
	update := bson.M{"$set": bson.M{"value": property["value"].(interface{})}}
	err = getCollection(session, DB_NAME, CONFIGURATION_COLLECTION).Update(query, update)
	if err != nil {
		return ConvertMongoError(err, "Failed to update property value")
	}

	return err
}

// GetProperty returns a single configuration property specified by name parameter.
// if succeed to get, returns an error as nil.
// otherwise, return error.
func (Executor) GetProperty(name string) (map[string]interface{}, error) {
	session, err := connect(DB_URL)
	if err != nil {
		return nil, err
	}
	defer close(session)

	prop := Property{}
	query := bson.M{"_id": name}
	err = getCollection(session, DB_NAME, CONFIGURATION_COLLECTION).Find(query).One(&prop)
	if err != nil {
		err = ConvertMongoError(err, "")
		return nil, err
	}

	result := prop.convertToMap()
	return result, err
}

// GetProperties returns a list of configurations stored in database.
// if succeed to get, return list of all configurations as slice.
// otherwise, return error.
func (Executor) GetProperties() ([]map[string]interface{}, error) {
	session, err := connect(DB_URL)
	if err != nil {
		return nil, err
	}
	defer close(session)

	props := []Property{}
	err = getCollection(session, DB_NAME, CONFIGURATION_COLLECTION).Find(nil).All(&props)
	if err != nil {
		err = ConvertMongoError(err, "Failed to get all apps")
		return nil, err
	}

	result := make([]map[string]interface{}, len(props))
	for i, prop := range props {
		result[i] = prop.convertToMap()
	}

	return result, err
}

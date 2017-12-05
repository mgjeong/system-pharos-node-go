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
	"crypto/sha1"
	. "db/mongo/common"
	. "db/mongo/wrapper"
	. "db/modelinterface"
	"encoding/hex"
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"strings"
)

const (
	DB_NAME        = "DeploymentAgentDB"
	APP_COLLECTION = "APP"
	SERVICES_FIELD = "services"
	IMAGE_FIELD    = "image"
)

type App struct {
	ID          string `bson:"_id,omitempty"`
	Description string
	State       string
}

type DBManager struct {
	Service
}

var mgoBuilder Builder

func init() {
	mgoBuilder = &MongoBuilder{}
}

// Try to connect with dbms.
// if succeed to connect with db server, return db instance,
// otherwise, return nil and error.
func getDBmanager() (*MongoDBManager, error) {
	// TODO: Should be updated to support different types of databases.
	url := "localhost:27017"

	err := mgoBuilder.Connect(url)
	if err != nil {
		return nil, err
	}

	dbManager, err := mgoBuilder.CreateDB()
	if err != nil {
		return nil, err
	}

	return dbManager, err
}

// Convert to map by object of struct App.
// will return App information as map.
func (app App) convertToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          app.ID,
		"description": app.Description,
		"state":       app.State,
	}
}

// Add app description to app collection in mongo server.
// if succeed to add, return app information as map.
// otherwise, return error.
func (DBManager) InsertComposeFile(description string) (map[string]interface{}, error) {
	id, err := generateID(description)
	if err != nil {
		return nil, err
	}

	db, err := getDBmanager()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	app := App{
		ID:          id,
		Description: description,
		State:       "DEPLOY",
	}

	err = db.GetCollection(DB_NAME, APP_COLLECTION).Insert(app)
	if err != nil {
		return nil, ConvertMongoError(err, "")
	}

	result := app.convertToMap()
	return result, err
}

// Getting all of app informations.
// if succeed to get, return list of all app information as slice.
// otherwise, return error.
func (DBManager) GetAppList() ([]map[string]interface{}, error) {
	db, err := getDBmanager()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	apps := []App{}
	err = db.GetCollection(DB_NAME, APP_COLLECTION).Find(nil).All(&apps)
	if err != nil {
		err = ConvertMongoError(err, "Failed to get all apps")
		return nil, err
	}

	result := make([]map[string]interface{}, len(apps))
	for i, app := range apps {
		result[i] = app.convertToMap()
	}

	return result, err
}

// Getting app information by app_id.
// if succeed to get, return app information as map.
// otherwise, return error.
func (DBManager) GetApp(app_id string) (map[string]interface{}, error) {
	db, err := getDBmanager()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return nil, err
	}

	app := App{}
	err = db.GetCollection(DB_NAME, APP_COLLECTION).Find(bson.M{"_id": app_id}).One(&app)
	if err != nil {
		errMsg := "Failed to find a app by " + app_id
		err = ConvertMongoError(err, errMsg)
		return nil, err
	}

	result := app.convertToMap()
	return result, err
}

// Updating app information by app_id.
// if succeed to update, return error as nil.
// otherwise, return error.
func (DBManager) UpdateAppInfo(app_id string, description string) error {
	db, err := getDBmanager()
	if err != nil {
		return err
	}
	defer db.Close()

	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return err
	}

	update := bson.M{"$set": bson.M{"description": description}}
	err = db.GetCollection(DB_NAME, APP_COLLECTION).Update(bson.M{"_id": app_id}, update)
	if err != nil {
		errMsg := "Failed to update a app by " + app_id
		err = ConvertMongoError(err, errMsg)
		return err
	}

	return err
}

// Deleting app collection by app_id.
// if succeed to delete, return error as nil.
// otherwise, return error.
func (DBManager) DeleteApp(app_id string) error {
	db, err := getDBmanager()
	if err != nil {
		return err
	}
	defer db.Close()

	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return err
	}

	err = db.GetCollection(DB_NAME, APP_COLLECTION).Remove(bson.M{"_id": app_id})
	if err != nil {
		errMsg := "Failed to remove a app by " + app_id
		err = ConvertMongoError(err, errMsg)
		return err
	}

	return err
}

// Getting app state by app_id.
// if succeed to get state, return state (e.g.DEPLOY, UP, STOP...).
// otherwise, return error.
func (DBManager) GetAppState(app_id string) (string, error) {
	db, err := getDBmanager()
	if err != nil {
		return "", err
	}
	defer db.Close()

	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return "", err
	}

	app := App{}
	err = db.GetCollection(DB_NAME, APP_COLLECTION).Find(bson.M{"_id": app_id}).One(&app)
	if err != nil {
		errMsg := "Failed to get app's state by " + app_id
		err = ConvertMongoError(err, errMsg)
		return "", err
	}

	return app.State, err
}

// Updating app state by app_id.
// if succeed to update state, return error as nil.
// otherwise, return error.
func (DBManager) UpdateAppState(app_id string, state string) error {
	db, err := getDBmanager()
	if err != nil {
		return err
	}
	defer db.Close()

	if len(state) == 0 {
		err := errors.InvalidParam{"Invalid param error : state is empty."}
		return err
	}
	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return err
	}

	update := bson.M{"$set": bson.M{"state": state}}
	err = db.GetCollection(DB_NAME, APP_COLLECTION).Update(bson.M{"_id": app_id}, update)
	if err != nil {
		errMsg := "Failed to update app's state by " + app_id
		err = ConvertMongoError(err, errMsg)
		return err
	}

	return err
}

// Generating app_id using hash of description
// if succeed to generate, return UUID (32bytes).
// otherwise, return error.
func generateID(description string) (string, error) {
	extractedValue, err := extractHashValue([]byte(description))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(makeHash(extractedValue)), nil
}

// Sorting string for generation of hash code.
// return sorted string.
func sortString(unsorted string) string {
	s := strings.Split(unsorted, "")
	sort.Strings(s)
	sorted := strings.Join(s, "")
	return sorted
}

// Making hash value by app description.
// if succeed to make, return hash value
// otherwise, return error.
func extractHashValue(source []byte) (string, error) {
	var targetValue string
	description := make(map[string]interface{})

	err := json.Unmarshal(source, &description)
	if err != nil {

		return "", convertJsonError(err)
	}

	if len(description[SERVICES_FIELD].(map[string]interface{})) == 0 || description[SERVICES_FIELD] == nil {
		return "", errors.InvalidYaml{"Invalid YAML error : description has not service information."}
	}

	for service_name, service_info := range description[SERVICES_FIELD].(map[string]interface{}) {
		targetValue += string(service_name)

		if service_info.(map[string]interface{})[IMAGE_FIELD] == nil {
			return "", errors.InvalidYaml{"Invalid YAML error : description has not image information."}
		}

		targetValue += service_info.(map[string]interface{})[IMAGE_FIELD].(string)
	}
	return sortString(targetValue), nil
}

// Making hash code by hash value.
// return hash code as slice of byte
func makeHash(source string) []byte {
	h := sha1.New()
	h.Write([]byte(source))
	return h.Sum(nil)
}

// Converting to commons/errors by Json error
func convertJsonError(jsonError error) (err error) {
	switch jsonError.(type) {
	case *json.SyntaxError,
		*json.InvalidUTF8Error,
		*json.InvalidUnmarshalError,
		*json.UnmarshalFieldError,
		*json.UnmarshalTypeError:
		return errors.InvalidYaml{}
	default:
		return errors.Unknown{}
	}
}

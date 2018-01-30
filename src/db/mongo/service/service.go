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
	. "db/mongo/wrapper"
	"encoding/hex"
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"strings"
)

// Interface of Service model's operations.
type Command interface {
	// InsertComposeFile insert docker-compose file for new service.
	InsertComposeFile(description string) (map[string]interface{}, error)

	// GetAppList returns all of app's IDs.
	GetAppList() ([]map[string]interface{}, error)

	// GetApp returns docker-compose data of target app.
	GetApp(app_id string) (map[string]interface{}, error)

	// UpdateAppInfo updates docker-compose data of target app.
	UpdateAppInfo(app_id string, description string) error

	// DeleteApp delete docker-compose data of target app.
	DeleteApp(app_id string) error

	// GetAppState returns app's state
	GetAppState(app_id string) (string, error)

	// UpdateAppState updates app's State.
	UpdateAppState(app_id string, state string) error

	// UpdateAppEvent updates the last received event from docker registry.
	UpdateAppEvent(app_id string, repo string, tag string, event string) error
}

const (
	DB_NAME        = "DeploymentNodeDB"
	APP_COLLECTION = "APP"
	SERVICES_FIELD = "services"
	IMAGE_FIELD    = "image"
	DB_URL         = "127.0.0.1:27017"
	EVENT_NONE     = "none"
)

type App struct {
	ID          string `bson:"_id,omitempty"`
	Description string
	State       string
	Images      []map[string]interface{}
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

// Convert to map by object of struct App.
// will return App information as map.
func (app App) convertToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          app.ID,
		"description": app.Description,
		"state":       app.State,
		"images":      app.Images,
	}
}

// Add app description to app collection in mongo server.
// if succeed to add, return app information as map.
// otherwise, return error.
func (Executor) InsertComposeFile(description string) (map[string]interface{}, error) {
	id, err := generateID(description)
	if err != nil {
		return nil, err
	}

	session, err := connect(DB_URL)
	if err != nil {
		return nil, err
	}
	defer close(session)

	images, err := getImageNames([]byte(description))
	if err != nil {
		return nil, err
	}

	app := App{
		ID:          id,
		Description: description,
		State:       "DEPLOY",
		Images:      images,
	}

	err = getCollection(session, DB_NAME, APP_COLLECTION).Insert(app)
	if err != nil {
		return nil, ConvertMongoError(err, "")
	}

	result := app.convertToMap()
	return result, err
}

// Getting all of app informations.
// if succeed to get, return list of all app information as slice.
// otherwise, return error.
func (Executor) GetAppList() ([]map[string]interface{}, error) {
	session, err := connect(DB_URL)
	if err != nil {
		return nil, err
	}
	defer close(session)

	apps := []App{}
	err = getCollection(session, DB_NAME, APP_COLLECTION).Find(nil).All(&apps)
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
func (Executor) GetApp(app_id string) (map[string]interface{}, error) {
	session, err := connect(DB_URL)
	if err != nil {
		return nil, err
	}
	defer close(session)

	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return nil, err
	}

	app := App{}
	err = getCollection(session, DB_NAME, APP_COLLECTION).Find(bson.M{"_id": app_id}).One(&app)
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
func (Executor) UpdateAppInfo(app_id string, description string) error {
	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return err
	}

	session, err := connect(DB_URL)
	if err != nil {
		return err
	}
	defer close(session)

	update := bson.M{"$set": bson.M{"description": description}}
	err = getCollection(session, DB_NAME, APP_COLLECTION).Update(bson.M{"_id": app_id}, update)
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
func (Executor) DeleteApp(app_id string) error {
	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return err
	}

	session, err := connect(DB_URL)
	if err != nil {
		return err
	}
	defer close(session)

	err = getCollection(session, DB_NAME, APP_COLLECTION).Remove(bson.M{"_id": app_id})
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
func (Executor) GetAppState(app_id string) (string, error) {
	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return "", err
	}

	session, err := connect(DB_URL)
	if err != nil {
		return "", err
	}
	defer close(session)

	app := App{}
	err = getCollection(session, DB_NAME, APP_COLLECTION).Find(bson.M{"_id": app_id}).One(&app)
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
func (Executor) UpdateAppState(app_id string, state string) error {
	if len(state) == 0 {
		err := errors.InvalidParam{"Invalid param error : state is empty."}
		return err
	}
	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return err
	}

	session, err := connect(DB_URL)
	if err != nil {
		return err
	}
	defer close(session)

	update := bson.M{"$set": bson.M{"state": state}}
	err = getCollection(session, DB_NAME, APP_COLLECTION).Update(bson.M{"_id": app_id}, update)
	if err != nil {
		errMsg := "Failed to update app's state by " + app_id
		err = ConvertMongoError(err, errMsg)
		return err
	}

	return err
}

func (Executor) UpdateAppEvent(app_id string, repo string, tag string, event string) error {
	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return err
	}

	session, err := connect(DB_URL)
	if err != nil {
		return err
	}
	defer close(session)

	app := App{}
	err = getCollection(session, DB_NAME, APP_COLLECTION).Find(bson.M{"_id": app_id}).One(&app)
	if err != nil {
		errMsg := "Failed to get app information by " + app_id
		err = ConvertMongoError(err, errMsg)
		return err
	}

	// Find image specified by repo parameter.
	for index, image := range app.Images {
		if strings.Compare(image["name"].(string), repo) == 0 {
			// If event type is none, delete 'changes' field.
			if event == EVENT_NONE {
				delete(app.Images[index], "changes")
			} else {
				newEvent := make(map[string]interface{})
				newEvent["tag"] = tag
				newEvent["status"] = event
				app.Images[index]["changes"] = newEvent
			}
		}

		// Save the changes to database.
		update := bson.M{"$set": bson.M{"images": app.Images}}
		err = getCollection(session, DB_NAME, APP_COLLECTION).Update(bson.M{"_id": app_id}, update)
		if err != nil {
			errMsg := "Failed to update app information" + app_id
			return ConvertMongoError(err, errMsg)
		}
		return nil
	}

	return errors.NotFound{Msg: "There is no matching image"}
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

		// Parse full image name to exclude tag when generating application id.
		fullImageName := service_info.(map[string]interface{})[IMAGE_FIELD].(string)
		words := strings.Split(fullImageName, "/")
		imageNameWithoutRepo := strings.Join(words[:len(words)-1], "/")
		repo := strings.Split(words[len(words)-1], ":")

		imageNameWithoutTag := imageNameWithoutRepo
		if len(words) > 1 {
			imageNameWithoutTag += "/"
		}
		imageNameWithoutTag += repo[0]
		targetValue += imageNameWithoutTag
	}
	return sortString(targetValue), nil
}

func getImageNames(source []byte) ([]map[string]interface{}, error) {
	description := make(map[string]interface{})

	err := json.Unmarshal(source, &description)
	if err != nil {
		return nil, convertJsonError(err)
	}

	if len(description[SERVICES_FIELD].(map[string]interface{})) == 0 || description[SERVICES_FIELD] == nil {
		return nil, errors.InvalidYaml{"Invalid YAML error : description has not service information."}
	}

	images := make([]map[string]interface{}, 0)
	for _, service_info := range description[SERVICES_FIELD].(map[string]interface{}) {
		if service_info.(map[string]interface{})[IMAGE_FIELD] == nil {
			return nil, errors.InvalidYaml{"Invalid YAML error : description has not image information."}
		}

		fullImageName := service_info.(map[string]interface{})[IMAGE_FIELD].(string)
		words := strings.Split(fullImageName, "/")
		imageNameWithoutRepo := strings.Join(words[:len(words)-1], "/")
		repo := strings.Split(words[len(words)-1], ":")

		imageNameWithoutTag := imageNameWithoutRepo
		if len(words) > 1 {
			imageNameWithoutTag += "/"
		}
		imageNameWithoutTag += repo[0]

		image := make(map[string]interface{})
		image["name"] = imageNameWithoutTag
		images = append(images, image)
	}
	return images, nil
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

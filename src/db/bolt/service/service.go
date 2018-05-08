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

package service

import (
	"commons/errors"
	"commons/logger"
	"crypto/sha1"
	. "db/bolt/wrapper"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

// Interface of Service model's operations.
type Command interface {
	// InsertComposeFile insert docker-compose file for new service.
	InsertComposeFile(description string, state string) (map[string]interface{}, error)

	// GetAppList returns all of app's IDs.
	GetAppList() ([]map[string]interface{}, error)

	// GetApp returns docker-compose data of target app.
	GetApp(app_id string) (map[string]interface{}, error)

	// UpdateAppInfo updates docker-compose data of target app.
	UpdateAppInfo(app_id string, description string) error

	// DeleteApp delete docker-compose data of target app.
	DeleteApp(app_id string) error

	// UpdateAppState updates app's State.
	UpdateAppState(app_id string, state string) error

	// UpdateAppEvent updates the last received event from docker registry.
	UpdateAppEvent(app_id string, repo string, tag string, event string) error
}

const (
	BUCKET_NAME    = "service"
	SERVICES_FIELD = "services"
	IMAGE_FIELD    = "image"
	EVENT_NONE     = "none"
)

type App struct {
	ID          string                   `json:"id"`
	Description string                   `json:"description"`
	State       string                   `json:"state"`
	Images      []map[string]interface{} `json:"images"`
}

type Executor struct {
}

var db Database

func init() {
	db = NewBoltDB(BUCKET_NAME)
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

func (app App) encode() ([]byte, error) {
	encoded, err := json.Marshal(app)
	if err != nil {
		return nil, errors.InvalidJSON{Msg: err.Error()}
	}
	return encoded, nil
}

func decode(data []byte) (*App, error) {
	var app *App
	err := json.Unmarshal(data, &app)
	if err != nil {
		return nil, errors.InvalidJSON{Msg: err.Error()}
	}
	return app, nil
}

// Add app description to app collection in mongo server.
// if succeed to add, return app information as map.
// otherwise, return error.
func (Executor) InsertComposeFile(description string, state string) (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	id, err := generateID(description)
	if err != nil {
		return nil, err
	}

	value, err := db.Get([]byte(id))
	if err == nil {
		app, err := decode(value)
		if err == nil {
			return app.convertToMap(), errors.AlreadyReported{Msg: id}
		}
	}

	images, err := getImageNames([]byte(description))
	if err != nil {
		return nil, err
	}

	installedApp := App{
		ID:          id,
		Description: description,
		State:       state,
		Images:      images,
	}

	encoded, err := installedApp.encode()
	if err != nil {
		return nil, err
	}

	err = db.Put([]byte(id), encoded)
	if err != nil {
		return nil, err
	}

	return installedApp.convertToMap(), nil
}

// Getting all of app informations.
// if succeed to get, return list of all app information as slice.
// otherwise, return error.
func (Executor) GetAppList() ([]map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	apps, err := db.List()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, value := range apps {
		app, err := decode([]byte(value.(string)))
		if err != nil {
			continue
		}
		result = append(result, app.convertToMap())
	}
	return result, nil
}

// Getting app information by app_id.
// if succeed to get, return app information as map.
// otherwise, return error.
func (Executor) GetApp(app_id string) (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return nil, err
	}

	value, err := db.Get([]byte(app_id))
	if err != nil {
		return nil, err
	}

	app, err := decode(value)
	if err != nil {
		return nil, err
	}
	return app.convertToMap(), nil
}

// Updating app information by app_id.
// if succeed to update, return error as nil.
// otherwise, return error.
func (Executor) UpdateAppInfo(app_id string, description string) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return err
	}

	id, err := generateID(description)
	if err != nil {
		return err
	}

	if strings.Compare(app_id, id) != 0 {
		err := errors.InvalidYaml{`the description is information that can not be reflected in the app that matches the appId .`}
		return err
	}

	value, err := db.Get([]byte(app_id))
	if err != nil {
		return err
	}

	app, err := decode(value)
	if err != nil {
		return err
	}

	app.Description = description
	encoded, err := app.encode()
	if err != nil {
		return err
	}

	return db.Put([]byte(app_id), encoded)
}

// Deleting app collection by app_id.
// if succeed to delete, return error as nil.
// otherwise, return error.
func (Executor) DeleteApp(app_id string) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return err
	}

	return db.Delete([]byte(app_id))
}

// Updating app state by app_id.
// if succeed to update state, return error as nil.
// otherwise, return error.
func (Executor) UpdateAppState(app_id string, state string) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	if len(state) == 0 {
		err := errors.InvalidParam{"Invalid param error : state is empty."}
		return err
	}
	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return err
	}

	value, err := db.Get([]byte(app_id))
	if err != nil {
		return err
	}

	app, err := decode(value)
	if err != nil {
		return err
	}

	app.State = state
	encoded, err := app.encode()
	if err != nil {
		return err
	}

	return db.Put([]byte(app_id), encoded)
}

func (Executor) UpdateAppEvent(app_id string, repo string, tag string, event string) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	if len(app_id) == 0 {
		err := errors.InvalidParam{"Invalid param error : app_id is empty."}
		return err
	}

	value, err := db.Get([]byte(app_id))
	if err != nil {
		return err
	}

	app, err := decode(value)
	if err != nil {
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
		encoded, err := app.encode()
		if err != nil {
			return err
		}

		return db.Put([]byte(app_id), encoded)
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

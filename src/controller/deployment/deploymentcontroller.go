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

// Package controller provides controllability of
// persistence database and docker(docker-compose).
package deployment

import (
	"commons/errors"
	"commons/logger"
	"controller/deployment/dockercontroller"
	"db/mongo/service"
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/ghodss/yaml"
)

const (
	COMPOSE_FILE = "docker-compose.yaml"
	ID           = "id"
	DESCRIPTION  = "description"
	SERVICES     = "services"
	IMAGE        = "image"
	NAME         = "name"
	STATE        = "state"
)

type Command interface {
	DeployApp(body string) (map[string]interface{}, error)
	Apps() (map[string]interface{}, error)
	App(appId string) (map[string]interface{}, error)
	UpdateAppInfo(appId string, body string) error
	DeleteApp(appId string) error
	StartApp(appId string) error
	StopApp(appId string) error
	UpdateApp(appId string) error
}

type depExecutorImpl struct{}

var Executor depExecutorImpl
var dockerExecutor dockercontroller.Command

var fileMode = os.FileMode(0755)
var dbExecutor service.Command

func init() {
	dockerExecutor = dockercontroller.Executor
	dbExecutor = service.Executor{}
}

// Deploy app to target by yaml description.
// yaml description will be inserted to db server
// and docker images in the service list of yaml description will be downloaded
// and create, start containers on the target.
// if succeed to deploy, return app_id
// otherwise, return error.
func (depExecutorImpl) DeployApp(body string) (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	err := ioutil.WriteFile(COMPOSE_FILE, []byte(body), fileMode)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, errors.IOError{Msg: "file io fail"}
	}
	defer os.RemoveAll(COMPOSE_FILE)

	convertedData, err := yaml.YAMLToJSON([]byte(body))
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, errors.InvalidYaml{Msg: "invalid yaml syntax"}
	}

	data, err := dbExecutor.InsertComposeFile(string(convertedData))
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, errors.Unknown{Msg: "db operation fail"}
	}

	err = dockerExecutor.Up(data[ID].(string), COMPOSE_FILE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		e := dockerExecutor.DownWithRemoveImages(data[ID].(string), COMPOSE_FILE)
		if e != nil {
			logger.Logging(logger.ERROR, e.Error())
		}
		dbExecutor.DeleteApp(data[ID].(string))
		return nil, err
	}

	res := make(map[string]interface{})
	res[ID] = data[ID].(string)

	return res, nil
}

// Getting all of app informations in the target.
// if succeed to get, return all of app informations as map
// otherwise, return error.
func (depExecutorImpl) Apps() (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	apps, err := dbExecutor.GetAppList()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, errors.Unknown{Msg: "db operation fail"}
	}

	yamlList := make([]map[string]interface{}, 0)
	for _, app := range apps {
		m := make(map[string]interface{})
		m[ID] = app[ID].(string)
		m[STATE] = app[STATE].(string)
		yamlList = append(yamlList, m)
	}

	res := make(map[string]interface{})
	res["apps"] = yamlList

	return res, nil
}

// Getting app information in the target by input appId.
// if succeed to get, return app information
// otherwise, return error.
func (depExecutorImpl) App(appId string) (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN", appId)
	defer logger.Logging(logger.DEBUG, "OUT")

	app, err := dbExecutor.GetApp(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, convertDBError(err, appId)
	}

	yaml, err := yaml.JSONToYAML([]byte(app[DESCRIPTION].(string)))
	if err != nil {
		return nil, errors.InvalidYaml{Msg: "invalid yaml syntax"}
	}

	err = ioutil.WriteFile(COMPOSE_FILE, yaml, fileMode)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, errors.IOError{Msg: "file io fail"}
	}
	defer os.RemoveAll(COMPOSE_FILE)

	description := make(map[string]interface{})
	err = json.Unmarshal([]byte(app[DESCRIPTION].(string)), &description)

	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, errors.IOError{Msg: "json unmarshal fail"}
	}

	if description[SERVICES] == nil || len(description[SERVICES].(map[string]interface{})) == 0 {
		return nil, errors.Unknown{Msg: "can't find application info"}
	}

	services := make([]map[string]interface{}, 0)
	for _, serviceName := range reflect.ValueOf(description[SERVICES].(map[string]interface{})).MapKeys() {
		service := make(map[string]interface{}, 0)

		state, err := getServiceState(appId, serviceName.String())
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return nil, errors.Unknown{Msg: "get state fail"}
		}

		service[NAME] = serviceName.String()
		service[STATE] = state
		services = append(services, service)
	}

	m := make(map[string]interface{})
	m[STATE] = app[STATE].(string)
	m[DESCRIPTION] = string(yaml)
	m[SERVICES] = services

	return m, nil
}

// Updating app information in the target by input appId and updated description.
// exclud restart of containers and pull the new images.
// only update yaml description on the db server.
// if succeed to update, return error as nil
// otherwise, return error.
func (depExecutorImpl) UpdateAppInfo(appId string, body string) error {
	logger.Logging(logger.DEBUG, "IN", appId)
	defer logger.Logging(logger.DEBUG, "OUT")

	convertedData, err := yaml.YAMLToJSON([]byte(body))
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return errors.InvalidYaml{Msg: "invalid yaml syntax"}
	}

	err = dbExecutor.UpdateAppInfo(appId, string(convertedData))
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	return nil
}

// Start app in the target by input appId.
// if starting is failed, Agent will make sure that only previous state.
// can not guarantee about valid operation of containers.
// if succeed to start, return error as nil
// otherwise, return error.
func (depExecutorImpl) StartApp(appId string) error {
	logger.Logging(logger.DEBUG, "IN", appId)
	defer logger.Logging(logger.DEBUG, "OUT")

	state, err := dbExecutor.GetAppState(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	if state == "START" {
		return errors.AlreadyReported{Msg: state}
	}

	err = setYamlFile(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	err = dockerExecutor.Start(appId, COMPOSE_FILE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		e := restoreState(appId, state)
		if e != nil {
			logger.Logging(logger.ERROR, err.Error())
		}
		return err
	}

	err = dbExecutor.UpdateAppState(appId, "START")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	return nil
}

// Stop app in the target by input appId.
// if succeed to stop, return app information
// otherwise, return error.
func (depExecutorImpl) StopApp(appId string) error {
	logger.Logging(logger.DEBUG, "IN", appId)
	defer logger.Logging(logger.DEBUG, "OUT")

	state, err := dbExecutor.GetAppState(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	if state == "STOP" {
		return errors.AlreadyReported{Msg: state}
	}

	err = setYamlFile(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	err = dockerExecutor.Stop(appId, COMPOSE_FILE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		e := restoreState(appId, state)
		if e != nil {
			logger.Logging(logger.ERROR, err.Error())
		}
		return err
	}

	err = dbExecutor.UpdateAppState(appId, "STOP")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	return nil
}

// Update images and restart containers in the target
// by input appId and stored yaml in db server.
// if you want to update images,
// yaml should be updated as controller.UpdateAppInfo()
// See also controller.UpdateAppInfo().
// and if failed to update images,
// Agent can make sure that previous imaes by digest.
// if succeed to update, return error as nil
// otherwise, return error.
func (depExecutorImpl) UpdateApp(appId string) error {
	logger.Logging(logger.DEBUG, "IN", appId)
	defer logger.Logging(logger.DEBUG, "OUT")

	err := setYamlFile(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	app, e := dbExecutor.GetApp(appId)
	if e != nil {
		logger.Logging(logger.DEBUG, e.Error())
		return convertDBError(e, appId)
	}

	err = dockerExecutor.Pull(appId, COMPOSE_FILE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		e := restoreRepoDigests(appId, app[DESCRIPTION].(string), app[STATE].(string))
		if e != nil {
			logger.Logging(logger.ERROR, e.Error())
		}
		return err
	}

	err = dockerExecutor.Up(appId, COMPOSE_FILE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		e := restoreRepoDigests(appId, app[DESCRIPTION].(string), app[STATE].(string))
		if e != nil {
			logger.Logging(logger.ERROR, e.Error())
		}
		return err
	}

	return nil
}

// Delete images and remove containers in the target by input appId.
// and delete yaml description on the target.
// containers should be stopped as controller.StopApp().
// See also controller.StopApp().
// if succeed to delete, return error as nil
// otherwise, return error.
func (depExecutorImpl) DeleteApp(appId string) error {
	logger.Logging(logger.DEBUG, "IN", appId)
	defer logger.Logging(logger.DEBUG, "OUT")

	err := setYamlFile(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	err = dockerExecutor.DownWithRemoveImages(appId, COMPOSE_FILE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		state, e := dbExecutor.GetAppState(appId)
		if e != nil {
			logger.Logging(logger.ERROR, e.Error())
			return err
		}
		e = restoreState(appId, state)
		if e != nil {
			logger.Logging(logger.ERROR, e.Error())
		}
		return err
	}

	err = dbExecutor.DeleteApp(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	return nil
}

// Restore app images by previous disgests.
// See also controller.UpdateApp()
// if succeed to restore, return error as nil
// otherwise, return error.
func restoreRepoDigests(appId, desc, state string) error {
	imageNames, err := getImageNames([]byte(desc))
	if err != nil {
		logger.Logging(logger.DEBUG, err.Error())
		return err
	}
	repoDigests := make([]string, 0)

	for _, imageName := range imageNames {
		digest, err := dockerExecutor.GetImageDigestByName(imageName)
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return err
		}
		repoDigests = append(repoDigests, digest)
	}

	description := make(map[string]interface{})

	err = json.Unmarshal([]byte(desc), &description)
	if err != nil {
		return errors.IOError{Msg: "json unmarshal fail"}
	}

	if len(description[SERVICES].(map[string]interface{})) == 0 || description[SERVICES] == nil {
		return errors.Unknown{Msg: "can't find application info"}
	}

	idx := 0
	for _, service_info := range description[SERVICES].(map[string]interface{}) {
		service_info.(map[string]interface{})[IMAGE] = repoDigests[idx]
		idx++
	}

	restoredDesc, err := json.Marshal(description)
	if err != nil {
		logger.Logging(logger.DEBUG, "json marshal fail")
		return err
	}
	yaml, err := yaml.JSONToYAML(restoredDesc)

	err = ioutil.WriteFile(COMPOSE_FILE, yaml, fileMode)
	if err != nil {
		return errors.IOError{Msg: "file io fail"}
	}

	err = dockerExecutor.Up(appId, COMPOSE_FILE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	err = restoreState(appId, state)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}
	return nil
}

// Restore app state by previous state.
// See also controller.StartApp(), controller.StopApp()
// if succeed to restore, return error as nil
// otherwise, return error.
func restoreState(appId, state string) error {
	var err error

	if len(state) == 0 {
		return errors.InvalidParam{Msg: "empty state"}
	}

	switch state {
	case "STOP":
		err = dockerExecutor.Stop(appId, COMPOSE_FILE)
	case "START":
		err = dockerExecutor.Up(appId, COMPOSE_FILE)
	case "UP":
		err = dockerExecutor.Up(appId, COMPOSE_FILE)
	case "DEPLOY":
		err = dockerExecutor.Up(appId, COMPOSE_FILE)
	}

	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
	}
	return err
}

// Set YAML file about an app on a path.
// The path is defined as contant
// if setting YAML is succeeded, return error as nil
// otherwise, return error.
func setYamlFile(appId string) error {

	app, err := dbExecutor.GetApp(appId)
	if err != nil {
		return convertDBError(err, appId)
	}

	yaml, err := yaml.JSONToYAML([]byte(app[DESCRIPTION].(string)))
	if err != nil {
		return errors.InvalidYaml{Msg: "invalid yaml syntax"}
	}

	err = ioutil.WriteFile(COMPOSE_FILE, yaml, fileMode)
	if err != nil {
		return errors.IOError{Msg: "file io fail"}
	}

	return nil
}

// Get image names from an JSON file.
// If getting image names is succeeded, return image names
// otherwise, return error.
func getImageNames(source []byte) ([]string, error) {
	imageNames := make([]string, 0)
	description := make(map[string]interface{})

	err := json.Unmarshal(source, &description)
	if err != nil {
		return nil, errors.IOError{Msg: "json unmarshal fail"}
	}

	if len(description[SERVICES].(map[string]interface{})) == 0 || description[SERVICES] == nil {
		return nil, errors.Unknown{Msg: "can't find application info"}
	}

	for _, service_info := range description[SERVICES].(map[string]interface{}) {
		if service_info.(map[string]interface{})[IMAGE] == nil {
			return nil, errors.Unknown{Msg: "can't find service info"}
		}
		imageNames = append(imageNames, service_info.(map[string]interface{})[IMAGE].(string))
	}

	return imageNames, nil
}

// Get service state by service name.
// First of all, get container name using docker-compose ps <service name>
// And then, get service state from using docker inspect <container name>
// if getting service state is succeed, return service state
// otherwise, return error.
func getServiceState(appId, serviceName string) (map[string]interface{}, error) {
	infos, err := dockerExecutor.Ps(appId, COMPOSE_FILE, serviceName)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}

	containerName := infos[0]["Name"]

	serviceInfo, err := dockerExecutor.GetContainerStateByName(containerName)
	if err != nil {
		return nil, err
	}
	return serviceInfo, nil
}

func convertDBError(err error, appId string) error {
	switch err.(type) {
	case errors.NotFound:
		return errors.InvalidAppId{Msg: "failed to find app id : " + appId}
	default:
		return errors.Unknown{Msg: "db operation fail"}
	}
}

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
	"commons/util"
	"controller/dockercontroller"
	"controller/monitoring/apps"
	"db/bolt/service"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	ID             = "id"
	CID            = "cid"
	DESCRIPTION    = "description"
	SERVICES       = "services"
	IMAGE          = "image"
	IMAGES         = "images"
	NAME           = "name"
	PORTS          = "ports"
	STATE          = "state"
	STATUS         = "status"
	EXIT_CODE      = "exitcode"
	EVENTS         = "events"
	TARGETINFO     = "target"
	REQUESTINFO    = "request"
	HOST           = "host"
	REPOSITORY     = "repository"
	TAG            = "tag"
	ACTION         = "action"
	PUSH           = "push"
	UPDATE         = "update"
	DELETE         = "delete"
	RUNNING_STATE  = "running"
	EXITED_STATE   = "exited"
	UPDATING_STATE = "updating"
	NONE           = "none"
	CHANGES        = "changes"
	EVENTID        = "eventId"
)

type Command interface {
	DeployApp(body string, query map[string]interface{}) (map[string]interface{}, error)
	Apps() (map[string]interface{}, error)
	App(appId string) (map[string]interface{}, error)
	UpdateAppInfo(appId string, body string) error
	DeleteApp(appId string) error
	StartApp(appId string) error
	StopApp(appId string) error
	HandleEvents(appId string, body string) error
	UpdateApp(appId string, query map[string]interface{}) error
}

type depExecutorImpl struct{}

var Executor depExecutorImpl
var dockerExecutor dockercontroller.Command
var appsMonitor apps.Command

var fileMode = os.FileMode(0755)
var dbExecutor service.Command

func init() {
	dockerExecutor = dockercontroller.Executor
	dbExecutor = service.Executor{}
	appsMonitor = apps.Executor{}

	restoreAllAppsState()
}

// Deploy app to target by yaml description.
// yaml description will be inserted to db server
// and docker images in the service list of yaml description will be downloaded
// and create, start containers on the target.
// if succeed to deploy, return app_id
// otherwise, return error.
func (executor depExecutorImpl) DeployApp(body string, query map[string]interface{}) (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	var description interface{}
	err := yaml.Unmarshal([]byte(body), &description)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, errors.Unknown{Msg: "db operation fail"}
	}

	description = convert(description)

	jsonData, err := json.Marshal(description)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, errors.InvalidYaml{Msg: "invalid yaml syntax"}
	}

	data, err := dbExecutor.InsertComposeFile(string(jsonData), RUNNING_STATE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		switch err.(type) {
		default:
			return nil, err
		case errors.AlreadyReported:
			deployedApp, err := executor.App(data[ID].(string))
			if err != nil {
				logger.Logging(logger.ERROR, err.Error())
				return nil, err
			}
			deployedApp[ID] = data[ID].(string)
			return deployedApp, err
		}
	}

	composeFile := genYamlFileName(data[ID].(string), "deploy")
	err = ioutil.WriteFile(composeFile, []byte(body), fileMode)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, errors.IOError{Msg: "file io fail"}
	}
	defer os.RemoveAll(composeFile)

	err = appsMonitor.EnableEventMonitoring(data[ID].(string), composeFile)
	if err != nil {
		dbExecutor.DeleteApp(data[ID].(string))
		return nil, err
	}

	if eventIds, exists := query[EVENTID]; exists {
		err = dockerExecutor.UpWithEvent(data[ID].(string), composeFile, eventIds.([]string)[0], appsMonitor.GetEventChannel())
	} else {
		err = dockerExecutor.Up(data[ID].(string), composeFile, true)
	}

	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		e := dockerExecutor.DownWithRemoveImages(data[ID].(string), composeFile)
		if e != nil {
			logger.Logging(logger.ERROR, e.Error())
		}
		dbExecutor.DeleteApp(data[ID].(string))
		return nil, err
	}

	deployedApp, err := executor.App(data[ID].(string))
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}
	deployedApp[ID] = data[ID].(string)

	return deployedApp, nil
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

	description := make(map[string]interface{})
	err = json.Unmarshal([]byte(app[DESCRIPTION].(string)), &description)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, errors.IOError{"json unmarshal fail"}
	}

	yaml, err := yaml.Marshal(description)
	if err != nil {
		return nil, errors.InvalidYaml{Msg: "invalid yaml syntax"}
	}

	composeFile := genYamlFileName(appId, "app")
	err = ioutil.WriteFile(composeFile, yaml, fileMode)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, errors.IOError{Msg: "file io fail"}
	}
	defer os.RemoveAll(composeFile)

	if description[SERVICES] == nil || len(description[SERVICES].(map[string]interface{})) == 0 {
		return nil, errors.Unknown{Msg: "can't find application info"}
	}

	services := make([]map[string]interface{}, 0)
	for _, serviceName := range reflect.ValueOf(description[SERVICES].(map[string]interface{})).MapKeys() {
		service := make(map[string]interface{}, 0)
		state := make(map[string]interface{}, 0)

		config, err := getServiceState(appId, composeFile, serviceName.String())
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return nil, errors.Unknown{Msg: "get state fail"}
		}

		service[NAME] = serviceName.String()
		service[CID] = config[CID]
		service[PORTS] = config[PORTS]
		state[STATUS] = config[STATUS]
		state[EXIT_CODE] = config[EXIT_CODE]
		service[STATE] = state
		services = append(services, service)
	}

	m := make(map[string]interface{})
	m[STATE] = app[STATE].(string)
	m[DESCRIPTION] = string(yaml)
	m[SERVICES] = services
	m[IMAGES] = app[IMAGES]

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

	var description interface{}
	err := yaml.Unmarshal([]byte(body), &description)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return errors.InvalidYaml{Msg: "invalid yaml syntax"}
	}

	description = convert(description)

	jsonData, err := json.Marshal(description)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return errors.InvalidYaml{"invalid yaml syntax"}
	}

	err = dbExecutor.UpdateAppInfo(appId, string(jsonData))
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	return nil
}

// Start app in the target by input appId.
// if starting is failed, Pharos Node will make sure that only previous state.
// can not guarantee about valid operation of containers.
// if succeed to start, return error as nil
// otherwise, return error.
func (depExecutorImpl) StartApp(appId string) error {
	logger.Logging(logger.DEBUG, "IN", appId)
	defer logger.Logging(logger.DEBUG, "OUT")

	app, err := dbExecutor.GetApp(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	state := app["state"].(string)
	if state == RUNNING_STATE {
		return errors.AlreadyReported{Msg: state}
	}

	composeFile, err := setYamlFile(appId, "start")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}
	defer os.RemoveAll(composeFile)

	appsMonitor.LockUpdateAppState()
	defer appsMonitor.UnlockUpdateAppState()

	err = dockerExecutor.Start(appId, composeFile)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		e := restoreState(appId, composeFile, state, true)
		if e != nil {
			logger.Logging(logger.ERROR, err.Error())
		}
		return err
	}

	err = dbExecutor.UpdateAppState(appId, RUNNING_STATE)
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

	app, err := dbExecutor.GetApp(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	state := app["state"].(string)
	if state == EXITED_STATE {
		return errors.AlreadyReported{Msg: state}
	}

	composeFile, err := setYamlFile(appId, "stop")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}
	defer os.RemoveAll(composeFile)

	appsMonitor.LockUpdateAppState()
	defer appsMonitor.UnlockUpdateAppState()

	err = dockerExecutor.Stop(appId, composeFile)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		e := restoreState(appId, composeFile, state, true)
		if e != nil {
			logger.Logging(logger.ERROR, err.Error())
		}
		return err
	}

	err = dbExecutor.UpdateAppState(appId, EXITED_STATE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	return nil
}

// Handle app's event in the target by input appId.
// Event information about the service of the app
// is stored in repository information and tag information.
// if succeed to update, return error as nil
// otherwise, return error.
func (depExecutorImpl) HandleEvents(appId string, body string) error {
	logger.Logging(logger.DEBUG, "IN", appId)
	defer logger.Logging(logger.DEBUG, "OUT")

	convertedBody, err := util.ConvertJsonToMap(body)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}
	events := convertedBody[EVENTS]

	for _, eventInfo := range events.([]interface{}) {
		parsedEvent := make(map[string]interface{})
		parsedEvent, err = parseEventInfo(eventInfo.(map[string]interface{}))
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return err
		}

		switch parsedEvent[ACTION] {
		case PUSH:
			err := updatedDockerImageFromRegistry(appId, parsedEvent)
			if err != nil {
				logger.Logging(logger.ERROR, err.Error())
				return err
			}
		case DELETE:
			err := deletedDockerImageFromRegistry(appId, parsedEvent)
			if err != nil {
				logger.Logging(logger.ERROR, err.Error())
				return err
			}
		}
	}

	return err
}

// Update images and restart containers in the target
// by input appId and stored yaml in db server.
// if you want to update images,
// yaml should be updated as controller.UpdateAppInfo()
// See also controller.UpdateAppInfo().
// and if failed to update images,
// Pharos Node can make sure that previous images by digest.
// if succeed to update, return error as nil
// otherwise, return error.
func (depExecutorImpl) UpdateApp(appId string, query map[string]interface{}) error {
	logger.Logging(logger.DEBUG, "IN", appId)
	defer logger.Logging(logger.DEBUG, "OUT")

	composeFile, err := setYamlFile(appId, "update")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}
	defer os.RemoveAll(composeFile)

	app, err := dbExecutor.GetApp(appId)
	if err != nil {
		logger.Logging(logger.DEBUG, err.Error())
		return convertDBError(err, appId)
	}

	appsMonitor.LockUpdateAppState()
	defer appsMonitor.UnlockUpdateAppState()

	err = dbExecutor.UpdateAppState(appId, UPDATING_STATE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	imageList, err := getImageNames([]byte(app[DESCRIPTION].(string)))
	if err != nil {
		logger.Logging(logger.DEBUG, err.Error())
		return err
	}

	repoDigests := make(map[string]string)
	for _, image := range imageList {
		repoDigest, err := dockerExecutor.GetImageDigestByName(image)
		if err != nil {
			logger.Logging(logger.DEBUG, err.Error())
			return err
		}
		repoDigests[image] = repoDigest
	}

	if query == nil {
		err = updateApp(appId, composeFile, app, repoDigests)
		if err != nil {
			logger.Logging(logger.DEBUG, err.Error())
			return err
		}
	} else {
		serviceName := ""
		images := query[IMAGES].([]string)
		updatedDescription := make(map[string]interface{})

		for _, imageName := range images {
			tagExist, repo, tag, err := extractQueryInfo(imageName)
			if err != nil {
				logger.Logging(logger.DEBUG, err.Error())
				return err
			}
			serviceName, err = getServiceName(repo, []byte(app[DESCRIPTION].(string)))
			if err != nil {
				logger.Logging(logger.DEBUG, err.Error())
				return err
			}
			if tagExist {
				updatedDescription, err = updateYamlFile(appId, composeFile, app[DESCRIPTION].(string), serviceName, repo+":"+tag)
				if err != nil {
					logger.Logging(logger.DEBUG, err.Error())
					return err
				}
			}
			err = updateService(appId, composeFile, app, repoDigests, serviceName)
			if err != nil {
				logger.Logging(logger.DEBUG, err.Error())
				return err
			}
			if tagExist {
				jsonDescription, err := json.Marshal(convert(updatedDescription))
				if err != nil {
					logger.Logging(logger.ERROR, err.Error())
					return errors.InvalidYaml{Msg: "invalid yaml syntax"}
				}

				err = dbExecutor.UpdateAppInfo(appId, string(jsonDescription))
				if err != nil {
					logger.Logging(logger.ERROR, err.Error())
					return convertDBError(err, appId)
				}
			}
		}
	}

	err = dbExecutor.UpdateAppState(appId, RUNNING_STATE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	err = updateAppEvent(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
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

	composeFile, err := setYamlFile(appId, "delete")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}
	defer os.RemoveAll(composeFile)

	err = dockerExecutor.DownWithRemoveImages(appId, composeFile)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		app, e := dbExecutor.GetApp(appId)
		if e != nil {
			logger.Logging(logger.ERROR, e.Error())
			return e
		}

		state := app["state"].(string)
		e = restoreState(appId, composeFile, state, true)
		if e != nil {
			logger.Logging(logger.ERROR, e.Error())
		}
		return err
	}

	err = appsMonitor.DisableEventMonitoring(appId, composeFile)
	if err != nil {
		dbExecutor.DeleteApp(appId)
		return err
	}

	err = dbExecutor.DeleteApp(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	return nil
}

func restoreRepoDigests(appId, composeFile string, repoDigests map[string]string, state string) error {
	for imageName, repoDigest := range repoDigests {
		err := dockerExecutor.ImagePull(repoDigest)
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return err
		}
		imageID, err := dockerExecutor.GetImageIDByRepoDigest(repoDigest)
		if err != nil {

			logger.Logging(logger.ERROR, err.Error())
			return err
		}
		err = dockerExecutor.ImageTag(imageID, imageName)
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return err
		}
	}

	err := restoreState(appId, composeFile, state, true)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}
	return nil
}

func updateApp(appId, composeFile string, app map[string]interface{}, repoDigests map[string]string) error {
	err := dockerExecutor.Pull(appId, composeFile)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		e := restoreRepoDigests(appId, composeFile, repoDigests, app[STATE].(string))
		if e != nil {
			logger.Logging(logger.ERROR, e.Error())
		}
		return err
	}
	err = dockerExecutor.Up(appId, composeFile, true)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		e := restoreRepoDigests(appId, composeFile, repoDigests, app[STATE].(string))
		if e != nil {
			logger.Logging(logger.ERROR, e.Error())
		}
		return err
	}
	return err
}

func updateService(appId, composeFile string, app map[string]interface{}, repoDigests map[string]string, services ...string) error {
	err := dockerExecutor.Pull(appId, composeFile, services...)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		e := restoreRepoDigests(appId, composeFile, repoDigests, app[STATE].(string))
		if e != nil {
			logger.Logging(logger.ERROR, e.Error())
		}
		return err
	}
	err = dockerExecutor.Up(appId, composeFile, true, services...)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		e := restoreRepoDigests(appId, composeFile, repoDigests, app[STATE].(string))
		if e != nil {
			logger.Logging(logger.ERROR, e.Error())
		}
		return err
	}
	return err
}

// Restore app state by previous state.
// See also controller.StartApp(), controller.StopApp()
// if succeed to restore, return error as nil
// otherwise, return error.
func restoreState(appId, composeFile, state string, forceRecreate bool) error {
	var err error

	if len(state) == 0 {
		logger.Logging(logger.ERROR, "empty state")
		return errors.InvalidParam{Msg: "empty state"}
	}

	switch state {
	case EXITED_STATE:
		err = dockerExecutor.Stop(appId, composeFile)
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return err
		}
		err = dbExecutor.UpdateAppState(appId, EXITED_STATE)
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
		}
	case RUNNING_STATE:
		err = dockerExecutor.Up(appId, composeFile, forceRecreate)
		if err != nil {
			if strings.Contains(err.Error(), "already exists in network") && forceRecreate == false {
				logger.Logging(logger.INFO, "It occurs when a service is already restarted by itself and expected result")
			} else {
				logger.Logging(logger.ERROR, err.Error())
				return err
			}
		}
		err = dbExecutor.UpdateAppState(appId, RUNNING_STATE)
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
		}
	}

	return err
}

// Set YAML file about an app on a path.
// The path is defined as contant
// if setting YAML is succeeded, return error as nil
// otherwise, return error.
func setYamlFile(appId, api string) (string, error) {
	app, err := dbExecutor.GetApp(appId)
	if err != nil {
		return "", convertDBError(err, appId)
	}
	description := make(map[string]interface{})
	err = json.Unmarshal([]byte(app[DESCRIPTION].(string)), &description)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return "", errors.IOError{"json unmarshal fail"}
	}
	yaml, err := yaml.Marshal(description)
	if err != nil {
		return "", errors.InvalidYaml{Msg: "invalid yaml syntax"}
	}

	composeFile := genYamlFileName(appId, api)
	err = ioutil.WriteFile(composeFile, yaml, fileMode)
	if err != nil {
		return "", errors.IOError{Msg: "file io fail"}
	}
	return composeFile, nil
}

func updateYamlFile(appId, composeFile, orginDescription, service, newImage string) (map[string]interface{}, error) {
	updatedDescription := make(map[string]interface{})

	err := json.Unmarshal([]byte(orginDescription), &updatedDescription)
	if err != nil {
		return nil, errors.IOError{Msg: "json unmarshal fail"}
	}

	if updatedDescription[SERVICES] == nil || len(updatedDescription[SERVICES].(map[string]interface{})) == 0 {
		return nil, errors.Unknown{Msg: "can't find application info"}
	}

	for serviceName, serviceInfo := range updatedDescription[SERVICES].(map[string]interface{}) {
		if serviceName == service {
			serviceInfo.(map[string]interface{})[IMAGE] = newImage
		}
	}

	yaml, err := yaml.Marshal(updatedDescription)
	if err != nil {
		return nil, errors.InvalidYaml{Msg: "invalid yaml syntax"}
	}
	err = ioutil.WriteFile(composeFile, yaml, fileMode)
	if err != nil {
		return nil, errors.IOError{Msg: "file io fail"}
	}
	return updatedDescription, err
}

// Get service state by service name.
// First of all, get container name using docker-compose ps <service name>
// And then, get service config from using docker inspect <container name>
// if getting service state is succeed, return service state
// otherwise, return error.
func getServiceState(appId, composeFile, serviceName string) (map[string]interface{}, error) {
	infos, err := dockerExecutor.Ps(appId, composeFile, serviceName)
	if len(infos) == 0 {
		logger.Logging(logger.ERROR, "no information about service")
		return nil, errors.Unknown{Msg: "no information about service"}
	}
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}

	containerName := infos[0]["Name"]

	serviceInfo, err := dockerExecutor.GetContainerConfigByName(containerName)
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

// convert function changes the type of key from interface{} to string.
// yaml package unmarshal key-value pairs with map[interface{}]interface{}.
// but map[interface{}]interface{} type is not supported in json package.
// this function is available to resolve the problem.
func convert(in interface{}) interface{} {
	switch x := in.(type) {
	case map[interface{}]interface{}:
		out := map[string]interface{}{}
		for key, value := range x {
			out[key.(string)] = convert(value)
		}
		return out
	case []interface{}:
		for key, value := range x {
			x[key] = convert(value)
		}
	}
	return in
}

// Update events received from the registry are reflected in the app collection.
func updatedDockerImageFromRegistry(appId string, imageInfo map[string]interface{}) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	repository := imageInfo[HOST].(string) + "/" + imageInfo[REPOSITORY].(string)

	err := dbExecutor.UpdateAppEvent(appId, repository, imageInfo[TAG].(string), UPDATE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	return nil
}

// Delete events received from the registry are reflected in the app collection.
func deletedDockerImageFromRegistry(appId string, imageInfo map[string]interface{}) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	repository := imageInfo[HOST].(string) + "/" + imageInfo[REPOSITORY].(string)

	err := dbExecutor.UpdateAppEvent(appId, repository, imageInfo[TAG].(string), DELETE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return convertDBError(err, appId)
	}

	return nil
}

// parseEventInfo parse data which is matched image-info on DB from event-notification.
func parseEventInfo(eventInfo map[string]interface{}) (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	targetInfoEvent := make(map[string]interface{})
	requestInfoEvent := make(map[string]interface{})
	parsedEvent := make(map[string]interface{})

	targetInfoEvent = eventInfo[TARGETINFO].(map[string]interface{})
	requestInfoEvent = eventInfo[REQUESTINFO].(map[string]interface{})

	parsedEvent[ACTION] = eventInfo[ACTION]
	parsedEvent[REPOSITORY] = targetInfoEvent[REPOSITORY]
	parsedEvent[TAG] = targetInfoEvent[TAG]
	parsedEvent[HOST] = requestInfoEvent[HOST]

	return parsedEvent, nil
}

func extractQueryInfo(imageName string) (bool, string, string, error) {
	words := strings.Split(imageName, "/")
	imageNameWithoutRepo := strings.Join(words[:len(words)-1], "/")
	repo := strings.Split(words[len(words)-1], ":")

	imageNameWithoutTag := imageNameWithoutRepo
	if len(words) > 1 {
		imageNameWithoutTag += "/"
	}
	imageNameWithoutTag += repo[0]

	tagInfo := strings.Replace(imageName, imageNameWithoutTag, "", -1)
	isTag := false
	if len(tagInfo) != 0 {
		isTag = true
		tagInfo = strings.Split(tagInfo, ":")[1]
		if len(tagInfo) == 0 {
			return false, "", "", errors.Unknown{Msg: "invalid repogitory"}
		}
	}

	return isTag, imageNameWithoutTag, tagInfo, nil
}

// Get name of service which use given imageName.
// If getting image names is succeeded, return name of service.
// otherwise, return error.
func getServiceName(repository string, desc []byte) (string, error) {
	description := make(map[string]interface{})
	err := json.Unmarshal(desc, &description)
	if err != nil {
		return "", errors.IOError{Msg: "json unmarshal fail"}
	}
	if description[SERVICES] == nil || len(description[SERVICES].(map[string]interface{})) == 0 {
		return "", errors.Unknown{Msg: "can't find application info"}
	}

	for serviceName, serviceInfo := range description[SERVICES].(map[string]interface{}) {
		fullImageName := serviceInfo.(map[string]interface{})[IMAGE].(string)
		words := strings.Split(fullImageName, "/")
		imageNameWithoutRepo := strings.Join(words[:len(words)-1], "/")
		repo := strings.Split(words[len(words)-1], ":")

		imageNameWithoutTag := imageNameWithoutRepo
		if len(words) > 1 {
			imageNameWithoutTag += "/"
		}
		imageNameWithoutTag += repo[0]

		if imageNameWithoutTag == repository {
			return serviceName, nil
		}
	}

	return "", errors.Unknown{Msg: "can't find matched service"}
}

// Get an image name list shown in app[DESCRIPTION]
// If getting an image name list is succeeded, return an image name list.
// otherwise, return error.
func getImageNames(desc []byte) ([]string, error) {
	description := make(map[string]interface{})
	err := json.Unmarshal(desc, &description)
	if err != nil {
		return nil, errors.IOError{Msg: "json unmarshal fail"}
	}
	if description[SERVICES] == nil {
		return nil, errors.Unknown{Msg: "No service in YAML description"}
	}

	imageList := make([]string, 0)
	for _, serviceInfo := range description[SERVICES].(map[string]interface{}) {
		imageList = append(imageList, serviceInfo.(map[string]interface{})[IMAGE].(string))
	}
	return imageList, nil
}

func updateAppEvent(appId string) error {
	app, err := dbExecutor.GetApp(appId)
	if err != nil {
		logger.Logging(logger.DEBUG, err.Error())
		return convertDBError(err, appId)
	}

	description := make(map[string]interface{})
	err = json.Unmarshal([]byte(app[DESCRIPTION].(string)), &description)
	if err != nil {
		logger.Logging(logger.DEBUG, err.Error())
		return errors.IOError{"json unmarshal fail"}
	}

	services := description[SERVICES].(map[string]interface{})
	images := app[IMAGES].([]map[string]interface{})
	for _, serviceInfo := range services {
		descImageName := serviceInfo.(map[string]interface{})[IMAGE].(string)
		for _, image := range images {
			if changes, exists := image[CHANGES]; exists {
				changesTag := changes.(map[string]interface{})[TAG].(string)
				if (image[NAME].(string) + ":" + changesTag) == descImageName {
					err = dbExecutor.UpdateAppEvent(appId, image[NAME].(string), changesTag, NONE)
					if err != nil {
						logger.Logging(logger.DEBUG, err.Error())
						return convertDBError(err, appId)
					}
				}
			}
		}
	}
	return err
}

func restoreAllAppsState() {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	apps, err := dbExecutor.GetAppList()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return
	}

	for _, app := range apps {
		appId := app[ID].(string)

		composeFile, err := setYamlFile(appId, "restoreAllAppsState")
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return
		}
		defer os.RemoveAll(composeFile)

		app, err := dbExecutor.GetApp(appId)
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return
		}

		state := app["state"].(string)
		restoreState(appId, composeFile, state, false)
	}
}

func genYamlFileName(appId string, api string) string {
	return appId + "_" + api + "_" + strconv.Itoa(genRandNumber(100000))
}

func genRandNumber(n int) int {
	source := rand.NewSource(time.Now().UnixNano())
	return rand.New(source).Intn(n)
}

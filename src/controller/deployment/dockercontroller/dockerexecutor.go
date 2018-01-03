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

// Package dockercontroller provide functionlity of docker commands.
package dockercontroller

import (
	"commons/errors"
	"commons/logger"
	"controller/deployment/dockercontroller/compose"
	shell "controller/shellcommand"
	"encoding/json"
	"reflect"
)

type Command interface {
	Create(path string) error
	Up(path string) error
	Down(path string) error
	DownWithRemoveImages(path string) error
	Start(path string) error
	Stop(path string) error
	Pause(path string) error
	Unpause(path string) error
	Pull(path string) error
	Ps(path string, args ...string) (string, error)
	Inspect(IdOrName string) (string, error)
	GetImageDigest(imageName string) (string, error)
}

type dockerExecutorImpl struct {
	dockerCommand string
}

var Executor dockerExecutorImpl
var shellExecutor shell.Command

func init() {
	shellExecutor = shell.Executor
	Executor.dockerCommand = "docker"
}
// Creating containers of service list in the yaml description.
// if succeed to create, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Create(path string) error {
	return compose.Executor.Create(path)
}

// Pulling images and creating containers and start containers
// of service list in the yaml description.
// if succeed to up, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Up(path string) error {
	return compose.Executor.Up(path)
}

// Stop and remove containers of service list in the yaml description.
// if succeed to down, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Down(path string) error {
	return compose.Executor.Down(path)
}

// Stop and remove containers, remove images of service list
// in the yaml description.
// if succeed to down with rmi option, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) DownWithRemoveImages(path string) error {
	return compose.Executor.DownWithRemoveImages(path)
}

// Starting containers of service list in the yaml description.
// if succeed to start, return error as nil
// otherwise, return error. (if contianers is not created, return error)
func (dockerExecutorImpl) Start(path string) error {
	return compose.Executor.Start(path)
}

// Stopping containers of service list in the yaml description.
// if succeed to stop, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Stop(path string) error {
	return compose.Executor.Stop(path)
}

// Pause containers of service list in the yaml description.
// if succeed to pause, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Pause(path string) error {
	return compose.Executor.Pause(path)
}

// Resume paused containers of service list in the yaml description.
// if succeed to resume, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Unpause(path string) error {
	return compose.Executor.Unpause(path)
}

// Pulling images of service list in the yaml description.
// if succeed to pull, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Pull(path string) error {
	return compose.Executor.Pull(path)
}

// Getting container informations of service list in the yaml description.
// (e.g. container name, command, state, port)
// if succeed to get, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Ps(path string, args ...string) (string, error) {
	return compose.Executor.Ps(path, args...)
}

// Getting image information by input image name.
// if succeed to get, return string of image information
// otherwise, return error.
func (d dockerExecutorImpl) Inspect(IdOrName string) (string, error) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	funcName := "inspect"

	args := []string{funcName, IdOrName}
	return d.executeCommand(args...)
}

// Getting image digest by input image name.
// if succeed to get, return string of image digest
// otherwise, return error.
func (d dockerExecutorImpl) GetImageDigest(imageName string) (string, error) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")
	var inspectMap []map[string]interface{}

	funcName := "inspect"

	args := []string{funcName, imageName}
	ret, err := d.executeCommand(args...)

	if err != nil {
		logger.Logging(logger.DEBUG, "executeCommand error")
		return ret, err
	}

	json.Unmarshal([]byte(ret), &inspectMap)
	digestsList := reflect.ValueOf(inspectMap[0]["RepoDigests"])
	if digestsList.Len() == 0 {
		return ret, errors.NotFoundImage{"RepoDigests"}
	} else {
		ret = digestsList.Index(0).Interface().(string)
		return ret, err
	}
}

// Executing shell command.
// if succeed to execute, return output of command
// otherwise, return error.
func (d dockerExecutorImpl) executeCommand(args ...string) (string, error) {
	var tmpArgs []string
	tmpArgs = append(tmpArgs, args...)
	logger.Logging(logger.DEBUG, tmpArgs...)

	return shellExecutor.ExecuteCommand(Executor.dockerCommand, tmpArgs...)
}

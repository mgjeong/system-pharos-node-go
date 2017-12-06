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

// Package compose provide functionlity of docker-compose commands.
package compose

import "commons/logger"
import "controller/shellcommand"

var Executor composeExecutorImpl
var shellExecutor func(command string, args ...string) (string, error)

type composeExecutorImpl struct {
	composeCommand string
	firstArg       string
}

func init() {
	Executor.composeCommand = "docker-compose"
	Executor.firstArg = "-f"
	shellExecutor = shellcommand.ExecuteCommand
}

// Implement of dockercontroller.Create().
// if succeed to create, return error as nil
// otherwise, return error.
func (c composeExecutorImpl) Create(path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	funcName := "create"

	args := []string{path, funcName}
	args = append(args, "--force-recreate")
	return c.executeCommand(args...)
}

// Implement of dockercontroller.Up().
// if succeed to up, return error as nil
// otherwise, return error.
func (c composeExecutorImpl) Up(path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	funcName := "up"

	args := []string{path, funcName}
	args = append(args, "-d", "--force-recreate")
	return c.executeCommand(args...)
}

// Implement of dockercontroller.Down().
// if succeed to down, return error as nil
// otherwise, return error.
func (c composeExecutorImpl) Down(path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	funcName := "down"

	args := []string{path, funcName}
	return c.executeCommand(args...)
}

// Implement of dockercontroller.DownWithRemoveImages().
// if succeed to down with rmi option, return error as nil
// otherwise, return error.
func (c composeExecutorImpl) DownWithRemoveImages(path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	funcName := "down"

	args := []string{path, funcName, "--rmi", "all"}
	return c.executeCommand(args...)
}

// Implement of dockercontroller.Start().
// if succeed to start, return error as nil
// otherwise, return error. (if contianers is not created, return error)
func (c composeExecutorImpl) Start(path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	funcName := "start"

	args := []string{path, funcName}
	return c.executeCommand(args...)
}

// Implement of dockercontroller.Stop().
// if succeed to stop, return error as nil
// otherwise, return error.
func (c composeExecutorImpl) Stop(path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	funcName := "stop"

	args := []string{path, funcName}
	return c.executeCommand(args...)
}

// Implement of dockercontroller.Pause().
// if succeed to pause, return error as nil
// otherwise, return error.
func (c composeExecutorImpl) Pause(path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	funcName := "pause"

	args := []string{path, funcName}
	return c.executeCommand(args...)
}

// Implement of dockercontroller.Resume().
// if succeed to resume, return error as nil
// otherwise, return error.
func (c composeExecutorImpl) Unpause(path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	funcName := "unpause"

	args := []string{path, funcName}
	return c.executeCommand(args...)
}

// Implement of dockercontroller.Pull().
// if succeed to pull, return error as nil
// otherwise, return error.
func (c composeExecutorImpl) Pull(path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	funcName := "pull"

	args := []string{path, funcName}
	return c.executeCommand(args...)
}

// Implement of dockercontroller.Ps().
// (e.g. container name, command, state, port)
// if succeed to get, return error as nil
// otherwise, return error.
func (c composeExecutorImpl) Ps(path string, services ...string) (string, error) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	funcName := "ps"

	args := []string{path, funcName}

	for _, service := range services {
		args = append(args, service)
	}

	return c.executeCommandWithMsg(args...)
}

// Executing shell command.
// if succeed to execute, return error as nil
// otherwise, return error.
func (c composeExecutorImpl) executeCommand(args ...string) error {
	tmpArgs := []string{c.firstArg}
	tmpArgs = append(tmpArgs, args...)
	logger.Logging(logger.DEBUG, tmpArgs...)

	_, err := shellExecutor(c.composeCommand, tmpArgs...)

	return err
}

// Executing shell command with output message.
// if succeed to execute, return output message
// otherwise, return error.
func (c composeExecutorImpl) executeCommandWithMsg(args ...string) (string, error) {
	tmpArgs := []string{c.firstArg}
	tmpArgs = append(tmpArgs, args...)
	logger.Logging(logger.DEBUG, tmpArgs...)

	msg, err := shellExecutor(c.composeCommand, tmpArgs...)

	return msg, err
}

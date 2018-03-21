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

// Package shellcommand provide functionlity of command to shell.
package shellcommand

import (
	"commons/errors"
	"commons/logger"
	"os/exec"
	"strings"
)

type Command interface {
	ExecuteCommand(command string, args ...string) (string, error)
}

type shellExecutorImpl struct{}

var Executor shellExecutorImpl

type shellInnerInterface interface {
	executeCommand(name string, arg ...string)
	getOutput() ([]byte, error)
}

type shellInnerExecutor struct {
	cmd *exec.Cmd
}

var shell shellInnerInterface

// Executing command to shell (private function).
func (e *shellInnerExecutor) executeCommand(name string, arg ...string) {
	e.cmd = exec.Command(name, arg...)
}

// Getting command response  shell.
// return stdout of shell command.
func (e *shellInnerExecutor) getOutput() ([]byte, error) {
	return e.cmd.CombinedOutput()
}

func init() {
	shell = &shellInnerExecutor{}
}

// Executing command to shell.
// if succeed to executing, return message of stdout
// otherwise, return error.
func (shellExecutorImpl) ExecuteCommand(command string, args ...string) (string, error) {
	logger.Logging(logger.DEBUG, args...)
	shell.executeCommand(command, args...)
	out, err := shell.getOutput()

	ret := string(out[:])
	if err == nil {
		logger.Logging(logger.DEBUG, ret)
		return ret, nil
	}

	switch {
	case isNotFoundDockerComposeFile(&ret):
		return ret, errors.NotFound{ret}
	case isNotFoundDockerEngine(&ret):
		return ret, errors.NotFound{ret}
	case isNotFoundStatFile(&ret):
		return ret, errors.NotFound{ret}
	case isNotFoundCPUInfoFile(&ret):
		return ret, errors.NotFound{ret}
	case isNotFoundMemInfoFile(&ret):
		return ret, errors.NotFound{ret}
	case isInvalidYaml(&ret):
		return ret, errors.InvalidYaml{ret}
	case isNotFoundDockerImage(&ret):
		return ret, errors.NotFoundImage{ret}
	case isAlreadyAllocatedPort(&ret):
		return ret, errors.AlreadyAllocatedPort{ret}
	case isAlreadyUsedContainerName(&ret):
		return ret, errors.AlreadyUsedName{ret}
	case isInvalidContainerName(&ret):
		return ret, errors.InvalidContainerName{ret}

	default:
		return ret, errors.Unknown{ret}
	}
}

var notFoundCPUInfoFile string = "/proc/cpuinfo: No such file or directory"
var notFoundMemInfoFile string = "/proc/meminfo: No such file or directory"
var notFoundStatFile string = "/proc/stat: No such file or directory"
var notFoundDockerComposeFile string = "Can't find a suitable configuration file in this directory or any" +
	"parent. Are you in the right directory?"
var notFoundFile string = ".IOError: [Errno 2] No such file or directory:"
var notFoundDockerEngine string = "Couldn't connect to Docker daemon"
var invalidYaml string = "is invalid because:"
var notFoundDockerImage string = "No such object:"
var alreayUsedContainerName string = "is already in use by container"
var alreadyAllocatedPort string = "port is already allocated"
var invalidContainerName string = "Invalid container name"

// Check output message for not found proc file.
// if output message has string such as "/proc/stat: no such file",  return true
// otherwise, return false.
func isNotFoundStatFile(msg *string) bool {
	return strings.Contains(*msg, notFoundStatFile)
}

// Check output message for not found proc file.
// if output message has string such as "/proc/cpuinfo: no such file",  return true
// otherwise, return false.
func isNotFoundCPUInfoFile(msg *string) bool {
	return strings.Contains(*msg, notFoundCPUInfoFile)
}

// Check output message for not found proc file.
// if output message has string such as "/proc/meminfo: no such file", return true
// otherwise, return false.
func isNotFoundMemInfoFile(msg *string) bool {
	return strings.Contains(*msg, notFoundMemInfoFile)
}

// Check output message for not found yaml file.
// if output message has string such as "not found yaml file", return true
// otherwise, return false.
func isNotFoundDockerComposeFile(msg *string) bool {
	return strings.Contains(*msg, notFoundDockerComposeFile) ||
		strings.Contains(*msg, notFoundFile)
}

// Check output message for not found docker engine.
// if output message has string such as "not found docker engine", return true
// otherwise, return false.
func isNotFoundDockerEngine(msg *string) bool {
	return strings.Contains(*msg, notFoundDockerEngine)
}

// Check output message for invalid yaml.
// if output message has string such as "invalid yaml form", return true
// otherwise, return false.
func isInvalidYaml(msg *string) bool {
	return strings.Contains(*msg, invalidYaml)
}

// Check output message for not found docker image.
// if output message has string such as "not found docker image", return true
// otherwise, return false.
func isNotFoundDockerImage(msg *string) bool {
	return strings.Contains(*msg, notFoundDockerImage)
}

// Check output message for already allocated port.
// if output message has string such as "already allocated port", return true
// otherwise, return false.
func isAlreadyAllocatedPort(msg *string) bool {
	return strings.Contains(*msg, alreadyAllocatedPort)
}

// Check output message for already used name.
// if output message has string such as "already used name", return true
// otherwise, return false.
func isAlreadyUsedContainerName(msg *string) bool {
	return strings.Contains(*msg, alreayUsedContainerName)
}

// Check output message for invalid container name.
// if output message has string such as "invalid container name", return true
// otherwise, return false.
func isInvalidContainerName(msg *string) bool {
	return strings.Contains(*msg, invalidContainerName)
}

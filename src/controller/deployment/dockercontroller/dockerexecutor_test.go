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
package dockercontroller

import (
	"commons/errors"
	"testing"
)

var doSomething func(incommand string, inargs ...string) (string, error)

func mockExecuteCommand(command string, args ...string) (string, error) {
	return doSomething(command, args...)
}

type shellFunc func(command string, args ...string) (string, error)

var oldShellInnerExecutor shellFunc

const (
	inspect = iota
	getImageDigests
)

type executor []func(string) (string, error)

func getCommandList(imageName string) map[int][]string {
	commandList := make(map[int][]string)
	commandList[inspect] = []string{"inspect", "inspect", imageName}
	commandList[getImageDigests] = []string{"getImageDigests", "inspect", imageName}

	return commandList
}

func getExecutor() map[int]func(string) (string, error) {
	executor := make(map[int]func(string) (string, error))
	executor[inspect] = Executor.Inspect
	executor[getImageDigests] = Executor.GetImageDigest

	return executor
}

func TestExpectEqualCommandList(t *testing.T) {
	imageName := "ubuntu"
	commandList := getCommandList(imageName)
	executor := getExecutor()

	for i, f := range executor {
		t.Run(commandList[i][0], func(t *testing.T) {
			doSomething = func(incommand string, inargs ...string) (string, error) {
				for ii, arg := range inargs {
					if arg != commandList[i][ii+1] {
						t.Error()
					}
				}
				return "[{\"RepoDigests\":[\"sha\"]}]", nil
			}
			f(imageName)
		})
	}

	imageName = "abc"
	commandList = getCommandList(imageName)
	t.Run(commandList[getImageDigests][0]+"/NoImage", func(t *testing.T) {
		doSomething = func(incommand string, inargs ...string) (string, error) {
			for ii, arg := range inargs {
				if arg != commandList[getImageDigests][ii+1] {
					t.Error()
				}
			}
			return "[]", errors.NotFound{imageName}
		}
		Executor.GetImageDigest(imageName)
	})

	t.Run(commandList[getImageDigests][0]+"/NoRepoDigests", func(t *testing.T) {
		doSomething = func(incommand string, inargs ...string) (string, error) {
			for ii, arg := range inargs {
				if arg != commandList[getImageDigests][ii+1] {
					t.Error()
				}
			}
			return "[{\"RepoDigests\":[]}]", nil
		}
		Executor.GetImageDigest(imageName)
	})
}
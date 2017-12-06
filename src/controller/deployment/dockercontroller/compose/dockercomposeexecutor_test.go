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
package compose

import "testing"

var doSomething func(incommand string, inargs ...string) (string, error)

func mockExecuteCommand(command string, args ...string) (string, error) {
	return doSomething(command, args...)
}

type shellFunc func(command string, args ...string) (string, error)

var oldShellExecutor shellFunc

type tearDown func(t *testing.T)

func setUp(t *testing.T) tearDown {
	oldShellExecutor = shellExecutor
	shellExecutor = mockExecuteCommand

	return func(t *testing.T) {
		shellExecutor = oldShellExecutor
	}
}

const (
	create = iota
	up
	down
	downWithRemoveImages
	start
	stop
	pause
	unpause
	pull
	ps
)

type executor []func(string) (string, error)

func getCommandList(path string) map[int][]string {
	commandList := make(map[int][]string)
	commandList[create] = []string{"-f", path, "create", "--force-recreate"}
	commandList[up] = []string{"-f", path, "up", "-d", "--force-recreate"}
	commandList[down] = []string{"-f", path, "down"}
	commandList[downWithRemoveImages] = []string{"-f", path, "down", "--rmi", "all"}
	commandList[start] = []string{"-f", path, "start"}
	commandList[stop] = []string{"-f", path, "stop"}
	commandList[pause] = []string{"-f", path, "pause"}
	commandList[unpause] = []string{"-f", path, "unpause"}
	commandList[pull] = []string{"-f", path, "pull"}

	return commandList
}

func getExecutor() map[int]func(string) error {
	executor := make(map[int]func(string) error)
	executor[create] = Executor.Create
	executor[up] = Executor.Up
	executor[down] = Executor.Down
	executor[downWithRemoveImages] = Executor.DownWithRemoveImages
	executor[start] = Executor.Start
	executor[stop] = Executor.Stop
	executor[pause] = Executor.Pause
	executor[unpause] = Executor.Unpause
	executor[pull] = Executor.Pull

	return executor
}

func TestExpectEqualCommandList(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	path := "./"
	commandList := getCommandList(path)
	executor := getExecutor()

	for i, f := range executor {
		t.Run(commandList[i][2], func(t *testing.T) {
			doSomething = func(incommand string, inargs ...string) (string, error) {
				for ii, arg := range inargs {
					if arg != commandList[i][ii] {
						t.Error()
					}
				}
				return "", nil
			}
			f(path)
		})
	}
}

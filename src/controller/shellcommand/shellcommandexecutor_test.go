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
package shellcommand

import (
	"commons/errors"
	"testing"
)

type testError struct {
	msg string
}

func (e testError) Error() string {
	return e.msg
}

var doSomething func() ([]byte, error)

var name string
var args []string

type mockShellExecutor struct{}

func (m *mockShellExecutor) executeCommand(n string, a ...string) {
	name = n
	args = make([]string, 5, 5)
	for i, arg := range a {
		args[i] = arg
	}
}

func (m *mockShellExecutor) getOutput() ([]byte, error) {
	return doSomething()
}

var mockShell mockShellExecutor
var oldShell shellInnerInterface

var testCommand string
var testArgs []string
var testRet []byte

type tearDown func(t *testing.T)

func setUp(t *testing.T) tearDown {
	mockShell := mockShellExecutor{}

	oldShell = shell
	shell = &mockShell

	testCommand = "testCommand"
	testArgs = []string{"1", "2", "3", "4", "5"}

	return func(t *testing.T) {
		shell = oldShell
	}
}

func runShellCommand(str string) (string, error) {
	var err error
	if str != "" {
		err = errors.Unknown{str}
	} else {
		err = nil
	}

	testRet = []byte(str)
	doSomething = func() ([]byte, error) {
		return testRet, err
	}
	return Executor.ExecuteCommand(testCommand, testArgs...)
}

func expectReturnErrorNotFoundError(t *testing.T, e string) {
	_, err := runShellCommand(e)

	switch err.(type) {
	default:
		t.Error()
	case errors.NotFound:
	}
}

func TestShellCommand(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	t.Run("ExpectSameCommandandArgs", func(t *testing.T) {
		ret, err := runShellCommand("")
		if testCommand != name || ret != string(testRet) || err != nil {
			t.Error()
		}
		for i, arg := range testArgs {
			if arg != args[i] {
				t.Error()
			}
		}
	})

	t.Run("ExpectReturnErrorNotFoundError1", func(t *testing.T) {
		expectReturnErrorNotFoundError(t, "test Can't find a suitable configuration file in this directory or any"+
			"parent. Are you in the right directory?...")
	})
	t.Run("ExpectReturnErrorNotFoundError2", func(t *testing.T) {
		expectReturnErrorNotFoundError(t, "tes .IOError: [Errno 2] No such file or directory: test")
	})
	t.Run("ExpectReturnErrorNotFoundError3", func(t *testing.T) {
		expectReturnErrorNotFoundError(t, "test Couldn't connect to Docker daemon test")
	})

	t.Run("ExpectReturnErrorInvalidYamlError", func(t *testing.T) {
		_, err := runShellCommand("test is invalid because: test")
		switch err.(type) {
		default:
			t.Error()
		case errors.InvalidYaml:
		}
	})

	t.Run("ExpectReturnErrorNotFoundImage", func(t *testing.T) {
		_, err := runShellCommand("test No such object: test")
		switch err.(type) {
		default:
			t.Error()
		case errors.NotFoundImage:
		}
	})

	t.Run("ExpectReturnErrorAlreadyAllocatedPort", func(t *testing.T) {
		_, err := runShellCommand("test port is already allocated")
		t.Log(err.Error())
		switch err.(type) {
		default:
			t.Error()
		case errors.AlreadyAllocatedPort:
		}
	})

	t.Run("ExpectReturnErrorAlreadyUsedName", func(t *testing.T) {
		_, err := runShellCommand("test is already in use by container")
		t.Log(err.Error())
		switch err.(type) {
		default:
			t.Error()
		case errors.AlreadyUsedName:
		}
	})

	t.Run("ExpectReturnErrorInvalidContainerName", func(t *testing.T) {
		_, err := runShellCommand("test Invalid container name test")
		switch err.(type) {
		default:
			t.Error()
		case errors.InvalidContainerName:
		}
	})

	t.Run("ExpectReturnErrorUnknownError", func(t *testing.T) {
		_, err := runShellCommand("unknown")
		switch err.(type) {
		default:
			t.Error()
		case errors.Unknown:
		}
	})

}

type boolTest struct {
	name         string
	expect       string
	inputMsg     string
	f            func(*string) bool
	expectResult bool
}

func TestBoolFunction(t *testing.T) {
	boolTestList := [7]boolTest{}
	str := "test"
	matchStr := "Match"
	unmatchStr := "Unmatch"

	for i, _ := range boolTestList {
		if i%2 == 1 {
			boolTestList[i].expect = unmatchStr
			boolTestList[i].inputMsg = str
			boolTestList[i].expectResult = false
		} else {
			boolTestList[i].expect = matchStr
			boolTestList[i].expectResult = true
		}
	}

	boolTestList[0].name = "isNotFoundDockerEngine/"
	boolTestList[0].inputMsg = notFoundDockerEngine
	boolTestList[0].f = isNotFoundDockerEngine

	boolTestList[1].name = "isNotFoundDockerEngine/"
	boolTestList[1].f = isNotFoundDockerEngine

	boolTestList[2].name = "isInvalidYaml/"
	boolTestList[2].inputMsg = invalidYaml
	boolTestList[2].f = isInvalidYaml

	boolTestList[3].name = "isInvalidYaml/"
	boolTestList[3].f = isInvalidYaml

	boolTestList[4].name = "isNotFoundDockerComposeFile/"
	boolTestList[4].expect += "1"
	boolTestList[4].inputMsg = notFoundDockerComposeFile
	boolTestList[4].f = isNotFoundDockerComposeFile

	boolTestList[5].name = "isNotFoundDockerComposeFile/"
	boolTestList[5].f = isNotFoundDockerComposeFile

	boolTestList[6].name = "isNotFoundDockerComposeFile/"
	boolTestList[6].expect += "2"
	boolTestList[6].inputMsg = notFoundFile
	boolTestList[6].f = isNotFoundDockerComposeFile

	for _, test := range boolTestList {
		t.Run(test.name+test.expect, func(t *testing.T) {
			if test.f(&test.inputMsg) != test.expectResult {
				t.Error()
			}
		})
	}
}

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

	"golang.org/x/net/context"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"

	"github.com/docker/libcompose/project"

	origineErr "errors"
	"strconv"
	"strings"
	"testing"
)

type tearDown func(t *testing.T)

func setUp(t *testing.T) tearDown {
	getComposeInstance = fakeGetComposeInstance
	getImageList = fakeImageList
	getContainerList = fakeContainerList
	getContainerInspect = fakeContainerExecInspect

	return func(t *testing.T) {
		getComposeInstance = getComposeInstanceImpl
		getImageList = (*docker.Client).ImageList
		getContainerList = (*docker.Client).ContainerList
		getContainerInspect = (*docker.Client).ContainerInspect
	}
}

var fakeGetComposeInstanceImpl func() (project.APIProject, error)
var fakeRunImageList func() ([]types.ImageSummary, error)
var fakeRunContainerList func() ([]types.Container, error)
var fakeRunContaienrInspect func() (types.ContainerJSON, error)

func fakeGetComposeInstance(string, string) (project.APIProject, error) {
	return fakeGetComposeInstanceImpl()
}

func fakeImageList(*docker.Client, context.Context, types.ImageListOptions) ([]types.ImageSummary, error) {
	return fakeRunImageList()
}

func fakeContainerList(*docker.Client, context.Context, types.ContainerListOptions) ([]types.Container, error) {
	return fakeRunContainerList()
}

func fakeContainerExecInspect(*docker.Client, context.Context, string) (types.ContainerJSON, error) {
	return fakeRunContaienrInspect()
}

func TestGetImageDigestByName(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	t.Run("ReturnErrorWhenReceiveErrorFromDockerEngine", func(t *testing.T) {
		fakeRunImageList = func() ([]types.ImageSummary, error) {
			return nil, origineErr.New("")
		}
		_, err := Executor.GetImageDigestByName("123")
		switch err.(type) {
		default:
			t.Error()
		case errors.Unknown:
		}
	})

	ret := []types.ImageSummary{
		{
			RepoDigests: []string{"", "", ""},
			RepoTags:    []string{"test:latest", "test:111", "test:123"},
		}}

	t.Run("ReturnErrorWhenNotFoundImageInList", func(t *testing.T) {
		fakeRunImageList = func() ([]types.ImageSummary, error) {
			return ret, nil
		}
		_, err := Executor.GetImageDigestByName("test:123")
		switch err.(type) {
		default:
			t.Error()
		case errors.NotFoundImage:
		}
	})

	expected := "ShouldBeReturned"
	t.Run("GetDigestSuccessful", func(t *testing.T) {
		ret[0].RepoDigests[0] = expected
		fakeRunImageList = func() ([]types.ImageSummary, error) {
			return ret, nil
		}
		digest, _ := Executor.GetImageDigestByName("test:123")
		if strings.Compare(digest, expected) != 0 {
			t.Error()
		}
	})
}

func TestGetContainerStateByName(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	t.Run("ReturnErrorWhenReceiveErrorFromDockerEngine", func(t *testing.T) {
		fakeRunContainerList = func() ([]types.Container, error) {
			return nil, origineErr.New("")
		}
		_, err := Executor.GetContainerStateByName("123")
		switch err.(type) {
		default:
			t.Error()
		case errors.Unknown:
		}
	})

	retContainers := []types.Container{
		{
			ID:    "containerId",
			State: "running",
			Names: []string{"/test_latest", "/test_111", "/test_123"},
		},
	}

	t.Run("ReturnErrorWhenNotFoundContainerInList", func(t *testing.T) {
		fakeRunContainerList = func() ([]types.Container, error) {
			return retContainers, nil
		}
		_, err := Executor.GetContainerStateByName("123")
		switch err.(type) {
		default:
			t.Error()
		case errors.NotFoundImage:
		}
	})

	state := types.ContainerState{ExitCode: 0}
	retContainerInspect := types.ContainerJSON{
		new(types.ContainerJSONBase),
		[]types.MountPoint{},
		new(container.Config),
		new(types.NetworkSettings),
	}
	retContainerInspect.State = &state

	t.Run("ReturnErrorWhenNotFoundContainerInfo", func(t *testing.T) {
		fakeRunContainerList = func() ([]types.Container, error) {
			return retContainers, nil
		}
		fakeRunContaienrInspect = func() (types.ContainerJSON, error) {
			return retContainerInspect, origineErr.New("")
		}
		_, err := Executor.GetContainerStateByName("test_123")
		switch err.(type) {
		default:
			t.Error()
		case errors.NotFoundImage:
		}
	})

	t.Run("GetStatusSuccessful", func(t *testing.T) {
		STATUS := "Status"
		EXITCODE := "ExitCode"

		fakeRunContainerList = func() ([]types.Container, error) {
			return retContainers, nil
		}
		fakeRunContaienrInspect = func() (types.ContainerJSON, error) {
			return retContainerInspect, nil
		}
		inspect, _ := Executor.GetContainerStateByName("test_123")
		if strings.Compare(inspect[STATUS].(string), retContainers[0].State) != 0 ||
			strings.Compare(inspect[EXITCODE].(string), strconv.Itoa(retContainerInspect.State.ExitCode)) != 0 {
			t.Error()
		}
	})
}

func runisContainedName(t *testing.T, source []string, input string, expected bool) {
	ret := isContainedStringInList(source, input)
	if ret != expected {
		t.Errorf("Expect %s, but returned %s", strconv.FormatBool(expected), strconv.FormatBool(ret))
	}
}

func TestIsContainedName(t *testing.T) {
	type testList struct {
		testType string
		expect   bool
		input    string
	}
	tests := [...]testList{{"Negative", false, "getcontiner"}, {"Positive", true, "test"}}

	source := []string{"test", "contained", "container", "name"}

	for _, test := range tests {
		t.Run(test.testType, func(t *testing.T) {
			input := test.input
			runisContainedName(t, source, input, test.expect)
		})
	}
}

func checkError(t *testing.T, err error) {
	switch err {
	case nil:
		t.Error()
	default:
	}
}

func TestComposeFunctionality(t *testing.T) {
	// TODO extending unit tests for compose.
	fakeGetComposeInstanceImpl = func() (project.APIProject, error) {
		return nil, origineErr.New("")
	}

	err := Executor.Create("", "")
	checkError(t, err)
	err = Executor.Down("", "")
	checkError(t, err)
	err = Executor.DownWithRemoveImages("", "")
	checkError(t, err)
	err = Executor.Pause("", "")
	checkError(t, err)
	_, err = Executor.Ps("", "")
	checkError(t, err)
	err = Executor.Pull("", "")
	checkError(t, err)
	err = Executor.Start("", "")
	checkError(t, err)
	err = Executor.Stop("", "")
	checkError(t, err)
	err = Executor.Unpause("", "")
	checkError(t, err)
	err = Executor.Up("", "")
	checkError(t, err)
}

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
	"bytes"
	"commons/errors"
	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"encoding/json"
	origineErr "errors"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type testObj struct {
	received events.Event
	expected Event
}

type tearDown func(t *testing.T)

func setUp(t *testing.T) tearDown {
	//client = nil
	getComposeInstance = fakeGetComposeInstance
	getImageList = fakeImageList
	getContainerList = fakeContainerList
	getContainerInspect = fakeContainerExecInspect
	getImagePull = fakeImagePull
	getImageTag = fakeImageTag
	getContainerStats = fakeContainerStats
	getPs = fakeComposePs
	getPull = fakeComposePull
	getUp = fakeComposeUp

	return func(t *testing.T) {
		client, _ = docker.NewEnvClient()
		getComposeInstance = getComposeInstanceImpl
		getImageList = (*docker.Client).ImageList
		getContainerList = (*docker.Client).ContainerList
		getContainerInspect = (*docker.Client).ContainerInspect
		getImagePull = (*docker.Client).ImagePull
		getImageTag = (*docker.Client).ImageTag
		getContainerStats = (*docker.Client).ContainerStats
		getPs = composePs
		getPull = composePull
		getUp = composeUp
	}
}

var fakeGetComposeInstanceImpl func() (project.APIProject, error)
var fakeRunImageList func() ([]types.ImageSummary, error)
var fakeRunContainerList func() ([]types.Container, error)
var fakeRunContaienrInspect func() (types.ContainerJSON, error)
var fakeRunImagePull func() (io.ReadCloser, error)
var fakeRunImageTag func() error
var fakeRunContainerStats func() (types.ContainerStats, error)
var fakeRunComposePs func() (project.InfoSet, error)
var fakeRunComposePull func() error
var fakeRunComposeUp func() error

func fakeImagePull(*docker.Client, context.Context, string, types.ImagePullOptions) (io.ReadCloser, error) {
	return fakeRunImagePull()
}

func fakeImageTag(*docker.Client, context.Context, string, string) error {
	return fakeRunImageTag()
}

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

func fakeContainerStats(*docker.Client, context.Context, string, bool) (types.ContainerStats, error) {
	return fakeRunContainerStats()
}

func fakeComposePs(instance project.APIProject, ctx context.Context, params ...string) (project.InfoSet, error) {
	return fakeRunComposePs()
}

func fakeComposePull(instance project.APIProject, ctx context.Context, services ...string) error {
	return fakeRunComposePull()
}

func fakeComposeUp(instance project.APIProject, ctx context.Context, opt options.Up, services ...string) error {
	return fakeRunComposeUp()
}

func TestIsDeployEvent(t *testing.T) {
	eventWithExpectedResult := map[events.EventType]bool{
		events.ServicePull:      true,
		events.ContainerCreated: true,
		events.ContainerStarted: true,
		events.ServiceAdd:       false,
	}

	t.Run("ReturnTrueOrFalse", func(t *testing.T) {
		for event, expectedRet := range eventWithExpectedResult {
			ret := isDeployEventType(event)
			if ret != expectedRet {
				t.Errorf("Expected err: %v, actual err: %v", expectedRet, ret)
			}
		}
	})
}

func TestMakeDeployEvent(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	t.Run("Success", func(t *testing.T) {
		testFileName := "test"
		description := make(map[string]interface{})
		description_str := `"services:\n  my-test:\n    image: ubuntu:latest\nversion: \'2\'\n"`
		json.Unmarshal([]byte(description_str), &description)
		yaml, _ := yaml.Marshal(description)
		ioutil.WriteFile(testFileName, yaml, os.FileMode(0755))

		testAppID := "appid"
		testEventID := "eventid"
		testServiceName := "servicename"

		testList := []testObj{
			{
				received: events.Event{events.ServicePull, testServiceName, nil},
				expected: Event{
					ID:          testEventID,
					Type:        IMAGE,
					AppID:       testAppID,
					ServiceName: testServiceName,
					Status:      PULLED,
				},
			},
			{
				received: events.Event{events.ContainerCreated, testServiceName, nil},
				expected: Event{
					ID:          testEventID,
					Type:        CONTAINER,
					AppID:       testAppID,
					ServiceName: testServiceName,
					Status:      CREATED,
					ContainerEvent: ContainerEvent{
						CID: "testcid",
					},
				},
			},
			{
				received: events.Event{events.ContainerStarted, testServiceName, nil},
				expected: Event{
					ID:          testEventID,
					Type:        CONTAINER,
					AppID:       testAppID,
					ServiceName: testServiceName,
					Status:      STARTED,
					ContainerEvent: ContainerEvent{
						CID: "testcid",
					},
				},
			},
		}
		getComposeInstance = getComposeInstanceImpl

		fakeRunComposePs = func() (project.InfoSet, error) {
			var infoset project.InfoSet
			testInfoSet := `[{"Command":"","Id":"testcid","Name":"","Ports":"","State":""}]`
			err := json.Unmarshal([]byte(testInfoSet), &infoset)
			if err != nil {
				return nil, err
			}
			return infoset, nil
		}

		for _, test := range testList {
			ret := makeDeployEvent(testAppID, testFileName, testEventID, test.received)
			if !reflect.DeepEqual(ret, test.expected) {
				t.Errorf("Expected result: %v, Actual result: %v", test.expected, ret)
			}
		}
		os.RemoveAll(testFileName)
	})
}

func TestUpWithEvent(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	t.Run("Success", func(t *testing.T) {
		testAppID := "appid"
		testEventID := "eventid"
		testFileName := "test"
		description := make(map[string]interface{})
		description_str := `"services:\n  my-test:\n    image: ubuntu:latest\nversion: \'2\'\n"`
		json.Unmarshal([]byte(description_str), &description)
		yaml, _ := yaml.Marshal(description)
		ioutil.WriteFile(testFileName, yaml, os.FileMode(0755))

		getComposeInstance = getComposeInstanceImpl

		fakeRunComposePull = func() error {
			return nil
		}
		fakeRunComposeUp = func() error {
			return nil
		}

		evt := make(chan Event)
		err := Executor.UpWithEvent(testAppID, testFileName, testEventID, evt)
		if err != nil {
			t.Errorf("Exepcted err : nil, Actual err : %s", err.Error())
		}
		close(evt)
		os.RemoveAll(testFileName)
	})
}

func TestGetImageIDByRepoDigest(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	testID := "abcd"
	testRepoDigest := "test@sha256"

	t.Run("ReturnErrorWhenReceiveErrorFromDockerEngine", func(t *testing.T) {
		fakeRunImageList = func() ([]types.ImageSummary, error) {
			return nil, origineErr.New("")
		}
		_, err := Executor.GetImageIDByRepoDigest(testRepoDigest)
		switch err.(type) {
		default:
			t.Error()
		case errors.Unknown:
		}
	})

	ret := []types.ImageSummary{
		{
			ID:          testID,
			RepoDigests: []string{"wrong"},
		}}

	t.Run("ReturnErrorWhenNotFoundImageInList", func(t *testing.T) {
		fakeRunImageList = func() ([]types.ImageSummary, error) {
			return ret, nil
		}
		_, err := Executor.GetImageIDByRepoDigest(testRepoDigest)
		switch err.(type) {
		default:
			t.Error()
		case errors.NotFoundImage:
		}
	})

	expected := testID
	t.Run("GetIDSuccessful", func(t *testing.T) {
		ret[0].RepoDigests[0] = testRepoDigest
		fakeRunImageList = func() ([]types.ImageSummary, error) {
			return ret, nil
		}
		id, _ := Executor.GetImageIDByRepoDigest(testRepoDigest)
		if strings.Compare(id, expected) != 0 {
			t.Error()
		}
	})
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

func TestGetAppStats(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	testFileName := "test"
	description := make(map[string]interface{})
	description_str := `"services:\n  my-test:\n    image: google/cadvisor:latest\nversion: \'2\'\n"`
	json.Unmarshal([]byte(description_str), &description)
	yaml, _ := yaml.Marshal(description)
	ioutil.WriteFile(testFileName, yaml, os.FileMode(0755))

	t.Run("GetContainerStats_ExpectReturnError", func(t *testing.T) {
		getComposeInstance = getComposeInstanceImpl
		fakeRunComposePs = func() (project.InfoSet, error) {
			var infoset project.InfoSet

			testInfoSet := `[{"Command":"/usr/bin/cadvisor -logtostderr","Id":"9081b5c76879096799265c62848e6ea798d107c689632229cfa63d4110849a4e","Name":"2277a03208c65ce497de317bd19015fd2a8fba15_one_1","Ports":"8080/tcp","State":"Up 2 hours"},{"Command":"/usr/bin/cadvisor","Id":"e176ed709b89322909d1eb4771cb8548d37dcf5932dde4bc4240706c1f350376","Name":"2277a03208c65ce497de317bd19015fd2a8fba15_two_1","Ports":"8080/tcp","State":"Up 2 hours"}]`
			err := json.Unmarshal([]byte(testInfoSet), &infoset)
			if err != nil {
				return nil, err
			}
			return infoset, nil
		}

		fakeRunContainerList = func() ([]types.Container, error) {
			var containers []types.Container

			testContainers := `[{"Id":"71a3a3c09149e8081c352bba7b62119bd06650f0e8d88ab24b5dd3cf922bd76a","Names":["/2277a03208c65ce497de317bd19015fd2a8fba15_two_1"],"Image":"google/cadvisor:0.1.0","ImageID":"sha256:88381b5edb12821b9098349171fbf885cb3a7de5d9045646a2ea258af28785e2","Command":"/usr/bin/cadvisor","Created":1521533754,"Ports":[{"PrivatePort":8080,"Type":"tcp"}],"Labels":{"com.docker.compose.config-hash":"a84a152bfaae6ed168b8dfc1d1ebaa82ba68ac5e","com.docker.compose.container-number":"1","com.docker.compose.oneoff":"False","com.docker.compose.project":"2277a03208c65ce497de317bd19015fd2a8fba15","com.docker.compose.service":"two","com.docker.compose.version":"1.5.0"},"State":"running","Status":"Up 13 minutes","HostConfig":{"NetworkMode":"2277a03208c65ce497de317bd19015fd2a8fba15_default"},"NetworkSettings":{"Networks":{"2277a03208c65ce497de317bd19015fd2a8fba15_default":{"IPAMConfig":null,"Links":null,"Aliases":null,"NetworkID":"60f931661fdf4fc75fcb2f3e94938693f1d81f5f65f121e5db58b96bde517bb3","EndpointID":"eddd972ce07853ab14ea11ee5bc92abae14e8130ddfd51d8233f5484b96d54b8","Gateway":"172.28.0.1","IPAddress":"172.28.0.3","IPPrefixLen":16,"IPv6Gateway":"","GlobalIPv6Address":"","GlobalIPv6PrefixLen":0,"MacAddress":"02:42:ac:1c:00:03","DriverOpts":null}}},"Mounts":[]},{"Id":"ab7abf724ed2251ecd9bf62eda091b4781907ce9cbdcf4c5bc39a9849138ab41","Names":["/2277a03208c65ce497de317bd19015fd2a8fba15_one_1"],"Image":"google/cadvisor:latest","ImageID":"sha256:75f88e3ec333cbb410297e4f40297ac615e076b4a50aeeae49f287093ff01ab1","Command":"/usr/bin/cadvisor -logtostderr","Created":1521533754,"Ports":[{"PrivatePort":8080,"Type":"tcp"}],"Labels":{"com.docker.compose.config-hash":"1e76922e07f43ff504ca62fd539be44b52cb204f","com.docker.compose.container-number":"1","com.docker.compose.oneoff":"False","com.docker.compose.project":"2277a03208c65ce497de317bd19015fd2a8fba15","com.docker.compose.service":"one","com.docker.compose.version":"1.5.0"},"State":"running","Status":"Up 13 minutes","HostConfig":{"NetworkMode":"2277a03208c65ce497de317bd19015fd2a8fba15_default"},"NetworkSettings":{"Networks":{"2277a03208c65ce497de317bd19015fd2a8fba15_default":{"IPAMConfig":null,"Links":null,"Aliases":null,"NetworkID":"60f931661fdf4fc75fcb2f3e94938693f1d81f5f65f121e5db58b96bde517bb3","EndpointID":"ae13a3fa33695c630057bc8d3f8fda15066fec8fce0ff006cd1b7f2da8c37ccb","Gateway":"172.28.0.1","IPAddress":"172.28.0.2","IPPrefixLen":16,"IPv6Gateway":"","GlobalIPv6Address":"","GlobalIPv6PrefixLen":0,"MacAddress":"02:42:ac:1c:00:02","DriverOpts":null}}},"Mounts":[]}]`
			err := json.Unmarshal([]byte(testContainers), &containers)
			if err != nil {
				return nil, err
			}
			return containers, nil
		}

		fakeRunContainerStats = func() (types.ContainerStats, error) {
			var stats types.ContainerStats
			stats.Body = nil
			return stats, errors.Unknown{}
		}

		_, err := Executor.GetAppStats("test", testFileName)
		switch err.(type) {
		default:
			t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
		case errors.Unknown:
		}
	})

	t.Run("GetContainerListError_ExpectReturnError", func(t *testing.T) {
		getComposeInstance = getComposeInstanceImpl
		fakeRunComposePs = func() (project.InfoSet, error) {
			var infoset project.InfoSet

			testInfoSet := `[{"Command":"/usr/bin/cadvisor -logtostderr","Id":"9081b5c76879096799265c62848e6ea798d107c689632229cfa63d4110849a4e","Name":"2277a03208c65ce497de317bd19015fd2a8fba15_one_1","Ports":"8080/tcp","State":"Up 2 hours"},{"Command":"/usr/bin/cadvisor","Id":"e176ed709b89322909d1eb4771cb8548d37dcf5932dde4bc4240706c1f350376","Name":"2277a03208c65ce497de317bd19015fd2a8fba15_two_1","Ports":"8080/tcp","State":"Up 2 hours"}]`
			err := json.Unmarshal([]byte(testInfoSet), &infoset)
			if err != nil {
				return nil, err
			}
			return infoset, nil
		}

		fakeRunContainerList = func() ([]types.Container, error) {
			return nil, errors.Unknown{}
		}
		_, err := Executor.GetAppStats("test", testFileName)
		switch err.(type) {
		default:
			t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
		case errors.Unknown:
		}
	})

	t.Run("PsError_ExpectReturnError", func(t *testing.T) {
		getComposeInstance = getComposeInstanceImpl
		fakeRunComposePs = func() (project.InfoSet, error) {
			return nil, errors.Unknown{}
		}

		_, err := Executor.GetAppStats("test", testFileName)
		switch err.(type) {
		default:
			t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
		case errors.Unknown:
		}
	})

	t.Run("Success", func(t *testing.T) {
		getComposeInstance = getComposeInstanceImpl
		fakeRunComposePs = func() (project.InfoSet, error) {
			var infoset project.InfoSet

			testInfoSet := `[{"Command":"/usr/bin/cadvisor -logtostderr","Id":"9081b5c76879096799265c62848e6ea798d107c689632229cfa63d4110849a4e","Name":"2277a03208c65ce497de317bd19015fd2a8fba15_one_1","Ports":"8080/tcp","State":"Up 2 hours"},{"Command":"/usr/bin/cadvisor","Id":"e176ed709b89322909d1eb4771cb8548d37dcf5932dde4bc4240706c1f350376","Name":"2277a03208c65ce497de317bd19015fd2a8fba15_two_1","Ports":"8080/tcp","State":"Up 2 hours"}]`
			err := json.Unmarshal([]byte(testInfoSet), &infoset)
			if err != nil {
				return nil, err
			}
			return infoset, nil
		}

		fakeRunContainerList = func() ([]types.Container, error) {
			var containers []types.Container

			testContainers := `[{"Id":"71a3a3c09149e8081c352bba7b62119bd06650f0e8d88ab24b5dd3cf922bd76a","Names":["/2277a03208c65ce497de317bd19015fd2a8fba15_two_1"],"Image":"google/cadvisor:0.1.0","ImageID":"sha256:88381b5edb12821b9098349171fbf885cb3a7de5d9045646a2ea258af28785e2","Command":"/usr/bin/cadvisor","Created":1521533754,"Ports":[{"PrivatePort":8080,"Type":"tcp"}],"Labels":{"com.docker.compose.config-hash":"a84a152bfaae6ed168b8dfc1d1ebaa82ba68ac5e","com.docker.compose.container-number":"1","com.docker.compose.oneoff":"False","com.docker.compose.project":"2277a03208c65ce497de317bd19015fd2a8fba15","com.docker.compose.service":"two","com.docker.compose.version":"1.5.0"},"State":"running","Status":"Up 13 minutes","HostConfig":{"NetworkMode":"2277a03208c65ce497de317bd19015fd2a8fba15_default"},"NetworkSettings":{"Networks":{"2277a03208c65ce497de317bd19015fd2a8fba15_default":{"IPAMConfig":null,"Links":null,"Aliases":null,"NetworkID":"60f931661fdf4fc75fcb2f3e94938693f1d81f5f65f121e5db58b96bde517bb3","EndpointID":"eddd972ce07853ab14ea11ee5bc92abae14e8130ddfd51d8233f5484b96d54b8","Gateway":"172.28.0.1","IPAddress":"172.28.0.3","IPPrefixLen":16,"IPv6Gateway":"","GlobalIPv6Address":"","GlobalIPv6PrefixLen":0,"MacAddress":"02:42:ac:1c:00:03","DriverOpts":null}}},"Mounts":[]},{"Id":"ab7abf724ed2251ecd9bf62eda091b4781907ce9cbdcf4c5bc39a9849138ab41","Names":["/2277a03208c65ce497de317bd19015fd2a8fba15_one_1"],"Image":"google/cadvisor:latest","ImageID":"sha256:75f88e3ec333cbb410297e4f40297ac615e076b4a50aeeae49f287093ff01ab1","Command":"/usr/bin/cadvisor -logtostderr","Created":1521533754,"Ports":[{"PrivatePort":8080,"Type":"tcp"}],"Labels":{"com.docker.compose.config-hash":"1e76922e07f43ff504ca62fd539be44b52cb204f","com.docker.compose.container-number":"1","com.docker.compose.oneoff":"False","com.docker.compose.project":"2277a03208c65ce497de317bd19015fd2a8fba15","com.docker.compose.service":"one","com.docker.compose.version":"1.5.0"},"State":"running","Status":"Up 13 minutes","HostConfig":{"NetworkMode":"2277a03208c65ce497de317bd19015fd2a8fba15_default"},"NetworkSettings":{"Networks":{"2277a03208c65ce497de317bd19015fd2a8fba15_default":{"IPAMConfig":null,"Links":null,"Aliases":null,"NetworkID":"60f931661fdf4fc75fcb2f3e94938693f1d81f5f65f121e5db58b96bde517bb3","EndpointID":"ae13a3fa33695c630057bc8d3f8fda15066fec8fce0ff006cd1b7f2da8c37ccb","Gateway":"172.28.0.1","IPAddress":"172.28.0.2","IPPrefixLen":16,"IPv6Gateway":"","GlobalIPv6Address":"","GlobalIPv6PrefixLen":0,"MacAddress":"02:42:ac:1c:00:02","DriverOpts":null}}},"Mounts":[]}]`
			err := json.Unmarshal([]byte(testContainers), &containers)
			if err != nil {
				return nil, err
			}
			return containers, nil
		}

		fakeRunContainerStats = func() (types.ContainerStats, error) {
			var stats types.ContainerStats

			testStats := `{"read":"2018-03-20T09:00:45.108700589Z","preread":"2018-03-20T09:00:44.108937037Z","pids_stats":{"current":9},"blkio_stats":{"io_service_bytes_recursive":[{"major":8,"minor":16,"op":"Read","value":8572928},{"major":8,"minor":16,"op":"Write","value":0},{"major":8,"minor":16,"op":"Sync","value":8572928},{"major":8,"minor":16,"op":"Async","value":0},{"major":8,"minor":16,"op":"Total","value":8572928},{"major":8,"minor":16,"op":"Read","value":8572928},{"major":8,"minor":16,"op":"Write","value":0},{"major":8,"minor":16,"op":"Sync","value":8572928},{"major":8,"minor":16,"op":"Async","value":0},{"major":8,"minor":16,"op":"Total","value":8572928}],"io_serviced_recursive":[{"major":8,"minor":16,"op":"Read","value":343},{"major":8,"minor":16,"op":"Write","value":0},{"major":8,"minor":16,"op":"Sync","value":343},{"major":8,"minor":16,"op":"Async","value":0},{"major":8,"minor":16,"op":"Total","value":343},{"major":8,"minor":16,"op":"Read","value":343},{"major":8,"minor":16,"op":"Write","value":0},{"major":8,"minor":16,"op":"Sync","value":343},{"major":8,"minor":16,"op":"Async","value":0},{"major":8,"minor":16,"op":"Total","value":343}],"io_queue_recursive":[{"major":8,"minor":16,"op":"Read","value":0},{"major":8,"minor":16,"op":"Write","value":0},{"major":8,"minor":16,"op":"Sync","value":0},{"major":8,"minor":16,"op":"Async","value":0},{"major":8,"minor":16,"op":"Total","value":0}],"io_service_time_recursive":[{"major":8,"minor":16,"op":"Read","value":115378760},{"major":8,"minor":16,"op":"Write","value":0},{"major":8,"minor":16,"op":"Sync","value":115378760},{"major":8,"minor":16,"op":"Async","value":0},{"major":8,"minor":16,"op":"Total","value":115378760}],"io_wait_time_recursive":[{"major":8,"minor":16,"op":"Read","value":28990215},{"major":8,"minor":16,"op":"Write","value":0},{"major":8,"minor":16,"op":"Sync","value":28990215},{"major":8,"minor":16,"op":"Async","value":0},{"major":8,"minor":16,"op":"Total","value":28990215}],"io_merged_recursive":[{"major":8,"minor":16,"op":"Read","value":0},{"major":8,"minor":16,"op":"Write","value":0},{"major":8,"minor":16,"op":"Sync","value":0},{"major":8,"minor":16,"op":"Async","value":0},{"major":8,"minor":16,"op":"Total","value":0}],"io_time_recursive":[{"major":8,"minor":16,"op":"","value":536060268}],"sectors_recursive":[{"major":8,"minor":16,"op":"","value":16744}]},"num_procs":0,"storage_stats":{},"cpu_stats":{"cpu_usage":{"total_usage":38900043200,"percpu_usage":[4969423189,4858004428,5210404191,4895648649,4474746326,4750894971,5001950610,4738970836],"usage_in_kernelmode":9640000000,"usage_in_usermode":21130000000},"system_cpu_usage":4893544710000000,"online_cpus":8,"throttling_data":{"periods":0,"throttled_periods":0,"throttled_time":0}},"precpu_stats":{"cpu_usage":{"total_usage":38890053594,"percpu_usage":[4969371267,4857320495,5210176736,4895157812,4471815759,4750602150,5001950610,4733658765],"usage_in_kernelmode":9640000000,"usage_in_usermode":21120000000},"system_cpu_usage":4893536750000000,"online_cpus":8,"throttling_data":{"periods":0,"throttled_periods":0,"throttled_time":0}},"memory_stats":{"usage":15187968,"max_usage":16859136,"stats":{"active_anon":1449984,"active_file":6008832,"cache":8282112,"dirty":0,"hierarchical_memory_limit":9223372036854771712,"hierarchical_memsw_limit":0,"inactive_anon":1642496,"inactive_file":2273280,"mapped_file":2564096,"pgfault":1740657,"pgmajfault":139,"pgpgin":667951,"pgpgout":665168,"rss":3117056,"rss_huge":0,"total_active_anon":1449984,"total_active_file":6008832,"total_cache":8282112,"total_dirty":0,"total_inactive_anon":1642496,"total_inactive_file":2273280,"total_mapped_file":2564096,"total_pgfault":1740657,"total_pgmajfault":139,"total_pgpgin":667951,"total_pgpgout":665168,"total_rss":3117056,"total_rss_huge":0,"total_unevictable":0,"total_writeback":0,"unevictable":0,"writeback":0},"limit":8317444096},"name":"/2277a03208c65ce497de317bd19015fd2a8fba15_two_1","id":"71a3a3c09149e8081c352bba7b62119bd06650f0e8d88ab24b5dd3cf922bd76a","networks":{"eth0":{"rx_bytes":93156,"rx_packets":739,"rx_errors":0,"rx_dropped":0,"tx_bytes":0,"tx_packets":0,"tx_errors":0,"tx_dropped":0}}}`
			stats.Body = ioutil.NopCloser(bytes.NewReader([]byte(testStats)))
			return stats, nil
		}

		_, err := Executor.GetAppStats("test", testFileName)
		if err != nil {
			t.Error("Expected nil error but error occured")
		}
	})
	os.Remove(testFileName)
}

func TestGetContainerConfigByName(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown(t)

	t.Run("ReturnErrorWhenReceiveErrorFromDockerEngine", func(t *testing.T) {
		fakeRunContainerList = func() ([]types.Container, error) {
			return nil, origineErr.New("")
		}
		_, err := Executor.GetContainerConfigByName("123")
		switch err.(type) {
		default:
			t.Error()
		case errors.Unknown:
		}
	})

	retContainers := []types.Container{
		{
			ID: "containerId",
			Ports: []types.Port{
				{
					IP:          "testIP",
					PrivatePort: 1234,
					PublicPort:  1234,
					Type:        "testType",
				},
			},
			State: "running",
			Names: []string{"/test_latest", "/test_111", "/test_123"},
		},
	}

	t.Run("ReturnErrorWhenNotFoundContainerInList", func(t *testing.T) {
		fakeRunContainerList = func() ([]types.Container, error) {
			return retContainers, nil
		}
		_, err := Executor.GetContainerConfigByName("123")
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
		_, err := Executor.GetContainerConfigByName("test_123")
		switch err.(type) {
		default:
			t.Error()
		case errors.NotFoundImage:
		}
	})

	t.Run("GetStatusSuccessful", func(t *testing.T) {
		STATUS := "status"
		EXITCODE := "exitcode"

		fakeRunContainerList = func() ([]types.Container, error) {
			return retContainers, nil
		}
		fakeRunContaienrInspect = func() (types.ContainerJSON, error) {
			return retContainerInspect, nil
		}
		inspect, _ := Executor.GetContainerConfigByName("test_123")
		if strings.Compare(inspect[STATUS].(string), retContainers[0].State) != 0 ||
			strings.Compare(inspect[EXITCODE].(string), strconv.Itoa(retContainerInspect.State.ExitCode)) != 0 {
			t.Error()
		}
	})
}

func TestCalcNetworkIO(t *testing.T) {
	var network map[string]types.NetworkStats = map[string]types.NetworkStats{
		"one": types.NetworkStats{
			RxBytes: 1,
			TxBytes: 2,
		},
		"two": types.NetworkStats{
			RxBytes: 10,
			TxBytes: 11,
		},
	}
	rx, tx := calcNetworkIO(network)
	if rx != 11.0 || tx != 13.0 {
		t.Errorf("Expected rx : 11, tx : 13, Actual rx : %f, tx : %f", rx, tx)
	}
}

func TestConvertToHumanReadableBinaryUnit(t *testing.T) {
	t.Run("ConvertToHumanReadableBinrayUnit_ReturnBSuccessful", func(t *testing.T) {
		res := convertToHumanReadableBinaryUnit(1023.0)
		if res != "1023.000B" {
			t.Errorf("Expected result : 1023B, Actual Result : %s", res)
		}
	})

	t.Run("ConvertToHumanReadableBinaryUnit_ReturnKiBSuccessful", func(t *testing.T) {
		res := convertToHumanReadableBinaryUnit(2.0 * 1024.0)
		if res != "2.000KiB" {
			t.Errorf("Expected result : 2.000KiB, Actual Result : %s", res)
		}
	})

	t.Run("ConvertToHumanReadableBinrayUnit_ReturnMiBSuccessful", func(t *testing.T) {
		res := convertToHumanReadableBinaryUnit(2.0 * 1024.0 * 1024.0)
		if res != "2.000MiB" {
			t.Errorf("Expected result : 2.000MiB, Actual Result : %s", res)
		}
	})

	t.Run("ConvertToHumanReadableBinaryUnit_ReturnGiBSuccessful", func(t *testing.T) {
		res := convertToHumanReadableBinaryUnit(2.0 * 1024.0 * 1024.0 * 1024.0)
		if res != "2.000GiB" {
			t.Errorf("Expected result : 2.000GiB, Actual Result : %s", res)
		}
	})
}

func TestConvertToHumanReadableUnit(t *testing.T) {
	t.Run("ConvertToHumanReadableUnit_ReturnBSuccessful", func(t *testing.T) {
		res := convertToHumanReadableUnit(99.0)
		if res != "99.000B" {
			t.Errorf("Expected result : 99.000B, Actual Result : %s", res)
		}
	})

	t.Run("ConvertToHumanReadableUnit_ReturnKBSuccessful", func(t *testing.T) {
		res := convertToHumanReadableUnit(2.0 * 1000.0)
		if res != "2.000KB" {
			t.Errorf("Expected result : 2.000KB, Actual Result : %s", res)
		}
	})

	t.Run("ConvertToHumanReadableUnit_ReturnMBSuccessful", func(t *testing.T) {
		res := convertToHumanReadableUnit(2.0 * 1000.0 * 1000.0)
		if res != "2.000MB" {
			t.Errorf("Expected result : 2.000MB, Actual Result : %s", res)
		}
	})

	t.Run("ConvertToHumanReadableUnit_ReturnGBSuccessful", func(t *testing.T) {
		res := convertToHumanReadableUnit(2.0 * 1000.0 * 1000.0 * 1000.0)
		if res != "2.000GB" {
			t.Errorf("Expected result : 2.000GB, Actual Result : %s", res)
		}
	})
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
	err = Executor.Up("", "", true)
	checkError(t, err)
}

func TestDockerFunctionality(t *testing.T) {
	// TODO extending unit tests for docker.
	fakeRunImagePull = func() (io.ReadCloser, error) {
		return nil, origineErr.New("")
	}
	err := Executor.ImagePull("")
	checkError(t, err)

	fakeRunImageTag = func() error {
		return origineErr.New("")
	}
	err = Executor.ImageTag("", "")
	checkError(t, err)
}

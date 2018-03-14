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
	"bytes"
	"commons/errors"
	"commons/logger"
	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"golang.org/x/net/context"
	"io"
	"strconv"
	"strings"

	dockercompose "github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	events "github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
)

type Event struct {
	Service string
	Event   string
}

type Command interface {
	Create(id, path string) error
	Up(id, path string, services ...string) error
	Down(id, path string) error
	DownWithRemoveImages(id, path string) error
	Start(id, path string) error
	Stop(id, path string) error
	Pause(id, path string) error
	Unpause(id, path string) error
	Pull(id, path string, services ...string) error
	Ps(id, path string, args ...string) ([]map[string]string, error)
	GetContainerConfigByName(containerName string) (map[string]interface{}, error)
	GetImageDigestByName(imageName string) (string, error)
	GetImageIDByRepoDigest(imageName string) (string, error)
	ImagePull(image string) error
	ImageTag(imageID string, repoTags string) error
	Events(id, path string, evt chan Event, services ...string) error
}

const (
	CID      string = "cid"
	PORTS    string = "ports"
	STATUS   string = "status"
	EXITCODE string = "exitcode"
)

var Executor dockerExecutorImpl

type dockerExecutorImpl struct{}

var client *docker.Client

type typeGetImageList func(*docker.Client, context.Context, types.ImageListOptions) ([]types.ImageSummary, error)
type typeGetContainerList func(*docker.Client, context.Context, types.ContainerListOptions) ([]types.Container, error)
type typeGetContainerInspect func(*docker.Client, context.Context, string) (types.ContainerJSON, error)
type typeGetImagePull func(*docker.Client, context.Context, string, types.ImagePullOptions) (io.ReadCloser, error)
type typeGetImageTag func(*docker.Client, context.Context, string, string) error

var getImageList typeGetImageList
var getContainerList typeGetContainerList
var getContainerInspect typeGetContainerInspect
var getImagePull typeGetImagePull
var getImageTag typeGetImageTag

type createType func(*project.APIProject, context.Context, options.Create, ...string) error

var getComposeInstance func(string, string) (project.APIProject, error)
var create createType

var evts map[string]chan events.ContainerEvent

func init() {
	evts = make(map[string]chan events.ContainerEvent, 0)

	getComposeInstance = getComposeInstanceImpl

	client, _ = docker.NewEnvClient()
	getImageList = (*docker.Client).ImageList
	getContainerList = (*docker.Client).ContainerList
	getContainerInspect = (*docker.Client).ContainerInspect
	getImagePull = (*docker.Client).ImagePull
	getImageTag = (*docker.Client).ImageTag
}

// Creating containers of service list in the yaml description.
// if succeed to create, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Create(id, path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	compose, err := getComposeInstance(id, path)
	if err != nil {
		return err
	}

	return compose.Create(context.Background(), options.Create{ForceRecreate: true})
}

// Pulling images and creating containers and start containers
// of service list in the yaml description.
// if succeed to up, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Up(id, path string, services ...string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	compose, err := getComposeInstance(id, path)
	if err != nil {
		return err
	}
	return compose.Up(context.Background(), options.Up{Create: options.Create{ForceRecreate: true}}, services...)
}

// Stop and remove containers of service list in the yaml description.
// if succeed to down, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Down(id, path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	compose, err := getComposeInstance(id, path)
	if err != nil {
		return err
	}
	return compose.Down(context.Background(), options.Down{})
}

// Stop and remove containers, remove images of service list
// in the yaml description.
// if succeed to down with rmi option, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) DownWithRemoveImages(id, path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	compose, err := getComposeInstance(id, path)
	if err != nil {
		return err
	}
	return compose.Down(context.Background(), options.Down{RemoveImages: "all"})
}

// Starting containers of service list in the yaml description.
// if succeed to start, return error as nil
// otherwise, return error. (if contianers is not created, return error)
func (dockerExecutorImpl) Start(id, path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	compose, err := getComposeInstance(id, path)
	if err != nil {
		return err
	}
	return compose.Start(context.Background())
}

// Stopping containers of service list in the yaml description.
// if succeed to stop, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Stop(id, path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	compose, err := getComposeInstance(id, path)
	if err != nil {
		return err
	}
	return compose.Stop(context.Background(), 10)
}

// Pause containers of service list in the yaml description.
// if succeed to pause, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Pause(id, path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	compose, err := getComposeInstance(id, path)
	if err != nil {
		return err
	}
	return compose.Pause(context.Background())
}

// Resume paused containers of service list in the yaml description.
// if succeed to resume, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Unpause(id, path string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	compose, err := getComposeInstance(id, path)
	if err != nil {
		return err
	}
	return compose.Unpause(context.Background())
}

// Pulling images of service list in the yaml description.
// if succeed to pull, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Pull(id, path string, services ...string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	compose, err := getComposeInstance(id, path)
	if err != nil {
		return err
	}
	return compose.Pull(context.Background(), services...)
}

// Pulling an image
// if succeed to pull, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) ImagePull(image string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	rc, err := getImagePull(client, context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		logger.Logging(logger.DEBUG, err.Error())
		return err
	}
	var buf1 bytes.Buffer
	io.Copy(&buf1, rc)
	return err
}

// Tagging an image with repoTags
// if succeed to tag, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) ImageTag(imageID string, repoTags string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	err := getImageTag(client, context.Background(), imageID, repoTags)
	if err != nil {
		logger.Logging(logger.DEBUG, err.Error())
		return err
	}
	return nil
}

// Getting container informations of service list in the yaml description.
// (e.g. container name, command, state, port)
// if succeed to get, return error as nil
// otherwise, return error.
func (dockerExecutorImpl) Ps(id, path string, args ...string) ([]map[string]string, error) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	compose, err := getComposeInstance(id, path)
	if err != nil {
		return nil, err
	}
	infos, retErr := compose.Ps(context.Background(), args...)
	retMap := make([]map[string]string, len(infos))

	for idx, info := range infos {
		retMap[idx] = make(map[string]string)
		for key, value := range info {
			retMap[idx][key] = value
		}
	}
	return retMap, retErr
}

// Getting container config in the docker engine by container name.
// if succeed to get, return state and exit code of container,
// othewise, return error.
func (d dockerExecutorImpl) GetContainerConfigByName(containerName string) (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	containers, err := getContainerList(client, context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		logger.Logging(logger.ERROR)
		return nil, errors.Unknown{Msg: "fail to get the container list from docker engine"}
	}

	for _, container := range containers {
		target_str := "/" + containerName
		if isContainedStringInList(container.Names, target_str) {
			ins, err := getContainerInspect(client, context.Background(), container.ID)
			if err != nil {
				logger.Logging(logger.ERROR, err.Error())
				continue
			}

			ret := make(map[string]interface{})
			ret[CID] = container.ID
			ret[PORTS] = container.Ports
			ret[STATUS] = container.State
			ret[EXITCODE] = strconv.Itoa(ins.State.ExitCode)

			return ret, nil
		}
	}
	return nil, errors.NotFoundImage{Msg: "can not found container"}
}

// Getting image digest in the docker engine by image name.
// if succeed to get, return digest of image,
// othewise, return error.
func (d dockerExecutorImpl) GetImageDigestByName(imageName string) (string, error) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	images, err := getImageList(client, context.Background(), types.ImageListOptions{})
	if err != nil {
		logger.Logging(logger.ERROR, "fail to get the image list from docker engine")
		return "", errors.Unknown{Msg: "fail to get the image list from docker engine"}
	}

	for _, image := range images {
		if isContainedStringInList(image.RepoTags, imageName) &&
			image.RepoDigests != nil && len(image.RepoDigests[0]) != 0 {
			return image.RepoDigests[0], nil
		}
	}
	return "", errors.NotFoundImage{Msg: "can not found image"}
}

func (dockerExecutorImpl) GetImageIDByRepoDigest(repoDigest string) (string, error) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	images, err := getImageList(client, context.Background(), types.ImageListOptions{})
	if err != nil {
		logger.Logging(logger.ERROR, "fail to get the image list from docker engine")
		return "", errors.Unknown{Msg: "fail to get the image list from docker engine"}
	}

	for _, image := range images {
		if isContainedStringInList(image.RepoDigests, repoDigest) && len(image.RepoTags) == 0 {
			return image.ID, nil
		}
	}
	return "", errors.NotFoundImage{Msg: "can not found image"}
}

func (dockerExecutorImpl) Events(id, path string, evt chan Event, services ...string) error {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if evt == nil {
		e := events.ContainerEvent{}
		evts[id] <- e
		delete(evts, id)
		return nil
	}

	compose, err := getComposeInstance(id, path)
	if err != nil {
		return err
	}

	ctx, cancelFun := context.WithCancel(context.Background())
	containerEvents, err := compose.Events(ctx)
	if err != nil {
		return err
	}
	evts[id] = containerEvents

	go func() {
		for {
			for event := range containerEvents {
				if _, exists := evts[id]; !exists {
					cancelFun()
					close(containerEvents)
					return
				}
				e := Event{
					Service: event.Service,
					Event:   event.Event,
				}
				evt <- e
			}
		}
	}()

	return nil
}

func isContainedStringInList(list []string, name string) bool {
	for _, str := range list {
		if strings.Compare(str, name) == 0 {
			return true
		}
	}
	return false
}

func getComposeInstanceImpl(id, path string) (project.APIProject, error) {
	return dockercompose.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeFiles: []string{path},
			ProjectName:  id,
		},
	}, nil)
}

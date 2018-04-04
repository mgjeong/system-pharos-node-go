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
	"commons/util"
	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"encoding/json"
	"fmt"
	dockercompose "github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	events "github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
	"golang.org/x/net/context"
	"io"
	"strconv"
	"strings"
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
	GetAppStats(id, path string) ([]map[string]interface{}, error)
	GetContainerConfigByName(containerName string) (map[string]interface{}, error)
	GetImageDigestByName(imageName string) (string, error)
	GetImageIDByRepoDigest(imageName string) (string, error)
	ImagePull(image string) error
	ImageTag(imageID string, repoTags string) error
	Events(id, path string, evt chan Event, services ...string) error
}

const (
	CID           string = "cid"
	PORTS         string = "ports"
	STATUS        string = "status"
	EXITCODE      string = "exitcode"
	CNAME         string = "cname"
	CPU           string = "cpu"
	MEM           string = "mem"
	MEMUSAGE      string = "memusage"
	MEMLIMIT      string = "memlimit"
	BLOCKINPUT    string = "blockinput"
	BLOCKOUTPUT   string = "blockoutput"
	NETWORKINPUT  string = "networkinput"
	NETWORKOUTPUT string = "networkoutput"
	PIDS          string = "pids"
)

var Executor dockerExecutorImpl

type dockerExecutorImpl struct{}

var client *docker.Client

var getImageList func(*docker.Client, context.Context, types.ImageListOptions) ([]types.ImageSummary, error)
var getImagePull func(*docker.Client, context.Context, string, types.ImagePullOptions) (io.ReadCloser, error)
var getImageTag func(*docker.Client, context.Context, string, string) error
var getContainerList func(*docker.Client, context.Context, types.ContainerListOptions) ([]types.Container, error)
var getContainerInspect func(*docker.Client, context.Context, string) (types.ContainerJSON, error)
var getContainerStats func(*docker.Client, context.Context, string, bool) (types.ContainerStats, error)
var getPs func(instance project.APIProject, ctx context.Context, params ...string) (project.InfoSet, error)
var getComposeInstance func(string, string) (project.APIProject, error)

//type createType func(*project.APIProject, context.Context, options.Create, ...string) error

//var create createType

var evts map[string]chan events.ContainerEvent

func dockerPs(instance project.APIProject, ctx context.Context, params ...string) (project.InfoSet, error) {
	return instance.Ps(ctx, params...)
}

func init() {
	evts = make(map[string]chan events.ContainerEvent, 0)

	getComposeInstance = getComposeInstanceImpl

	client, _ = docker.NewEnvClient()
	getImageList = (*docker.Client).ImageList
	getContainerList = (*docker.Client).ContainerList
	getContainerInspect = (*docker.Client).ContainerInspect
	getImagePull = (*docker.Client).ImagePull
	getImageTag = (*docker.Client).ImageTag
	getContainerStats = (*docker.Client).ContainerStats
	getPs = dockerPs
}

func (dockerExecutorImpl) GetAppStats(id, path string) ([]map[string]interface{}, error) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	compose, err := getComposeInstance(id, path)
	if err != nil {
		return nil, err
	}

	appContainers, err := getPs(compose, context.Background())
	if err != nil {
		logger.Logging(logger.ERROR, "fail to execute dockercompose ps")
		return nil, errors.Unknown{Msg: "fail to execute dockercompose ps"}
	}

	appContainersNames := make([]string, 0)
	for _, appContainer := range appContainers {
		appContainersNames = append(appContainersNames, "/"+appContainer["Name"])
	}

	containers, err := getContainerList(client, context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		logger.Logging(logger.ERROR)
		return nil, errors.Unknown{Msg: "fail to get the container list from docker engine"}
	}

	result := make([]map[string]interface{}, 0)
	for _, container := range containers {
		if util.IsContainedStringInList(appContainersNames, container.Names[0]) {
			cStats, err := getContainerStats(client, context.Background(), container.ID, false)
			if err != nil {
				logger.Logging(logger.ERROR, err.Error())
				return nil, errors.Unknown{Msg: "fail to get ContainerStats from docker engine"}
			}

			decoder := json.NewDecoder(cStats.Body)

			var statsJSON *types.StatsJSON
			err = decoder.Decode(&statsJSON)
			if err != nil {
				logger.Logging(logger.ERROR)
				return nil, errors.Unknown{Msg: "fail to decode types.StatsJSON"}
			}
			cpuPercent := calcCPUPercent(statsJSON)
			memPercent := 0.0
			memUsage := float64(statsJSON.MemoryStats.Usage)
			memLimit := float64(statsJSON.MemoryStats.Limit)
			memPercent = memUsage / memLimit * 100.0
			bi, bo := calcBlockIO(statsJSON.BlkioStats)
			ni, no := calcNetworkIO(statsJSON.Networks)

			stats := make(map[string]interface{})
			stats[CID] = container.ID
			stats[CNAME] = strings.Replace(container.Names[0], "/", "", -1)
			stats[CPU] = fmt.Sprintf("%.3f", cpuPercent) + "%%"
			stats[MEM] = fmt.Sprintf("%.3f", memPercent) + "%%"
			stats[MEMUSAGE] = convertToHumanReadableBinaryUnit(float64(statsJSON.MemoryStats.Usage))
			stats[MEMLIMIT] = convertToHumanReadableBinaryUnit(float64(statsJSON.MemoryStats.Limit))
			stats[BLOCKINPUT] = convertToHumanReadableUnit(float64(bi))
			stats[BLOCKOUTPUT] = convertToHumanReadableUnit(float64(bo))
			stats[NETWORKINPUT] = convertToHumanReadableUnit(ni)
			stats[NETWORKOUTPUT] = convertToHumanReadableUnit(no)
			stats[PIDS] = statsJSON.PidsStats.Current
			result = append(result, stats)
		}
	}
	return result, nil
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
	infos, retErr := getPs(compose, context.Background(), args...)
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
		if util.IsContainedStringInList(container.Names, target_str) {
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
		if util.IsContainedStringInList(image.RepoTags, imageName) &&
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
		if util.IsContainedStringInList(image.RepoDigests, repoDigest) && len(image.RepoTags) == 0 {
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

func getComposeInstanceImpl(id, path string) (project.APIProject, error) {
	return dockercompose.NewProject(&ctx.Context{

		Context: project.Context{
			ComposeFiles: []string{path},
			ProjectName:  id,
		},
	}, nil)
}

func calcCPUPercent(stats *types.StatsJSON) float64 {
	cpuPercent := 0.0
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

func calcBlockIO(blockio types.BlkioStats) (blockRead uint64, blockWrite uint64) {
	for _, bio := range blockio.IoServiceBytesRecursive {
		switch strings.ToLower(bio.Op) {
		case "read":
			blockRead = blockRead + bio.Value
		case "write":
			blockWrite = blockWrite + bio.Value
		}
	}
	return
}

func calcNetworkIO(network map[string]types.NetworkStats) (float64, float64) {
	var rx, tx float64
	for _, v := range network {
		rx += float64(v.RxBytes)
		tx += float64(v.TxBytes)
	}
	return rx, tx
}

func convertToHumanReadableBinaryUnit(num float64) string {
	if num > 1024*1024*1024 {
		return fmt.Sprintf("%.3f", num/1024/1024/1024) + "GiB"
	} else if num > 1024*1024 {
		return fmt.Sprintf("%.3f", num/1024/1024) + "MiB"
	} else if num > 1024 {
		return fmt.Sprintf("%.3f", num/1024) + "KiB"
	} else {
		return fmt.Sprintf("%.3f", num) + "B"
	}
}

func convertToHumanReadableUnit(num float64) string {
	if num > 1000*1000*1000 {
		return fmt.Sprintf("%.3f", num/1000/1000/1000) + "GB"
	} else if num > 1000*1000 {
		return fmt.Sprintf("%.3f", num/1000/1000) + "MB"
	} else if num > 1000 {
		return fmt.Sprintf("%.3f", num/1000) + "KB"
	} else {
		return fmt.Sprintf("%.3f", num) + "B"
	}
}

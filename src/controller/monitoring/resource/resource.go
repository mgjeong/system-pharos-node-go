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
package resource

import (
	"commons/errors"
	"commons/logger"
	"commons/url"
	"commons/util"
	"controller/dockercontroller"
	"db/bolt/service"
	"encoding/json"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"messenger"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	SYSTEMCONTAINER = "SYSTEMCONTAINER"
	COMPOSE_FILE    = "docker-compose.yaml"
	DESCRIPTION     = "description"
	CPU             = "cpu"
	MEM             = "mem"
	DISK            = "disk"
	NETWORK         = "network"
	INTERFACENAME   = "interfacename"
	BYTESSENT       = "bytessent"
	BYTESRECV       = "bytesrecv"
	PACKETSSENT     = "packetssent"
	PACKETSRECV     = "packetsrecv"
	TOTAL           = "total"
	FREE            = "free"
	USED            = "used"
	USEDPERCENT     = "usedpercent"
	PATH            = "path"
	SERVICES        = "services"
)

type Command interface {
	GetHostResourceInfo() (map[string]interface{}, error)
	GetAppResourceInfo(appId string) (map[string]interface{}, error)
}

type networkTraffic struct {
	InterfaceName string
	BytesSent     string
	BytesRecv     string
	PacketsSent   string
	PacketsRecv   string
}

type memoryUsage struct {
	Total       string
	Free        string
	Used        string
	UsedPercent string
}

type diskUsage struct {
	Path        string
	Total       string
	Free        string
	Used        string
	UsedPercent string
}

type resExecutorImpl struct{}

var dockerExecutor dockercontroller.Command
var dbExecutor service.Command
var httpExecutor messenger.Command
var Executor resExecutorImpl
var fileMode = os.FileMode(0755)

func init() {
	dockerExecutor = dockercontroller.Executor
	dbExecutor = service.Executor{}
	httpExecutor = messenger.NewExecutor()
}

func (resExecutorImpl) GetAppResourceInfo(appId string) (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	err := setYamlFile(appId)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}
	defer os.RemoveAll(COMPOSE_FILE)

	appStats, err := dockerExecutor.GetAppStats(appId, COMPOSE_FILE)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}

	results := make(map[string]interface{})
	results[SERVICES] = appStats
	return results, err
}

func (resExecutorImpl) GetHostResourceInfo() (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	cpu, err := getCPUUsage()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}
	mem, err := getMemUsage()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}
	disk, err := getDiskUsage()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}
	network, err := getNetworkTrafficInfo()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}

	resource := make(map[string]interface{})
	resource[CPU] = cpu
	resource[DISK] = disk
	resource[MEM] = mem
	resource[NETWORK] = network

	return resource, err
}

func getCPUUsage() ([]string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	percent, err := cpu.Percent(time.Second, true)
	if err != nil {
		logger.Logging(logger.DEBUG, "gopsutil cpu.Percent() error")
		return nil, errors.Unknown{"gopsutil cpu.Percent() error"}
	}

	result := make([]string, 0)
	for _, float := range percent {
		result = append(result, strconv.FormatFloat(float, 'f', 2, 64)+"%%")
	}
	return result, nil
}

func getMemUsage() (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	mem_v, err := mem.VirtualMemory()
	if err != nil {
		logger.Logging(logger.DEBUG, "gopsutil mem.VirtualMemory() error")
		return nil, errors.Unknown{"gopsutil mem.VirtualMemory() error"}
	}
	mem := memoryUsage{}
	mem.Total = strconv.FormatUint(mem_v.Total/1024, 10) + "KB"
	mem.Free = strconv.FormatUint(mem_v.Free/1024, 10) + "KB"
	mem.Used = strconv.FormatUint(mem_v.Used/1024, 10) + "KB"
	mem.UsedPercent = strconv.FormatFloat(mem_v.UsedPercent, 'f', 2, 64) + "%%"

	return convertToMemUsageMap(mem), err
}

func getDiskUsage() ([]map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	diskInfoList := make([]map[string]interface{}, 0)
	if scIP, exists := os.LookupEnv(SYSTEMCONTAINER); exists {
		scUrl := util.MakeSCRequestUrl(scIP, url.Disk())
		_, disk, err := httpExecutor.SendHttpRequest("GET", scUrl)
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return nil, err
		}
		disk = strings.Replace(disk, "%", "%%", -1)

		diskMap := make(map[string]interface{})
		err = json.Unmarshal([]byte(disk), &diskMap)
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return nil, err
		}

		for _, info := range diskMap[DISK].([]interface{}) {
			diskInfoList = append(diskInfoList, info.(map[string]interface{}))
		}

		return diskInfoList, err
	} else {
		return diskInfoList, nil
	}
}

func getNetworkTrafficInfo() ([]map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	result := make([]map[string]interface{}, 0)
	IOCountersStats, err := net.IOCounters(true)
	for _, IOCountersStat := range IOCountersStats {
		network := networkTraffic{}
		network.InterfaceName = IOCountersStat.Name
		network.BytesSent = strconv.FormatUint(IOCountersStat.BytesSent, 10)
		network.BytesRecv = strconv.FormatUint(IOCountersStat.BytesRecv, 10)
		network.PacketsSent = strconv.FormatUint(IOCountersStat.PacketsSent, 10)
		network.PacketsRecv = strconv.FormatUint(IOCountersStat.PacketsRecv, 10)
		result = append(result, convertToNetworkTrafficMap(network))
	}
	return result, err
}

func convertToNetworkTrafficMap(network networkTraffic) map[string]interface{} {
	return map[string]interface{}{
		INTERFACENAME: network.InterfaceName,
		BYTESSENT:     network.BytesSent,
		BYTESRECV:     network.BytesRecv,
		PACKETSSENT:   network.PacketsSent,
		PACKETSRECV:   network.PacketsRecv,
	}
}
func convertToMemUsageMap(mem memoryUsage) map[string]interface{} {
	return map[string]interface{}{
		TOTAL:       mem.Total,
		FREE:        mem.Free,
		USED:        mem.Used,
		USEDPERCENT: mem.UsedPercent,
	}
}

func convertToDiskUsageMap(disk diskUsage) map[string]interface{} {
	return map[string]interface{}{
		PATH:        disk.Path,
		TOTAL:       disk.Total,
		FREE:        disk.Free,
		USED:        disk.Used,
		USEDPERCENT: disk.UsedPercent,
	}
}

// Set YAML file about an app on a path.
// The path is defined as contant
// if setting YAML is succeeded, return error as nil
// otherwise, return error.
func setYamlFile(appId string) error {
	app, err := dbExecutor.GetApp(appId)
	if err != nil {
		return convertDBError(err, appId)
	}
	description := make(map[string]interface{})
	err = json.Unmarshal([]byte(app[DESCRIPTION].(string)), &description)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return errors.IOError{"json unmarshal fail"}
	}
	yaml, err := yaml.Marshal(description)
	if err != nil {
		return errors.InvalidYaml{Msg: "invalid yaml syntax"}
	}
	err = ioutil.WriteFile(COMPOSE_FILE, yaml, fileMode)
	if err != nil {
		return errors.IOError{Msg: "file io fail"}
	}
	return nil
}

func convertDBError(err error, appId string) error {
	switch err.(type) {
	case errors.NotFound:
		return errors.InvalidAppId{Msg: "failed to find app id : " + appId}
	default:
		return errors.Unknown{Msg: "db operation fail"}
	}
}

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

// Package configuration provide virtual functionality of configuration.
package configuration

import (
	"commons/errors"
	"commons/logger"
	"commons/util"
	"controller/dockercontroller"
	"db/bolt/configuration"
	"github.com/shirou/gopsutil/cpu"
	"os"
	"strconv"
)

// Interface of configuration operations.
type Command interface {
	// GetConfiguration returns a map of configuration stored in database.
	GetConfiguration() (map[string]interface{}, error)

	// SetConfiguration updates configuration sets.
	SetConfiguration(body string) error
}

type Executor struct{}

type (
	platformInfo struct {
		Platform string
		Family   string
		Version  string
	}

	processorInfo struct {
		CPU       string
		ModelName string
	}
)

const (
	PROPERTIES            = "properties"
	NAME                  = "name"
	VALUE                 = "value"
	READONLY              = "readOnly"
	DEFAULT_DEVICE_NAME   = "EdgeDevice"
	DEFAULT_PING_INTERVAL = "10"
)

var dbExecutor configuration.Command
var dockerExecutor dockercontroller.Command

func init() {
	dbExecutor = configuration.Executor{}
	dockerExecutor = dockercontroller.Executor
	// Initialize configuration before loading pharos-node.
	initConfiguration()
}

func initConfiguration() {
	anchoraddress := os.Getenv("ANCHOR_ADDRESS")
	nodeaddress := os.Getenv("NODE_ADDRESS")

	os, platform, err := getOSInfo()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
	}

	processor, err := getProcessorInfo()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
	}

	deviceName := DEFAULT_DEVICE_NAME
	prop, err := dbExecutor.GetProperty("devicename")
	if err == nil {
		deviceName = prop["value"].(string)
	}

	interval := DEFAULT_PING_INTERVAL
	prop, err = dbExecutor.GetProperty("pinginterval")
	if err == nil {
		interval = prop["value"].(string)
	}

	properties := make([]map[string]interface{}, 0)
	properties = append(properties, makeProperty("anchoraddress", anchoraddress, true))
	properties = append(properties, makeProperty("nodeaddress", nodeaddress, true))
	properties = append(properties, makeProperty("devicename", deviceName, false))
	properties = append(properties, makeProperty("pinginterval", interval, false))
	properties = append(properties, makeProperty("os", os, true))
	properties = append(properties, makeProperty("platform", platform, true))
	properties = append(properties, makeProperty("processor", processor, true))
	properties = append(properties, makeProperty("nodeid", "", false))

	for _, prop := range properties {
		err = dbExecutor.SetProperty(prop)
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
		}
	}
}

func (Executor) GetConfiguration() (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	props, err := dbExecutor.GetProperties()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}

	values := make([]map[string]interface{}, 0)
	for _, prop := range props {
		value := make(map[string]interface{})
		value[prop["name"].(string)] = prop["value"]
		value["readOnly"] = prop["readOnly"]
		values = append(values, value)
	}

	res := make(map[string]interface{})
	res[PROPERTIES] = values

	return res, nil
}

func (configurator Executor) SetConfiguration(body string) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	bodyMap, err := util.ConvertJsonToMap(body)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	for _, prop := range bodyMap[PROPERTIES].([]interface{}) {
		for key, value := range prop.(map[string]interface{}) {
			property, err := dbExecutor.GetProperty(key)
			if err != nil {
				logger.Logging(logger.ERROR, err.Error())
				return errors.InvalidJSON{"not supported property"}
			}

			if property[READONLY].(bool) {
				return errors.InvalidJSON{"read only property"}
			}

			property[VALUE] = value
			err = dbExecutor.SetProperty(property)
			if err != nil {
				logger.Logging(logger.ERROR, err.Error())
				return convertDBError(err)
			}
		}
	}

	return nil
}

func makeProperty(name string, value interface{}, readOnly bool) map[string]interface{} {
	prop := make(map[string]interface{})
	prop[NAME] = name
	prop[VALUE] = value
	prop[READONLY] = readOnly
	return prop
}

func getOSInfo() (string, string, error) {
	infoMap, err := dockerExecutor.Info()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return "", "", err
	}

	return infoMap["OSType"].(string), infoMap["OperatingSystem"].(string), nil
}

//
//func getPlatformInfo() (map[string]interface{}, error) {
//	platform, family, version, err := host.PlatformInformation()
//	if err != nil {
//		logger.Logging(logger.ERROR, err.Error())
//		return nil, errors.Unknown{"gopsutil host.PlatformInformation() error"}
//	}
//
//	info := platformInfo{}
//	info.Platform = platform
//	info.Family = family
//	info.Version = version
//
//	return convertToPlatformMap(info), nil
//}

func getProcessorInfo() ([]map[string]interface{}, error) {
	infos, err := cpu.Info()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, errors.Unknown{"cpu host.Info() error"}
	}

	result := make([]map[string]interface{}, 0)
	for _, info := range infos {
		procs := processorInfo{}
		procs.CPU = strconv.FormatInt(int64(info.CPU), 10)
		procs.ModelName = info.ModelName
		result = append(result, convertToProcessorMap(procs))
	}

	return result, err
}

func convertToPlatformMap(info platformInfo) map[string]interface{} {
	return map[string]interface{}{
		"platform": info.Platform,
		"family":   info.Family,
		"version":  info.Version,
	}
}

func convertToProcessorMap(info processorInfo) map[string]interface{} {
	return map[string]interface{}{
		"cpu":       info.CPU,
		"modelname": info.ModelName,
	}
}

func convertDBError(err error) error {
	switch err.(type) {
	case errors.NotFound:
		return errors.NotFound{}
	default:
		return errors.Unknown{Msg: "db operation fail"}
	}
}

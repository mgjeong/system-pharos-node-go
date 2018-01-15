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
	shell "controller/shellcommand"
	"strings"
)

type Command interface {
	GetResourceInfo() (map[string]interface{}, error)
	GetPerformanceInfo() (map[string]interface{}, error)
}

type resExecutorImpl struct{}

var Executor resExecutorImpl
var shellExecutor shell.Command

func init() {
	shellExecutor = shell.Executor
}

func (resExecutorImpl) GetResourceInfo() (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	processor, err := getProcessorModel()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}
	os, err := getOS()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}
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

	resources := make(map[string]interface{})
	resources["processor"] = processor
	resources["os"] = os
	resources["cpu"] = cpu
	resources["disk"] = disk
	resources["mem"] = mem

	return resources, err
}

func (resExecutorImpl) GetPerformanceInfo() (map[string]interface{}, error) {
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

	usage := make(map[string]interface{})
	usage["cpu"] = cpu
	usage["disk"] = disk
	usage["mem"] = mem

	return usage, err
}

func getProcessorModel() (string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	modelName, err := shellExecutor.ExecuteCommand("bash", "-c", "grep -m1 ^'model name' /proc/cpuinfo")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return "", err
	}
	if len(modelName) == 0 {
		logger.Logging(logger.ERROR, "can't find cpu model name info")
		return "", errors.Unknown{"can't find cpu model name info"}
	}
	modelInfo := strings.Split(modelName, ":")
	_, model := modelInfo[0], modelInfo[1]

	return strings.TrimSpace(model), err
}

func getOS() (string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	os, err := shellExecutor.ExecuteCommand("bash", "-c", "uname -mrs")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return "", err
	}

	return strings.TrimSpace(os), err
}

func getCPUUsage() (string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	procStatCPU, err := shellExecutor.ExecuteCommand("bash", "-c", "cat /proc/stat | grep cpu")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return "", err
	}
	if len(procStatCPU) == 0 {
		logger.Logging(logger.ERROR, "can't find cpu usage info")
		return "", errors.Unknown{"can't find cpu usage info"}
	}
	
	return procStatCPU, err
}

func getMemUsage() (string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	procMeminfo, err := shellExecutor.ExecuteCommand("bash", "-c", "cat /proc/meminfo")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return "", err
	}
	if len(procMeminfo) == 0 {
		logger.Logging(logger.ERROR, "can't find total memory info")
		return "", errors.Unknown{"can't find total memory info"}
	}
	
	return procMeminfo, err
}

func getDiskUsage() (string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	df, err := shellExecutor.ExecuteCommand("bash", "-c", "df -m")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return "", err
	}
	if len(df) == 0 {
		logger.Logging(logger.ERROR, "can't find total disk info")
		return "", errors.Unknown{"can't find total disk info"}
	}
	
	return df, err
}
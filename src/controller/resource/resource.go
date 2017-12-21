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
	"strconv"
	"strings"
)

type ResourceInterface interface {
	GetResourceInfo() (map[string]interface{}, error)
	GetPerformanceInfo() (map[string]interface{}, error)
}

type resource struct{}

var Resource resource

var shellExecutor shell.ShellInterface

func init() {
	shellExecutor = shell.Executor
}

func (resource) GetResourceInfo() (map[string]interface{}, error) {
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

func (resource) GetPerformanceInfo() (map[string]interface{}, error) {
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
		return "", errors.Unknown{"can't find cpu usage info"}
	}

	procStatCPUSlice := strings.Split(procStatCPU, " ")

	user, _ := strconv.Atoi(procStatCPUSlice[1])
	nice, _ := strconv.Atoi(procStatCPUSlice[2])
	system, _ := strconv.Atoi(procStatCPUSlice[3])
	idle, _ := strconv.Atoi(procStatCPUSlice[4])
	iowait, _ := strconv.Atoi(procStatCPUSlice[5])
	irq, _ := strconv.Atoi(procStatCPUSlice[6])
	softirq, _ := strconv.Atoi(procStatCPUSlice[7])
	steal, _ := strconv.Atoi(procStatCPUSlice[8])

	totalTime := user + nice + system + idle + iowait + irq + softirq + steal
	idleTime := idle + iowait

	cpuUsagePerc := 100 * (totalTime - idleTime) / totalTime
	return strconv.Itoa(cpuUsagePerc) + "%%", err
}

func getMemUsage() (string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	total, err := shellExecutor.ExecuteCommand("bash", "-c", "cat /proc/meminfo | grep MemTotal: | awk '{print $2}'")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return "", err
	}
	if len(total) == 0 {
		return "", errors.Unknown{"can't find total memory info"}
	}

	free, err := shellExecutor.ExecuteCommand("bash", "-c", "cat /proc/meminfo | grep MemFree: | awk '{print $2}'")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return "", err
	}
	if len(free) == 0 {
		return "", errors.Unknown{"can't find used memory info"}
	}

	memTotal, _ := strconv.Atoi(strings.TrimSpace(total))
	memFree, _ := strconv.Atoi(strings.TrimSpace(free))

	memUsagePerc := 100 * (memTotal - memFree) / memTotal
	return strconv.Itoa(memUsagePerc) + "%%", err
}

func getDiskUsage() (string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	total, err := shellExecutor.ExecuteCommand("bash", "-c", "df -m | awk '{print $2}'")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return "", err
	}
	if len(total) == 0 {
		return "", errors.Unknown{"can't find total disk info"}
	}

	available, err := shellExecutor.ExecuteCommand("bash", "-c", "df -m | awk '{print $4}'")
	if err != nil {
		return "", err
	}
	if len(available) == 0 {
		return "", errors.Unknown{"can't find availble disk info"}
	}

	totalSlice := strings.Split(total, "\n")
	availableSlice := strings.Split(available, "\n")

	diskTotalSum := 0
	diskAvailableSum := 0

	for idx, value := range totalSlice {
		if idx != 0 {
			diskSize, _ := strconv.Atoi(value)
			diskAvailable, _ := strconv.Atoi(availableSlice[idx])
			diskTotalSum += diskSize
			diskAvailableSum += diskAvailable
		}
	}

	diskUsagePerc := 100 * (diskTotalSum - diskAvailableSum) / diskTotalSum
	return strconv.Itoa(diskUsagePerc) + "%%", err
}
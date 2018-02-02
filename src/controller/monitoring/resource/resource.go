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
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"strconv"
	"time"
)

type Command interface {
	GetResourceInfo() (map[string]interface{}, error)
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

var Executor resExecutorImpl

func (resExecutorImpl) GetResourceInfo() (map[string]interface{}, error) {
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

	resources := make(map[string]interface{})
	resources["cpu"] = cpu
	resources["disk"] = disk
	resources["mem"] = mem

	return resources, err
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

	parts, err := disk.Partitions(false)
	if err != nil {
		logger.Logging(logger.DEBUG, "gopsutil disk.Partitions() error")
		return nil, errors.Unknown{"gopsutil  disk.Partitions() error"}
	}

	result := make([]map[string]interface{}, 0)
	for _, part := range parts {
		du, err := disk.Usage(part.Mountpoint)
		if err != nil {
			logger.Logging(logger.DEBUG, "gopsutil disk.Usage() error")
			return nil, errors.Unknown{"gopsutil  disk.Usage() error"}
		}
		disk := diskUsage{}
		disk.Path = du.Path
		disk.Free = strconv.FormatUint(du.Free/1024/1024/1024, 10) + "G"
		disk.Total = strconv.FormatUint(du.Total/1024/1024/1024, 10) + "G"
		disk.Used = strconv.FormatUint(du.Used/1024/1024/1024, 10) + "G"
		disk.UsedPercent = strconv.FormatFloat(du.UsedPercent, 'f', 2, 64) + "%%"
		result = append(result, convertToDiskUsageMap(disk))
	}
	return result, err
}

func convertToMemUsageMap(mem memoryUsage) map[string]interface{} {
	return map[string]interface{}{
		"total":       mem.Total,
		"free":        mem.Free,
		"used":        mem.Used,
		"usedpercent": mem.UsedPercent,
	}
}

func convertToDiskUsageMap(disk diskUsage) map[string]interface{} {
	return map[string]interface{}{
		"path":        disk.Path,
		"total":       disk.Total,
		"free":        disk.Free,
		"used":        disk.Used,
		"usedpercent": disk.UsedPercent,
	}
}

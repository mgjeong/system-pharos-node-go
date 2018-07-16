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

// Package health provides logic of checking health with system-edge-manager service.
package health

import (
	"commons/errors"
	"commons/logger"
	"commons/url"
	"commons/util"
	"controller/configuration"
	configDB "db/bolt/configuration"
	"db/bolt/service"
	"messenger"
	"runtime"
	"time"
)

const (
	HTTP_TAG               = "http://"
	IP                     = "ip"
	MANAGER                = "manager"
	NODE                   = "node"
	INTERVAL               = "interval"
	HEALTH_CHECK           = "healthCheck"
	DEFAULT_RETRY_INTERVAL = 1
	TIME_UNIT              = time.Minute
)

type Command interface {
	Unregister() error
}

type Executor struct{}

var httpExecutor messenger.Command
var configurator configuration.Command
var srvDbExecutor service.Command
var configDbExecutor configDB.Command

func init() {
	httpExecutor = messenger.NewExecutor()
	configurator = configuration.Executor{}
	srvDbExecutor = service.Executor{}
	configDbExecutor = configDB.Executor{}

	// Request to register new pharos node.
	err := register(true)
	if err != nil {
		quit := make(chan bool)
		ticker := time.NewTicker(time.Duration(DEFAULT_RETRY_INTERVAL) * TIME_UNIT)
		go func() {
			for {
				select {
				case <-ticker.C:
					err := register(true)
					if err != nil {
						logger.Logging(logger.ERROR, err.Error())
					} else {
						logger.Logging(logger.ERROR, "Successfully registered")
						ticker.Stop()
						close(quit)
						return
					}
				}
				runtime.Gosched()
			}
		}()
	}
}

// register to pharos-anchor service.
// should know the pharos-anchor address(ip:port)
// if succeed to register, return error as nil
// otherwise, return error.
func register(enableHealthCheck bool) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	config, err := configurator.GetConfiguration()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	// Make a request body for registration.
	body := makeRegistrationBody(config)

	code, respStr, err := sendRegisterRequest(body)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	respMap, err := common.convertRespToMap(respStr)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	if code != 200 {
		message := respMap["message"].(string)
		return errors.Unknown{"received error message from system-edge-manager" + message}
	}

	// Insert deviceId in configuration db.
	property, err := configDbExecutor.GetProperty("deviceid")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return errors.InvalidJSON{"not supported property"}
	}

	property["value"] = respMap["id"]
	err = configDbExecutor.SetProperty(property)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	// Start a new ticker and send a ping message repeatedly at regular intervals.
	if enableHealthCheck {
		startHealthCheck()
	}
	return nil
}

// Unregister to pharos-anchor service.
// if succeed to unregister, return error as nil
// otherwise, return error.
func (Executor) Unregister() error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	// Reset node id.
	property, err := configDbExecutor.GetProperty("deviceid")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return errors.InvalidJSON{"not supported property"}
	}

	property["value"] = ""
	err = configDbExecutor.SetProperty(property)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	// Stop a ticker to send ping request.
	if common.quit != nil {
		common.quit <- true
	}
	return nil
}

func sendRegisterRequest(body map[string]interface{}) (int, string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	url, err := util.MakeAnchorRequestUrl(url.Management(), url.Nodes(), url.Register())
	if err != nil {
		logger.Logging(logger.ERROR, "failed to make anchor request url")
	}

	jsonData, err := util.ConvertMapToJson(body)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return 500, "", err
	}
	return httpExecutor.SendHttpRequest("POST", url, []byte(jsonData))
}

func sendUnregisterRequest(nodeID string) (int, string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	url, err := util.MakeAnchorRequestUrl(url.Management(), url.Nodes(), "/", nodeID, url.Unregister())
	if err != nil {
		logger.Logging(logger.ERROR, "failed to make anchor request url")
		return 500, "", err
	}
	return httpExecutor.SendHttpRequest("POST", url)
}

func makeRegistrationBody(config map[string]interface{}) map[string]interface{} {
	data := make(map[string]interface{})

	// Set pharos-node address from configuration.
	properties := config["properties"].([]map[string]interface{})
	for _, prop := range properties {
		if value, exists := prop["nodeaddress"]; exists {
			data["ip"] = value
		}
	}

	// Remove unnecessary property from configuration.
	filteredProps := make([]map[string]interface{}, 0)
	for _, prop := range properties {
		if _, exists := prop["anchorendpoint"]; exists {
			continue
		}
		if _, exists := prop["anchoraddress"]; exists {
			continue
		}
		if _, exists := prop["nodeaddress"]; exists {
			continue
		}
		filteredProps = append(filteredProps, prop)
	}

	// Set configuration information in request body.
	configData := make(map[string]interface{})
	configData["properties"] = filteredProps

	// Set application information in request body.
	apps, err := srvDbExecutor.GetAppList()
	appIds := make([]string, 0)
	if err == nil {
		for _, app := range apps {
			appIds = append(appIds, app["id"].(string))
		}
	}

	data["config"] = configData
	data["apps"] = appIds
	return data
}

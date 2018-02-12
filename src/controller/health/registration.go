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
	"controller/configuration"
	"encoding/json"
	"messenger"
	"strings"
	"time"
)

const (
	HTTP_TAG          = "http://"
	IP                = "ip"
	MANAGER           = "manager"
	NODE              = "node"
	INTERVAL          = "interval"
	HEALTH_CHECK      = "healthCheck"
	DEFAULT_SDAM_PORT = "48099"
	TIME_UNIT         = time.Minute
)

type Command interface {
	Unregister() error
}

type Executor struct{}

var httpExecutor messenger.Command
var configurator configuration.Command

func init() {
	httpExecutor = messenger.NewExecutor()
	configurator = configuration.Executor{}

	// Register
	err := register()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
	}
}

// register to pharos-anchor service.
// should know the pharos-anchor address(ip:port)
// if succeed to register, return error as nil
// otherwise, return error.
func register() error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	config, err := configurator.GetConfiguration()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	// Get pharos-anchor address from configuration.
	for _, prop := range config["properties"].([]map[string]interface{}) {
		if strings.Compare(prop["name"].(string), "anchoraddress") == 0 {
			common.managerAddress = prop["value"].(string)
		}
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

	// Insert node id in configuration db.
	updatedProp := make(map[string]interface{})
	updatedProp["name"] = "nodeid"
	updatedProp["value"] = respMap["id"]

	updatedProperties := make(map[string]interface{})
	updatedProperties["properties"] = []map[string]interface{}{updatedProp}

	jsonString, _ := json.Marshal(updatedProperties)
	err = configurator.SetConfiguration(string(jsonString))
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	// Start a new ticker and send a ping message repeatedly at regular intervals.
	startHealthCheck(respMap["id"].(string))
	return nil
}

// Unregister to pharos-anchor service.
// if succeed to unregister, return error as nil
// otherwise, return error.
func (Executor) Unregister() error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	// Reset node id.
	updatedProp := make(map[string]interface{})
	updatedProp["name"] = "nodeid"
	updatedProp["value"] = ""

	updatedProperties := make(map[string]interface{})
	updatedProperties["properties"] = []map[string]interface{}{updatedProp}

	jsonString, _ := json.Marshal(updatedProperties)
	err := configurator.SetConfiguration(string(jsonString))
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

	url := common.makeRequestUrl(url.Nodes(), url.Register())

	jsonData, err := common.convertMapToJson(body)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return 500, "", err
	}
	return httpExecutor.SendHttpRequest("POST", url, []byte(jsonData))
}

func sendUnregisterRequest(nodeID string) (int, string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	url := common.makeRequestUrl(url.Nodes(), "/", nodeID, url.Unregister())
	return httpExecutor.SendHttpRequest("POST", url)
}

func makeRegistrationBody(config map[string]interface{}) map[string]interface{} {
	data := make(map[string]interface{})

	// Set pharos-node address from configuration.
	properties := config["properties"].([]map[string]interface{})
	for _, prop := range properties {
		if strings.Compare(prop["name"].(string), "nodeaddress") == 0 {
			data["ip"] = prop["value"].(string)
		}
	}

	// Remove unnecessary property from configuration.
	filteredProps := make([]map[string]interface{}, 0)
	for _, prop := range properties {
		if strings.Compare(prop["name"].(string), "nodeid") != 0 &&
			strings.Compare(prop["name"].(string), "anchoraddress") != 0 &&
			strings.Compare(prop["name"].(string), "nodeaddress") != 0 {
			filteredProps = append(filteredProps, prop)
		}
	}

	// Set configuration information in request body.
	configData := make(map[string]interface{})
	configData["properties"] = filteredProps

	data["config"] = configData
	return data
}

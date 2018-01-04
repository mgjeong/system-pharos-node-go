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

// Package registration provides logic of checking health with system-edge-manager service.
package registration

import (
	"bytes"
	"commons/errors"
	"commons/logger"
	"commons/url"
	"controller/configuration"
	"encoding/json"
	"messenger"
	"strconv"
	"time"
)

const (
	HTTP_TAG          = "http://"
	IP                = "ip"
	MANAGER           = "manager"
	AGENT             = "agent"
	INTERVAL          = "interval"
	HEALTH_CHECK      = "healthCheck"
	DEFAULT_SDAM_PORT = "48099"
	TIME_UNIT         = time.Minute
)

var (
	quit           chan bool
	ticker         *time.Ticker
	managerAddress string
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

// register to system-edge-manager service.
// should know the system-edge-manager address(ip:port)
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

	// Get system-edge-manager address from configuration.
	managerAddress = config["serveraddress"].(string)

	// Make a request body for registration.
	body := makeRegistrationBody(config)

	code, respStr, err := sendRegisterRequest(body)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	respMap, err := convertRespToMap(respStr)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	if code != 200 {
		message := respMap["message"].(string)
		return errors.Unknown{"received error message from system-edge-manager" + message}
	}

	// Insert agent id in configuration file.
	newConfig := make(map[string]interface{})
	newConfig["agentid"] = respMap["id"]

	err = configurator.SetConfiguration(newConfig)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	// Start a new ticker and send a ping message repeatedly at regular intervals.
	startHealthCheck(respMap["id"].(string))
	return nil
}

// Unregister to system-edge-manager service.
// if succeed to unregister, return error as nil
// otherwise, return error.
func (Executor) Unregister() error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	// Reset agent id.
	newConfig := make(map[string]interface{})
	newConfig["agentid"] = ""

	err := configurator.SetConfiguration(newConfig)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	// Stop a ticker to send ping request.
	if quit != nil {
		quit <- true
	}
	return nil
}

func stopHealthCheck() {
	close(quit)
	quit = nil
}

func startHealthCheck(agentID string) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	// Get interval from configuration file.
	config, err := configurator.GetConfiguration()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
	}
	interval := config["pinginterval"].(string)

	quit = make(chan bool)
	intervalInt, _ := strconv.Atoi(interval)
	ticker = time.NewTicker(time.Duration(intervalInt) * TIME_UNIT)
	go func() {
		for {
			select {
			case <-ticker.C:
				sendPingRequest(agentID, interval)
			case <-quit:
				ticker.Stop()
				stopHealthCheck()
				return
			}
		}
	}()
}

func sendPingRequest(agentID string, interval string) (int, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	data := make(map[string]interface{})
	data[INTERVAL] = interval

	jsonData, err := convertMapToJson(data)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return 500, err
	}

	logger.Logging(logger.DEBUG, "try to send ping request")

	url := makeRequestUrl(url.Agents(), "/", agentID, url.Ping())
	code, _, err := httpExecutor.SendHttpRequest("POST", url, []byte(jsonData))
	if err != nil {
		logger.Logging(logger.ERROR, "failed to send ping request")
		return code, err
	}

	logger.Logging(logger.DEBUG, "receive pong response, code["+strconv.Itoa(code)+"]")
	return code, nil
}

func sendRegisterRequest(body map[string]interface{}) (int, string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	url := makeRequestUrl(url.Agents(), url.Register())

	jsonData, err := convertMapToJson(body)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return 500, "", err
	}
	return httpExecutor.SendHttpRequest("POST", url, []byte(jsonData))
}

func sendUnregisterRequest(agentID string) (int, string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	url := makeRequestUrl(url.Agents(), "/", agentID, url.Unregister())
	return httpExecutor.SendHttpRequest("POST", url)
}

func makeRequestUrl(api_parts ...string) string {
	var full_url bytes.Buffer
	full_url.WriteString(HTTP_TAG + managerAddress + ":" + DEFAULT_SDAM_PORT + url.Base())
	for _, api_part := range api_parts {
		full_url.WriteString(api_part)
	}

	logger.Logging(logger.DEBUG, full_url.String())
	return full_url.String()
}

func convertJsonToMap(jsonStr string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, errors.InvalidParam{"json unmarshalling failed"}
	}
	return result, err
}

func convertMapToJson(data map[string]interface{}) (string, error) {
	result, err := json.Marshal(data)
	if err != nil {
		return "", errors.Unknown{"json marshalling failed"}
	}
	return string(result), nil
}

func convertRespToMap(respStr string) (map[string]interface{}, error) {
	resp, err := convertJsonToMap(respStr)
	if err != nil {
		logger.Logging(logger.ERROR, "Failed to convert response from string to map")
		return nil, errors.Unknown{"Json Converting Failed"}
	}
	return resp, err
}

func makeRegistrationBody(config map[string]interface{}) map[string]interface{} {
	data := make(map[string]interface{})

	// Set device address from configuration.
	data["ip"] = config["deviceaddress"].(string)

	// Delete unused field.
	delete(config, "serveraddress")
	delete(config, "deviceaddress")
	delete(config, "agentid")

	// Set configuration information in request body.
	data["config"] = config

	return data
}

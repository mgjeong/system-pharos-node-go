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
	// TODO: agentId should be managed in configuration file.
	agentId        string
	quit           chan bool
	ticker         *time.Ticker
	interval       string
	managerAddress string
	agentAddress   string
)

type RegistrationInterface interface {
	Register(body string) error
	Unregister() error
}

type Registration struct{}

var httpRequester messenger.MessengerInterface

func init() {
	httpRequester = messenger.NewMessenger()
}

// Register to system-edge-manger service.
// should know the system-edge-manger address(ip:port)
// if succeed to register, return error as nil
// otherwise, return error.
func (Registration) Register(body string) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	if agentId != "" {
		return errors.Unknown{"already registered, unregister it and try registration again"}
	}

	bodyMap, err := convertJsonToMap(body)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	if _, exists := bodyMap[IP]; !exists {
		return errors.InvalidParam{"ip field is required"}
	}

	if _, exists := bodyMap[HEALTH_CHECK]; !exists {
		return errors.InvalidParam{"healthCheck field is required"}
	}

	if _, exists := bodyMap[HEALTH_CHECK].(map[string]interface{})[INTERVAL]; !exists {
		return errors.InvalidParam{"interval field is required"}
	}

	managerAddress = bodyMap[IP].(map[string]interface{})[MANAGER].(string)
	agentAddress = bodyMap[IP].(map[string]interface{})[AGENT].(string)
	healthCheck := bodyMap[HEALTH_CHECK].(map[string]interface{})
	interval = healthCheck[INTERVAL].(string)

	code, respStr, err := sendRegisterRequest()
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
		return errors.Unknown{"received error message from sdam" + message}
	}

	// Start a new ticker and send a ping message repeatedly at regular intervals.
	agentId = respMap["id"].(string)
	startHealthCheck(interval)
	return nil
}

// Unregister to system-edge-manger service.
// if succeed to unregister, return error as nil
// otherwise, return error.
func (Registration) Unregister() error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	if agentId == "" {
		return errors.Unknown{"already unregistered"}
	}

	code, respStr, err := sendUnregisterRequest()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	if code != 200 {
		respMap, err := convertJsonToMap(respStr)
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return err
		}

		message := respMap["message"].(string)
		return errors.Unknown{"received error message from sdam:" + message}
	}

	agentId = ""
	if quit != nil {
		quit <- true
	}
	return nil
}

func stopHealthCheck() {
	close(quit)
	quit = nil
}

func startHealthCheck(interval string) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	quit = make(chan bool)
	intervalInt, _ := strconv.Atoi(interval)
	ticker = time.NewTicker(time.Duration(intervalInt) * TIME_UNIT)
	go func() {
		for {
			select {
			case <-ticker.C:
				sendPingRequest(interval)
			case <-quit:
				ticker.Stop()
				stopHealthCheck()
				return
			}
		}
	}()
}

func sendPingRequest(interval string) (int, error) {
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

	url := makeRequestUrl(url.Agents(), "/", agentId, url.Ping())
	code, _, err := httpRequester.SendHttpRequest("POST", url, []byte(jsonData))
	if err != nil {
		logger.Logging(logger.ERROR, "failed to send ping request")
		return code, err
	}

	logger.Logging(logger.DEBUG, "receive pong response, code["+strconv.Itoa(code)+"]")
	return code, nil
}

func sendRegisterRequest() (int, string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	url := makeRequestUrl(url.Agents(), url.Register())

	data := make(map[string]interface{})
	data["ip"] = agentAddress

	jsonData, err := convertMapToJson(data)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return 500, "", err
	}
	return httpRequester.SendHttpRequest("POST", url, []byte(jsonData))
}

func sendUnregisterRequest() (int, string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	url := makeRequestUrl(url.Agents(), "/", agentId, url.Unregister())
	return httpRequester.SendHttpRequest("POST", url)
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

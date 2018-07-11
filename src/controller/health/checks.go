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
	"strconv"
	"time"
)

func startHealthCheck() {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	// Get interval from configuration file.
	config, err := configurator.GetConfiguration()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
	}

	var interval string
	for _, prop := range config["properties"].([]map[string]interface{}) {
		if value, exists := prop["pinginterval"]; exists {
			interval = value.(string)
		}
	}

	common.quit = make(chan bool)
	intervalInt, _ := strconv.Atoi(interval)
	common.ticker = time.NewTicker(time.Duration(intervalInt) * TIME_UNIT)
	go func() {
		sendPingRequest(interval)
		for {
			select {
			case <-common.ticker.C:
				code, _ := sendPingRequest(interval)
				if code == 404 {
					logger.Logging(logger.ERROR, "received 'not found' error, re-registration is required")
					common.ticker.Stop()

					err := register(false)
					if err != nil {
						logger.Logging(logger.ERROR, err.Error())
					}
					common.ticker = time.NewTicker(time.Duration(intervalInt) * TIME_UNIT)
				}
			case <-common.quit:
				common.ticker.Stop()
				stopHealthCheck()
				return
			}
		}
	}()
}

func stopHealthCheck() {
	close(common.quit)
	common.quit = nil
}

func sendPingRequest(interval string) (int, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	property, err := configDbExecutor.GetProperty("deviceid")
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return 500, errors.InvalidJSON{"not supported property"}
	}
	nodeID := property["value"].(string)

	data := make(map[string]interface{})
	data[INTERVAL] = interval

	jsonData, err := util.ConvertMapToJson(data)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return 500, err
	}

	logger.Logging(logger.DEBUG, "try to send ping request")

	reqUrl, err := util.MakeAnchorRequestUrl(url.Management(), url.Nodes(), "/", nodeID, url.Ping())
	if err != nil {
		logger.Logging(logger.ERROR, "failed to make anchor request url")
		return 500, err
	}

	code, _, err := httpExecutor.SendHttpRequest("POST", reqUrl, []byte(jsonData))
	if err != nil {
		logger.Logging(logger.ERROR, "failed to send ping request")
		return code, err
	}

	logger.Logging(logger.DEBUG, "receive pong response, code["+strconv.Itoa(code)+"]")
	return code, nil
}

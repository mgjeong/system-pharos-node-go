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
	"commons/logger"
	"commons/url"
	"strconv"
	"time"
)

func startHealthCheck(nodeID string) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	// Get interval from configuration file.
	config, err := configurator.GetConfiguration()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
	}
	interval := config["pinginterval"].(string)

	common.quit = make(chan bool)
	intervalInt, _ := strconv.Atoi(interval)
	common.ticker = time.NewTicker(time.Duration(intervalInt) * TIME_UNIT)
	go func() {
		for {
			select {
			case <-common.ticker.C:
				sendPingRequest(nodeID, interval)
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

func sendPingRequest(nodeID string, interval string) (int, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	data := make(map[string]interface{})
	data[INTERVAL] = interval

	jsonData, err := common.convertMapToJson(data)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return 500, err
	}

	logger.Logging(logger.DEBUG, "try to send ping request")

	url := common.makeRequestUrl(url.Nodes(), "/", nodeID, url.Ping())
	code, _, err := httpExecutor.SendHttpRequest("POST", url, []byte(jsonData))
	if err != nil {
		logger.Logging(logger.ERROR, "failed to send ping request")
		return code, err
	}

	logger.Logging(logger.DEBUG, "receive pong response, code["+strconv.Itoa(code)+"]")
	return code, nil
}
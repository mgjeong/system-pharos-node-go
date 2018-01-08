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
	"bytes"
	"commons/errors"
	"commons/logger"
	"commons/url"
	"encoding/json"
	"time"
)

var common context

type context struct {
	quit           chan bool
	ticker         *time.Ticker
	managerAddress string
}

func (ctx context) makeRequestUrl(api_parts ...string) string {
	var full_url bytes.Buffer
	full_url.WriteString(HTTP_TAG + ctx.managerAddress + ":" + DEFAULT_SDAM_PORT + url.Base() + url.Management())
	for _, api_part := range api_parts {
		full_url.WriteString(api_part)
	}

	logger.Logging(logger.DEBUG, full_url.String())
	return full_url.String()
}

func (ctx context) convertJsonToMap(jsonStr string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, errors.InvalidParam{"json unmarshalling failed"}
	}
	return result, err
}

func (ctx context) convertMapToJson(data map[string]interface{}) (string, error) {
	result, err := json.Marshal(data)
	if err != nil {
		return "", errors.Unknown{"json marshalling failed"}
	}
	return string(result), nil
}

func (ctx context) convertRespToMap(respStr string) (map[string]interface{}, error) {
	resp, err := ctx.convertJsonToMap(respStr)
	if err != nil {
		logger.Logging(logger.ERROR, "Failed to convert response from string to map")
		return nil, errors.Unknown{"Json Converting Failed"}
	}
	return resp, err
}
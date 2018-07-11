/*******************************************************************************
 * Copyright 2018 Samsung Electronics All Rights Reserved.
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

// Package commons/util defines utility functions used by Pharos Node.
package util

import (
	"bytes"
	"commons/errors"
	"commons/logger"
	"commons/url"
	"encoding/json"
	"net"
	"os"
	"strings"
)

const (
	DEFAULT_ANCHOR_PORT                      = "48099"
	UNSECURED_ANCHOR_PORT_WITH_REVERSE_PROXY = "80"
)

// convertJsonToMap converts JSON data into a map.
// If successful, this function returns an error as nil.
// otherwise, an appropriate error will be returned.
func ConvertJsonToMap(jsonStr string) (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, errors.InvalidJSON{"Unmarshalling Failed"}
	}
	return result, err
}

// ConvertMapToJson converts Map data to json data.
// If successful, this function returns an error as nil.
// otherwise, an appropriate error will be returned.
func ConvertMapToJson(data map[string]interface{}) (string, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	result, err := json.Marshal(data)
	if err != nil {
		return "", errors.Unknown{"json marshalling failed"}
	}
	return string(result), nil
}

func IsContainedStringInList(list []string, str string) bool {
	for _, value := range list {
		if strings.Compare(value, str) == 0 {
			return true
		}
	}
	return false
}

// MakeAnchorRequestUrl makes url which is used to send request to Pharos Anchor.
func MakeAnchorRequestUrl(api_parts ...string) (string, error) {
	var full_url bytes.Buffer

	anchorIP := os.Getenv("ANCHOR_ADDRESS")
	if len(anchorIP) == 0 {
		logger.Logging(logger.ERROR, "No anchor address environment")
		return "", errors.NotFound{"No anchor address environment"}
	}

	ipTest := net.ParseIP(anchorIP)
	if ipTest == nil {
		logger.Logging(logger.ERROR, "Anchor address's validation check failed")
		return "", errors.InvalidParam{"Anchor address's validation check failed"}
	}

	anchorProxy := os.Getenv("ANCHOR_REVERSE_PROXY")
	if len(anchorProxy) == 0 || anchorProxy == "false" {
		full_url.WriteString("http://" + anchorIP + ":" + DEFAULT_ANCHOR_PORT + url.Base())
	} else if anchorProxy == "true" {
		full_url.WriteString("http://" + anchorIP + ":" + UNSECURED_ANCHOR_PORT_WITH_REVERSE_PROXY + url.PharosAnchor() + url.Base())
	} else {
		logger.Logging(logger.ERROR, "Invalid value for ANCHOR_REVERSE_PROXY")
		return "", errors.InvalidParam{"Invalid value for ANCHOR_REVERSE_PROXY"}
	}

	for _, api_part := range api_parts {
		full_url.WriteString(api_part)
	}

	logger.Logging(logger.DEBUG, full_url.String())
	return full_url.String(), nil
}

// MakeSCRequestUrl makes url which is used to send request to system continaer to control device.
func MakeSCRequestUrl(scIP string, api_parts ...string) string {
	var full_url bytes.Buffer

	full_url.WriteString("http://" + scIP + url.Base() + url.Device() + url.Management())
	for _, api_part := range api_parts {
		full_url.WriteString(api_part)
	}

	logger.Logging(logger.DEBUG, full_url.String())
	return full_url.String()
}

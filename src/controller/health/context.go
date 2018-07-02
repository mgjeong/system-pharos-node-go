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
	"commons/util"
	"time"
)

var common context

type context struct {
	quit           chan bool
	ticker         *time.Ticker
	managerAddress string
}

func (ctx context) convertRespToMap(respStr string) (map[string]interface{}, error) {
	resp, err := util.ConvertJsonToMap(respStr)
	if err != nil {
		logger.Logging(logger.ERROR, "Failed to convert response from string to map")
		return nil, errors.Unknown{"Json Converting Failed"}
	}
	return resp, err
}

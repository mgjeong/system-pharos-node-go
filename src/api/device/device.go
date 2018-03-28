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
package device

import (
	"api/common"
	"commons/errors"
	"commons/logger"
	"commons/url"
	"controller/device"
	"net/http"
	"strings"
)

const (
	GET    string = "GET"
	PUT    string = "PUT"
	POST   string = "POST"
	DELETE string = "DELETE"
)

type Command interface {
	Handle(w http.ResponseWriter, req *http.Request)
}

type apiInnerCommand interface {
	reboot(w http.ResponseWriter, req *http.Request)
	restore(w http.ResponseWriter, req *http.Request)
}

type Executor struct{}
type innerExecutorImpl struct{}

var apiInnerExecutor apiInnerCommand
var deviceExecutor device.Command

func init() {
	apiInnerExecutor = innerExecutorImpl{}
	deviceExecutor = device.Executor{}
}

func (Executor) Handle(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	switch reqUrl, _ := req.URL.Path, strings.Split(req.URL.Path, "/"); {
	case strings.Contains(reqUrl, url.Restore()):
		apiInnerExecutor.restore(w, req)
	case strings.Contains(reqUrl, url.Reboot()):
		apiInnerExecutor.reboot(w, req)
	default:
		logger.Logging(logger.DEBUG, "Unmatched url")
		common.MakeErrorResponse(w, errors.NotFoundURL{reqUrl})
	}
}

// reboot handles requests which is used to reboot a device.
func (innerExecutorImpl) reboot(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, POST) {
		return
	}

	e := deviceExecutor.Reboot()
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}

	response := make(map[string]interface{})
	response["result"] = "success"
	common.MakeResponse(w, common.ChangeToJson(response))
}

// restore handles requests which is used to reset a device to initail state.
func (innerExecutorImpl) restore(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, POST) {
		return
	}

	e := deviceExecutor.Restore()
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}

	response := make(map[string]interface{})
	response["result"] = "success"
	common.MakeResponse(w, common.ChangeToJson(response))
}

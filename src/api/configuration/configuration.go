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
package configuration

import (
	"api/common"
	"commons/errors"
	"commons/logger"
	"controller/configuration"
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
	configuration(w http.ResponseWriter, req *http.Request)
}

type Executor struct{}
type innerExecutorImpl struct{}

var apiInnerExecutor apiInnerCommand
var configurationExecutor configuration.Command

func init() {
	apiInnerExecutor = innerExecutorImpl{}
	configurationExecutor = configuration.Executor{}
}

func (Executor) Handle(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	switch reqUrl, split := req.URL.Path, strings.Split(req.URL.Path, "/"); {
	case len(split) == 6:
		apiInnerExecutor.configuration(w, req)
	default:
		logger.Logging(logger.DEBUG, "Unmatched url")
		common.MakeErrorResponse(w, errors.NotFoundURL{reqUrl})
	}
}

// configuration handles requests which is used to get/set a node configuration.
func (innerExecutorImpl) configuration(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, GET, POST) {
		return
	}

	response := make(map[string]interface{})
	var e error
	switch req.Method {
	case GET:
		response, e = configurationExecutor.GetConfiguration()
	case POST:
		var bodyStr string
		bodyStr, e = common.GetBodyFromReq(req)
		if e != nil {
			common.MakeErrorResponse(w, errors.InvalidYaml{"body is empty"})
			return
		}
		e = configurationExecutor.SetConfiguration(bodyStr)
	}
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}

	if req.Method != GET {
		response["result"] = "success"
	}

	common.MakeResponse(w, common.ChangeToJson(response))
}

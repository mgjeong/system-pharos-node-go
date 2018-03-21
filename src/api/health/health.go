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
package health

import (
	"api/common"
	"commons/logger"
	"commons/url"
	"controller/health"

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
	unregister(w http.ResponseWriter, req *http.Request)
}

type Executor struct{}
type innerExecutorImpl struct{}

var apiInnerExecutor apiInnerCommand
var healthExecutor health.Command

func init() {
	apiInnerExecutor = innerExecutorImpl{}
	healthExecutor = health.Executor{}
}

// Handling requests which is related to deployment functions.
func (Executor) Handle(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	switch reqUrl := req.URL.Path; {
	case strings.Contains(reqUrl, url.Unregister()):
		apiInnerExecutor.unregister(w, req)
	}
}

// Handling requests which is to unregister to manager service.
func (innerExecutorImpl) unregister(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, POST) {
		return
	}

	e := healthExecutor.Unregister()
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}

	response := make(map[string]interface{})
	response["result"] = "success"
	common.MakeResponse(w, common.ChangeToJson(response))
}

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
package resource

import (
	"api/common"
	"commons/errors"
	"commons/logger"
	"controller/monitoring/resource"
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
	hostResource(w http.ResponseWriter, req *http.Request)
	appResource(w http.ResponseWriter, req *http.Request, appId string)
}

type Executor struct{}
type innerExecutorImpl struct{}

var apiInnerExecutor apiInnerCommand
var resourceExecutor resource.Command

func init() {
	apiInnerExecutor = innerExecutorImpl{}
	resourceExecutor = resource.Executor
}

// Handling requests which is getting device resource or app's resource information
func (Executor) Handle(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	switch reqUrl, split := req.URL.Path, strings.Split(req.URL.Path, "/"); {
	case len(split) == 5: ///api/v1/monitoring/resource
		apiInnerExecutor.hostResource(w, req)
	case len(split) == 7: ///api/v1/monitoring/apps/{appid}/resource
		apiInnerExecutor.appResource(w, req, split[len(split)-2])
	default:
		logger.Logging(logger.DEBUG, "Unmatched url")
		common.MakeErrorResponse(w, errors.NotFoundURL{reqUrl})
	}
}

// Handling requests which is getting resources information
func (innerExecutorImpl) hostResource(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, GET) {
		return
	}

	response, e := resourceExecutor.GetHostResourceInfo()
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}
	common.MakeResponse(w, common.ChangeToJson(response))
}

// Handling requests which is getting app's resource information
func (innerExecutorImpl) appResource(w http.ResponseWriter, req *http.Request, appId string) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, GET) {
		return
	}

	response, e := resourceExecutor.GetAppResourceInfo(appId)
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}
	common.MakeResponse(w, common.ChangeToJson(response))
}

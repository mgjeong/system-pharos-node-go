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

// Package api provides web server for pharos-node
// and also provides functionality of request processing and response making.
package api

import (
	"api/common"
	deploymentapi "api/deployment"
	healthapi "api/health"
	resourceapi "api/monitoring/resource"
	configurationapi "api/configuration"
	"commons/errors"
	"commons/logger"
	"commons/url"
	"net/http"
	"strconv"
	"strings"
)

// Starting Web server service with address and port.
func RunNodeWebServer(addr string, port int) {
	logger.Logging(logger.DEBUG, "Start Pharos Node Web Server")
	logger.Logging(logger.DEBUG, "Listening "+addr+":"+strconv.Itoa(port))
	http.ListenAndServe(addr+":"+strconv.Itoa(port), &NodeApis)
}

var deploymentApiExecutor deploymentapi.Command
var healthApiExecutor healthapi.Command
var resourceApiExecutor resourceapi.Command
var configurationApiExecutor configurationapi.Command
var NodeApis Executor

type Executor struct{}

func init() {
	deploymentApiExecutor = deploymentapi.Executor{}
	healthApiExecutor = healthapi.Executor{}
	resourceApiExecutor = resourceapi.Executor{}
	configurationApiExecutor = configurationapi.Executor{}
}

// Implements of http serve interface.
// All of request is handled by this function.
func (Executor) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG, "receive msg", req.Method, req.URL.Path)
	defer logger.Logging(logger.DEBUG, "OUT")

	switch reqUrl := req.URL.Path; {
	default:
		logger.Logging(logger.DEBUG, "Unknown URL")
		common.MakeErrorResponse(w, errors.NotFoundURL{reqUrl})

	case !(strings.Contains(reqUrl, (url.Base()+url.Management())) ||
		strings.Contains(reqUrl, (url.Base()+url.Monitoring()))):
		logger.Logging(logger.DEBUG, "Unknown URL")
		common.MakeErrorResponse(w, errors.NotFoundURL{reqUrl})

	case strings.Contains(reqUrl, url.Unregister()):
		healthApiExecutor.Handle(w, req)

	case strings.Contains(reqUrl, url.Apps()):
		deploymentApiExecutor.Handle(w, req)

	case strings.Contains(reqUrl, url.Resource()):
		resourceApiExecutor.Handle(w, req)

	case strings.Contains(reqUrl, url.Configuration()):
		configurationApiExecutor.Handle(w, req)
	}
}

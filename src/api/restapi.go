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
	configurationapi "api/configuration"
	deploymentapi "api/deployment"
	deviceapi "api/device"
	healthapi "api/health"
	resourceapi "api/monitoring/resource"
	notificationapi "api/notification"
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
	http.ListenAndServe(addr+":"+strconv.Itoa(port), &NodeAPIs)
}

var deploymentAPIExecutor deploymentapi.Command
var healthAPIExecutor healthapi.Command
var resourceAPIExecutor resourceapi.Command
var configurationAPIExecutor configurationapi.Command
var deviceAPIExecutor deviceapi.Command
var notificationAPIExecutor notificationapi.Command
var NodeAPIs Executor

type Executor struct{}

func init() {
	deploymentAPIExecutor = deploymentapi.Executor{}
	healthAPIExecutor = healthapi.Executor{}
	resourceAPIExecutor = resourceapi.Executor{}
	configurationAPIExecutor = configurationapi.Executor{}
	deviceAPIExecutor = deviceapi.Executor{}
	notificationAPIExecutor = notificationapi.Executor{}
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
		strings.Contains(reqUrl, (url.Base()+url.Monitoring())) ||
		strings.Contains(reqUrl, (url.Base()+url.Notification()))):
		logger.Logging(logger.DEBUG, "Unknown URL")
		common.MakeErrorResponse(w, errors.NotFoundURL{reqUrl})

	case strings.Contains(reqUrl, url.Unregister()):
		healthAPIExecutor.Handle(w, req)

	case strings.Contains(reqUrl, url.Management()) &&
		strings.Contains(reqUrl, url.Apps()):
		deploymentAPIExecutor.Handle(w, req)

	case strings.Contains(reqUrl, url.Resource()):
		resourceAPIExecutor.Handle(w, req)

	case strings.Contains(reqUrl, url.Configuration()):
		configurationAPIExecutor.Handle(w, req)

	case strings.Contains(reqUrl, url.Device()):
		deviceAPIExecutor.Handle(w, req)

	case strings.Contains(reqUrl, url.Notification()):
		notificationAPIExecutor.Handle(w, req)
	}
}

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

package notification

import (
	"api/common"
	"api/notification/apps"
	"commons/errors"
	"commons/logger"
	URL "commons/url"
	"net/http"
	"strings"
)

type Command interface {
	Handle(w http.ResponseWriter, req *http.Request)
}

type Executor struct{}

var appsNotificationHandler apps.Command

func init() {
	appsNotificationHandler = apps.Executor{}
}

func (Executor) Handle(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG, "receive msg", req.Method, req.URL.Path)
	defer logger.Logging(logger.DEBUG, "OUT")

	switch url := req.URL.Path; {
	default:
		logger.Logging(logger.DEBUG, "Unknown URL")
		common.MakeErrorResponse(w, errors.NotFoundURL{url})

	case !strings.Contains(url, URL.Base()):
		logger.Logging(logger.DEBUG, "Unknown URL")
		common.MakeErrorResponse(w, errors.NotFoundURL{url})

	case strings.Contains(url, URL.Apps()):
		logger.Logging(logger.DEBUG, "Request Apps APIs")
		appsNotificationHandler.Handle(w, req)
	}
}

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
package apps

import (
	"api/common"
	"commons/errors"
	"commons/logger"
	URL "commons/url"
	"controller/notification/apps"
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
	subscribeEvent(w http.ResponseWriter, req *http.Request)
	unsubscribeEvent(w http.ResponseWriter, req *http.Request)
}

type Executor struct{}
type innerExecutorImpl struct{}

var apiInnerExecutor apiInnerCommand

var appsExecutor apps.Command

func init() {
	apiInnerExecutor = innerExecutorImpl{}
	appsExecutor = apps.Executor{}
}

func (Executor) Handle(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	url := strings.Replace(req.URL.Path, URL.Base()+URL.Notification()+URL.Apps(), "", -1)
	split := strings.Split(url, "/")

	switch len(split) {
	case 2:
		if "/"+split[1] == URL.Watch() {
			if req.Method == POST {
				apiInnerExecutor.subscribeEvent(w, req)
			} else if req.Method == DELETE {
				apiInnerExecutor.unsubscribeEvent(w, req)
			} else {
				common.MakeErrorResponse(w, errors.InvalidMethod{req.Method})
			}
		} else {
			common.MakeErrorResponse(w, errors.NotFoundURL{url})
		}
	default:
		logger.Logging(logger.DEBUG, "Unmatched url")
		common.MakeErrorResponse(w, errors.NotFoundURL{url})
	}
}

func (innerExecutorImpl) subscribeEvent(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	bodyStr, err := common.GetBodyFromReq(req)
	if err != nil {
		common.MakeErrorResponse(w, errors.InvalidYaml{"body is empty"})
		return
	}

	response, e := appsExecutor.SubscribeEvent(bodyStr)
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}
	common.MakeResponse(w, common.ChangeToJson(response))
}

func (innerExecutorImpl) unsubscribeEvent(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	bodyStr, err := common.GetBodyFromReq(req)
	if err != nil {
		common.MakeErrorResponse(w, errors.InvalidYaml{"body is empty"})
		return
	}

	response := make(map[string]interface{})
	e := appsExecutor.UnsubscribeEvent(bodyStr)
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}
	common.MakeResponse(w, common.ChangeToJson(response))
}

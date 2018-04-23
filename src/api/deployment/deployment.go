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
package deployment

import (
	"api/common"
	"commons/errors"
	"commons/logger"
	"commons/url"
	"controller/deployment"
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
	deploy(w http.ResponseWriter, req *http.Request)
	app(w http.ResponseWriter, req *http.Request, appId string)
	apps(w http.ResponseWriter, req *http.Request)
	update(w http.ResponseWriter, req *http.Request, appId string)
	stop(w http.ResponseWriter, req *http.Request, appId string)
	start(w http.ResponseWriter, req *http.Request, appId string)
	events(w http.ResponseWriter, req *http.Request, appId string)
}

type Executor struct{}
type innerExecutorImpl struct{}

var apiInnerExecutor apiInnerCommand
var deploymentExecutor deployment.Command

func init() {
	apiInnerExecutor = innerExecutorImpl{}
	deploymentExecutor = deployment.Executor
}

// Handling requests which is related to deployment functions.
func (Executor) Handle(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	switch reqUrl, split := req.URL.Path, strings.Split(req.URL.Path, "/"); {
	case len(split) == 7:
		switch appId := split[len(split)-2]; {
		case strings.HasSuffix(reqUrl, url.Start()):
			apiInnerExecutor.start(w, req, appId)

		case strings.HasSuffix(reqUrl, url.Stop()):
			apiInnerExecutor.stop(w, req, appId)

		case strings.HasSuffix(reqUrl, url.Update()):
			apiInnerExecutor.update(w, req, appId)

		case strings.HasSuffix(reqUrl, url.Events()):
			apiInnerExecutor.events(w, req, appId)

		default:
			logger.Logging(logger.DEBUG, "Unmatched url")
			common.MakeErrorResponse(w, errors.NotFoundURL{reqUrl})
		}
	case len(split) == 6:
		if strings.Contains(req.URL.Path, url.Deploy()) {
			apiInnerExecutor.deploy(w, req)
		} else {
			apiInnerExecutor.app(w, req, split[len(split)-1])
		}
	case len(split) == 5:
		apiInnerExecutor.apps(w, req)
	default:
		logger.Logging(logger.DEBUG, "Unmatched url")
		common.MakeErrorResponse(w, errors.NotFoundURL{reqUrl})
	}
}

// Handling requests which is deploy(pulling images) app to the target.
func (innerExecutorImpl) deploy(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, POST) {
		return
	}

	bodyStr, err := common.GetBodyFromReq(req)
	if err != nil {
		common.MakeErrorResponse(w, errors.InvalidYaml{"body is empty"})
		return
	}

	response, e := deploymentExecutor.DeployApp(bodyStr, parseQuery(req))
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}

	appId := response["id"].(string)
	w.Header().Set("Location", url.Base()+url.Management()+url.Apps()+"/"+appId)

	common.MakeResponse(w, common.ChangeToJson(response))
}

// Handling requests which is getting app information
// and update app description, delete app on the target.
func (innerExecutorImpl) app(w http.ResponseWriter, req *http.Request, appId string) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, GET, POST, DELETE) {
		return
	}

	response := make(map[string]interface{})
	var e error
	switch req.Method {
	case GET:
		response, e = deploymentExecutor.App(appId)
	case POST:
		var bodyStr string
		bodyStr, e = common.GetBodyFromReq(req)
		if e != nil {
			common.MakeErrorResponse(w, errors.InvalidYaml{"body is empty"})
			return
		}
		e = deploymentExecutor.UpdateAppInfo(appId, bodyStr)
	case DELETE:
		e = deploymentExecutor.DeleteApp(appId)
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

// Handling requests which is getting all of app informations.
func (innerExecutorImpl) apps(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, GET) {
		return
	}
	response, e := deploymentExecutor.Apps()
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}

	common.MakeResponse(w, common.ChangeToJson(response))
}

// Handling requests which is updating image from registry.
func (innerExecutorImpl) update(w http.ResponseWriter, req *http.Request, appId string) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, POST) {
		return
	}

	e := deploymentExecutor.UpdateApp(appId, parseQuery(req))
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}

	response := make(map[string]interface{})
	response["result"] = "success"
	common.MakeResponse(w, common.ChangeToJson(response))
}

// Handling requests which is stop the app.
func (innerExecutorImpl) stop(w http.ResponseWriter, req *http.Request, appId string) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, POST) {
		return
	}
	e := deploymentExecutor.StopApp(appId)
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}

	response := make(map[string]interface{})
	response["result"] = "success"
	common.MakeResponse(w, common.ChangeToJson(response))
}

// Handling requests which is start the app.
func (innerExecutorImpl) start(w http.ResponseWriter, req *http.Request, appId string) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, POST) {
		return
	}
	e := deploymentExecutor.StartApp(appId)
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}

	response := make(map[string]interface{})
	response["result"] = "success"
	common.MakeResponse(w, common.ChangeToJson(response))
}

// Handling requests which is event the app.
func (innerExecutorImpl) events(w http.ResponseWriter, req *http.Request, appId string) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !common.CheckSupportedMethod(w, req.Method, POST) {
		return
	}

	var bodyStr string
	bodyStr, e := common.GetBodyFromReq(req)
	if e != nil {
		common.MakeErrorResponse(w, errors.InvalidYaml{"body is empty"})
		return
	}

	e = deploymentExecutor.HandleEvents(appId, bodyStr)
	if e != nil {
		common.MakeErrorResponse(w, e)
		return
	}

	response := make(map[string]interface{})
	response["result"] = "success"
	common.MakeResponse(w, common.ChangeToJson(response))
}

func parseQuery(req *http.Request) map[string]interface{} {
	query := make(map[string]interface{})

	keys := req.URL.Query()
	if len(keys) == 0 {
		return nil
	}

	for key, value := range req.URL.Query() {
		query[key] = value
	}

	return query
}
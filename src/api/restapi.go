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

// Package api provides web server for Service Deployment Agent
// and also provides functionality of request processing and response making.
package api

import (
	"commons/errors"
	"commons/logger"
	"commons/url"
	dep "controller/deployment"
	reg "controller/registration"
	res "controller/resource"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type ResponseType map[string]interface{}

// Starting Web server service with address and port.
func RunSDAWebServer(addr string, port int) {
	logger.Logging(logger.DEBUG, "Start Management Agent Web Server")
	logger.Logging(logger.DEBUG, "Listening "+addr+":"+strconv.Itoa(port))
	http.ListenAndServe(addr+":"+strconv.Itoa(port), &_SDAApis)
}

var deploymentExecutor dep.Command
var registrationExecutor reg.Command
var resourceExecutor res.Command

var _SDAApis _SDAApisHandler

type _SDAApisHandler struct{}

const (
	GET    string = "GET"
	PUT    string = "PUT"
	POST   string = "POST"
	DELETE string = "DELETE"
)

func init() {
	deploymentExecutor = dep.Executor
	registrationExecutor = reg.Executor{}
	resourceExecutor = res.Executor
}

// Implements of http serve interface.
// All of request is handled by this function.
func (sda *_SDAApisHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG, "receive msg", req.Method, req.URL.Path)
	defer logger.Logging(logger.DEBUG, "OUT")

	switch reqUrl := req.URL.Path; {
	default:
		logger.Logging(logger.DEBUG, "Unknown URL")
		makeErrorResponse(w, errors.NotFoundURL{reqUrl})

	case !strings.Contains(reqUrl, url.Base()):
		logger.Logging(logger.DEBUG, "Unknown URL")
		makeErrorResponse(w, errors.NotFoundURL{reqUrl})

	case strings.Contains(reqUrl, url.Unregister()):
		sda.handleUnregister(w, req)

	case strings.Contains(reqUrl, url.Deploy()):
		sda.handleDeploy(w, req)

	case strings.Contains(reqUrl, url.Apps()):
		sda.handleApps(w, req)

	case strings.Contains(reqUrl, url.Resource()):
		sda.handleResource(w, req)
	}
}

// Handling requests which is to unregister to manager service.
func (sda *_SDAApisHandler) handleUnregister(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !checkSupportedMethod(w, req.Method, POST) {
		return
	}

	e := registrationExecutor.Unregister()
	if e != nil {
		makeErrorResponse(w, e)
		return
	}

	response := make(map[string]interface{})
	response["result"] = "success"
	makeResponse(w, changeToJson(response))
}

// Handling requests which is deploy(pulling images) app to the target.
func (sda *_SDAApisHandler) handleDeploy(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !checkSupportedMethod(w, req.Method, POST) {
		return
	}

	bodyStr, err := getBodyFromReq(req)
	if err != nil {
		makeErrorResponse(w, errors.InvalidYaml{"body is empty"})
		return
	}

	response, e := deploymentExecutor.DeployApp(bodyStr)
	if e != nil {
		makeErrorResponse(w, e)
		return
	}

	appId := response["id"].(string)
	w.Header().Set("Location", url.Base()+url.Apps()+"/"+appId)

	makeResponse(w, changeToJson(response))
}

// Handling requests which is start and stop, update the apps.
func (sda *_SDAApisHandler) handleApps(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	switch reqUrl, split := req.URL.Path, strings.Split(req.URL.Path, "/"); {
	case len(split) == 6:
		switch appId := split[len(split)-2]; {
		case strings.HasSuffix(reqUrl, url.Start()):
			sda.start(w, req, appId)

		case strings.HasSuffix(reqUrl, url.Stop()):
			sda.stop(w, req, appId)

		case strings.HasSuffix(reqUrl, url.Update()):
			sda.update(w, req, appId)

		default:
			logger.Logging(logger.DEBUG, "Unmatched url")
			makeErrorResponse(w, errors.NotFoundURL{reqUrl})
		}
	case len(split) == 5:
		sda.app(w, req, split[len(split)-1])

	case len(split) == 4:
		sda.apps(w, req)

	default:
		logger.Logging(logger.DEBUG, "Unmatched url")
		makeErrorResponse(w, errors.NotFoundURL{reqUrl})
	}
}

// Handling requests which is getting device resource or performance information
func (sda *_SDAApisHandler) handleResource(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	switch reqUrl, split := req.URL.Path, strings.Split(req.URL.Path, "/"); {
	case len(split) == 4:
		sda.resource(w, req)
	case len(split) == 5:
		sda.performance(w, req)

	default:
		logger.Logging(logger.DEBUG, "Unmatched url")
		makeErrorResponse(w, errors.NotFoundURL{reqUrl})
	}
}

// Handling requests which is getting app information
// and update app description, delete app on the target.
func (sda *_SDAApisHandler) app(w http.ResponseWriter, req *http.Request, appId string) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !checkSupportedMethod(w, req.Method, GET, POST, DELETE) {
		return
	}

	response := make(map[string]interface{})
	var e error
	switch req.Method {
	case GET:
		response, e = deploymentExecutor.App(appId)
	case POST:
		var bodyStr string
		bodyStr, e = getBodyFromReq(req)
		if e != nil {
			makeErrorResponse(w, errors.InvalidYaml{"body is empty"})
			return
		}
		e = deploymentExecutor.UpdateAppInfo(appId, bodyStr)
	case DELETE:
		e = deploymentExecutor.DeleteApp(appId)
	}
	if e != nil {
		makeErrorResponse(w, e)
		return
	}

	if req.Method != GET {
		response["result"] = "success"
	}

	makeResponse(w, changeToJson(response))
}

// Handling requests which is getting all of app informations.
func (sda *_SDAApisHandler) apps(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !checkSupportedMethod(w, req.Method, GET) {
		return
	}
	response, e := deploymentExecutor.Apps()
	if e != nil {
		makeErrorResponse(w, e)
		return
	}

	makeResponse(w, changeToJson(response))
}

// Handling requests which is updating image from registry.
func (sda *_SDAApisHandler) update(w http.ResponseWriter, req *http.Request, appId string) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !checkSupportedMethod(w, req.Method, POST) {
		return
	}
	e := deploymentExecutor.UpdateApp(appId)
	if e != nil {
		makeErrorResponse(w, e)
		return
	}

	response := make(map[string]interface{})
	response["result"] = "success"
	makeResponse(w, changeToJson(response))
}

// Handling requests which is stop the app.
func (sda *_SDAApisHandler) stop(w http.ResponseWriter, req *http.Request, appId string) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !checkSupportedMethod(w, req.Method, POST) {
		return
	}
	e := deploymentExecutor.StopApp(appId)
	if e != nil {
		makeErrorResponse(w, e)
		return
	}

	response := make(map[string]interface{})
	response["result"] = "success"
	makeResponse(w, changeToJson(response))
}

// Handling requests which is start the app.
func (sda *_SDAApisHandler) start(w http.ResponseWriter, req *http.Request, appId string) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !checkSupportedMethod(w, req.Method, POST) {
		return
	}
	e := deploymentExecutor.StartApp(appId)
	if e != nil {
		makeErrorResponse(w, e)
		return
	}

	response := make(map[string]interface{})
	response["result"] = "success"
	makeResponse(w, changeToJson(response))
}

// Handling requests which is getting resources information
func (sda *_SDAApisHandler) resource(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !checkSupportedMethod(w, req.Method, GET) {
		return
	}

	response, e := resourceExecutor.GetResourceInfo()

	if e != nil {
		makeErrorResponse(w, e)
		return
	}
	makeResponse(w, changeToJson(response))
}

// Handling requests which is getting performance information
func (sda *_SDAApisHandler) performance(w http.ResponseWriter, req *http.Request) {
	logger.Logging(logger.DEBUG)
	defer logger.Logging(logger.DEBUG, "OUT")

	if !checkSupportedMethod(w, req.Method, GET) {
		return
	}

	response, e := resourceExecutor.GetPerformanceInfo()

	if e != nil {
		makeErrorResponse(w, e)
		return
	}
	makeResponse(w, changeToJson(response))
}

// Making non succeed response by error type.
func makeErrorResponse(w http.ResponseWriter, err error) {
	var code int

	switch err.(type) {

	case errors.NotFoundURL:
		code = http.StatusNotFound
		
	case errors.InvalidMethod:
		code = http.StatusMethodNotAllowed

	case errors.InvalidYaml, errors.InvalidAppId,
		errors.InvalidParam, errors.NotFoundImage,
		errors.AlreadyAllocatedPort, errors.AlreadyUsedName,
		errors.InvalidContainerName:
		code = http.StatusBadRequest

	case errors.IOError:
		code = http.StatusInternalServerError

	case errors.ConnectionError, errors.NotFound:
		code = http.StatusServiceUnavailable

	case errors.AlreadyReported:
		code = http.StatusAlreadyReported

	default:
		code = http.StatusInternalServerError
	}

	logger.Logging(logger.DEBUG, "Send response", strconv.Itoa(code), err.Error())

	response := make(map[string]string)
	response["message"] = err.Error()
	data, err := json.Marshal(response)

	w.WriteHeader(code)
	w.Write(data)
}

// Making response for succeed case.
func makeResponse(w http.ResponseWriter, data []byte) {
	if data == nil {
		retOk := make(map[string]string)
		retOk["message"] = "OK"
		var err error
		data, err = json.Marshal(retOk)
		if err != nil {
			makeErrorResponse(w, errors.IOError{"data convert fail"})
			return
		}
	}
	logger.Logging(logger.DEBUG, "Send response : 200")
	w.WriteHeader(http.StatusOK)
	writeSuccess(w, data)
}

// Setting body of response.
func writeSuccess(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(data))
}

// Checking the can handle the request method, if not, make error response.
// Will return
//  true with support method and
//  false with non-support method.
func checkSupportedMethod(w http.ResponseWriter, reqMethod string, methods ...string) bool {
	for _, method := range methods {
		if method == reqMethod {
			return true
		}
	}
	logger.Logging(logger.DEBUG, "UnSupported method")
	makeErrorResponse(w, errors.InvalidMethod{reqMethod})
	return false
}

// Convert to Json format by map.
func changeToJson(src ResponseType) []byte {
	dst, err := json.Marshal(src)
	if err != nil {
		logger.Logging(logger.DEBUG, "Can't convert to Json")
		return nil
	}
	return dst
}

// Parsing body from request.
func getBodyFromReq(req *http.Request) (string, error) {
	if req.Body == nil {
		logger.Logging(logger.DEBUG, "Body is empty")
		return "", errors.InvalidParam{}
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Logging(logger.DEBUG, "Can't parse requested body")
		return "", errors.InvalidParam{}
	}
	return string(body), nil
}

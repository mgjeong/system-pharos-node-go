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
package common

import (
	"commons/errors"
	"commons/logger"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	GET    string = "GET"
	PUT    string = "PUT"
	POST   string = "POST"
	DELETE string = "DELETE"
)

type ResponseType map[string]interface{}

// Making non succeed response by error type.
func MakeErrorResponse(w http.ResponseWriter, err error) {
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
func MakeResponse(w http.ResponseWriter, data []byte) {
	if data == nil {
		retOk := make(map[string]string)
		retOk["message"] = "OK"
		var err error
		data, err = json.Marshal(retOk)
		if err != nil {
			MakeErrorResponse(w, errors.IOError{"data convert fail"})
			return
		}
	}
	logger.Logging(logger.DEBUG, "Send response : 200")
	w.WriteHeader(http.StatusOK)
	WriteSuccess(w, data)
}

// Setting body of response.
func WriteSuccess(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(data))
}

// Checking the can handle the request method, if not, make error response.
// Will return
//  true with support method and
//  false with non-support method.
func CheckSupportedMethod(w http.ResponseWriter, reqMethod string, methods ...string) bool {
	for _, method := range methods {
		if method == reqMethod {
			return true
		}
	}
	logger.Logging(logger.DEBUG, "UnSupported method")
	MakeErrorResponse(w, errors.InvalidMethod{reqMethod})
	return false
}

// Convert to Json format by map.
func ChangeToJson(src ResponseType) []byte {
	dst, err := json.Marshal(src)
	if err != nil {
		logger.Logging(logger.DEBUG, "Can't convert to Json")
		return nil
	}
	return dst
}

// Parsing body from request.
func GetBodyFromReq(req *http.Request) (string, error) {
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

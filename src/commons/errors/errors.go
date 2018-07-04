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

// Package commons/errors defines error cases of Pharos Node.
package errors

// Struct InvalidParam will be used for return case of error
// which value of unknown or invalid type, range in the parameters.
type InvalidParam struct {
	Msg string
}

// Set error message of InvalidParam.
func (e *InvalidParam) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of InvalidParam for error interface.
func (e InvalidParam) Error() string { return "invalid parameter : " + e.Msg }

// Struct InvalidJSON will be used for return case of error
// which value of malformed json format.
type InvalidJSON struct {
	Msg string
}

// Set error message of InvalidJSON.
func (e *InvalidJSON) SetMsg(msg string) {
	e.Msg = msg
}

// Error sets an error message of InvalidJSON.
func (e InvalidJSON) Error() string {
	return "invalid json format: " + e.Msg
}

// Struct InvalidJSON will be used for return case of error
// which value of invalid ObjectId.
type InvalidObjectId struct {
	Msg string
}

// Error sets an error message of InvalidObjectId.
func (e InvalidObjectId) Error() string {
	return "invalid objectId: " + e.Msg
}

// Set error message of InvalidObjectId.
func (e *InvalidObjectId) SetMsg(msg string) {
	e.Msg = msg
}

// Struct NotFound will be used for return case of error
// which object or target can not found.
type NotFound struct {
	Msg string
}

// Set error message of NotFound.
func (e *NotFound) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of NotFound for error interface.
func (e NotFound) Error() string {
	return "not find target : " + e.Msg
}

// Struct InvalidYaml will be used for return case of error
// which input yaml form is invalid.
type InvalidYaml struct {
	Msg string
}

// Set error message of InvalidYaml.
func (e *InvalidYaml) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of InvalidYaml for error interface.
func (e InvalidYaml) Error() string {
	return "invalid yaml file : " + e.Msg
}

// Struct ConnectionError will be used for return case of error
// which connection failed with db server and docker daemon.
type ConnectionError struct {
	Msg string
}

// Set error message of ConnectionError.
func (e *ConnectionError) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of ConnectionError for error interface.
func (e ConnectionError) Error() string {
	return "can not connect : " + e.Msg
}

// Struct InvalidAppId will be used for return case of error
// which AppId of request is not valid.
type InvalidAppId struct {
	Msg string
}

// Set error message of InvalidAppId.
func (e *InvalidAppId) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of InvalidAppId for error interface.
func (e InvalidAppId) Error() string {
	return "invalid app id : " + e.Msg
}

// Struct InvalidMethod will be used for return case of error
// which method of request is not provide.
type InvalidMethod struct {
	Msg string
}

// Set error message of InvalidMethod.
func (e *InvalidMethod) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of InvalidMethod for error interface.
func (e InvalidMethod) Error() string {
	return "invalid method : " + e.Msg
}

// Struct NotFoundURL will be used for return case of error
// which url of request is not exist.
type NotFoundURL struct {
	Msg string
}

// Set error message of NotFoundURL.
func (e *NotFoundURL) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of NotFoundURL for error interface.
func (e NotFoundURL) Error() string {
	return "unsupported url : " + e.Msg
}

// Struct IOError will be used for return case of error
// which IO operaion fail like file operation failed or json marshalling failed.
type IOError struct {
	Msg string
}

// Set error message of IOError.
func (e *IOError) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of IOError for error interface.
func (e IOError) Error() string {
	return "io error : " + e.Msg
}

// Struct NotFoundImage will be used for return case of error
// which image can not found on docker daemon or docker registry.
type NotFoundImage struct {
	Msg string
}

// Set error message of NotFoundImage.
func (e *NotFoundImage) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of NotFoundImage for error interface.
func (e NotFoundImage) Error() string {
	return "unsupported url : " + e.Msg
}

// Struct AlreadyReported will be used for return case of error
// when operation alreay processing before.
type AlreadyReported struct {
	Msg string
}

// Set error message of AlreadyReported.
func (e *AlreadyReported) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of AlreadyReported for error interface.
func (e AlreadyReported) Error() string {
	return "already done processing : " + e.Msg
}

// Struct AlreadyUsedName will be used for return case of error
// when container name within yaml description already used.
type AlreadyUsedName struct {
	Msg string
}

// Set error message of AlreadyUseName.
func (e *AlreadyUsedName) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of AlreadyUsedName for error interface.
func (e AlreadyUsedName) Error() string {
	return "already used container name : " + e.Msg
}

// Struct InvalidContainerName will be used for return case of error
// when container name within yaml description violated with naming rule.
type InvalidContainerName struct {
	Msg string
}

// Set error message of InvalidContainerName.
func (e *InvalidContainerName) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of InvalidContainerName for error interface.
func (e InvalidContainerName) Error() string {
	return "invalid container name : " + e.Msg
}

// Struct InvalidContainerName will be used for return case of error
// when port number within yaml description is already binded.
type AlreadyAllocatedPort struct {
	Msg string
}

// Set error message of AlreadyAllocatedPort.
func (e *AlreadyAllocatedPort) SetMsg(msg string) {
	e.Msg = msg
}

// Implements of Error functionality of AlreadyAllocatedPort for error interface.
func (e AlreadyAllocatedPort) Error() string {
	return "already allocated port : " + e.Msg
}

// Struct Unknown will be used for return case of error
// which not a defined errors.
type Unknown struct {
	Msg string
}

// Implements of Error functionality of Unknown for error interface.
func (e Unknown) Error() string {
	return "unknown error : " + e.Msg
}

// Set error message of Unknown.
func (e *Unknown) SetMsg(msg string) {
	e.Msg = msg
}

// Struct InternalServerError will be used for return case of error
// which a generic error, given when an unexpected condition was encountered
// and no more specific message is suitable.
type InternalServerError struct {
	Msg string
}

// Error sets an error message of InternalServerError.
func (e InternalServerError) Error() string {
	return "internal server error: " + e.Msg
}

// Set error message of InternalServerError.
func (e *InternalServerError) SetMsg(msg string) {
	e.Msg = msg
}

// Struct DBConnectionError will be used for return case of error
// which connection failed with db server.
type DBConnectionError struct {
	Msg string
}

// Error sets an error message of DBConnectionError.
func (e DBConnectionError) Error() string {
	return "db connection failed: " + e.Msg
}

// Set error message of DBConnectionError.
func (e *DBConnectionError) SetMsg(msg string) {
	e.Msg = msg
}

// Struct DBOperationError will be used for return case of error
// which db operation failed(e.g., insert, update, delete).
type DBOperationError struct {
	Msg string
}

// Error sets an error message of DBOperationError.
func (e DBOperationError) Error() string {
	return "db operation failed: " + e.Msg
}

// Set error message of DBOperationError.
func (e *DBOperationError) SetMsg(msg string) {
	e.Msg = msg
}

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
	"commons/errors"
	dockermocks "controller/dockercontroller/mocks"
	appmocks "controller/monitoring/apps/mocks"
	dbmocks "db/bolt/service/mocks"
	"github.com/golang/mock/gomock"
	"os"
	"reflect"
	"testing"
)

const (
	COMPOSE_FILE                        = "docker-compose.yaml"
	APP_ID                              = "000000000000000000000000"
	DESCRIPTION_JSON_WITHOUT_SERVICE    = "{\"no_services\":{\"test_service\":{\"image\":\"test_image:0.2\"}},\"version\":\"2\"}"
	WRONG_DESCRIPTION_JSON              = "{{{{services:\n  test_service:\n    image: test_image:0.2\nversion: \"2\""
	WRONG_INSPECT_RETURN_MSG            = "error_[{\"State\": {\"Status\": \"running\", \"ExitCode\": \"0\"}}]"
	OLD_TAG                             = "1.0"
	NEW_TAG                             = "2.0"
	REPOSITORY_WITH_PORT_IMAGE          = "test_url:5000/test"
	REPOSITORY_WITH_PORT_IMAGE_WITH_TAG = "test_url:5000/test" + ":" + OLD_TAG
	APP_STATE                           = "STATE"
	SERVICE_NAME                        = "test_service"
	CONTAINER_NAME                      = "test_container"
	DESCRIPTION_JSON                    = "{\"services\":{\"" + SERVICE + "\":{\"image\":\"" + REPOSITORY_WITH_PORT_IMAGE + ":" + OLD_TAG + "\"}},\"version\":\"2\"}"
	DESCRIPTION_YAML                    = "services:\n  " + SERVICE + ":\n    image: " + REPOSITORY_WITH_PORT_IMAGE + ":" + OLD_TAG + "\nversion: \"2\"\n"
	REPOSITORY_WITH_PORT_IMAGE_DIGEST   = REPOSITORY_WITH_PORT_IMAGE + "@" + "sha256:1234567890"
	SERVICE                             = "test_service"
	CONTAINER                           = "test_container"
	ORIGIN_DESCRIPTION_JSON             = "{\"services\":{\"" + SERVICE + "\":{\"image\":\"" + REPOSITORY_WITH_PORT_IMAGE + ":" + OLD_TAG + "\"}},\"version\":\"2\"}"
	UPDATED_DESCRIPTION_JSON            = "{\"services\":{\"" + SERVICE + "\":{\"image\":\"" + REPOSITORY_WITH_PORT_IMAGE + ":" + NEW_TAG + "\"}},\"version\":\"2\"}"
	FULL_IMAGE_NAME                     = REPOSITORY_WITH_PORT_IMAGE + ":" + NEW_TAG
	NONE_EVENT                          = "none"
	CONTAINER_ID                        = 1234
	SERVICE_PORT                        = 1234
	SERVICE_STATUS                      = "running"
	EXIT_CODE_VALUE                     = "0"
	EVENT_REPOSITORY                    = "localhost:5000/test_repo"
	EVENT_TAG                           = "latest"
	UPDATE_EVENTS_JSON                  = `{"events":[{"action": "push","target": {"repository": "test_repo","tag": "latest"},"request": {"addr": "0.0.0.0:8888","host": "localhost:5000"}}]}`
	DELETE_EVENTS_JSON                  = `{"events":[{"action": "delete","target": {"repository": "test_repo","tag": "latest"},"request": {"addr": "0.0.0.0:8888","host": "localhost:5000"}}]}`
	INVALID_JSON_FORMAT                 = "invalid_json_format"
	REPODIGEST                          = "test@sha256test"
	IMAGE_ID                            = "abcd"
)

var (
	INSPECT_RETURN_MSG = map[string]interface{}{
		"cid":      CONTAINER_ID,
		"ports":    SERVICE_PORT,
		"status":   SERVICE_STATUS,
		"exitcode": EXIT_CODE_VALUE,
	}

	PS_EXPECT_RETURN = []map[string]string{
		{
			"Name": CONTAINER,
		},
	}

	DB_GET_APP_WITH_EXITED_STATE_OBJ = map[string]interface{}{
		"id":          APP_ID,
		"state":       EXITED_STATE,
		"description": ORIGIN_DESCRIPTION_JSON,
		"images": []map[string]interface{}{
			{
				"name": REPOSITORY_WITH_PORT_IMAGE,
				"changes": map[string]interface{}{
					"tag":   NEW_TAG,
					"state": "update",
				},
			},
		},
	}

	DB_GET_APP_UPDATING_OBJ = map[string]interface{}{
		"id":          APP_ID,
		"state":       UPDATING_STATE,
		"description": ORIGIN_DESCRIPTION_JSON,
		"images": []map[string]interface{}{
			{
				"name": REPOSITORY_WITH_PORT_IMAGE,
				"changes": map[string]interface{}{
					"tag":   NEW_TAG,
					"state": "update",
				},
			},
		},
	}

	DB_GET_APP_OBJ = map[string]interface{}{
		"id":          APP_ID,
		"state":       RUNNING_STATE,
		"description": ORIGIN_DESCRIPTION_JSON,
		"images": []map[string]interface{}{
			{
				"name": REPOSITORY_WITH_PORT_IMAGE,
				"changes": map[string]interface{}{
					"tag":   NEW_TAG,
					"state": "update",
				},
			},
		},
	}

	DB_GET_APP_UPDATED_OBJ = map[string]interface{}{
		"id":          APP_ID,
		"state":       RUNNING_STATE,
		"description": UPDATED_DESCRIPTION_JSON,
		"images": []map[string]interface{}{
			{
				"name": REPOSITORY_WITH_PORT_IMAGE,
				"changes": map[string]interface{}{
					"tag":   NEW_TAG,
					"state": "update",
				},
			},
		},
	}

	APP_OBJ = map[string]interface{}{
		"state":       RUNNING_STATE,
		"description": DESCRIPTION_YAML,
		"images": []map[string]interface{}{
			{
				"name": REPOSITORY_WITH_PORT_IMAGE,
				"changes": map[string]interface{}{
					"tag":   NEW_TAG,
					"state": "update",
				},
			},
		},
		"services": []map[string]interface{}{
			{
				"name":  SERVICE,
				"cid":   CONTAINER_ID,
				"ports": SERVICE_PORT,
				"state": map[string]interface{}{
					"status":   SERVICE_STATUS,
					"exitcode": EXIT_CODE_VALUE,
				},
			},
		},
	}

	DB_OBJ = map[string]interface{}{
		"id": APP_ID,
	}

	DB_GET_OBJ = map[string]interface{}{
		"description": DESCRIPTION_JSON,
		"state":       RUNNING_STATE,
	}

	DB_OBJs = []map[string]interface{}{
		map[string]interface{}{
			"id":          APP_ID,
			"state":       RUNNING_STATE,
			"description": DESCRIPTION_JSON,
		},
	}

	WRONG_DB_OBJ = map[string]interface{}{
		"id": APP_ID,
	}

	WRONG_DB_GET_OBJ = map[string]interface{}{
		"description": WRONG_DESCRIPTION_JSON,
		"state":       RUNNING_STATE,
	}

	WRONG_DB_OBJs = []map[string]interface{}{
		map[string]interface{}{
			"id":    APP_ID,
			"state": RUNNING_STATE,
		},
	}

	DB_GET_OBJ_WITHOUT_SERVICE = map[string]interface{}{
		"description": DESCRIPTION_JSON_WITHOUT_SERVICE,
		"state":       RUNNING_STATE,
	}

	NotFoundError    = errors.NotFound{}
	ConnectionError  = errors.ConnectionError{}
	InvalidYamlError = errors.InvalidYaml{}
	UnknownError     = errors.Unknown{}
)

func TestCalledDeployApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON, RUNNING_STATE).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().EnableEventMonitoring(gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), true).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(APP_ID, gomock.Any(), SERVICE_NAME).Return(PS_EXPECT_RETURN, nil),
		dockerExecutorMockObj.EXPECT().GetContainerConfigByName(gomock.Any()).Return(INSPECT_RETURN_MSG, nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	res, err := Executor.DeployApp(DESCRIPTION_YAML, nil)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
	compareReturnVal := map[string]interface{}{
		"id":          APP_ID,
		"state":       RUNNING_STATE,
		"description": DESCRIPTION_YAML,
		"images": []map[string]interface{}{
			{
				"name": REPOSITORY_WITH_PORT_IMAGE,
				"changes": map[string]interface{}{
					"tag":   NEW_TAG,
					"state": "update",
				},
			},
		},
		"services": []map[string]interface{}{
			{
				"name":  SERVICE,
				"cid":   CONTAINER_ID,
				"ports": SERVICE_PORT,
				"state": map[string]interface{}{
					"status":   SERVICE_STATUS,
					"exitcode": EXIT_CODE_VALUE,
				},
			},
		},
	}

	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Errorf("Expected result : %v, Actual Result : %v", compareReturnVal, res)
	}

	os.RemoveAll(COMPOSE_FILE)
}

func TestCalledDeployAppWithEventQuery_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testEventID := "testeventid"
	testQuery := map[string]interface{}{
		EVENTID: []string{testEventID},
	}

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON, RUNNING_STATE).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().EnableEventMonitoring(gomock.Any(), gomock.Any()).Return(nil),
		appExecutorMockObj.EXPECT().GetEventChannel().Return(nil),
		dockerExecutorMockObj.EXPECT().UpWithEvent(gomock.Any(), gomock.Any(), testEventID, nil).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(APP_ID, gomock.Any(), SERVICE_NAME).Return(PS_EXPECT_RETURN, nil),
		dockerExecutorMockObj.EXPECT().GetContainerConfigByName(gomock.Any()).Return(INSPECT_RETURN_MSG, nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	res, err := Executor.DeployApp(DESCRIPTION_YAML, testQuery)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
	compareReturnVal := map[string]interface{}{
		"id":          APP_ID,
		"state":       RUNNING_STATE,
		"description": DESCRIPTION_YAML,
		"images": []map[string]interface{}{
			{
				"name": REPOSITORY_WITH_PORT_IMAGE,
				"changes": map[string]interface{}{
					"tag":   NEW_TAG,
					"state": "update",
				},
			},
		},
		"services": []map[string]interface{}{
			{
				"name":  SERVICE,
				"cid":   CONTAINER_ID,
				"ports": SERVICE_PORT,
				"state": map[string]interface{}{
					"status":   SERVICE_STATUS,
					"exitcode": EXIT_CODE_VALUE,
				},
			},
		},
	}

	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Errorf("Expected result : %v, Actual Result : %v", compareReturnVal, res)
	}

	os.RemoveAll(COMPOSE_FILE)
}

func TestCalledDeployAppWhenAlreadyInstalled_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON, RUNNING_STATE).Return(DB_GET_APP_OBJ, errors.AlreadyReported{}),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(APP_ID, gomock.Any(), SERVICE_NAME).Return(PS_EXPECT_RETURN, nil),
		dockerExecutorMockObj.EXPECT().GetContainerConfigByName(gomock.Any()).Return(INSPECT_RETURN_MSG, nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	res, err := Executor.DeployApp(DESCRIPTION_YAML, nil)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
	compareReturnVal := map[string]interface{}{
		"id":          APP_ID,
		"state":       RUNNING_STATE,
		"description": DESCRIPTION_YAML,
		"images": []map[string]interface{}{
			{
				"name": REPOSITORY_WITH_PORT_IMAGE,
				"changes": map[string]interface{}{
					"tag":   NEW_TAG,
					"state": "update",
				},
			},
		},
		"services": []map[string]interface{}{
			{
				"name":  SERVICE,
				"cid":   CONTAINER_ID,
				"ports": SERVICE_PORT,
				"state": map[string]interface{}{
					"status":   SERVICE_STATUS,
					"exitcode": EXIT_CODE_VALUE,
				},
			},
		},
	}

	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Errorf("Expected result : %v, Actual Result : %v", compareReturnVal, res)
	}

	os.RemoveAll(COMPOSE_FILE)
}

func TestCalledDeployAppWhenFailedToSetEventChannelFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON, RUNNING_STATE).Return(DB_OBJ, nil),
		appExecutorMockObj.EXPECT().EnableEventMonitoring(gomock.Any(), gomock.Any()).Return(UnknownError),
		dbExecutorMockObj.EXPECT().DeleteApp(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	_, err := Executor.DeployApp(DESCRIPTION_YAML, nil)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknowError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE)
}

func TestCalledDeployAppWhenComposeUpFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON, RUNNING_STATE).Return(DB_OBJ, nil),
		appExecutorMockObj.EXPECT().EnableEventMonitoring(gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), true).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().DeleteApp(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	_, err := Executor.DeployApp(DESCRIPTION_YAML, nil)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknowError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE)
}

func TestCalledDeployAppWhenYAMLToJSONFailed_ExpectErrorReturn(t *testing.T) {
	_, err := Executor.DeployApp(WRONG_DESCRIPTION_JSON, nil)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYAMLError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE)
}

func TestCalledDeployAppWhenInsertComposeFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON, RUNNING_STATE).Return(nil, UnknownError),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	_, err := Executor.DeployApp(DESCRIPTION_YAML, nil)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InsertComposeFileFailed", "nil")
	}

	os.RemoveAll(COMPOSE_FILE)
}

func TestCalledApps_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppList().Return(DB_OBJs, nil),
	)

	dbExecutor = dbExecutorMockObj

	res, err := Executor.Apps()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	// Make Compare Value
	yamlList := []map[string]interface{}{
		map[string]interface{}{
			"id":    DB_OBJs[0]["id"],
			"state": DB_OBJs[0]["state"],
		},
	}
	compareReturnVal := make(map[string]interface{})
	compareReturnVal["apps"] = yamlList

	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Error()
	}
}

func TestCalledAppsWhenGetAppListFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppList().Return(nil, UnknownError),
	)

	dbExecutor = dbExecutorMockObj

	_, err := Executor.Apps()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(APP_ID, gomock.Any(), SERVICE).Return(PS_EXPECT_RETURN, nil),
		dockerExecutorMockObj.EXPECT().GetContainerConfigByName(CONTAINER).Return(INSPECT_RETURN_MSG, nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	res, err := Executor.App(APP_ID)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if !reflect.DeepEqual(res, APP_OBJ) {
		t.Error()
	}
}

func TestCalledAppWhenGetAppFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbExecutorMockObj

	_, err := Executor.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenJSONToYAMLFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(WRONG_DB_GET_OBJ, nil),
	)

	dbExecutor = dbExecutorMockObj

	_, err := Executor.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenNoServiceFiledinYAML_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ_WITHOUT_SERVICE, nil),
	)

	dbExecutor = dbExecutorMockObj

	_, err := Executor.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenGetServiceStateComposePsFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(APP_ID, gomock.Any(), SERVICE_NAME).Return(nil, UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	_, err := Executor.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenGetServiceStateComposeInspectFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(APP_ID, gomock.Any(), SERVICE_NAME).Return(PS_EXPECT_RETURN, nil),
		dockerExecutorMockObj.EXPECT().GetContainerConfigByName(CONTAINER_NAME).Return(nil, UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	_, err := Executor.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledHandleEventsWithUpdateEvent_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppEvent(APP_ID, EVENT_REPOSITORY, EVENT_TAG, UPDATE).Return(nil),
	)

	dbExecutor = dbExecutorMockObj

	err := Executor.HandleEvents(APP_ID, UPDATE_EVENTS_JSON)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledHandleEventsWithDeleteEvent_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppEvent(APP_ID, EVENT_REPOSITORY, EVENT_TAG, DELETE).Return(nil),
	)

	dbExecutor = dbExecutorMockObj

	err := Executor.HandleEvents(APP_ID, DELETE_EVENTS_JSON)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledHandleEventsWithInvalidJSONFormat_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	err := Executor.HandleEvents(APP_ID, INVALID_JSON_FORMAT)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidJSON", "nil")
	}
}

func TestCalledUpdateAppInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppInfo(APP_ID, DESCRIPTION_JSON).Return(nil),
	)

	dbExecutor = dbExecutorMockObj

	err := Executor.UpdateAppInfo(APP_ID, DESCRIPTION_YAML)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledUpdateAppInfoWhenYAMLToJSON_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	err := Executor.UpdateAppInfo(APP_ID, WRONG_DESCRIPTION_JSON)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYAMLError", "nil")
	}
}

func TestCalledUpdateAppInfoWhenUpdateAppInfoFailed_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppInfo(APP_ID, DESCRIPTION_JSON).Return(InvalidYamlError),
	)

	dbExecutor = dbExecutorMockObj

	err := Executor.UpdateAppInfo(APP_ID, DESCRIPTION_YAML)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYamlError", "nil")
	}
}

func TestCalledStartApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_WITH_EXITED_STATE_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_WITH_EXITED_STATE_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dockerExecutorMockObj.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.StartApp(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestStartAppWhenGetAppFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(nil, errors.Unknown{}),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	err := Executor.StartApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}
	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", err.Error())
	case errors.Unknown:
	}
}

func TestCalledStartAppWhenSetYAMLFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbExecutorMockObj

	err := Executor.StartApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledStartAppWhenComposeStartFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_WITH_EXITED_STATE_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_WITH_EXITED_STATE_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),

		dockerExecutorMockObj.EXPECT().Start(gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().Stop(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, EXITED_STATE).Return(nil),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.StartApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestStartAppWhenUpdateAppStateFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_WITH_EXITED_STATE_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_WITH_EXITED_STATE_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dockerExecutorMockObj.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(errors.Unknown{}),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.StartApp(APP_ID)

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", err.Error())
	case errors.Unknown:
	}
}

func TestCalledStopApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dockerExecutorMockObj.EXPECT().Stop(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, EXITED_STATE).Return(nil),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.StopApp(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestStopAppWhenGetAppFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(nil, errors.Unknown{}),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	err := Executor.StopApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", "nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: %s, actual err: %s", "Unknown", err.Error())
	case errors.Unknown:
	}
}

func TestCalledStopAppWhenSetYAMLFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbExecutorMockObj

	err := Executor.StopApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledStopAppWhenComposeStopFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dockerExecutorMockObj.EXPECT().Stop(gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.StopApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledDeleteApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any(), gomock.Any()).Return(nil),
		appExecutorMockObj.EXPECT().DisableEventMonitoring(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().DeleteApp(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.DeleteApp(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledDeleteAppWhenSetYAMLFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbExecutorMockObj

	err := Executor.DeleteApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledDeleteAppWhenComposeDeleteFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any(), gomock.Any()).Return(UnknownError),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	err := Executor.DeleteApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledDeleteAppWhenFailedToUnsetEventChannel_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any(), gomock.Any()).Return(nil),
		appExecutorMockObj.EXPECT().DisableEventMonitoring(gomock.Any(), gomock.Any()).Return(UnknownError),
		dbExecutorMockObj.EXPECT().DeleteApp(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.DeleteApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledDeleteAppWhenDBDeleteAppFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any(), gomock.Any()).Return(nil),
		appExecutorMockObj.EXPECT().DisableEventMonitoring(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().DeleteApp(gomock.Any()).Return(UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.DeleteApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestUpdateAppWithoutQuery_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATED_OBJ, nil),
		dbExecutorMockObj.EXPECT().UpdateAppEvent(APP_ID, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestUpdateAppWithoutQueryWhenUpdateAppStateToupdatingFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(UnknownError),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestUpdateAppWithoutQueryWhenGetAppFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestUpdateAppWithoutQueryWhenGetImageDigestByNameFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return("", UnknownError),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dbExecutor = dbExecutorMockObj
	dockerExecutor = dockerExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestUpdateAppWithoutQueryWhenPullFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, gomock.Any(), true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestUpdateAppWithoutQueryWhenPullAndImagePullFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(NotFoundError),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestUpdateAppWithoutQueryWhenPullAndGetImageIDByRepoDigestFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return("", NotFoundError),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestUpdateAppWithoutQueryWhenPullAndImageTagFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(NotFoundError),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestUpdateAppWithoutQueryWhenPullAndUpFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, gomock.Any(), true).Return(NotFoundError),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestUpdateAppWithoutQueryWhenUpFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), true).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, gomock.Any(), true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestUpdateAppWithoutQueryWhenUpdateAppStateTorunningFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATING_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATING_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(UnknownError),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestUpdateAppWithoutQueryWhenUpdateAppEventFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATED_OBJ, nil),
		dbExecutorMockObj.EXPECT().UpdateAppEvent(APP_ID, REPOSITORY_WITH_PORT_IMAGE, NEW_TAG, NONE_EVENT).Return(UnknownError),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestUpdateAppWithQueryWithTag_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	QUERY_IMAGE_LIST := []string{FULL_IMAGE_NAME}
	QUERY := map[string]interface{}{
		"images": QUERY_IMAGE_LIST,
	}

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), true, gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppInfo(APP_ID, UPDATED_DESCRIPTION_JSON).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATED_OBJ, nil),
		dbExecutorMockObj.EXPECT().UpdateAppEvent(APP_ID, REPOSITORY_WITH_PORT_IMAGE, NEW_TAG, NONE_EVENT).Return(nil),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, QUERY)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestUpdateAppWithQueryWithTagWhenUpdateAppInfoFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	QUERY_IMAGE_LIST := []string{FULL_IMAGE_NAME}
	QUERY := map[string]interface{}{
		"images": QUERY_IMAGE_LIST,
	}

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	appExecutorMockObj := appmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		appExecutorMockObj.EXPECT().LockUpdateAppState(),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE_WITH_TAG).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), true, gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppInfo(APP_ID, UPDATED_DESCRIPTION_JSON).Return(UnknownError),
		appExecutorMockObj.EXPECT().UnlockUpdateAppState(),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj
	appsMonitor = appExecutorMockObj

	err := Executor.UpdateApp(APP_ID, QUERY)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

/* Inner Funtion TEST */

func TestCalledSetYamlFile_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
	)

	dbExecutor = dbExecutorMockObj

	composeFile, err := setYamlFile(APP_ID, "api")
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	os.RemoveAll(composeFile)
}

func TestCalledSetYamlFileWhenGetAppFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbExecutorMockObj

	composeFile, err := setYamlFile(APP_ID, "api")
	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}

	os.RemoveAll(composeFile)
}

func TestCalledSetYamlFileWhenJSONToYAMLFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(WRONG_DB_GET_OBJ, nil),
	)

	dbExecutor = dbExecutorMockObj

	composeFile, err := setYamlFile(APP_ID, "api")
	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}

	os.RemoveAll(composeFile)
}

func TestRestoreRepoDigests_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := restoreRepoDigests(APP_ID, COMPOSE_FILE, repoDigests, RUNNING_STATE)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestRestoreRepoDigestsWhenImagePullFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(UnknownError),
	)

	dockerExecutor = dockerExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := restoreRepoDigests(APP_ID, COMPOSE_FILE, repoDigests, RUNNING_STATE)
	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestRestoreRepoDigestsWhenGetImageIDByRepoDigestFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return("", UnknownError),
	)

	dockerExecutor = dockerExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := restoreRepoDigests(APP_ID, COMPOSE_FILE, repoDigests, RUNNING_STATE)
	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestRestoreRepoDigestsWhenImagTagFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(UnknownError),
	)

	dockerExecutor = dockerExecutorMockObj

	composeFile := "random"
	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := restoreRepoDigests(composeFile, APP_ID, repoDigests, RUNNING_STATE)
	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestRestoreRepoDigestsWhenUpFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true).Return(UnknownError),
	)

	dockerExecutor = dockerExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := restoreRepoDigests(APP_ID, COMPOSE_FILE, repoDigests, RUNNING_STATE)
	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestInnerUpdateApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Pull(APP_ID, COMPOSE_FILE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := updateApp(APP_ID, COMPOSE_FILE, DB_GET_APP_OBJ, repoDigests)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestInnerUpdateAppWhenPullFailed_ExpectReturnErrror(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Pull(APP_ID, COMPOSE_FILE).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := updateApp(APP_ID, COMPOSE_FILE, DB_GET_APP_OBJ, repoDigests)
	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestInnerUpdateAppWhenUpFailed_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Pull(APP_ID, COMPOSE_FILE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := updateApp(APP_ID, COMPOSE_FILE, DB_GET_APP_OBJ, repoDigests)
	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestUpdateService_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Pull(APP_ID, COMPOSE_FILE, SERVICE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true, SERVICE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := updateService(APP_ID, COMPOSE_FILE, DB_GET_APP_OBJ, repoDigests, SERVICE)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestUpdateServiceWhenPullFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Pull(APP_ID, COMPOSE_FILE, SERVICE).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := updateService(APP_ID, COMPOSE_FILE, DB_GET_APP_OBJ, repoDigests, SERVICE)
	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestUpdateServiceWhenUpFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Pull(APP_ID, COMPOSE_FILE, SERVICE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true, SERVICE).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := updateService(APP_ID, COMPOSE_FILE, DB_GET_APP_OBJ, repoDigests, SERVICE)
	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestRestoreStateWithRUNNING_STATE_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	err := restoreState(APP_ID, COMPOSE_FILE, RUNNING_STATE, true)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestRestoreStateWhenUpFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true).Return(UnknownError),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState(APP_ID, COMPOSE_FILE, RUNNING_STATE, true)
	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestRestoreStateWhenUpdateAppStateToRunningFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(errors.Unknown{}),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	err := restoreState(APP_ID, COMPOSE_FILE, RUNNING_STATE, true)
	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestRestoreStateWithEXITED_STATE_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Stop(APP_ID, COMPOSE_FILE).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, EXITED_STATE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	err := restoreState(APP_ID, COMPOSE_FILE, EXITED_STATE, true)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestRestoreStateWhenStopFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Stop(APP_ID, COMPOSE_FILE).Return(UnknownError),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState(APP_ID, COMPOSE_FILE, EXITED_STATE, true)

	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestRestoreStateWhenUpdateAppStateToExitedFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, true).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(errors.Unknown{}),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	err := restoreState(APP_ID, COMPOSE_FILE, RUNNING_STATE, true)
	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestUpdateYamlFile_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := updateYamlFile(APP_ID, COMPOSE_FILE, ORIGIN_DESCRIPTION_JSON, SERVICE, REPOSITORY_WITH_PORT_IMAGE+":"+NEW_TAG)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestUpdateYamlFileWithInvalidJSON_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := updateYamlFile(APP_ID, COMPOSE_FILE, WRONG_DESCRIPTION_JSON, SERVICE, REPOSITORY_WITH_PORT_IMAGE+":"+NEW_TAG)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "json unmarshal", "nil")
	}
}

func TestUpdateYamlFileWithInvalidDescription_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := updateYamlFile(APP_ID, COMPOSE_FILE, DESCRIPTION_JSON_WITHOUT_SERVICE, SERVICE, REPOSITORY_WITH_PORT_IMAGE+":"+NEW_TAG)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "can't find application info unmarshal", "nil")
	}
}

func TestExtractQueryInfoWithRepoWithPortAndTag_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tagExist, repo, tag, err := extractQueryInfo(REPOSITORY_WITH_PORT_IMAGE + ":" + OLD_TAG)

	if tagExist == false || repo != REPOSITORY_WITH_PORT_IMAGE || tag != OLD_TAG || err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestExtractQueryInfoWithRepoWithPortAndNoTag_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tagExist, repo, tag, err := extractQueryInfo(REPOSITORY_WITH_PORT_IMAGE)

	if tagExist == true || repo != REPOSITORY_WITH_PORT_IMAGE || tag != "" || err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestExtractQueryInfoWithRepoWithoutPortAndTag_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	REPOSITORY_WITHOUT_PORT := "docker"

	tagExist, repo, tag, err := extractQueryInfo(REPOSITORY_WITHOUT_PORT + ":" + OLD_TAG)

	if tagExist == false || repo != REPOSITORY_WITHOUT_PORT || tag != OLD_TAG || err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestExtractQueryInfoWithRepoWithPortAndInvalidTag_ExpectReturnFalse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tagExist, _, _, err := extractQueryInfo(REPOSITORY_WITH_PORT_IMAGE + ":")

	if tagExist != false {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if err == nil {
		t.Error("Expected err : Unknown - invalid repository, Actual err : nil")
	}
}

func TestExtractQueryInfoWithRepoWithoutPortAndNoTag_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	REPOSITORY_WITHOUT_PORT := "docker"

	tagExist, repo, tag, err := extractQueryInfo(REPOSITORY_WITHOUT_PORT)

	if tagExist == true || repo != REPOSITORY_WITHOUT_PORT || tag != "" || err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestGetServiceName_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceName, err := getServiceName(REPOSITORY_WITH_PORT_IMAGE, []byte(UPDATED_DESCRIPTION_JSON))

	if serviceName != SERVICE {
		t.Errorf("Expected service name: %s, actual service name: %s", SERVICE_NAME, serviceName)
	}

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestGetServiceNameWithNoServiceDescription_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := getServiceName(REPOSITORY_WITH_PORT_IMAGE, []byte(DESCRIPTION_JSON_WITHOUT_SERVICE))

	if err == nil {
		t.Error("Expected err: Unknown, actual err : nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: Unknown, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestGetServiceNameWithInvalidDescription_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := getServiceName(REPOSITORY_WITH_PORT_IMAGE, []byte(INVALID_JSON_FORMAT))

	if err == nil {
		t.Error("Expected err: IOError, actual err : nil")
	}

	switch err.(type) {
	default:
		t.Errorf("Expected err: IOError, actual err: %s", err.Error())
	case errors.IOError:
	}
}

func TestGetServiceNameWithNoPortRepository_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	NO_PORT_REPOSITORY := "test"

	DESCRIPTION_JSON_WITH_NO_PORT_REPOSITORY := "{\"services\":{\"" + SERVICE + "\":{\"image\":\"" + NO_PORT_REPOSITORY + ":" + OLD_TAG + "\"}},\"version\":\"2\"}"
	serviceName, err := getServiceName(NO_PORT_REPOSITORY, []byte(DESCRIPTION_JSON_WITH_NO_PORT_REPOSITORY))

	if serviceName != SERVICE {
		t.Errorf("Expected service name: %s, actual service name: %s", SERVICE_NAME, serviceName)
	}

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestGetServiceNameWithInvalidImageName_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	NO_PORT_REPOSITORY := "test"
	INVALID_IMAGE := "wrong_img"
	DESCRIPTION_JSON_WITH_NO_PORT_REPOSITORY := "{\"services\":{\"" + SERVICE + "\":{\"image\":\"" + NO_PORT_REPOSITORY + ":" + OLD_TAG + "\"}},\"version\":\"2\"}"

	_, err := getServiceName(INVALID_IMAGE, []byte(DESCRIPTION_JSON_WITH_NO_PORT_REPOSITORY))

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown can't find matched service", "nil")
	}
}

func TestUpdateAppEvent_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATED_OBJ, nil),
		dbExecutorMockObj.EXPECT().UpdateAppEvent(APP_ID, gomock.Any(), gomock.Any(), "none").Return(nil),
	)

	dbExecutor = dbExecutorMockObj

	err := updateAppEvent(APP_ID)

	if err != nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown error", "nil")
	}
}

func TestUpdateAppEventWhenGetAppFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbExecutorMockObj

	err := updateAppEvent(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown error", "nil")
	}
}

func TestUpdateAppEventWhenJsonUnmarshalFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	DB_GET_APP_WRONG_DESCRIPTION_OBJ := map[string]interface{}{
		"id":          APP_ID,
		"state":       "UP",
		"description": WRONG_DESCRIPTION_JSON,
		"images": []map[string]interface{}{
			{
				"name": REPOSITORY_WITH_PORT_IMAGE,
				"changes": map[string]interface{}{
					"tag":   NEW_TAG,
					"state": "update",
				},
			},
		},
	}

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_WRONG_DESCRIPTION_OBJ, nil),
	)

	dbExecutor = dbExecutorMockObj

	err := updateAppEvent(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "json unmarshal", "nil")
	}
}

func TestUpdateAppEventWhenUpdateAppEventFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATED_OBJ, nil),
		dbExecutorMockObj.EXPECT().UpdateAppEvent(APP_ID, gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
	)

	dbExecutor = dbExecutorMockObj

	err := updateAppEvent(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "Unknown error", "nil")
	}
}

func TestRestoreAllAppsState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppList().Return(DB_OBJs, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, gomock.Any(), false).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE),
	)

	dbExecutor = dbExecutorMockObj
	dockerExecutor = dockerExecutorMockObj

	restoreAllAppsState()
}

func TestRestoreAllAppsStateGetAppListFailed_ExpectReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppList().Return(nil, errors.Unknown{}),
	)

	dbExecutor = dbExecutorMockObj

	restoreAllAppsState()
}

func TestRestoreAllAppsStateGetAppFailed_ExpectReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppList().Return(DB_OBJs, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(nil, errors.Unknown{}),
	)

	dbExecutor = dbExecutorMockObj
	dockerExecutor = dockerExecutorMockObj

	restoreAllAppsState()
}

func TestRestoreAllAppsStateUpFailed_ExpectReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppList().Return(DB_OBJs, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, gomock.Any(), false).Return(errors.Unknown{}),
	)

	dbExecutor = dbExecutorMockObj
	dockerExecutor = dockerExecutorMockObj

	restoreAllAppsState()
}

func TestRestoreAllAppsStateStopFailed_ExpectReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppList().Return(DB_OBJs, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_WITH_EXITED_STATE_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_WITH_EXITED_STATE_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Stop(APP_ID, gomock.Any()).Return(errors.Unknown{}),
	)

	dbExecutor = dbExecutorMockObj
	dockerExecutor = dockerExecutorMockObj

	restoreAllAppsState()
}

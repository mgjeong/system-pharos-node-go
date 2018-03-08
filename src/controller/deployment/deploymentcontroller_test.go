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
	dockermocks "controller/deployment/dockercontroller/mocks"
	dbmocks "db/mongo/service/mocks"
	"github.com/golang/mock/gomock"
	"os"
	"reflect"
	"testing"
)

const (
	COMPOSE_FILE_PATH                 = "docker-compose.yaml"
	APP_ID                            = "000000000000000000000000"
	DESCRIPTION_JSON_WITHOUT_SERVICE  = "{\"no_services\":{\"test_service\":{\"image\":\"test_image:0.2\"}},\"version\":\"2\"}"
	WRONG_DESCRIPTION_JSON            = "{{{{services:\n  test_service:\n    image: test_image:0.2\nversion: \"2\""
	WRONG_INSPECT_RETURN_MSG          = "error_[{\"State\": {\"Status\": \"running\", \"ExitCode\": \"0\"}}]"
	OLD_TAG                           = "1.0"
	NEW_TAG                           = "2.0"
	REPOSITORY_WITH_PORT_IMAGE        = "test_url:5000/test"
	APP_STATE                         = "STATE"
	SERVICE_NAME                      = "test_service"
	CONTAINER_NAME                    = "test_container"
	DESCRIPTION_JSON                  = "{\"services\":{\"" + SERVICE + "\":{\"image\":\"" + REPOSITORY_WITH_PORT_IMAGE + ":" + OLD_TAG + "\"}},\"version\":\"2\"}"
	DESCRIPTION_YAML                  = "services:\n  " + SERVICE + ":\n    image: " + REPOSITORY_WITH_PORT_IMAGE + ":" + OLD_TAG + "\nversion: \"2\"\n"
	REPOSITORY_WITH_PORT_IMAGE_DIGEST = REPOSITORY_WITH_PORT_IMAGE + "@" + "sha256:1234567890"
	SERVICE                           = "test_service"
	CONTAINER                         = "test_container"
	ORIGIN_DESCRIPTION_JSON           = "{\"services\":{\"" + SERVICE + "\":{\"image\":\"" + REPOSITORY_WITH_PORT_IMAGE + ":" + OLD_TAG + "\"}},\"version\":\"2\"}"
	UPDATED_DESCRIPTION_JSON          = "{\"services\":{\"" + SERVICE + "\":{\"image\":\"" + REPOSITORY_WITH_PORT_IMAGE + ":" + NEW_TAG + "\"}},\"version\":\"2\"}"
	FULL_IMAGE_NAME                   = REPOSITORY_WITH_PORT_IMAGE + ":" + NEW_TAG
	NONE_EVENT                        = "none"
	CONTAINER_ID                      = 1234
	SERVICE_PORT                      = 1234
	SERVICE_STATUS                    = "running"
	EXIT_CODE_VALUE                   = "0"
	EVENT_REPOSITORY                  = "localhost:5000/test_repo"
	EVENT_TAG                         = "latest"
	UPDATE_EVENTS_JSON                = `{"events":[{"action": "push","target": {"repository": "test_repo","tag": "latest"},"request": {"addr": "0.0.0.0:8888","host": "localhost:5000"}}]}`
	DELETE_EVENTS_JSON                = `{"events":[{"action": "delete","target": {"repository": "test_repo","tag": "latest"},"request": {"addr": "0.0.0.0:8888","host": "localhost:5000"}}]}`
	INVALID_JSON_FORMAT               = "invalid_json_format"
	REPODIGEST                        = "test@sha256test"
	IMAGE_ID                          = "abcd"
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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON, RUNNING_STATE).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(gomock.Any(), COMPOSE_FILE_PATH, SERVICE_NAME).Return(PS_EXPECT_RETURN, nil),
		dockerExecutorMockObj.EXPECT().GetContainerConfigByName(gomock.Any()).Return(INSPECT_RETURN_MSG, nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	res, err := Executor.DeployApp(DESCRIPTION_YAML)

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

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledDeployAppWhenComposeUpFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON, RUNNING_STATE).Return(DB_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().DeleteApp(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	_, err := Executor.DeployApp(DESCRIPTION_YAML)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknowError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledDeployAppWhenYAMLToJSONFailed_ExpectErrorReturn(t *testing.T) {
	_, err := Executor.DeployApp(WRONG_DESCRIPTION_JSON)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYAMLError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
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

	_, err := Executor.DeployApp(DESCRIPTION_YAML)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InsertComposeFileFailed", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
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
		dockerExecutorMockObj.EXPECT().Ps(gomock.Any(), COMPOSE_FILE_PATH, SERVICE).Return(PS_EXPECT_RETURN, nil),
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
		dockerExecutorMockObj.EXPECT().Ps(gomock.Any(), COMPOSE_FILE_PATH, SERVICE_NAME).Return(nil, UnknownError),
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
		dockerExecutorMockObj.EXPECT().Ps(gomock.Any(), COMPOSE_FILE_PATH, SERVICE_NAME).Return(PS_EXPECT_RETURN, nil),
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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	err := Executor.StartApp(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledStartAppWhenSetYAMLFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Start(gomock.Any(), gomock.Any()).Return(UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	err := Executor.StartApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledStopApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Stop(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, EXITED_STATE).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	err := Executor.StopApp(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledStopAppWhenSetYAMLFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Stop(gomock.Any(), gomock.Any()).Return(UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().DeleteApp(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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
		dbExecutorMockObj.EXPECT().GetAppState(APP_ID).Return(RUNNING_STATE, nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().DeleteApp(gomock.Any()).Return(UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATING_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATING_OBJ, nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATED_OBJ, nil),
		dbExecutorMockObj.EXPECT().UpdateAppEvent(APP_ID, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestUpdateAppWithoutQueryWhenUpdateAppStateToupdatingFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(UnknownError),
	)

	dbExecutor = dbExecutorMockObj

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
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE).Return("", UnknownError),
	)

	dbExecutor = dbExecutorMockObj
	dockerExecutor = dockerExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(NotFoundError),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return("", NotFoundError),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(NotFoundError),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(NotFoundError),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATING_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATING_OBJ, nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(UnknownError),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATED_OBJ, nil),
		dbExecutorMockObj.EXPECT().UpdateAppEvent(APP_ID, REPOSITORY_WITH_PORT_IMAGE, NEW_TAG, NONE_EVENT).Return(UnknownError),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATING_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATING_OBJ, nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppInfo(APP_ID, UPDATED_DESCRIPTION_JSON).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, RUNNING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATED_OBJ, nil),
		dbExecutorMockObj.EXPECT().UpdateAppEvent(APP_ID, REPOSITORY_WITH_PORT_IMAGE, NEW_TAG, NONE_EVENT).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().UpdateAppState(APP_ID, UPDATING_STATE).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(REPOSITORY_WITH_PORT_IMAGE).Return(REPODIGEST, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppInfo(APP_ID, UPDATED_DESCRIPTION_JSON).Return(UnknownError),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

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

	err := setYamlFile(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledSetYamlFileWhenGetAppFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbExecutorMockObj

	err := setYamlFile(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledSetYamlFileWhenJSONToYAMLFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(WRONG_DB_GET_OBJ, nil),
	)

	dbExecutor = dbExecutorMockObj

	err := setYamlFile(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestRestoreRepoDigests_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := restoreRepoDigests(APP_ID, repoDigests, RUNNING_STATE)
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

	err := restoreRepoDigests(APP_ID, repoDigests, RUNNING_STATE)
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

	err := restoreRepoDigests(APP_ID, repoDigests, RUNNING_STATE)
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

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := restoreRepoDigests(APP_ID, repoDigests, RUNNING_STATE)
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
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(UnknownError),
	)

	dockerExecutor = dockerExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := restoreRepoDigests(APP_ID, repoDigests, RUNNING_STATE)
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
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := updateApp(APP_ID, DB_GET_APP_OBJ, repoDigests)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestInnerUpdateAppWhenPullFailed_ExpectReturnErrror(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Pull(APP_ID, COMPOSE_FILE).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := updateApp(APP_ID, DB_GET_APP_OBJ, repoDigests)
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

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Pull(APP_ID, COMPOSE_FILE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := updateApp(APP_ID, DB_GET_APP_OBJ, repoDigests)
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

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Pull(APP_ID, COMPOSE_FILE, SERVICE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, SERVICE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := updateService(APP_ID, DB_GET_APP_OBJ, repoDigests, SERVICE)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestUpdateServiceWhenPullFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Pull(APP_ID, COMPOSE_FILE, SERVICE).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := updateService(APP_ID, DB_GET_APP_OBJ, repoDigests, SERVICE)
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

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Pull(APP_ID, COMPOSE_FILE, SERVICE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE, SERVICE).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().ImagePull(REPODIGEST).Return(nil),
		dockerExecutorMockObj.EXPECT().GetImageIDByRepoDigest(REPODIGEST).Return(IMAGE_ID, nil),
		dockerExecutorMockObj.EXPECT().ImageTag(IMAGE_ID, REPOSITORY_WITH_PORT_IMAGE).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	repoDigests := make(map[string]string, 0)
	repoDigests[REPOSITORY_WITH_PORT_IMAGE] = REPODIGEST

	err := updateService(APP_ID, DB_GET_APP_OBJ, repoDigests, SERVICE)
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

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState(APP_ID, RUNNING_STATE)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestRestoreStateWhenUpFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(APP_ID, COMPOSE_FILE).Return(UnknownError),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState(APP_ID, RUNNING_STATE)
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

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Stop(APP_ID, COMPOSE_FILE).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState(APP_ID, EXITED_STATE)
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

	err := restoreState(APP_ID, EXITED_STATE)
	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestUpdateYamlFile_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := updateYamlFile(APP_ID, ORIGIN_DESCRIPTION_JSON, SERVICE, REPOSITORY_WITH_PORT_IMAGE+":"+NEW_TAG)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestUpdateYamlFileWithInvalidJSON_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := updateYamlFile(APP_ID, WRONG_DESCRIPTION_JSON, SERVICE, REPOSITORY_WITH_PORT_IMAGE+":"+NEW_TAG)
	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "json unmarshal", "nil")
	}
}

func TestUpdateYamlFileWithInvalidDescription_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := updateYamlFile(APP_ID, DESCRIPTION_JSON_WITHOUT_SERVICE, SERVICE, REPOSITORY_WITH_PORT_IMAGE+":"+NEW_TAG)
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

func TestExtractQueryInfoWithRepoWithoutPortAndNoTag_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	REPOSITORY_WITHOUT_PORT := "docker"

	tagExist, repo, tag, err := extractQueryInfo(REPOSITORY_WITHOUT_PORT)
	if tagExist == true || repo != REPOSITORY_WITHOUT_PORT || tag != "" || err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestExtractQueryInfoWithInvalidRepository_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	INVALID_REPOSITORY := "docker:abcd:efsd:"

	tagExist, repo, tag, err := extractQueryInfo(INVALID_REPOSITORY)
	if err == nil {
		t.Errorf("Expected err: %s, actual err: %t %s %s %s", "invalid repository", tagExist, repo, tag, "nil")
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

func TestGetServiceNameWithNoPortRepository_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	NO_PORT_REPOSITORY := "test"

	DESCRIPTION_JSON_WITH_NO_PORT_REPOSITORY := "{\"services\":{\"" + SERVICE + "\":{\"image\":\"" + NO_PORT_REPOSITORY + ":" + OLD_TAG + "\"}},\"version\":\"2\"}"
	serviceName, err := getServiceName(NO_PORT_REPOSITORY, []byte(DESCRIPTION_JSON_WITH_NO_PORT_REPOSITORY))
	if serviceName != SERVICE {
		t.Errorf("Expected service name: %s, actual service name: %s", SERVICE_NAME, serviceName)
	}
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

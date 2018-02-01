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
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

const (
	COMPOSE_FILE_PATH                = "docker-compose.yaml"
	APP_ID                           = "000000000000000000000000"
	DESCRIPTION_JSON                 = "{\"services\":{\"test_service\":{\"image\":\"test_image:0.2\"}},\"version\":\"2\"}"
	DESCRIPTION_JSON_WITHOUT_SERVICE = "{\"no_services\":{\"test_service\":{\"image\":\"test_image\"}},\"version\":\"2\"}"
	WRONG_DESCRIPTION_JSON           = "{{{{services:\n  test_service:\n    image: test_image\nversion: \"2\""
	DESCRIPTION_YAML                 = "services:\n  test_service:\n    image: test_image\nversion: \"2\"\n"
	IMAGE_NAME                       = "test_image"
	IMAGE_DIGEST                     = "test_image@sha256:1234567890"
	APP_STATE                        = "STATE"
	SERVICE_NAME                     = "test_service"
	CONTAINER_NAME                   = "test_container"
	WRONG_INSPECT_RETURN_MSG         = "error_[{\"State\": {\"Status\": \"running\", \"ExitCode\": \"0\"}}]"

	OLD_TAG                  = "1.0"
	NEW_TAG                  = "2.0"
	REPOSITORY_WITH_PORT     = "test_url:5000/test"
	SERVICE                  = "test_service"
	ORIGIN_DESCRIPTION_JSON  = "{\"services\":{\"" + SERVICE + "\":{\"image\":\"" + REPOSITORY_WITH_PORT + ":" + OLD_TAG + "\"}},\"version\":\"2\"}"
	UPDATED_DESCRIPTION_JSON = "{\"services\":{\"" + SERVICE + "\":{\"image\":\"" + REPOSITORY_WITH_PORT + ":" + NEW_TAG + "\"}},\"version\":\"2\"}"
	FULL_IMAGE_NAME          = REPOSITORY_WITH_PORT + ":" + NEW_TAG
	NONE_EVENT               = "none"
)

var (
	DB_GET_APP_OBJ = map[string]interface{}{
		"id":          APP_ID,
		"state":       "UP",
		"description": ORIGIN_DESCRIPTION_JSON,
		"images": []map[string]interface{}{
			{
				"name": REPOSITORY_WITH_PORT,
				"changes": map[string]interface{}{
					"tag":   NEW_TAG,
					"state": "update",
				},
			},
		},
	}

	DB_GET_APP_UPDATED_OBJ = map[string]interface{}{
		"id":          APP_ID,
		"state":       "UP",
		"description": UPDATED_DESCRIPTION_JSON,
		"images": []map[string]interface{}{
			{
				"name": REPOSITORY_WITH_PORT,
				"changes": map[string]interface{}{
					"tag":   NEW_TAG,
					"state": "update",
				},
			},
		},
	}

	DB_OBJ = map[string]interface{}{
		"id": APP_ID,
	}

	DB_GET_OBJ = map[string]interface{}{
		"description": DESCRIPTION_JSON,
		"state":       "UP",
	}

	DB_OBJs = []map[string]interface{}{
		map[string]interface{}{
			"id":          APP_ID,
			"state":       "UP",
			"description": DESCRIPTION_JSON,
		},
	}

	WRONG_DB_OBJ = map[string]interface{}{
		"id": APP_ID,
	}

	WRONG_DB_GET_OBJ = map[string]interface{}{
		"description": WRONG_DESCRIPTION_JSON,
		"state":       "UP",
	}

	WRONG_DB_OBJs = []map[string]interface{}{
		map[string]interface{}{
			"id":    APP_ID,
			"state": "UP",
		},
	}

	DB_GET_OBJ_WITHOUT_SERVICE = map[string]interface{}{
		"description": DESCRIPTION_JSON_WITHOUT_SERVICE,
		"state":       "UP",
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
	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON).Return(DB_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbManagerMockObj

	res, err := Executor.DeployApp(DESCRIPTION_YAML)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	compareReturnVal := DB_OBJ

	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Error()
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledDeployAppWhenComposeUpFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON).Return(DB_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any(), gomock.Any()).Return(nil),
		dbManagerMockObj.EXPECT().DeleteApp(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbManagerMockObj

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

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON).Return(nil, UnknownError),
	)

	// pass mockObj to a real object.
	dbExecutor = dbManagerMockObj

	_, err := Executor.DeployApp(DESCRIPTION_YAML)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InsertComposeFileFailed", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledApps_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppList().Return(DB_OBJs, nil),
	)

	dbExecutor = dbManagerMockObj

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

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppList().Return(nil, UnknownError),
	)

	dbExecutor = dbManagerMockObj

	_, err := Executor.Apps()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	INSPECT_RETURN_MSG := make(map[string]interface{})
	INSPECT_RETURN_MSG["Status"] = "running"
	INSPECT_RETURN_MSG["ExitCode"] = "0"

	PS_EXPECT_RETURN := make([]map[string]string, 1)
	PS_EXPECT_RETURN[0] = make(map[string]string)
	PS_EXPECT_RETURN[0]["Name"] = "mongo_container"

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(gomock.Any(), COMPOSE_FILE_PATH, SERVICE_NAME).Return(PS_EXPECT_RETURN, nil),
		dockerExecutorMockObj.EXPECT().GetContainerStateByName(CONTAINER_NAME).Return(INSPECT_RETURN_MSG, nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbManagerMockObj

	res, err := Executor.App(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	// Make Compare Value
	services := make([]map[string]interface{}, 0)
	service := make(map[string]interface{}, 0)
	state := make(map[string]interface{}, 0)

	state["Status"] = "running"
	state["ExitCode"] = "0"
	service["state"] = state
	service["name"] = SERVICE_NAME
	services = append(services, service)

	compareReturnVal := map[string]interface{}{
		"state":       "UP",
		"description": DESCRIPTION_YAML,
		"services":    services,
	}

	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Error()
	}
}

func TestCalledAppWhenGetAppFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbManagerMockObj

	_, err := Executor.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenJSONToYAMLFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(WRONG_DB_GET_OBJ, nil),
	)

	dbExecutor = dbManagerMockObj

	_, err := Executor.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenNoServiceFiledinYAML_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ_WITHOUT_SERVICE, nil),
	)

	dbExecutor = dbManagerMockObj

	_, err := Executor.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenGetServiceStateComposePsFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(gomock.Any(), COMPOSE_FILE_PATH, SERVICE_NAME).Return(nil, UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbManagerMockObj

	_, err := Executor.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenGetServiceStateComposeInspectFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	PS_EXPECT_RETURN := make([]map[string]string, 1)
	PS_EXPECT_RETURN[0] = make(map[string]string)
	PS_EXPECT_RETURN[0]["Name"] = "mongo_container"

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(gomock.Any(), COMPOSE_FILE_PATH, SERVICE_NAME).Return(PS_EXPECT_RETURN, nil),
		dockerExecutorMockObj.EXPECT().GetContainerStateByName(CONTAINER_NAME).Return(nil, UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbManagerMockObj

	_, err := Executor.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledUpdateAppInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().UpdateAppInfo(APP_ID, DESCRIPTION_JSON).Return(nil),
	)

	dbExecutor = dbManagerMockObj

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

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().UpdateAppInfo(APP_ID, DESCRIPTION_JSON).Return(InvalidYamlError),
	)

	dbExecutor = dbManagerMockObj

	err := Executor.UpdateAppInfo(APP_ID, DESCRIPTION_YAML)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYamlError", "nil")
	}
}

func TestCalledStartApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil),
		dbManagerMockObj.EXPECT().UpdateAppState(APP_ID, gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbManagerMockObj

	err := Executor.StartApp(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledStartAppWhenSetYAMLFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbManagerMockObj

	err := Executor.StartApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledStartAppWhenComposeStartFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Start(gomock.Any(), gomock.Any()).Return(UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbManagerMockObj

	err := Executor.StartApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledStopApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Stop(gomock.Any(), gomock.Any()).Return(nil),
		dbManagerMockObj.EXPECT().UpdateAppState(APP_ID, gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbManagerMockObj

	err := Executor.StopApp(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledStopAppWhenSetYAMLFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbManagerMockObj

	err := Executor.StopApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledStopAppWhenComposeStopFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Stop(gomock.Any(), gomock.Any()).Return(UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbManagerMockObj

	err := Executor.StopApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledDeleteApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any(), gomock.Any()).Return(nil),
		dbManagerMockObj.EXPECT().DeleteApp(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbManagerMockObj

	err := Executor.DeleteApp(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledDeleteAppWhenSetYAMLFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbManagerMockObj

	err := Executor.DeleteApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledDeleteAppWhenComposeDeleteFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any(), gomock.Any()).Return(UnknownError),
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return("START", nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbManagerMockObj

	err := Executor.DeleteApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledDeleteAppWhenDBDeleteAppFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)
	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any(), gomock.Any()).Return(nil),
		dbManagerMockObj.EXPECT().DeleteApp(gomock.Any()).Return(UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbManagerMockObj

	err := Executor.DeleteApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

/* Inner Funtion TEST */

func TestCalledSetYamlFile_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
	)

	dbExecutor = dbManagerMockObj

	err := setYamlFile(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledSetYamlFileWhenGetAppFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	dbExecutor = dbManagerMockObj

	err := setYamlFile(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledSetYamlFileWhenJSONToYAMLFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(WRONG_DB_GET_OBJ, nil),
	)

	dbExecutor = dbManagerMockObj

	err := setYamlFile(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledRestoreRepoDigests_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	test_state := "UP"
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().GetImageDigestByName(IMAGE_NAME).Return(IMAGE_DIGEST, nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreRepoDigests("", DESCRIPTION_JSON, test_state)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledRestoreStateInputSTOP_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	test_state := "STOP"
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Stop(gomock.Any(), gomock.Any()).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState("", test_state)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledRestoreStateInputSTART_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	test_state := "START"
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState("", test_state)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledRestoreStateInputUP_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	test_state := "UP"
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState("", test_state)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledRestoreStateInputDEPLOY_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	test_state := "DEPLOY"
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState("", test_state)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestUpdateAppWithoutQuery_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any()).Return(nil),
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

func TestUpdateAppWithoutQueryWhenPullFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
	)

	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	err := Executor.UpdateApp(APP_ID, nil)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestUpdateAppWithoutQueryWhenUpFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), gomock.Any()).Return(UnknownError),
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
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		dbExecutorMockObj.EXPECT().UpdateAppInfo(APP_ID, UPDATED_DESCRIPTION_JSON).Return(nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_UPDATED_OBJ, nil),
		dbExecutorMockObj.EXPECT().UpdateAppEvent(APP_ID, REPOSITORY_WITH_PORT, NEW_TAG, NONE_EVENT).Return(nil),
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
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
		dbExecutorMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_APP_OBJ, nil),
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

func TestUpdateYamlFile_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := updateYamlFile(APP_ID, ORIGIN_DESCRIPTION_JSON, SERVICE, REPOSITORY_WITH_PORT+":"+NEW_TAG)
	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestUpdateYamlFileWithInvalidJSON_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := updateYamlFile(APP_ID, WRONG_DESCRIPTION_JSON, SERVICE, REPOSITORY_WITH_PORT+":"+NEW_TAG)
	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "json unmarshal", "nil")
	}
}

func TestUpdateYamlFileWithInvalidDescription_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := updateYamlFile(APP_ID, DESCRIPTION_JSON_WITHOUT_SERVICE, SERVICE, REPOSITORY_WITH_PORT+":"+NEW_TAG)
	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "can't find application info unmarshal", "nil")
	}
}

func TestExtractQueryInfoWithRepoWithPortAndTag_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tagExist, repo, tag, err := extractQueryInfo(REPOSITORY_WITH_PORT + ":" + OLD_TAG)
	if tagExist == false || repo != REPOSITORY_WITH_PORT || tag != OLD_TAG || err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestExtractQueryInfoWithRepoWithPortAndNoTag_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tagExist, repo, tag, err := extractQueryInfo(REPOSITORY_WITH_PORT)
	if tagExist == true || repo != REPOSITORY_WITH_PORT || tag != "" || err != nil {
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
		t.Errorf("Expected err: %s, actual err: %t %s %s %s", "invalid repogitory", tagExist, repo, tag, "nil")
	}
}

func TestGetServiceName_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceName, err := getServiceName(REPOSITORY_WITH_PORT, []byte(UPDATED_DESCRIPTION_JSON))
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
				"name": REPOSITORY_WITH_PORT,
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

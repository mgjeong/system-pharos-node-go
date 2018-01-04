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
	dbmocks "db/mongo/service/mocks"
	dockermocks "controller/deployment/dockercontroller/mocks"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

const (
	COMPOSE_FILE_PATH                = "docker-compose.yaml"
	APP_ID                           = "000000000000000000000000"
	DESCRIPTION_JSON                 = "{\"services\":{\"mongo\":{\"image\":\"nginx\"}},\"version\":\"2\"}"
	DESCRIPTION_JSON_WITHOUT_SERVICE = "{\"no_services\":{\"mongo\":{\"image\":\"nginx\"}},\"version\":\"2\"}"
	WRONG_DESCRIPTION_JSON           = "{{{{services:\n  mongo:\n    image: nginx\nversion: \"2\""
	DESCRIPTION_YAML                 = "services:\n  mongo:\n    image: nginx\nversion: \"2\"\n"
	IMAGE_NAME                       = "nginx"
	IMAGE_DIGEST                     = "nginx@sha256:1234567890"
	APP_STATE                        = "STATE"
	SERVICE_NAME                     = "mongo"
	PS_RETURN_MSG                    = "name command state port\n--------------------------\nmongo_container test_comman test_state test_port\n"
	CONTAINER_NAME                   = "mongo_container"
	INSPECT_RETURN_MSG               = "[{\"State\": {\"Status\": \"running\", \"ExitCode\": \"0\"}}]"
	WRONG_INSPECT_RETURN_MSG         = "error_[{\"State\": {\"Status\": \"running\", \"ExitCode\": \"0\"}}]"
)

var (
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

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
		dbManagerMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON).Return(DB_OBJ, nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	
	res, err := Controller.DeployApp(DESCRIPTION_YAML)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	compareReturnVal := DB_OBJ

	if !reflect.DeepEqual(res, compareReturnVal) {
		t.Error()
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledDeployAppWhenDBNotConnected_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := Controller.DeployApp(DESCRIPTION_YAML)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "ConnectionError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledDeployAppWhenComposeUpFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	_, err := Controller.DeployApp(DESCRIPTION_YAML)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknowError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledDeployAppWhenYAMLToJSONFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	
	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	
	_, err := Controller.DeployApp(WRONG_DESCRIPTION_JSON)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYAMLError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledDeployAppWhenInsertComposeFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)


	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
		dbManagerMockObj.EXPECT().InsertComposeFile(DESCRIPTION_JSON).Return(nil, UnknownError),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	
	_, err := Controller.DeployApp(DESCRIPTION_YAML)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InsertComposeFileFailed", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledApps_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppList().Return(DB_OBJs, nil),
	)

	res, err := Controller.Apps()

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

	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppList().Return(nil, UnknownError),
	)

	_, err := Controller.Apps()

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)


	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(COMPOSE_FILE_PATH, SERVICE_NAME).Return(PS_RETURN_MSG, nil),
		dockerExecutorMockObj.EXPECT().Inspect(CONTAINER_NAME).Return(INSPECT_RETURN_MSG, nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	res, err := Controller.App(APP_ID)

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

	dbManagerMockObj := dbmocks.MockCommand(ctrl)


	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	_, err := Controller.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenJSONToYAMLFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(WRONG_DB_GET_OBJ, nil),
	)

	_, err := Controller.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenNoServiceFiledinYAML_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ_WITHOUT_SERVICE, nil),
	)

	_, err := Controller.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenGetServiceStateComposePsFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(COMPOSE_FILE_PATH, SERVICE_NAME).Return("", UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	_, err := Controller.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenGetServiceStateComposeInspectFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)


	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(COMPOSE_FILE_PATH, SERVICE_NAME).Return(PS_RETURN_MSG, nil),
		dockerExecutorMockObj.EXPECT().Inspect(CONTAINER_NAME).Return("", UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	_, err := Controller.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledAppWhenGetServiceStateUnmarshalFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)


	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Ps(COMPOSE_FILE_PATH, SERVICE_NAME).Return(PS_RETURN_MSG, nil),
		dockerExecutorMockObj.EXPECT().Inspect(CONTAINER_NAME).Return(WRONG_INSPECT_RETURN_MSG, nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	_, err := Controller.App(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledUpdateAppInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.MockCommand(ctrl)


	gomock.InOrder(
		dbManagerMockObj.EXPECT().UpdateAppInfo(APP_ID, DESCRIPTION_JSON).Return(nil),
	)

	err := Controller.UpdateAppInfo(APP_ID, DESCRIPTION_YAML)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledUpdateAppInfoWhenYAMLToJSON_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	err := Controller.UpdateAppInfo(APP_ID, WRONG_DESCRIPTION_JSON)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYAMLError", "nil")
	}
}

func TestCalledUpdateAppInfoWhenUpdateAppInfoFailed_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().UpdateAppInfo(APP_ID, DESCRIPTION_JSON).Return(InvalidYamlError),
	)

	err := Controller.UpdateAppInfo(APP_ID, DESCRIPTION_YAML)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "InvalidYamlError", "nil")
	}
}

func TestCalledStartApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Start(gomock.Any()).Return(nil),
		dbManagerMockObj.EXPECT().UpdateAppState(APP_ID, gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	err := Controller.StartApp(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledStartAppWhenSetYAMLFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	err := Controller.StartApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledStartAppWhenComposeStartFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Start(gomock.Any()).Return(UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	err := Controller.StartApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledStopApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Stop(gomock.Any()).Return(nil),
		dbManagerMockObj.EXPECT().UpdateAppState(APP_ID, gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	err := Controller.StopApp(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledStopAppWhenSetYAMLFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	err := Controller.StopApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledStopAppWhenComposeStopFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)
	
	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return(APP_STATE, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Stop(gomock.Any()).Return(UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	
	err := Controller.StopApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledUpdateApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	
	err := Controller.UpdateApp(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledUpdateAppWhenSetYAMLFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	err := Controller.UpdateApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledUpdateAppWhenComposePullFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().GetImageDigest(IMAGE_NAME).Return(IMAGE_DIGEST, nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	err := Controller.UpdateApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledUpdateAppWhenComposeUpFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().Pull(gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(UnknownError),
		dockerExecutorMockObj.EXPECT().GetImageDigest(IMAGE_NAME).Return(IMAGE_DIGEST, nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	err := Controller.UpdateApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledDeleteApp_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any()).Return(nil),
		dbManagerMockObj.EXPECT().DeleteApp(gomock.Any()).Return(nil),)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	err := Controller.DeleteApp(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledDeleteAppWhenSetYAMLFileFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	err := Controller.DeleteApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledDeleteAppWhenComposeDeleteFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any()).Return(UnknownError),
		dbManagerMockObj.EXPECT().GetAppState(APP_ID).Return("START", nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	err := Controller.DeleteApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

func TestCalledDeleteAppWhenDBDeleteAppFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)
	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
		dockerExecutorMockObj.EXPECT().DownWithRemoveImages(gomock.Any()).Return(nil),
		dbManagerMockObj.EXPECT().DeleteApp(gomock.Any()).Return(UnknownError),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj

	err := Controller.DeleteApp(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}
}

/* Inner Funtion TEST */

func TestCalledSetYamlFile_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(DB_GET_OBJ, nil),
	)

	err := setYamlFile(APP_ID)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledSetYamlFileWhenGetAppFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(nil, UnknownError),
	)

	err := setYamlFile(APP_ID)

	if err == nil {
		t.Errorf("Expected err: %s, actual err: %s", "UnknownError", "nil")
	}

	os.RemoveAll(COMPOSE_FILE_PATH)
}

func TestCalledSetYamlFileWhenJSONToYAMLFailed_ExpectErrorReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbManagerMockObj := dbmocks.MockCommand(ctrl)

	gomock.InOrder(
		dbManagerMockObj.EXPECT().GetApp(APP_ID).Return(WRONG_DB_GET_OBJ, nil),
	)

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
	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().GetImageDigest(IMAGE_NAME).Return(IMAGE_DIGEST, nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreRepoDigests(DESCRIPTION_JSON, test_state)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledRestoreStateInputSTOP_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	test_state := "STOP"
	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Stop(gomock.Any()).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState(test_state)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledRestoreStateInputSTART_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	test_state := "START"
	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState(test_state)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledRestoreStateInputUP_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	test_state := "UP"
	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState(test_state)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestCalledRestoreStateInputDEPLOY_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	test_state := "DEPLOY"
	dockerExecutorMockObj := dockermocks.NewMockDockerExecutorInterface(ctrl)

	gomock.InOrder(
		dockerExecutorMockObj.EXPECT().Up(gomock.Any()).Return(nil),
	)

	dockerExecutor = dockerExecutorMockObj

	err := restoreState(test_state)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}
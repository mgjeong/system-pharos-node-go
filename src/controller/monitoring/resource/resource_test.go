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
package resource

import (
	"commons/errors"
	dockermocks "controller/dockercontroller/mocks"
	dbmocks "db/bolt/service/mocks"
	"github.com/golang/mock/gomock"
	"testing"
)

const (
	appId                   = "000000000000000000000000"
	oldTag                  = "1.0"
	repositoryWithPortImage = "test_url:5000/test"
	descriptionYaml         = "services:\n  " + testService + ":\n    image: " + repositoryWithPortImage + ":" + oldTag + "\nversion: \"2\"\n"
	originDescriptionJson   = "{\"services\":{\"" + testService + "\":{\"image\":\"" + repositoryWithPortImage + ":" + oldTag + "\"}},\"version\":\"2\"}"
	servicePort             = 1234
	serviceStatus           = "running"
	exitCodeValue           = "0"
	testNumStr              = "0"
	testNum                 = 0
	testService             = "test_service"
	testContainerId         = "test_container_id"
	testContainerName       = "test_container_name"
	runningState            = "running"
)

var (
	service1 = map[string]interface{}{
		"blockinput":    testNumStr,
		"blockoutput":   testNumStr,
		"cid":           testContainerId,
		"cname":         testContainerName,
		"cpu":           testNumStr,
		"mem":           testNumStr,
		"memlimit":      testNumStr,
		"memusage":      testNumStr,
		"networkinput":  testNumStr,
		"networkoutput": testNumStr,
		"pids":          testNum,
	}

	service2 = map[string]interface{}{
		"blockinput":    testNumStr,
		"blockoutput":   testNumStr,
		"cid":           testContainerId,
		"cname":         testContainerName,
		"cpu":           testNumStr,
		"mem":           testNumStr,
		"memlimit":      testNumStr,
		"memusage":      testNumStr,
		"networkinput":  testNumStr,
		"networkoutput": testNumStr,
		"pids":          testNum,
	}

	serviceList = []map[string]interface{}{
		service1,
		service2,
	}

	dbGetAppObj = map[string]interface{}{
		"id":          appId,
		"state":       runningState,
		"description": originDescriptionJson,
		"images": []map[string]interface{}{
			{
				"name": repositoryWithPortImage,
				"changes": map[string]interface{}{
					"tag":   oldTag,
					"state": "update",
				},
			},
		},
	}

	UnknownError = errors.Unknown{}
)

func TestGetAppResourceInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(appId).Return(dbGetAppObj, nil),
		dockerExecutorMockObj.EXPECT().GetAppStats(appId, COMPOSE_FILE).Return(serviceList, nil),
	)

	// pass mockObj to a real object.
	dockerExecutor = dockerExecutorMockObj
	dbExecutor = dbExecutorMockObj

	result, err := Executor.GetAppResourceInfo(appId)

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if _, exist := result[SERVICES]; !exist {
		t.Errorf("Unexpected err: " + SERVICES + " key does not exist")
	}
}

func TestGetAppResourceInfoWhenGetAppFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(appId).Return(nil, UnknownError),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj

	_, err := Executor.GetAppResourceInfo(appId)

	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestGetAppResourceInfoWhenGetAppStatsFailed_ExpectReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbExecutorMockObj := dbmocks.NewMockCommand(ctrl)
	dockerExecutorMockObj := dockermocks.NewMockCommand(ctrl)

	gomock.InOrder(
		dbExecutorMockObj.EXPECT().GetApp(appId).Return(dbGetAppObj, nil),
		dockerExecutorMockObj.EXPECT().GetAppStats(appId, COMPOSE_FILE).Return(nil, UnknownError),
	)

	// pass mockObj to a real object.
	dbExecutor = dbExecutorMockObj
	dockerExecutor = dockerExecutorMockObj

	_, err := Executor.GetAppResourceInfo(appId)

	switch err.(type) {
	default:
		t.Errorf("Expected err: UnknownError, actual err: %s", err.Error())
	case errors.Unknown:
	}
}

func TestGetHostResourceInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result, err := Executor.GetHostResourceInfo()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if _, exist := result[CPU]; !exist {
		t.Errorf("Unexpected err: " + CPU + " key does not exist")
	}

	if _, exist := result[DISK]; !exist {
		t.Errorf("Unexpected err: " + DISK + " key does not exist")
	}

	if _, exist := result[MEM]; !exist {
		t.Errorf("Unexpected err: " + MEM + " key does not exist")
	}

	if _, exist := result[NETWORK]; !exist {
		t.Errorf("Unexpected err: " + NETWORK + " key does not exist")
	}
}

func TestGetCPUUsage_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result, err := getCPUUsage()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if result == nil || len(result) == 0 {
		t.Errorf("Unexpected err : " + CPU + " usage array is empty")

	}
}

func TestGetMemUsage_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result, err := getMemUsage()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if _, exist := result[TOTAL]; !exist {
		t.Errorf("Unexpected err: " + TOTAL + " key does not exist")
	}

	if _, exist := result[FREE]; !exist {
		t.Errorf("Unexpected err: " + FREE + " key does not exist")
	}

	if _, exist := result[USED]; !exist {
		t.Errorf("Unexpected err: " + USED + " key does not exist")
	}

	if _, exist := result[USEDPERCENT]; !exist {
		t.Errorf("Unexpected err: " + USEDPERCENT + " key does not exist")
	}
}

func TestGetDiskUsage_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := getDiskUsage()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}
}

func TestGetNetworkTrafficInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result, err := getNetworkTrafficInfo()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	for _, value := range result {
		if _, exist := value[INTERFACENAME]; !exist {
			t.Errorf("Unexpected err: " + INTERFACENAME + " key does not exist")
		}

		if _, exist := value[BYTESSENT]; !exist {
			t.Errorf("Unexpected err: " + BYTESSENT + " key does not exist")
		}

		if _, exist := value[BYTESRECV]; !exist {
			t.Errorf("Unexpected err: " + BYTESRECV + "key does not exist")
		}

		if _, exist := value[PACKETSSENT]; !exist {
			t.Errorf("Unexpected err: " + PACKETSSENT + " key does not exist")
		}

		if _, exist := value[PACKETSRECV]; !exist {
			t.Errorf("Unexpected err: " + PACKETSRECV + " key does not exist")
		}
	}
}

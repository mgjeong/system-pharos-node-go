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
	"github.com/golang/mock/gomock"
	"testing"
)

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

	result, err := getDiskUsage()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	for _, value := range result {
		if _, exist := value[PATH]; !exist {
			t.Errorf("Unexpected err: " + PATH + " key does not exist")
		}

		if _, exist := value[TOTAL]; !exist {
			t.Errorf("Unexpected err: " + TOTAL + " key does not exist")
		}

		if _, exist := value[FREE]; !exist {
			t.Errorf("Unexpected err: " + FREE + " key does not exist")
		}

		if _, exist := value[USED]; !exist {
			t.Errorf("Unexpected err: " + USED + " key does not exist")
		}

		if _, exist := value[USEDPERCENT]; !exist {
			t.Errorf("Unexpected err: " + USEDPERCENT + " key does not exist")
		}
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

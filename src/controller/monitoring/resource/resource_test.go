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

func TestGetResrouceInfo_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result, err := Executor.GetResourceInfo()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if _, ok := result["CPU"]; !ok {
		t.Errorf("Unexpected err: CPU key does not exist")
	}

	if _, ok := result["DISK"]; !ok {
		t.Errorf("Unexpected err: DISK key does not exist")
	}

	if _, ok := result["MEM"]; !ok {
		t.Errorf("Unexpected err: MEM key does not exist")
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
		t.Errorf("Unexpected err : CPU usage array is empty")

	}
}

func TestGetMemUsage_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result, err := getMemUsage()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if _, ok := result["Total"]; !ok {
		t.Errorf("Unexpected err: Total key does not exist")
	}

	if _, ok := result["Free"]; !ok {
		t.Errorf("Unexpected err: Free key does not exist")
	}

	if _, ok := result["Used"]; !ok {
		t.Errorf("Unexpected err: Used key does not exist")
	}

	if _, ok := result["UsedPercent"]; !ok {
		t.Errorf("Unexpected err: UsedPercent key does not exist")
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
		if _, ok := value["Path"]; !ok {
			t.Errorf("Unexpected err: Path key does not exist")
		}

		if _, ok := value["Total"]; !ok {
			t.Errorf("Unexpected err: Total key does not exist")
		}

		if _, ok := value["Free"]; !ok {
			t.Errorf("Unexpected err: Free key does not exist")
		}

		if _, ok := value["Used"]; !ok {
			t.Errorf("Unexpected err: Used key does not exist")
		}

		if _, ok := value["UsedPercent"]; !ok {
			t.Errorf("Unexpected err: UsedPercent key does not exist")
		}
	}

}

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

	if _, ok := result["cpu"]; !ok {
		t.Errorf("Unexpected err: cpu key does not exist")
	}

	if _, ok := result["disk"]; !ok {
		t.Errorf("Unexpected err: disk key does not exist")
	}

	if _, ok := result["mem"]; !ok {
		t.Errorf("Unexpected err: mem key does not exist")
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
		t.Errorf("Unexpected err : cpu usage array is empty")

	}
}

func TestGetMemUsage_ExpectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result, err := getMemUsage()

	if err != nil {
		t.Errorf("Unexpected err: %s", err.Error())
	}

	if _, ok := result["total"]; !ok {
		t.Errorf("Unexpected err: total key does not exist")
	}

	if _, ok := result["free"]; !ok {
		t.Errorf("Unexpected err: free key does not exist")
	}

	if _, ok := result["used"]; !ok {
		t.Errorf("Unexpected err: used key does not exist")
	}

	if _, ok := result["usedpercent"]; !ok {
		t.Errorf("Unexpected err: usedpercent key does not exist")
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
		if _, ok := value["path"]; !ok {
			t.Errorf("Unexpected err: path key does not exist")
		}

		if _, ok := value["total"]; !ok {
			t.Errorf("Unexpected err: total key does not exist")
		}

		if _, ok := value["free"]; !ok {
			t.Errorf("Unexpected err: free key does not exist")
		}

		if _, ok := value["used"]; !ok {
			t.Errorf("Unexpected err: used key does not exist")
		}

		if _, ok := value["usedpercent"]; !ok {
			t.Errorf("Unexpected err: usedpercent key does not exist")
		}
	}
}
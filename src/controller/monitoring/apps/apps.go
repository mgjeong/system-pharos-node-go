/*******************************************************************************
 * Copyright 2018 Samsung Electronics All Rights Reserved.
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
package apps

import (
	"commons/logger"
	"controller/dockercontroller"
	"controller/notification/apps"
	"db/bolt/service"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	EXITED_STATE           = "exited"
	RUNNING_STATE          = "running"
	PARTIALLY_EXITED_STATE = "partially exited"
	START                  = "start"
	DIE                    = "die"
)

type Command interface {
	EnableEventMonitoring(appId, path string) error
	DisableEventMonitoring(appId, path string) error
	LockUpdateAppState()
	UnlockUpdateAppState()
	GetEventChannel() chan dockercontroller.Event
}

type Executor struct{}

var dbExecutor service.Command
var dockerExecutor dockercontroller.Command
var notiExecutor apps.Command
var events chan dockercontroller.Event
var appStateMutex = &sync.Mutex{}

func init() {
	dockerExecutor = dockercontroller.Executor
	dbExecutor = service.Executor{}
	notiExecutor = apps.Executor{}

	events = make(chan dockercontroller.Event)
	startEventMonitoring()
}

func (Executor) GetEventChannel() chan dockercontroller.Event {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	return events
}

func (Executor) LockUpdateAppState() {
	appStateMutex.Lock()
}

func (Executor) UnlockUpdateAppState() {
	appStateMutex.Unlock()
}

func (Executor) EnableEventMonitoring(appId, path string) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	err := dockerExecutor.Events(appId, path, events)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}
	return nil
}

func (Executor) DisableEventMonitoring(appId, path string) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	err := dockerExecutor.Events(appId, path, nil)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}
	return nil
}

func startEventMonitoring() {
	go func() {
		for {
			select {
			case event := <-events:
				notiExecutor.SendNotification(event)
				if event.Status == START ||
					event.Status == DIE {
					appStateMutex.Lock()
					updateAppState(event)
					appStateMutex.Unlock()
				}
			}
		}
	}()
}

func updateAppState(event dockercontroller.Event) {
	app, err := dbExecutor.GetApp(event.AppID)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return
	}

	if app["state"] == nil {
		logger.Logging(logger.DEBUG, "There is no state information. It must be Deploy API events")
		return
	}

	if app["state"].(string) == EXITED_STATE {
		logger.Logging(logger.DEBUG, "App state is exited")
		return
	}

	description := make(map[string]interface{})
	err = json.Unmarshal([]byte(app["description"].(string)), &description)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return
	}

	yaml, err := yaml.Marshal(description)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return
	}

	err = ioutil.WriteFile("docker-compose.yml", yaml, os.FileMode(0755))
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return
	}
	defer os.RemoveAll("docker-compose.yml")

	if description["services"] == nil || len(description["services"].(map[string]interface{})) == 0 {
		return
	}

	exitedServiceCnt := 0
	serviceCnt := len(description["services"].(map[string]interface{}))

	for _, serviceName := range reflect.ValueOf(description["services"].(map[string]interface{})).MapKeys() {
		infos, err := dockerExecutor.Ps(event.AppID, "docker-compose.yml", serviceName.String())
		if len(infos) == 0 {
			logger.Logging(logger.ERROR, "no information about service")
			return
		}
		if err != nil {
			logger.Logging(logger.ERROR, err.Error())
			return
		}

		if strings.Contains(infos[0]["State"], "Exited") {
			exitCode, _ := strconv.ParseUint(extractStringInParenthesis(infos[0]["State"]), 10, 32)
			if exitCode != 0 {
				exitedServiceCnt++
			}
		}
	}

	if exitedServiceCnt == 0 {
	} else if exitedServiceCnt < serviceCnt {
		dbExecutor.UpdateAppState(event.AppID, PARTIALLY_EXITED_STATE)
	} else if exitedServiceCnt == serviceCnt {
		dbExecutor.UpdateAppState(event.AppID, EXITED_STATE)
	}
}

func extractStringInParenthesis(s string) string {
	i := strings.Index(s, "(")
	if i >= 0 {
		j := strings.Index(s[i:], ")")
		if j >= 0 {
			return s[i+1 : j+i]
		}
	}
	return ""
}

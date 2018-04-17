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
)

type Command interface {
	EnableEventMonitoring(appId, path string) error
	DisableEventMonitoring(appId, path string) error
	GetEventChannel() chan dockercontroller.Event
}

type Executor struct{}

var dockerExecutor dockercontroller.Command
var notiExecutor apps.Command
var events chan dockercontroller.Event

func init() {
	dockerExecutor = dockercontroller.Executor
	notiExecutor = apps.Executor{}

	events = make(chan dockercontroller.Event)
	startEventMonitoring()
}

func (Executor) GetEventChannel() chan dockercontroller.Event {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	return events
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
			for event := range events {
				notiExecutor.SendNotification(event)
			}
		}
	}()
}
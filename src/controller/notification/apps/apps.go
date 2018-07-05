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
	"commons/errors"
	"commons/logger"
	"commons/url"
	"commons/util"
	"controller/configuration"
	"controller/dockercontroller"
	"db/bolt/event"
	"db/bolt/service"
	"encoding/json"
	"messenger"
	"strings"
)

const (
	COMPOSE_FILE        = "docker-compose.yaml"
	DESCRIPTION         = "description"
	SERVICES            = "services"
	IMAGE               = "image"
	HTTP_TAG            = "http://"
	DEFAULT_ANCHOR_PORT = "48099"
)

type Command interface {
	SubscribeEvent(body string) (map[string]interface{}, error)
	SendNotification(event dockercontroller.Event)
	UnsubscribeEvent(body string) error
}

type Executor struct{}

var httpExecutor messenger.Command
var dockerExecutor dockercontroller.Command
var dbExecutor event.Command
var serviceExecutor service.Command
var configurator configuration.Command

var Events chan dockercontroller.Event

func init() {
	httpExecutor = messenger.NewExecutor()
	dockerExecutor = dockercontroller.Executor
	dbExecutor = event.Executor{}
	serviceExecutor = service.Executor{}
	configurator = configuration.Executor{}

	Events = make(chan dockercontroller.Event)
}

func (Executor) SubscribeEvent(body string) (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	bodyMap, err := convertJsonToMap(body)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return nil, err
	}

	var appId, imageName string
	if value, exists := bodyMap["appid"]; exists {
		appId = value.(string)
	}
	if value, exists := bodyMap["imagename"]; exists {
		imageName = value.(string)
	}

	event, err := dbExecutor.InsertEvent(bodyMap["eventid"].(string), appId, imageName)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		switch err.(type) {
		default:
			return nil, err
		case errors.AlreadyReported:
			return event, err
		}
	}

	return event, err
}

func (Executor) SendNotification(e dockercontroller.Event) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	cid := ""
	timestamp := ""

	if e.Type == dockercontroller.IMAGE {
		logger.Logging(logger.DEBUG, "received event info: e.ID=", e.ID, "appId="+e.AppID+", serviceName="+e.ServiceName+", status="+e.Status)
	} else if e.Type == dockercontroller.CONTAINER {
		logger.Logging(logger.DEBUG, "received event info: e.ID=", e.ID, "appId="+e.AppID+", serviceName="+e.ServiceName+", cid="+e.CID+", status="+e.Status+", timestamp="+e.Timestamp)
		cid = e.CID
		timestamp = e.Timestamp
	}

	// Get docker image name from service name.
	imageName := getImageNameByServiceName(e.AppID, e.ServiceName)

	ids := make([]string, 0)
	if len(e.ID) == 0 {
		evts, _ := dbExecutor.GetEvents(e.AppID, imageName)
		if len(evts) == 0 {
			logger.Logging(logger.DEBUG, "There is no subscribers.")
			return
		}
		for _, evt := range evts {
			ids = append(ids, evt["id"].(string))
		}
	} else {
		ids = append(ids, e.ID)
	}

	nodeId := ""
	config, err := configurator.GetConfiguration()
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return
	}

	for _, prop := range config["properties"].([]map[string]interface{}) {
		if value, exists := prop["deviceid"]; exists {
			nodeId = value.(string)
		}
	}

	eventInfo := make(map[string]interface{})
	eventInfo["nodeid"] = nodeId
	eventInfo["appid"] = e.AppID
	eventInfo["status"] = e.Status
	eventInfo["imagename"] = imageName
	eventInfo["cid"] = cid
	eventInfo["timestamp"] = timestamp

	notiInfo := make(map[string]interface{})
	notiInfo["eventid"] = ids
	notiInfo["event"] = eventInfo

	// Notify container event to pharos-anchor.
	url, err := util.MakeAnchorRequestUrl(url.Notification(), url.Events())
	if err != nil {
		logger.Logging(logger.ERROR, "failed to make anchor request url")
		return
	}
	jsonData, _ := convertMapToJson(notiInfo)
	httpExecutor.SendHttpRequest("POST", url, []byte(jsonData))
}

func (Executor) UnsubscribeEvent(body string) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	bodyMap, err := convertJsonToMap(body)
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	err = dbExecutor.DeleteEvent(bodyMap["eventid"].(string))
	if err != nil {
		logger.Logging(logger.ERROR, err.Error())
		return err
	}

	return err
}

func getImageNameByServiceName(appId string, serviceName string) string {
	app, err := serviceExecutor.GetApp(appId)
	if err != nil {
		logger.Logging(logger.DEBUG, err.Error())
	}

	description := make(map[string]interface{})
	json.Unmarshal([]byte(app[DESCRIPTION].(string)), &description)

	for name, info := range description[SERVICES].(map[string]interface{}) {
		if strings.Compare(name, serviceName) == 0 {
			// Parse full image name to exclude tag.
			fullImageName := info.(map[string]interface{})[IMAGE].(string)
			words := strings.Split(fullImageName, "/")
			imageNameWithoutRepo := strings.Join(words[:len(words)-1], "/")
			repo := strings.Split(words[len(words)-1], ":")

			imageNameWithoutTag := imageNameWithoutRepo
			if len(words) > 1 {
				imageNameWithoutTag += "/"
			}
			imageNameWithoutTag += repo[0]
			return imageNameWithoutTag
		}
	}
	return ""
}

func convertJsonToMap(jsonStr string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, errors.InvalidJSON{"Unmarshalling Failed"}
	}
	return result, err
}

func convertMapToJson(data map[string]interface{}) (string, error) {
	result, err := json.Marshal(data)
	if err != nil {
		return "", errors.Unknown{"json marshalling failed"}
	}
	return string(result), nil
}

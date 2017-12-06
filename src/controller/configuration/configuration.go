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

// Package configuration provide virtual functionality of configuration.
package main

import (
	"commons/errors"
	"commons/logger"
	"encoding/json"
	"io/ioutil"
)

var configurationFileName = "./configuration.json"

// Interface of configuration operations.
type Command interface {
	// GetConfiguration returns a map of configuration stored in predefined configuration file.
	GetConfiguration() (map[string]interface{}, error)
	
	// SetConfiguration updates one of configurations
	SetConfiguration(map[string]interface{}) error
}

// Configuration schema
type Configuration struct {
    ServerAddress   string `json:"serveraddress"`
    DeviceName		string `json:"devicename"`
    DeviceID		string `json:"deviceid"`
    Manufacturer	string `json:"manufacturer"`
    ModelNumber		string `json:"modelnumber"`
    SerialNumber	string `json:"serialnumber"`
    Platform		string `json:"platform"`
    OS				string `json:"os"`
    Location		string `json:"location"`
}

func (conf Configuration) convertToMap() map[string]interface{} {
	return map[string]interface{}{
		"serveraddress"  : conf.ServerAddress,
		"devicename"     : conf.DeviceName,
		"deviceid"       : conf.DeviceID,
		"manufacturer"   : conf.Manufacturer,
		"modelnumber"    : conf.ModelNumber,
		"serialnumber"   : conf.SerialNumber,
		"platform"       : conf.Platform,
		"os"             : conf.OS,
		"location"       : conf.Location,
	}
}

func GetConfiguration() (map[string]interface{}, error) {
	
	raw, err := ioutil.ReadFile(configurationFileName)
    if err != nil {
        logger.Logging(logger.DEBUG, "Configuration file is not found.")
		return nil, errors.NotFound{configurationFileName}
    }

    var conf map[string]interface{}
    res := json.Unmarshal(raw, &conf)

    if res != nil {
        logger.Logging(logger.DEBUG, "Unmarshaling is failed")
		return nil, errors.Unknown{"Unmarshaling is failed"}
    }

    return conf, nil
}

func SetConfiguration (conf map[string]interface{}) error {
	
	_, err := json.Marshal(conf)
    if err != nil {
        logger.Logging(logger.DEBUG, "Converting map to JSON is failed")
		return errors.InvalidParam{"Converting map to JSON is failed"}
    }

    // TODO: Configuration file update may be needed

    return err
}
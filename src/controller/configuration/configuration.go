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
	//"commons/errors"
	//"commons/logger"
	"os"
	"fmt"
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
    ServerAddress   string `json:"server"`
    DeviceName		string `json:"devicename"`
}

func (conf Configuration) convertToMap() map[string]interface{} {
	return map[string]interface{}{
		"server":    		conf.ServerAddress,
		"devicename":		conf.DeviceName,
	}
}

func main() {
	
	conf, _:= GetConfiguration();   
    
    fmt.Println(conf["server"])
    
    newConf := Configuration{"10.113.65.119", "Jihun's Windows Desktop"}
	
    _ = SetConfiguration(newConf.convertToMap())
}

func GetConfiguration() (map[string]interface{}, error) {
	
	raw, err := ioutil.ReadFile(configurationFileName)
    if err != nil {
    	/*
        logger.Logging(logger.DEBUG, "Configuration file is not found.")		
		return nil, errors.NotFound{configurationFileName}
		*/
    	os.Exit(1);
    }
    
    var conf map[string]interface{}
    _= json.Unmarshal(raw, &conf)
    
    return conf, nil
}

func SetConfiguration (conf map[string]interface{}) error {
	
	jsonBytes, err := json.Marshal(conf)
    if err != nil {
    	/*
        logger.Logging(logger.DEBUG, "Converting map to JSON is failed")		
		return errors.InvalidParam{"Converting map to JSON is failed"}
		*/
        os.Exit(1);
    }
    
    jsonString := string(jsonBytes)
    fmt.Println(jsonString)
    
    return nil	
}

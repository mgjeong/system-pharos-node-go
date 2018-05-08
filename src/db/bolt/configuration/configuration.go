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

package configuration

import (
	"commons/errors"
	"commons/logger"
	. "db/bolt/wrapper"
	"encoding/json"
)

// Interface of Service model's operations.
type Command interface {
	// SetProperty updates configuration sets.
	SetProperty(property map[string]interface{}) error

	// GetProperty returns a single configuration property specified by name parameter.
	GetProperty(name string) (map[string]interface{}, error)

	// GetProperties returns a list of configurations stored in database.
	GetProperties() ([]map[string]interface{}, error)
}

const (
	BUCKET_NAME = "configuration"
)

type Property struct {
	Name     string      `json:"name"`
	Value    interface{} `json:"value"`
	ReadOnly bool        `json:"readonly"`
}

type Executor struct {
}

var db Database

func init() {
	db = NewBoltDB(BUCKET_NAME)
}

// Convert to map by object of struct Configuration.
// will return App information as map.
func (prop Property) convertToMap() map[string]interface{} {
	return map[string]interface{}{
		"name":     prop.Name,
		"value":    prop.Value,
		"readOnly": prop.ReadOnly,
	}
}

func (prop Property) encode() ([]byte, error) {
	encoded, err := json.Marshal(prop)
	if err != nil {
		return nil, errors.InvalidJSON{Msg: err.Error()}
	}
	return encoded, nil
}

func decode(data []byte) (*Property, error) {
	var prop *Property
	err := json.Unmarshal(data, &prop)
	if err != nil {
		return nil, errors.InvalidJSON{Msg: err.Error()}
	}
	return prop, nil
}

// SetProperty inserts a map of configuration into the database.
// if succeed to add new configuration sets, returns an error as nil.
// otherwise, return error.
func (Executor) SetProperty(property map[string]interface{}) error {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	value, err := db.Get([]byte(property["name"].(string)))
	if err != nil {
		switch err.(type) {
		default:
			return err
		case errors.NotFound:
			prop := Property{
				Name:     property["name"].(string),
				Value:    property["value"].(interface{}),
				ReadOnly: property["readOnly"].(bool),
			}

			encoded, err := prop.encode()
			if err != nil {
				return err
			}

			err = db.Put([]byte(property["name"].(string)), encoded)
			if err != nil {
				return err
			}
			return nil
		}
	}

	prop, err := decode(value)
	if err != nil {
		return err
	}

	prop.Value = property["value"]
	encoded, err := prop.encode()
	if err != nil {
		return err
	}

	return db.Put([]byte(property["name"].(string)), encoded)
}

// GetProperty returns a single configuration property specified by name parameter.
// if succeed to get, returns an error as nil.
// otherwise, return error.
func (Executor) GetProperty(name string) (map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	value, err := db.Get([]byte(name))
	if err != nil {
		return nil, err
	}

	prop, err := decode(value)
	if err != nil {
		return nil, err
	}

	return prop.convertToMap(), nil
}

// GetProperties returns a list of configurations stored in database.
// if succeed to get, return list of all configurations as slice.
// otherwise, return error.
func (Executor) GetProperties() ([]map[string]interface{}, error) {
	logger.Logging(logger.DEBUG, "IN")
	defer logger.Logging(logger.DEBUG, "OUT")

	props, err := db.List()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, value := range props {
		prop, err := decode([]byte(value.(string)))
		if err != nil {
			continue
		}
		result = append(result, prop.convertToMap())
	}
	return result, nil
}

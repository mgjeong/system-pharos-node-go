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

package modelinterface

// Interface of Service model's operations.
type Service interface {
	// InsertComposeFile insert docker-compose file for new service.
	InsertComposeFile(description string) (map[string]interface{}, error)

	// GetAppList returns all of app's IDs.
	GetAppList() ([]map[string]interface{}, error)

	// GetApp returns docker-compose data of target app.
	GetApp(app_id string) (map[string]interface{}, error)

	// UpdateAppInfo updates docker-compose data of target app.
	UpdateAppInfo(app_id string, description string) error

	// DeleteApp delete docker-compose data of target app.
	DeleteApp(app_id string) error

	// GetAppState returns app's state
	GetAppState(app_id string) (string, error)

	// UpdateAppState updates app's State.
	UpdateAppState(app_id string, state string) error
}

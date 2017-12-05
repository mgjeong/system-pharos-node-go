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

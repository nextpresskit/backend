package config

// AppConfig holds the application-level configuration.
type AppConfig struct {
	Name            string // Application name
	Env             string // Environment: local, dev, staging, production
	Port            string // Port on which the API will run
	APIBasePath     string // Optional API path prefix (e.g. /v1)
	LogIdentifier   string // Startup/shutdown log field "service" (APP_LOG_IDENTIFIER)
}

// LoadAppConfig reads environment variables and returns AppConfig.
func LoadAppConfig() AppConfig {
	return AppConfig{
		Name:          GetEnv("APP_NAME", "NextPressKit"),
		Env:           GetEnv("APP_ENV", "local"),
		Port:          GetEnv("APP_PORT", "9090"),
		APIBasePath:   normalizeBasePath(GetEnv("API_BASE_PATH", "")),
		LogIdentifier: GetEnv("APP_LOG_IDENTIFIER", "nextpresskit-backend"),
	}
}

func normalizeBasePath(raw string) string {
	path := raw
	for len(path) > 0 && path[len(path)-1] == ' ' {
		path = path[:len(path)-1]
	}
	for len(path) > 0 && path[0] == ' ' {
		path = path[1:]
	}
	if path == "" || path == "/" {
		return ""
	}
	if path[0] != '/' {
		path = "/" + path
	}
	for len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return path
}

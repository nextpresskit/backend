package config

// AppConfig holds the application-level configuration.
type AppConfig struct {
	Name string // Application name
	Env  string // Environment: local, dev, staging, production
	Port string // Port on which the API will run
}

// LoadAppConfig reads environment variables and returns AppConfig.
func LoadAppConfig() AppConfig {
	return AppConfig{
		Name: GetEnv("APP_NAME", "NextPress"),
		Env:  GetEnv("APP_ENV", "local"),
		Port: GetEnv("APP_PORT", "9090"),
	}
}

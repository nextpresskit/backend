package config

type AppConfig struct {
	Name string
	Env  string
	Port string
}

func LoadAppConfig() AppConfig {
	return AppConfig{
		Name: GetEnv("APP_NAME", "NextPress"),
		Env:  GetEnv("APP_ENV", "development"),
		Port: GetEnv("APP_PORT", "9090"),
	}
}

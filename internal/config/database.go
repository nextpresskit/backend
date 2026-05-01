package config

type DBConfig struct {
	Driver   string
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

func LoadDBConfig() DBConfig {
	return DBConfig{
		Driver:   GetEnv("DB_DRIVER", "postgres"),
		Host:     GetEnv("DB_HOST", "127.0.0.1"),
		Port:     GetEnv("DB_PORT", "5432"),
		Name:     GetEnv("DB_NAME", "nextpresskit"),
		User:     GetEnv("DB_USER", "postgres"),
		Password: GetEnv("DB_PASSWORD", "secret"),
		SSLMode:  GetEnv("DB_SSLMODE", "disable"),
	}
}

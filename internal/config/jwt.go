package config

import "time"

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func LoadJWTConfig() JWTConfig {
	return JWTConfig{
		Secret:     GetEnv("JWT_SECRET", "dev-secret-change-me"),
		AccessTTL:  mustParseDuration(GetEnv("JWT_ACCESS_TTL", "15m")),
		RefreshTTL: mustParseDuration(GetEnv("JWT_REFRESH_TTL", "168h")), // 7 days
	}
}

func mustParseDuration(v string) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		// Fallback to 15 minutes if misconfigured
		return 15 * time.Minute
	}
	return d
}

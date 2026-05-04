package config

import "time"

type RateLimitConfig struct {
	Enabled bool

	PublicMaxPerMinute int
	AuthMaxPerMinute   int
	AdminMaxPerMinute  int

	Window time.Duration

	RedisEnabled  bool
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisPrefix   string
}

func LoadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:          parseBool(GetEnv("RATE_LIMIT_ENABLED", "true")),
		PublicMaxPerMinute: parseInt(GetEnv("RATE_LIMIT_PUBLIC_MAX_PER_MINUTE", "120"), 120),
		AuthMaxPerMinute:   parseInt(GetEnv("RATE_LIMIT_AUTH_MAX_PER_MINUTE", "30"), 30),
		AdminMaxPerMinute:  parseInt(GetEnv("RATE_LIMIT_ADMIN_MAX_PER_MINUTE", "60"), 60),
		Window:            time.Minute,
		RedisEnabled:      parseBool(GetEnv("RATE_LIMIT_REDIS_ENABLED", "false")),
		RedisAddr:         GetEnv("RATE_LIMIT_REDIS_ADDR", ""),
		RedisPassword:     GetEnv("RATE_LIMIT_REDIS_PASSWORD", ""),
		RedisDB:           parseInt(GetEnv("RATE_LIMIT_REDIS_DB", "0"), 0),
		RedisPrefix:       GetEnv("RATE_LIMIT_REDIS_PREFIX", "nextpresskit:ratelimit"),
	}
}

func parseBool(v string) bool {
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func parseInt(v string, fallback int) int {
	n := 0
	seenDigit := false
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c < '0' || c > '9' {
			continue
		}
		seenDigit = true
		n = n*10 + int(c-'0')
	}
	if !seenDigit {
		return fallback
	}
	return n
}


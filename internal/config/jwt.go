package config

import (
	"strings"
	"time"
)

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration

	// AuthSource controls where the backend reads access JWTs from.
	// Supported values: "cookie", "header" (case-insensitive).
	AuthSource string

	AccessCookieName  string
	RefreshCookieName string

	// CookieSecure forces the Secure attribute (recommended: true in production).
	CookieSecure bool
	// CookieSameSite should be one of: "lax", "strict", "none" (case-insensitive).
	CookieSameSite string
	CookiePath     string
	CookieDomain   string
}

func LoadJWTConfig() JWTConfig {
	authSource := strings.ToLower(strings.TrimSpace(GetEnv("JWT_AUTH_SOURCE", "cookie")))
	if authSource != "cookie" && authSource != "header" {
		authSource = "cookie"
	}

	return JWTConfig{
		Secret:     GetEnv("JWT_SECRET", "dev-secret-change-me"),
		AccessTTL:  mustParseDuration(GetEnv("JWT_ACCESS_TTL", "15m")),
		RefreshTTL: mustParseDuration(GetEnv("JWT_REFRESH_TTL", "168h")), // 7 days

		AuthSource: authSource,

		AccessCookieName:  GetEnv("JWT_ACCESS_COOKIE_NAME", "access_token"),
		RefreshCookieName: GetEnv("JWT_REFRESH_COOKIE_NAME", "refresh_token"),

		CookieSecure:   parseBool(GetEnv("JWT_COOKIE_SECURE", "true")),
		CookieSameSite: strings.ToLower(strings.TrimSpace(GetEnv("JWT_COOKIE_SAME_SITE", "none"))),
		CookiePath:     GetEnv("JWT_COOKIE_PATH", "/"),
		CookieDomain:   GetEnv("JWT_COOKIE_DOMAIN", ""),
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

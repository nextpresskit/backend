package jwtcookie

import (
	"net/http"
	"strings"
	"time"

	"github.com/nextpresskit/backend/internal/config"
)

func cookieSameSiteMode(s string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		// Default to "none" because this project targets cross-site cookies by default.
		return http.SameSiteNoneMode
	}
}

// SetAuthCookies sets both access and refresh JWT cookies.
// It uses secure defaults intended for cross-site usage.
func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken string, jwtCfg config.JWTConfig) {
	now := time.Now().UTC()

	// SameSite=None requires Secure in modern browsers.
	secure := jwtCfg.CookieSecure
	if cookieSameSiteMode(jwtCfg.CookieSameSite) == http.SameSiteNoneMode {
		secure = true
	}

	path := jwtCfg.CookiePath
	if strings.TrimSpace(path) == "" {
		path = "/"
	}

	// Access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     jwtCfg.AccessCookieName,
		Value:    accessToken,
		Path:     path,
		Domain:   strings.TrimSpace(jwtCfg.CookieDomain),
		HttpOnly: true,
		Secure:   secure,
		SameSite: cookieSameSiteMode(jwtCfg.CookieSameSite),
		Expires:  now.Add(jwtCfg.AccessTTL),
		MaxAge:   int(jwtCfg.AccessTTL.Seconds()),
	})

	// Refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     jwtCfg.RefreshCookieName,
		Value:    refreshToken,
		Path:     path,
		Domain:   strings.TrimSpace(jwtCfg.CookieDomain),
		HttpOnly: true,
		Secure:   secure,
		SameSite: cookieSameSiteMode(jwtCfg.CookieSameSite),
		Expires:  now.Add(jwtCfg.RefreshTTL),
		MaxAge:   int(jwtCfg.RefreshTTL.Seconds()),
	})
}

func ClearAuthCookies(w http.ResponseWriter, jwtCfg config.JWTConfig) {
	// Clear by setting MaxAge=-1 and an expiration in the past.
	past := time.Unix(0, 0).UTC()
	secure := jwtCfg.CookieSecure
	if cookieSameSiteMode(jwtCfg.CookieSameSite) == http.SameSiteNoneMode {
		secure = true
	}

	path := jwtCfg.CookiePath
	if strings.TrimSpace(path) == "" {
		path = "/"
	}

	http.SetCookie(w, &http.Cookie{
		Name:     jwtCfg.AccessCookieName,
		Value:    "",
		Path:     path,
		Domain:   strings.TrimSpace(jwtCfg.CookieDomain),
		HttpOnly: true,
		Secure:   secure,
		SameSite: cookieSameSiteMode(jwtCfg.CookieSameSite),
		Expires:  past,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     jwtCfg.RefreshCookieName,
		Value:    "",
		Path:     path,
		Domain:   strings.TrimSpace(jwtCfg.CookieDomain),
		HttpOnly: true,
		Secure:   secure,
		SameSite: cookieSameSiteMode(jwtCfg.CookieSameSite),
		Expires:  past,
		MaxAge:   -1,
	})
}

func GetCookieValue(r *http.Request, cookieName string) (string, bool) {
	if r == nil {
		return "", false
	}
	ck, err := r.Cookie(cookieName)
	if err != nil || ck == nil {
		return "", false
	}
	if strings.TrimSpace(ck.Value) == "" {
		return "", false
	}
	return ck.Value, true
}


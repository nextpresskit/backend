package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/nextpresskit/backend/internal/config"
)

type dummyAccessTokenParser struct{}

func (p dummyAccessTokenParser) ParseAccessToken(token string) (string, error) {
	if token == "good" {
		return "user-123", nil
	}
	return "", errors.New("invalid token")
}

func TestAuthRequired_MissingAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	jwtCfg := config.JWTConfig{
		AuthSource:       "header",
		AccessCookieName: "access_token",
	}
	r.GET("/private", AuthRequired(dummyAccessTokenParser{}, jwtCfg), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAuthRequired_InvalidAuthorizationHeaderScheme(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	jwtCfg := config.JWTConfig{
		AuthSource:       "header",
		AccessCookieName: "access_token",
	}
	r.GET("/private", AuthRequired(dummyAccessTokenParser{}, jwtCfg), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Basic abc")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	jwtCfg := config.JWTConfig{
		AuthSource:       "header",
		AccessCookieName: "access_token",
	}
	r.GET("/private", AuthRequired(dummyAccessTokenParser{}, jwtCfg), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer bad")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAuthRequired_ValidTokenSetsUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	jwtCfg := config.JWTConfig{
		AuthSource:       "header",
		AccessCookieName: "access_token",
	}
	r.GET("/private", AuthRequired(dummyAccessTokenParser{}, jwtCfg), func(c *gin.Context) {
		v, ok := c.Get(ContextUserIDKey)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "missing_user_context"})
			return
		}
		userID, _ := v.(string)
		c.JSON(http.StatusOK, gin.H{"userId": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer good")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestAuthRequired_ValidAccessTokenCookieSetsUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	jwtCfg := config.JWTConfig{
		AuthSource:       "cookie",
		AccessCookieName: "access_token",
	}

	r.GET("/private", AuthRequired(dummyAccessTokenParser{}, jwtCfg), func(c *gin.Context) {
		v, ok := c.Get(ContextUserIDKey)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "missing_user_context"})
			return
		}
		userID, _ := v.(string)
		c.JSON(http.StatusOK, gin.H{"userId": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "good", Path: "/"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

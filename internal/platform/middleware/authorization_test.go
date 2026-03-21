package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockPermissionChecker struct {
	allowed bool
	err     error
}

func (m mockPermissionChecker) UserHasPermission(_ context.Context, _ string, _ string) (bool, error) {
	return m.allowed, m.err
}

func TestRequirePermission_MissingUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	checker := mockPermissionChecker{allowed: true}

	r.GET("/admin", RequirePermission(checker, "perm"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestRequirePermission_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	checker := mockPermissionChecker{allowed: false}

	r.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, "u1")
		c.Next()
	})

	r.GET("/admin", RequirePermission(checker, "perm"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestRequirePermission_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	checker := mockPermissionChecker{allowed: true}

	r.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, "u1")
		c.Next()
	})

	r.GET("/admin", RequirePermission(checker, "perm"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestRequirePermission_CheckerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	checker := mockPermissionChecker{allowed: false, err: errors.New("boom")}

	r.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, "u1")
		c.Next()
	})

	r.GET("/admin", RequirePermission(checker, "perm"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}


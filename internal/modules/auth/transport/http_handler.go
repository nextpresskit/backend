package transport

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nextpresskit/backend/internal/config"
	"github.com/nextpresskit/backend/internal/modules/auth/application"
	userdomain "github.com/nextpresskit/backend/internal/modules/user/domain"
	platformmw "github.com/nextpresskit/backend/internal/platform/middleware"
	jwtcookie "github.com/nextpresskit/backend/internal/platform/jwtcookie"
)

type Handler struct {
	svc            *application.Service
	authMiddleware gin.HandlerFunc

	jwtCfg config.JWTConfig
}

func NewHandler(svc *application.Service, authMiddleware gin.HandlerFunc, jwtCfg config.JWTConfig) *Handler {
	return &Handler{svc: svc, authMiddleware: authMiddleware, jwtCfg: jwtCfg}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/auth")
	g.POST("/register", h.register)
	g.POST("/login", h.login)
	g.POST("/refresh", h.refresh)
	g.POST("/logout", h.logout)
	g.GET("/me", h.authMiddleware, h.me)
}

type registerRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
}

func (h *Handler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	u, err := h.svc.Register(c.Request.Context(), req.FirstName, req.LastName, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        u.ID,
		"uuid":      u.UUID,
		"firstName": u.FirstName,
		"lastName":  u.LastName,
		"email":     u.Email,
	})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	u, access, refresh, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}
	relations, err := h.svc.Relations(c.Request.Context(), fmt.Sprintf("%d", u.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	if strings.EqualFold(h.jwtCfg.AuthSource, "cookie") {
		jwtcookie.SetAuthCookies(c.Writer, access, refresh, h.jwtCfg)
		c.JSON(http.StatusOK, gin.H{
			"user": userToJSON(u, relations),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": gin.H{
			"accessToken":  access,
			"refreshToken": refresh,
		},
		"user": userToJSON(u, relations),
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshRequest
	// In cookie auth mode, the refresh token is read from cookies (request body may be empty).
	if strings.EqualFold(h.jwtCfg.AuthSource, "header") || c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
			return
		}
	}

	refreshToken := strings.TrimSpace(req.RefreshToken)
	if strings.EqualFold(h.jwtCfg.AuthSource, "cookie") {
		if v, ok := jwtcookie.GetCookieValue(c.Request, h.jwtCfg.RefreshCookieName); ok {
			refreshToken = v
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh_token"})
			return
		}
	}
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	u, access, refresh, err := h.svc.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh_token"})
		return
	}
	relations, err := h.svc.Relations(c.Request.Context(), fmt.Sprintf("%d", u.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	if strings.EqualFold(h.jwtCfg.AuthSource, "cookie") {
		jwtcookie.SetAuthCookies(c.Writer, access, refresh, h.jwtCfg)
		c.JSON(http.StatusOK, gin.H{
			"user": userToJSON(u, relations),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": gin.H{
			"accessToken":  access,
			"refreshToken": refresh,
		},
		"user": userToJSON(u, relations),
	})
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) logout(c *gin.Context) {
	var req logoutRequest
	// In cookie auth mode, the refresh token is read from cookies.
	if strings.EqualFold(h.jwtCfg.AuthSource, "header") || c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
			return
		}
	}

	refreshToken := strings.TrimSpace(req.RefreshToken)
	if strings.EqualFold(h.jwtCfg.AuthSource, "cookie") {
		if v, ok := jwtcookie.GetCookieValue(c.Request, h.jwtCfg.RefreshCookieName); ok {
			refreshToken = v
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh_token"})
			return
		}
	}
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	if err := h.svc.Logout(c.Request.Context(), refreshToken); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh_token"})
		return
	}

	if strings.EqualFold(h.jwtCfg.AuthSource, "cookie") {
		jwtcookie.ClearAuthCookies(c.Writer, h.jwtCfg)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Signed out."})
}

func (h *Handler) me(c *gin.Context) {
	v, ok := c.Get(platformmw.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_or_expired_token"})
		return
	}
	uid, _ := v.(string)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_or_expired_token"})
		return
	}

	u, err := h.svc.Me(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, application.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	relations, err := h.svc.Relations(c.Request.Context(), fmt.Sprintf("%d", u.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": userToJSON(u, relations)})
}

func userToJSON(u *userdomain.User, relations application.UserRelations) gin.H {
	if u == nil {
		return nil
	}
	out := gin.H{
		"id":        u.ID,
		"uuid":      string(u.UUID),
		"firstName": u.FirstName,
		"lastName":  u.LastName,
		"email":     u.Email,
		"active":    u.Active,
		"rbac": gin.H{
			"roles":       relations.RoleNames,
			"permissions": relations.PermissionCodes,
		},
	}
	if !u.CreatedAt.IsZero() {
		out["createdAt"] = u.CreatedAt.UTC().Format(time.RFC3339Nano)
	}
	if !u.UpdatedAt.IsZero() {
		out["updatedAt"] = u.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	if u.DeletedAt != nil && !u.DeletedAt.IsZero() {
		out["deletedAt"] = u.DeletedAt.UTC().Format(time.RFC3339Nano)
	}
	return out
}

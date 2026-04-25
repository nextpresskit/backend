package transport

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/auth/application"
	userdomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/user/domain"
	platformmw "github.com/Petar-V-Nikolov/nextpress-backend/internal/platform/middleware"
)

type Handler struct {
	svc            *application.Service
	authMiddleware gin.HandlerFunc
}

func NewHandler(svc *application.Service, authMiddleware gin.HandlerFunc) *Handler {
	return &Handler{svc: svc, authMiddleware: authMiddleware}
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

	c.JSON(http.StatusOK, gin.H{
		"tokens": gin.H{
			"accessToken":  access,
			"refreshToken": refresh,
		},
		"user": userToJSON(u, relations),
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	u, access, refresh, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh_token"})
		return
	}
	relations, err := h.svc.Relations(c.Request.Context(), fmt.Sprintf("%d", u.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
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
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh_token"})
		return
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

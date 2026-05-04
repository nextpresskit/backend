package transport

import (
	"net/http"

	"github.com/gin-gonic/gin"

	rbacApp "github.com/nextpresskit/backend/internal/modules/rbac/application"
)

type Handler struct {
	svc *rbacApp.Service
}

func NewHandler(svc *rbacApp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/roles", h.listRoles)
	rg.POST("/roles", h.createRole)

	rg.GET("/permissions", h.listPermissions)
	rg.POST("/permissions", h.createPermission)

	rg.POST("/roles/:role_id/permissions", h.grantPermissionToRole)
	rg.POST("/users/:user_id/roles", h.assignRoleToUser)
}

type createRoleRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) createRole(c *gin.Context) {
	var req createRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	role, err := h.svc.CreateRole(c.Request.Context(), req.Name)
	if err != nil {
		switch err {
		case rbacApp.ErrInvalidNameOrCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case rbacApp.ErrAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":   role.ID,
		"name": role.Name,
	})
}

func (h *Handler) listRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]gin.H, 0, len(roles))
	for _, r := range roles {
		out = append(out, gin.H{
			"id":   r.ID,
			"name": r.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{"roles": out})
}

type createPermissionRequest struct {
	Code string `json:"code" binding:"required"`
}

func (h *Handler) createPermission(c *gin.Context) {
	var req createPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	perm, err := h.svc.CreatePermission(c.Request.Context(), req.Code)
	if err != nil {
		switch err {
		case rbacApp.ErrInvalidNameOrCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case rbacApp.ErrAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":   perm.ID,
		"code": perm.Code,
	})
}

func (h *Handler) listPermissions(c *gin.Context) {
	perms, err := h.svc.ListPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]gin.H, 0, len(perms))
	for _, p := range perms {
		out = append(out, gin.H{
			"id":   p.ID,
			"code": p.Code,
		})
	}

	c.JSON(http.StatusOK, gin.H{"permissions": out})
}

type grantPermissionRequest struct {
	PermissionID string `json:"permission_id" binding:"required"`
}

func (h *Handler) grantPermissionToRole(c *gin.Context) {
	roleID := c.Param("role_id")

	var req grantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	if err := h.svc.GrantPermissionToRole(c.Request.Context(), roleID, req.PermissionID); err != nil {
		switch err {
		case rbacApp.ErrInvalidNameOrCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type assignRoleRequest struct {
	RoleID string `json:"role_id" binding:"required"`
}

func (h *Handler) assignRoleToUser(c *gin.Context) {
	userID := c.Param("user_id")

	var req assignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	if err := h.svc.AssignRoleToUser(c.Request.Context(), userID, req.RoleID); err != nil {
		switch err {
		case rbacApp.ErrInvalidNameOrCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}


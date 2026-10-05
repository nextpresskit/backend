package transport

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	userApp "github.com/nextpresskit/backend/internal/modules/user/application"
)

// UsersReader is the admin read surface of the user application service.
type UsersReader interface {
	List(ctx context.Context, query string, limit, offset int) (userApp.ListResult, error)
	Get(ctx context.Context, rawID string) (*userApp.UserWithRoles, error)
}

// ReadPermission guards the admin user endpoints.
const ReadPermission = "users:read"

type Handler struct {
	svc UsersReader
}

func NewHandler(svc UsersReader) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts GET /users and GET /users/:id on an authenticated admin group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requirePerm func(string) gin.HandlerFunc) {
	g := rg.Group("/users", requirePerm(ReadPermission))
	g.GET("", h.list)
	g.GET("/:id", h.get)
}

func (h *Handler) list(c *gin.Context) {
	limit, ok := optionalNonNegativeInt(c.Query("limit"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": userApp.ErrInvalidListArg.Error()})
		return
	}
	offset, ok := optionalNonNegativeInt(c.Query("offset"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": userApp.ErrInvalidListArg.Error()})
		return
	}

	res, err := h.svc.List(c.Request.Context(), c.Query("q"), limit, offset)
	if err != nil {
		respondError(c, err)
		return
	}

	out := make([]gin.H, 0, len(res.Users))
	for i := range res.Users {
		out = append(out, userToJSON(&res.Users[i]))
	}
	c.JSON(http.StatusOK, gin.H{
		"users":  out,
		"total":  res.Total,
		"limit":  res.Limit,
		"offset": res.Offset,
	})
}

func (h *Handler) get(c *gin.Context) {
	u, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": userToJSON(u)})
}

func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, userApp.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, userApp.ErrInvalidUserID), errors.Is(err, userApp.ErrInvalidListArg):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}

func optionalNonNegativeInt(raw string) (int, bool) {
	if raw == "" {
		return 0, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// userToJSON mirrors the /auth/me user field names; the password hash is never exposed.
func userToJSON(u *userApp.UserWithRoles) gin.H {
	out := gin.H{
		"id":        u.User.ID,
		"uuid":      u.User.UUID,
		"firstName": u.User.FirstName,
		"lastName":  u.User.LastName,
		"email":     u.User.Email,
		"active":    u.User.Active,
		"roles":     u.Roles,
	}
	if !u.User.CreatedAt.IsZero() {
		out["createdAt"] = u.User.CreatedAt.UTC().Format(time.RFC3339Nano)
	}
	if !u.User.UpdatedAt.IsZero() {
		out["updatedAt"] = u.User.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	return out
}

package transport

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	taxApp "github.com/nextpresskit/backend/internal/modules/taxonomy/application"
	taxDomain "github.com/nextpresskit/backend/internal/modules/taxonomy/domain"
)

type Handler struct {
	svc *taxApp.Service
}

func NewHandler(svc *taxApp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc, requirePerm func(string) gin.HandlerFunc) {
	cat := rg.Group("/categories")
	cat.GET("", auth, requirePerm("categories:read"), h.listCategories)
	cat.POST("", auth, requirePerm("categories:write"), h.createCategory)
	cat.PUT("/:id", auth, requirePerm("categories:write"), h.updateCategory)
	cat.DELETE("/:id", auth, requirePerm("categories:write"), h.deleteCategory)

	tags := rg.Group("/tags")
	tags.GET("", auth, requirePerm("tags:read"), h.listTags)
	tags.POST("", auth, requirePerm("tags:write"), h.createTag)
	tags.PUT("/:id", auth, requirePerm("tags:write"), h.updateTag)
	tags.DELETE("/:id", auth, requirePerm("tags:write"), h.deleteTag)
}

type createTaxRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

func (h *Handler) createCategory(c *gin.Context) {
	var req createTaxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	out, err := h.svc.CreateCategory(c.Request.Context(), req.Name, req.Slug)
	if err != nil {
		writeTaxError(c, err)
		return
	}
	c.JSON(http.StatusCreated, categoryToJSON(out))
}

func (h *Handler) listCategories(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	out, err := h.svc.ListCategories(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	items := make([]gin.H, 0, len(out))
	for i := range out {
		v := out[i]
		items = append(items, categoryToJSON(&v))
	}
	c.JSON(http.StatusOK, gin.H{"categories": items})
}

type updateTaxRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *Handler) updateCategory(c *gin.Context) {
	id := c.Param("id")
	var req updateTaxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	out, err := h.svc.UpdateCategory(c.Request.Context(), id, req.Name, req.Slug)
	if err != nil {
		writeTaxError(c, err)
		return
	}
	c.JSON(http.StatusOK, categoryToJSON(out))
}

func (h *Handler) deleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteCategory(c.Request.Context(), id); err != nil {
		writeTaxError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) createTag(c *gin.Context) {
	var req createTaxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	out, err := h.svc.CreateTag(c.Request.Context(), req.Name, req.Slug)
	if err != nil {
		writeTaxError(c, err)
		return
	}
	c.JSON(http.StatusCreated, tagToJSON(out))
}

func (h *Handler) listTags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	out, err := h.svc.ListTags(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	items := make([]gin.H, 0, len(out))
	for i := range out {
		v := out[i]
		items = append(items, tagToJSON(&v))
	}
	c.JSON(http.StatusOK, gin.H{"tags": items})
}

func (h *Handler) updateTag(c *gin.Context) {
	id := c.Param("id")
	var req updateTaxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	out, err := h.svc.UpdateTag(c.Request.Context(), id, req.Name, req.Slug)
	if err != nil {
		writeTaxError(c, err)
		return
	}
	c.JSON(http.StatusOK, tagToJSON(out))
}

func (h *Handler) deleteTag(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteTag(c.Request.Context(), id); err != nil {
		writeTaxError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func writeTaxError(c *gin.Context, err error) {
	switch err {
	case taxApp.ErrInvalidInput:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case taxApp.ErrAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case taxApp.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}

func categoryToJSON(v *taxDomain.Category) gin.H {
	return gin.H{"id": v.ID, "uuid": v.UUID, "name": v.Name, "slug": v.Slug}
}

func tagToJSON(v *taxDomain.Tag) gin.H {
	return gin.H{"id": v.ID, "uuid": v.UUID, "name": v.Name, "slug": v.Slug}
}


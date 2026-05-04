package transport

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	pageApp "github.com/nextpresskit/backend/internal/modules/pages/application"
	pageDomain "github.com/nextpresskit/backend/internal/modules/pages/domain"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
)

type Handler struct {
	svc *pageApp.Service
}

func NewHandler(svc *pageApp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc, requirePerm func(string) gin.HandlerFunc) {
	pages := rg.Group("/pages")

	pages.GET("", auth, requirePerm("pages:read"), h.list)
	pages.GET("/:id", auth, requirePerm("pages:read"), h.getByID)
	pages.POST("", auth, requirePerm("pages:write"), h.create)
	pages.PUT("/:id", auth, requirePerm("pages:write"), h.update)
	pages.DELETE("/:id", auth, requirePerm("pages:write"), h.delete)
}

func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	pages := rg.Group("/pages")
	pages.GET("/:slug", h.publicGetBySlug)
}

type createPageRequest struct {
	Title   string `json:"title" binding:"required"`
	Slug    string `json:"slug" binding:"required"`
	Content string `json:"content"`
}

func (h *Handler) create(c *gin.Context) {
	var req createPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	v, _ := c.Get(platformMiddleware.ContextUserIDKey)
	authorID, _ := v.(string)

	p, err := h.svc.Create(c.Request.Context(), authorID, req.Title, req.Slug, req.Content)
	if err != nil {
		switch err {
		case pageApp.ErrInvalidPage:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case pageApp.ErrSlugTaken:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusCreated, pageToJSON(p))
}

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")
	p, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == pageApp.ErrPageNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, pageToJSON(p))
}

func (h *Handler) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	status := c.Query("status")
	authorID := c.Query("author_id")
	q := c.Query("q")

	pages, err := h.svc.ListFiltered(c.Request.Context(), limit, offset, status, authorID, q)
	if err != nil {
		if err == pageApp.ErrInvalidStatus {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]gin.H, 0, len(pages))
	for i := range pages {
		p := pages[i]
		out = append(out, pageToJSON(&p))
	}
	c.JSON(http.StatusOK, gin.H{"pages": out})
}

func (h *Handler) publicGetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	p, err := h.svc.PublicGetBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == pageApp.ErrPageNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, pageToJSON(p))
}

type updatePageRequest struct {
	Title   string `json:"title"`
	Slug    string `json:"slug"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

func (h *Handler) update(c *gin.Context) {
	id := c.Param("id")

	var req updatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	p, err := h.svc.Update(c.Request.Context(), id, req.Title, req.Slug, req.Content, req.Status)
	if err != nil {
		switch err {
		case pageApp.ErrPageNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case pageApp.ErrSlugTaken, pageApp.ErrInvalidStatus:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusOK, pageToJSON(p))
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func pageToJSON(p *pageDomain.Page) gin.H {
	var publishedAt any
	if p.PublishedAt != nil {
		publishedAt = p.PublishedAt
	}

	return gin.H{
		"id":          p.ID,
		"authorId":    p.AuthorID,
		"title":       p.Title,
		"slug":        p.Slug,
		"content":     p.Content,
		"status":      p.Status,
		"publishedAt": publishedAt,
		"createdAt":   p.CreatedAt,
		"updatedAt":   p.UpdatedAt,
	}
}

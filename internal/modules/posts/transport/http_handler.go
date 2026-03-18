package transport

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	postApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/application"
	postDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain"
	platformMiddleware "github.com/Petar-V-Nikolov/nextpress-backend/internal/platform/middleware"
)

type Handler struct {
	svc *postApp.Service
}

func NewHandler(svc *postApp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc, requirePerm func(string) gin.HandlerFunc) {
	posts := rg.Group("/posts")

	posts.GET("",
		auth,
		requirePerm("posts:read"),
		h.list,
	)
	posts.GET("/:id",
		auth,
		requirePerm("posts:read"),
		h.getByID,
	)
	posts.POST("",
		auth,
		requirePerm("posts:write"),
		h.create,
	)
	posts.PUT("/:id",
		auth,
		requirePerm("posts:write"),
		h.update,
	)
	posts.DELETE("/:id",
		auth,
		requirePerm("posts:write"),
		h.delete,
	)

	posts.PUT("/:id/categories",
		auth,
		requirePerm("posts:write"),
		h.setCategories,
	)
	posts.PUT("/:id/tags",
		auth,
		requirePerm("posts:write"),
		h.setTags,
	)
}

func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	posts := rg.Group("/posts")
	posts.GET("", h.publicList)
	posts.GET("/:slug", h.publicGetBySlug)
}

type createPostRequest struct {
	Title   string `json:"title" binding:"required"`
	Slug    string `json:"slug" binding:"required"`
	Content string `json:"content"`
}

func (h *Handler) create(c *gin.Context) {
	var req createPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	v, _ := c.Get(platformMiddleware.ContextUserIDKey)
	authorID, _ := v.(string)

	p, err := h.svc.Create(c.Request.Context(), authorID, req.Title, req.Slug, req.Content)
	if err != nil {
		switch err {
		case postApp.ErrInvalidPost:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case postApp.ErrSlugTaken:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusCreated, postToJSON(p))
}

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")
	p, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == postApp.ErrPostNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, postToJSON(p))
}

func (h *Handler) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	status := c.Query("status")
	authorID := c.Query("author_id")
	q := c.Query("q")

	posts, err := h.svc.ListFiltered(c.Request.Context(), limit, offset, status, authorID, q)
	if err != nil {
		if err == postApp.ErrInvalidStatus {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]gin.H, 0, len(posts))
	for i := range posts {
		p := posts[i]
		out = append(out, postToJSON(&p))
	}
	c.JSON(http.StatusOK, gin.H{"posts": out})
}

func (h *Handler) publicList(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	q := c.Query("q")
	categoryID := c.Query("category_id")
	tagID := c.Query("tag_id")

	posts, err := h.svc.PublicList(c.Request.Context(), limit, offset, q, categoryID, tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]gin.H, 0, len(posts))
	for i := range posts {
		p := posts[i]
		out = append(out, postToJSON(&p))
	}
	c.JSON(http.StatusOK, gin.H{"posts": out})
}

func (h *Handler) publicGetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	p, err := h.svc.PublicGetBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == postApp.ErrPostNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, postToJSON(p))
}

type updatePostRequest struct {
	Title   string `json:"title"`
	Slug    string `json:"slug"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

func (h *Handler) update(c *gin.Context) {
	id := c.Param("id")

	var req updatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	p, err := h.svc.Update(c.Request.Context(), id, req.Title, req.Slug, req.Content, req.Status)
	if err != nil {
		switch err {
		case postApp.ErrPostNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case postApp.ErrSlugTaken, postApp.ErrInvalidStatus:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusOK, postToJSON(p))
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type setIDsRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

func (h *Handler) setCategories(c *gin.Context) {
	id := c.Param("id")

	var req setIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	if err := h.svc.SetCategories(c.Request.Context(), id, req.IDs); err != nil {
		if err == postApp.ErrPostNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) setTags(c *gin.Context) {
	id := c.Param("id")

	var req setIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	if err := h.svc.SetTags(c.Request.Context(), id, req.IDs); err != nil {
		if err == postApp.ErrPostNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func postToJSON(p *postDomain.Post) gin.H {
	var publishedAt any
	if p.PublishedAt != nil {
		publishedAt = p.PublishedAt
	}

	return gin.H{
		"id":           p.ID,
		"author_id":    p.AuthorID,
		"title":        p.Title,
		"slug":         p.Slug,
		"content":      p.Content,
		"status":       p.Status,
		"published_at": publishedAt,
		"created_at":   p.CreatedAt,
		"updated_at":   p.UpdatedAt,
	}
}


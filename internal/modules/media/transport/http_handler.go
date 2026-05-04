package transport

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"

	mediaApp "github.com/nextpresskit/backend/internal/modules/media/application"
	mediaDomain "github.com/nextpresskit/backend/internal/modules/media/domain"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
)

type Handler struct {
	svc *mediaApp.Service
}

func NewHandler(svc *mediaApp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc, requirePerm func(string) gin.HandlerFunc) {
	media := rg.Group("/media")

	media.GET("", auth, requirePerm("media:read"), h.list)
	media.GET("/:id", auth, requirePerm("media:read"), h.getByID)
	media.POST("", auth, requirePerm("media:write"), h.upload)
}

func (h *Handler) upload(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_file"})
		return
	}

	file, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_file"})
		return
	}
	defer file.Close()

	v, _ := c.Get(platformMiddleware.ContextUserIDKey)
	uploaderID, _ := v.(string)

	originalName := filepath.Base(fh.Filename)
	mimeType := fh.Header.Get("Content-Type")

	m, err := h.svc.Upload(c.Request.Context(), uploaderID, originalName, mimeType, file)
	if err != nil {
		if err == mediaApp.ErrInvalidUpload {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload_failed"})
		return
	}

	c.JSON(http.StatusCreated, mediaToJSON(m))
}

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")
	m, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == mediaApp.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, mediaToJSON(m))
}

func (h *Handler) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	items, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]gin.H, 0, len(items))
	for i := range items {
		m := items[i]
		out = append(out, mediaToJSON(&m))
	}
	c.JSON(http.StatusOK, gin.H{"media": out})
}

func mediaToJSON(m *mediaDomain.Media) gin.H {
	return gin.H{
		"id":           m.ID,
		"uploaderId":   m.UploaderID,
		"originalName": m.OriginalName,
		"mimeType":     m.MimeType,
		"sizeBytes":    m.SizeBytes,
		"publicUrl":    m.PublicURL,
		"createdAt":    m.CreatedAt,
	}
}

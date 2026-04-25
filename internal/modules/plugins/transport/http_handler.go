package transport

import (
	"net/http"

	"github.com/gin-gonic/gin"

	pluginsApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/plugins/application"
	pluginsDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/plugins/domain"
)

type Handler struct {
	svc *pluginsApp.Service
}

func NewHandler(svc *pluginsApp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc, requirePerm func(string) gin.HandlerFunc) {
	plugins := rg.Group("/plugins")

	plugins.GET("", auth, requirePerm("plugins:manage"), h.list)
	plugins.POST("", auth, requirePerm("plugins:manage"), h.register)
	plugins.PUT("/:id", auth, requirePerm("plugins:manage"), h.update)
}

func (h *Handler) list(c *gin.Context) {
	out, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	resp := make([]gin.H, 0, len(out))
	for i := range out {
		p := out[i]
		resp = append(resp, pluginToJSON(&p))
	}

	c.JSON(http.StatusOK, gin.H{"plugins": resp})
}

type registerPluginRequest struct {
	Name    string         `json:"name" binding:"required"`
	Slug    string         `json:"slug" binding:"required"`
	Enabled bool           `json:"enabled"`
	Version string         `json:"version"`
	Config  map[string]any `json:"config"`
}

func (h *Handler) register(c *gin.Context) {
	var req registerPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	p, err := h.svc.Register(
		c.Request.Context(),
		req.Name,
		req.Slug,
		req.Enabled,
		req.Version,
		req.Config,
	)
	if err != nil {
		switch err {
		case pluginsApp.ErrInvalidPluginInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case pluginsApp.ErrPluginAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusCreated, pluginToJSON(p))
}

type updatePluginRequest struct {
	Enabled *bool           `json:"enabled"`
	Version *string         `json:"version"`
	Config  *map[string]any `json:"config"`
}

func (h *Handler) update(c *gin.Context) {
	id := c.Param("id")

	var req updatePluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	p, err := h.svc.Update(c.Request.Context(), id, req.Enabled, req.Version, req.Config)
	if err != nil {
		switch err {
		case pluginsApp.ErrPluginNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case pluginsApp.ErrPluginAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}

	c.JSON(http.StatusOK, pluginToJSON(p))
}

func pluginToJSON(p *pluginsDomain.Plugin) gin.H {
	return gin.H{
		"id":        p.ID,
		"name":      p.Name,
		"slug":      p.Slug,
		"enabled":   p.Enabled,
		"version":   p.Version,
		"config":    p.Config,
		"createdAt": p.CreatedAt,
		"updatedAt": p.UpdatedAt,
	}
}

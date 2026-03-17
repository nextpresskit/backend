package transport

import "github.com/gin-gonic/gin"

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	// /v1/auth/register, /v1/auth/login, /v1/auth/refresh will be added later
}

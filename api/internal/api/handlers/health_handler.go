package handlers

import (
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// RegisterRoutes registers all routes for health checks
func (h *HealthHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/health", h.Check)
}

// @Summary Health check
// @Description Check if the API is running
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	Ok(c, HealthResponse{
		Status: "healthy",
	})
}

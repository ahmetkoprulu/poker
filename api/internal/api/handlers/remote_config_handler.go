package handlers

import (
	"github.com/ahmetkoprulu/rtrp/internal/services"
	"github.com/ahmetkoprulu/rtrp/models"

	"github.com/gin-gonic/gin"
)

type RemoteConfigHandler struct {
	service *services.RemoteConfigService
}

func NewRemoteConfigHandler(service *services.RemoteConfigService) *RemoteConfigHandler {
	return &RemoteConfigHandler{service: service}
}

func (h *RemoteConfigHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/remote-configs/:version", h.GetRemoteConfig)
	router.POST("/remote-configs", h.CreateRemoteConfig)
}

func (h *RemoteConfigHandler) GetRemoteConfig(c *gin.Context) {
	version := c.Param("version")
	configs, err := h.service.GetAllRemoteConfigsByVersion(c.Request.Context(), version)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Ok(c, configs)
}

func (h *RemoteConfigHandler) CreateRemoteConfig(c *gin.Context) {
	var config models.RemoteConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		BadRequest(c, err.Error())
		return
	}

	err := h.service.SaveRemoteConfig(c.Request.Context(), &config)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Ok(c, config)
}

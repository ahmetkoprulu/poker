package handlers

import (
	"github.com/ahmetkoprulu/rtrp/internal/services"
	"github.com/gin-gonic/gin"
)

type LobbyHandler struct {
	lobbyService *services.LobbyService
}

func NewLobbyHandler(lobbyService *services.LobbyService) *LobbyHandler {
	return &LobbyHandler{lobbyService: lobbyService}
}

func (h *LobbyHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	lobby := router.Group("/lobby")
	{
		lobby.GET("/state", authMiddleware, h.GetLobbyState)
	}
}

func (h *LobbyHandler) GetLobbyState(c *gin.Context) {
	playerID := c.GetString("playerID")
	if playerID == "" {
		BadRequest(c, "player_id is required")
		return
	}

	lobbyState, err := h.lobbyService.GetState(c.Request.Context(), playerID)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Ok(c, lobbyState)
}

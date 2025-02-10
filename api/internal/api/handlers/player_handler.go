package handlers

import (
	"github.com/ahmetkoprulu/rtrp/internal/services"
	"github.com/gin-gonic/gin"
)

type PlayerHandler struct {
	playerService *services.PlayerService
}

func NewPlayerHandler(playerService *services.PlayerService) *PlayerHandler {
	return &PlayerHandler{
		playerService: playerService,
	}
}

func (h *PlayerHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc, serverToServerAuthMiddleware gin.HandlerFunc) {
	player := router.Group("/players")
	{
		player.GET("/me", authMiddleware, h.GetMyPlayer)
		player.PUT("/chips", serverToServerAuthMiddleware, h.IncrementChips)
	}
}

func (h *PlayerHandler) GetMyPlayer(c *gin.Context) {
	userId, playerId := c.GetString("userID"), c.GetString("playerID")
	if userId == "" || playerId == "" {
		BadRequest(c, "userID and playerID are required")
		return
	}

	if playerId = c.GetString("playerID"); playerId == "" {
		BadRequest(c, "playerID is required")
		return
	}

	player, err := h.playerService.GetPlayerByID(c.Request.Context(), playerId)
	if err != nil {
		BadRequest(c, err.Error())
		return
	} else if player == nil {
		NotFound(c, "player not found")
		return
	} else if player.UserID != userId {
		Unauthorized(c, "player not found")
		return
	}

	Ok(c, player)
}

func (h *PlayerHandler) IncrementChips(c *gin.Context) {
	model := BindModel[IncrementChipsRequest](c)
	if model == nil {
		return
	}

	chips, err := h.playerService.IncrementChips(c.Request.Context(), model.ID, model.Amount)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, chips)
}

type IncrementChipsRequest struct {
	ID     string `json:"id"`
	Amount int    `json:"amount"`
}

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

// @Summary Get current player
// @Description Mevcut Player'in bilgilerini almak icin kullanilir.
// @Tags players
// @Produce json
// @Security Bearer
// @Success 200 {object} models.Player
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Player not found"
// @Router /players/me [get]
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

// @Summary Increment player chips
// @Description Increment a player's chips balance (server-to-server only)
// @Tags players
// @Accept json
// @Produce json
// @Param request body IncrementChipsRequest true "Increment chips request"
// @Success 200 {integer} int64 "Updated chips balance"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /players/chips [put]
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

// IncrementChipsRequest represents a request to increment a player's chips
type IncrementChipsRequest struct {
	// Player ID
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Amount of chips to increment (can be negative for decrement)
	Amount int `json:"amount" example:"1000"`
}

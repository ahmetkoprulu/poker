package handlers

import (
	"math/rand"

	"github.com/ahmetkoprulu/rtrp/internal/services"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/gin-gonic/gin"
)

type PlayerHandler struct {
	playerService  *services.PlayerService
	productService *services.ProductService
}

func NewPlayerHandler(playerService *services.PlayerService, productService *services.ProductService) *PlayerHandler {
	return &PlayerHandler{
		playerService:  playerService,
		productService: productService,
	}
}

func (h *PlayerHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc, serverToServerAuthMiddleware gin.HandlerFunc) {
	player := router.Group("/players")
	{
		player.GET("/me", authMiddleware, h.GetMyPlayer)
		player.PUT("/chips", serverToServerAuthMiddleware, h.IncrementChips)
		player.POST("/free-spin-wheel", authMiddleware, h.FreeSpinWheel)
		player.POST("/spin-wheel", authMiddleware, h.SpinWheel)
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

type SpinWheelResponse struct {
	Index  int         `json:"index"`
	Reward models.Item `json:"reward"`
}

func (h *PlayerHandler) FreeSpinWheel(c *gin.Context) {
	userId, playerId := c.GetString("userID"), c.GetString("playerID")
	if userId == "" || playerId == "" {
		BadRequest(c, "userID and playerID are required")
		return
	}

	rewards := []models.Item{
		{
			Type:   models.ItemTypeChips,
			Amount: 1000,
		},
		{
			Type:   models.ItemTypeChips,
			Amount: 1100,
		},
		{
			Type:   models.ItemTypeChips,
			Amount: 1200,
		},
		{
			Type:   models.ItemTypeChips,
			Amount: 1300,
		},
		{
			Type:   models.ItemTypeChips,
			Amount: 1400,
		},
		{
			Type:   models.ItemTypeChips,
			Amount: 1500,
		},
		{
			Type:   models.ItemTypeChips,
			Amount: 1600,
		},
		{
			Type:   models.ItemTypeGold,
			Amount: 1700,
		},
	}

	index := rand.Intn(len(rewards))
	reward := rewards[index]

	err := h.productService.GiveRewardToPlayer(c.Request.Context(), []models.Item{reward}, playerId)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, SpinWheelResponse{
		Index:  index,
		Reward: reward,
	})
}

func (h *PlayerHandler) SpinWheel(c *gin.Context) {
	userId, playerId := c.GetString("userID"), c.GetString("playerID")
	if userId == "" || playerId == "" {
		BadRequest(c, "userID and playerID are required")
		return
	}

	rewards := []models.Item{
		{
			Type:   models.ItemTypeChips,
			Amount: 1000 * 10,
		},
		{
			Type:   models.ItemTypeChips,
			Amount: 1000 * 11,
		},
		{
			Type:   models.ItemTypeChips,
			Amount: 1000 * 12,
		},
		{
			Type:   models.ItemTypeChips,
			Amount: 1000 * 13,
		},
		{
			Type:   models.ItemTypeChips,
			Amount: 1000 * 14,
		},
		{
			Type:   models.ItemTypeChips,
			Amount: 1000 * 15,
		},
		{
			Type:   models.ItemTypeChips,
			Amount: 1000 * 16,
		},
		{
			Type:   models.ItemTypeGold,
			Amount: 1000 * 17,
		},
	}

	index := rand.Intn(len(rewards))
	reward := rewards[index]

	err := h.productService.GiveRewardToPlayer(c.Request.Context(), []models.Item{reward}, playerId)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, SpinWheelResponse{
		Index:  index,
		Reward: reward,
	})
}

// IncrementChipsRequest represents a request to increment a player's chips
type IncrementChipsRequest struct {
	// Player ID
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Amount of chips to increment (can be negative for decrement)
	Amount int `json:"amount" example:"1000"`
}

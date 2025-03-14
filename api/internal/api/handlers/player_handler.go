package handlers

import (
	"github.com/ahmetkoprulu/rtrp/internal/services"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/gin-gonic/gin"
)

type PlayerHandler struct {
	playerService   *services.PlayerService
	productService  *services.ProductService
	miniGameService *services.MiniGameService
}

func NewPlayerHandler(playerService *services.PlayerService, productService *services.ProductService, miniGameService *services.MiniGameService) *PlayerHandler {
	return &PlayerHandler{
		playerService:   playerService,
		productService:  productService,
		miniGameService: miniGameService,
	}
}

func (h *PlayerHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc, serverToServerAuthMiddleware gin.HandlerFunc) {
	player := router.Group("/players")
	{
		player.GET("/me", authMiddleware, h.GetMyPlayer)
		player.PUT("/chips", serverToServerAuthMiddleware, h.IncrementChips)
		player.POST("/free-spin-wheel", authMiddleware, h.FreeSpinWheel)
		player.POST("/spin-wheel", authMiddleware, h.SpinWheel)
		player.POST("/spin-slot", authMiddleware, h.SpinSlot)
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

	index, reward, err := h.miniGameService.SpinWheel(c.Request.Context(), playerId)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, SpinWheelResponse{
		Index:  index,
		Reward: *reward,
	})
}

func (h *PlayerHandler) SpinWheel(c *gin.Context) {
	userId, playerId := c.GetString("userID"), c.GetString("playerID")
	if userId == "" || playerId == "" {
		BadRequest(c, "userID and playerID are required")
		return
	}

	index, reward, err := h.miniGameService.SpinGoldWheel(c.Request.Context(), playerId)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, SpinWheelResponse{
		Index:  index,
		Reward: *reward,
	})
}

type SpinSlotResponse struct {
	Index  []models.ItemType `json:"index"`
	Reward models.Item       `json:"reward"`
}

func (h *PlayerHandler) SpinSlot(c *gin.Context) {
	userId, playerId := c.GetString("userID"), c.GetString("playerID")
	if userId == "" || playerId == "" {
		BadRequest(c, "userID and playerID are required")
		return
	}

	reward, indices, err := h.miniGameService.SpinSlot(c.Request.Context(), playerId)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, SpinSlotResponse{
		Index:  indices,
		Reward: *reward,
	})
}

// IncrementChipsRequest represents a request to increment a player's chips
type IncrementChipsRequest struct {
	// Player ID
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	// Amount of chips to increment (can be negative for decrement)
	Amount int `json:"amount" example:"1000"`
}

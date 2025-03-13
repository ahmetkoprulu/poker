package handlers

import (
	"time"

	"github.com/ahmetkoprulu/rtrp/internal/services"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Message string `json:"message"`
}

type BattlePassHandler struct {
	battlePassService *services.BattlePassService
}

func NewBattlePassHandler(battlePassService *services.BattlePassService) *BattlePassHandler {
	return &BattlePassHandler{
		battlePassService: battlePassService,
	}
}

// RegisterRoutes registers all battle pass routes
func (h *BattlePassHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	battlePass := router.Group("/battle-pass")
	{
		// Player endpoints (require auth)
		battlePass.GET("/active", h.GetActiveBattlePass)
		battlePass.GET("/progress", authMiddleware, h.GetPlayerProgress)
		battlePass.GET("/details", authMiddleware, h.GetPlayerBattlePassDetails)
		battlePass.POST("/claim-reward", authMiddleware, h.ClaimReward)
		battlePass.POST("/upgrade", authMiddleware, h.UpgradeToPremium)

		// Test endpoints (require auth)
		test := battlePass.Group("/test")
		{
			test.POST("/add-xp", authMiddleware, h.AddXP)
		}

		// Admin endpoints (should be protected)
		admin := battlePass.Group("/admin")
		{
			admin.POST("/create", h.CreateBattlePass)
			admin.POST("/create-from-template", h.CreateBattlePassFromTemplate)
		}
	}
}

// GetActiveBattlePass godoc
// @Summary Get active battle pass
// @Description Get the currently active battle pass
// @Tags Battle Pass
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.BattlePassResponse
// @Failure 404 {object} models.StringResponse
// @Router /battle-pass/active [get]
func (h *BattlePassHandler) GetActiveBattlePass(c *gin.Context) {
	battlePass, err := h.battlePassService.GetActiveBattlePass(c)
	if err != nil {
		NotFound(c, "No active battle pass found")
		return
	}

	Ok(c, battlePass)
}

// GetPlayerProgress godoc
// @Summary Get player's battle pass progress
// @Description Returns the player's current progress in the active battle pass, including current level, XP, and premium status
// @Tags Battle Pass
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.ApiResponse[models.PlayerBattlePass] "Player's battle pass progress"
// @Failure 400 {object} models.ApiResponse[string] "Invalid request or missing player ID"
// @Failure 404 {object} models.ApiResponse[string] "No active battle pass found"
// @Router /battle-pass/progress [get]
func (h *BattlePassHandler) GetPlayerProgress(c *gin.Context) {
	playerID := c.GetString("playerID")
	if playerID == "" {
		BadRequest(c, "Player ID is required")
		return
	}

	// Get active battle pass
	battlePass, err := h.battlePassService.GetActiveBattlePass(c)
	if err != nil {
		NotFound(c, "No active battle pass found")
		return
	}

	// Get or create player progress
	progress, err := h.battlePassService.GetOrCreatePlayerBattlePass(c, playerID, battlePass.ID)
	if err != nil {
		InternalServerError(c, "Failed to get player progress")
		return
	}

	Ok(c, progress)
}

type ClaimRewardRequest struct {
	Level     int  `json:"level" binding:"required,min=1"`
	IsPremium bool `json:"is_premium"`
}

// ClaimReward godoc
// @Summary Claim battle pass level reward
// @Description Claim a reward for a completed battle pass level
// @Tags Battle Pass
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body ClaimRewardRequest true "Claim reward request"
// @Success 200 {object} models.StringResponse
// @Failure 400 {object} models.StringResponse
// @Failure 404 {object} models.StringResponse
// @Router /battle-pass/claim-reward [post]
func (h *BattlePassHandler) ClaimReward(c *gin.Context) {
	playerID := c.GetString("playerID")
	if playerID == "" {
		BadRequest(c, "Player ID is required")
		return
	}

	model := BindModel[ClaimRewardRequest](c)
	if model == nil {
		return
	}

	// Get active battle pass
	battlePass, err := h.battlePassService.GetActiveBattlePass(c)
	if err != nil {
		NotFound(c, "No active battle pass found")
		return
	}

	// Get player battle pass
	playerBattlePass, err := h.battlePassService.GetOrCreatePlayerBattlePass(c, playerID, battlePass.ID)
	if err != nil {
		InternalServerError(c, "Failed to get player battle pass")
		return
	}

	// Claim reward
	reward, err := h.battlePassService.ClaimReward(c, playerBattlePass.ID, model.Level, model.IsPremium)
	if err != nil {
		switch err {
		case services.ErrInsufficientLevel:
			BadRequest(c, "Insufficient level")
		case services.ErrRewardAlreadyClaimed:
			BadRequest(c, "Reward already claimed")
		case services.ErrPremiumRequired:
			BadRequest(c, "Premium battle pass required")
		default:
			InternalServerError(c, "Failed to claim reward")
		}
		return
	}

	Ok(c, reward)
}

type UpgradeToPremiumRequest struct {
	PlayerID string `json:"player_id" binding:"required"`
}

// UpgradeToPremium godoc
// @Summary Upgrade to premium battle pass
// @Description Upgrades a player's battle pass to premium status, unlocking premium rewards
// @Tags Battle Pass
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.ApiResponse[string] "Upgraded to premium successfully"
// @Failure 400 {object} models.ApiResponse[string] "Invalid request or missing player ID"
// @Failure 404 {object} models.ApiResponse[string] "No active battle pass found"
// @Router /battle-pass/upgrade [post]
func (h *BattlePassHandler) UpgradeToPremium(c *gin.Context) {
	playerID := c.GetString("playerID")
	if playerID == "" {
		BadRequest(c, "Player ID is required")
		return
	}

	// Get active battle pass
	battlePass, err := h.battlePassService.GetActiveBattlePass(c)
	if err != nil {
		NotFound(c, "No active battle pass found")
		return
	}

	// Get player battle pass
	playerBattlePass, err := h.battlePassService.GetOrCreatePlayerBattlePass(c, playerID, battlePass.ID)
	if err != nil {
		InternalServerError(c, "Failed to get player battle pass")
		return
	}

	// Upgrade to premium
	err = h.battlePassService.UpgradeToPremium(c, playerBattlePass.ID)
	if err != nil {
		InternalServerError(c, "Failed to upgrade to premium")
		return
	}

	Ok(c, nil)
}

type CreateBattlePassRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description" binding:"required"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
	MaxLevel    int       `json:"max_level" binding:"required,min=1"`
}

// CreateBattlePass godoc
// @Summary Create a new battle pass (admin only)
// @Description Create a new battle pass with specified parameters
// @Tags Battle Pass Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateBattlePassRequest true "Create battle pass request"
// @Success 200 {object} models.BattlePassResponse
// @Failure 400 {object} models.StringResponse
// @Router /battle-pass/admin/create [post]
func (h *BattlePassHandler) CreateBattlePass(c *gin.Context) {
	model := BindModel[CreateBattlePassRequest](c)
	if model == nil {
		return
	}

	battlePass := &models.BattlePass{
		Name:        model.Name,
		Description: model.Description,
		StartTime:   model.StartTime,
		EndTime:     model.EndTime,
		Status:      models.BattlePassStatusUpcoming,
		MaxLevel:    model.MaxLevel,
	}

	err := h.battlePassService.CreateBattlePass(c, battlePass)
	if err != nil {
		InternalServerError(c, "Failed to create battle pass")
		return
	}

	Ok(c, battlePass)
}

// CreateBattlePassFromTemplate godoc
// @Summary Create a battle pass from template (admin only)
// @Description Create a new battle pass using a predefined template
// @Tags Battle Pass Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.BattlePassResponse
// @Failure 400 {object} models.StringResponse
// @Router /battle-pass/admin/create-from-template [post]
func (h *BattlePassHandler) CreateBattlePassFromTemplate(c *gin.Context) {
	model := BindModel[CreateBattlePassRequest](c)
	if model == nil {
		return
	}

	err := h.battlePassService.CreateBattlePassFromTemplate(c, model.Name, model.Description, model.StartTime, model.EndTime, model.MaxLevel)
	if err != nil {
		InternalServerError(c, "Failed to create battle pass from template")
		return
	}

	Ok(c, nil)
}

// GetPlayerBattlePassDetails godoc
// @Summary Get player's battle pass details
// @Description Get detailed information about a player's battle pass progress
// @Tags Battle Pass
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.BattlePassProgressDetailsResponse
// @Failure 400 {object} models.StringResponse
// @Failure 404 {object} models.StringResponse
// @Router /battle-pass/details [get]
func (h *BattlePassHandler) GetPlayerBattlePassDetails(c *gin.Context) {
	playerID := c.GetString("playerID")
	if playerID == "" {
		BadRequest(c, "Player ID is required")
		return
	}

	// Get active battle pass
	battlePass, err := h.battlePassService.GetActiveBattlePass(c)
	if err != nil {
		NotFound(c, "No active battle pass found")
		return
	}

	// Get or create player progress
	// playerBattlePass, err := h.battlePassService.GetOrCreatePlayerBattlePass(c, playerID, battlePass.ID)
	// if err != nil {
	// 	InternalServerError(c, "Failed to get player battle pass")
	// 	return
	// }

	// Get detailed progress
	details, err := h.battlePassService.GetPlayerBattlePassDetails(c, playerID, battlePass.ID)
	if err != nil {
		InternalServerError(c, "Failed to get battle pass details")
		return
	}

	Ok(c, details)
}

type AddXPRequest struct {
	Amount   int                    `json:"amount" binding:"required,min=1"`
	Source   string                 `json:"source" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AddXP godoc
// @Summary Add XP to player's battle pass (test endpoint)
// @Description Add XP to a player's battle pass for testing purposes
// @Tags Battle Pass
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body AddXPRequest true "Add XP request"
// @Success 200 {object} models.StringResponse
// @Failure 400 {object} models.StringResponse
// @Failure 404 {object} models.StringResponse
// @Router /battle-pass/add-xp [post]
func (h *BattlePassHandler) AddXP(c *gin.Context) {
	playerID := c.GetString("playerID")
	if playerID == "" {
		BadRequest(c, "Player ID is required")
		return
	}

	model := BindModel[AddXPRequest](c)
	if model == nil {
		return
	}

	// Get active battle pass
	battlePass, err := h.battlePassService.GetActiveBattlePass(c)
	if err != nil {
		NotFound(c, "No active battle pass found")
		return
	}

	// Get player battle pass
	playerBattlePass, err := h.battlePassService.GetOrCreatePlayerBattlePass(c, playerID, battlePass.ID)
	if err != nil {
		InternalServerError(c, "Failed to get player battle pass")
		return
	}

	// Add XP
	err = h.battlePassService.AddXP(c, playerBattlePass.ID, model.Amount, model.Source, model.Metadata)
	if err != nil {
		InternalServerError(c, "Failed to add XP")
		return
	}

	Ok(c, nil)
}

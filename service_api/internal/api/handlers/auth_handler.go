package handlers

import (
	"context"

	"github.com/ahmetkoprulu/rtrp/internal/services"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRoutes registers all routes for authentication
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.GET("/user/", authMiddleware, h.GetUser)
	}
}

// @Summary Register a new user
// @Description Email ve Parola ile kayit olmak icin kullanilir. Provider 1 Email anlamina gelir. Identifier alanina email adresi girilir. Secret alanina parola girilir.
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User registration details"
// @Success 200 {object} models.ApiResponse[string] "User created successfully"
// @Failure 400 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	req := BindModel[models.RegisterRequest](c)
	if req == nil {
		return
	}

	err := h.authService.Register(context.Background(), req)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, nil)
}

// @Summary Login user
// @Description Kullanici girisi yapmak icin kullanilir. Provider: 0: Guest, 1 Email, 2 Google
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "User login credentials"
// @Success 200 {object} models.ApiResponse[models.UserPlayer] "Login successful"
// @Failure 400 {object} ErrorResponse "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	req := BindModel[models.LoginRequest](c)
	if req == nil {
		return
	}

	client, err := h.authService.Login(context.Background(), req)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, client)
}

// @Summary Get user
// @Description Kullanici bilgilerini almak icin kullanilir.
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse[models.UserPlayer] "User details"
// @Failure 400 {object} ErrorResponse "User not found"
// @Router /auth/user/{user_id} [get]
func (h *AuthHandler) GetUser(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		BadRequest(c, "User ID is required")
		return
	}

	user, err := h.authService.GetUser(context.Background(), userID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, user)
}

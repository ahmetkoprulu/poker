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
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}
}

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

	Ok(c, "User created successfully")
}

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

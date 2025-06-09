package auth

import (
	"context"

	"github.com/ahmetkoprulu/rtrp/models"
)

// AuthProvider defines the interface for different authentication methods
type AuthProvider interface {
	Authenticate(ctx context.Context, request *models.LoginRequest) (*models.User, error)
	ValidateRequest(request *models.LoginRequest) error
}

type AuthResponse struct {
	Identifier string `json:"identifier"`
	Secret     string `json:"secret"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	ProfilePic string `json:"profile_pic"`
}

// UserStore interface for database operations
type UserStore interface {
	GetUserByIdentifier(ctx context.Context, provider models.SocialNetwork, identifier string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	CreatePlayer(ctx context.Context, player *models.Player) (*models.Player, error)
}

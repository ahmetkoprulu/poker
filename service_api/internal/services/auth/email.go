package auth

import (
	"context"
	"fmt"

	"github.com/ahmetkoprulu/rtrp/models"
	"golang.org/x/crypto/bcrypt"
)

// EmailAuthProvider handles email/password authentication
type EmailAuthProvider struct {
	userStore UserStore
}

func NewEmailAuthProvider(userStore UserStore) *EmailAuthProvider {
	return &EmailAuthProvider{userStore: userStore}
}

func (p *EmailAuthProvider) ValidateRequest(request *models.LoginRequest) error {
	if request.Secret == "" || request.Identifier == "" {
		return fmt.Errorf("credentials are required for email authentication")
	}
	return nil
}

func (p *EmailAuthProvider) Authenticate(ctx context.Context, request *models.LoginRequest) (*models.User, error) {
	user, err := p.userStore.GetUserByIdentifier(ctx, models.Email, request.Identifier)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Secret))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	if user.Player == nil {
		player := models.NewGuestPlayer(user.ID)
		player, err = p.userStore.CreatePlayer(ctx, player)
		if err != nil {
			return nil, fmt.Errorf("failed to create player: %w", err)
		}

		user.Player = player
	}

	return user, nil
}

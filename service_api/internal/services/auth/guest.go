package auth

import (
	"context"
	"fmt"

	"github.com/ahmetkoprulu/rtrp/models"
)

// GuestAuthProvider handles guest authentication
type GuestAuthProvider struct {
	userStore UserStore
}

func NewGuestAuthProvider(userStore UserStore) *GuestAuthProvider {
	return &GuestAuthProvider{userStore: userStore}
}

func (p *GuestAuthProvider) ValidateRequest(request *models.LoginRequest) error {
	if request.Identifier == "" {
		return fmt.Errorf("credentials are required for guest authentication")
	}
	return nil
}

func (p *GuestAuthProvider) Authenticate(ctx context.Context, request *models.LoginRequest) (*models.User, error) {
	user, err := p.userStore.GetUserByIdentifier(ctx, models.Guest, request.Identifier)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user = models.NewGuestUser(request.Identifier)
		err = p.userStore.CreateUser(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
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

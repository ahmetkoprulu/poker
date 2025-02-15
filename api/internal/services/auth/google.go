package auth

import (
	"context"
	"fmt"

	"github.com/ahmetkoprulu/rtrp/internal/config"
	"github.com/ahmetkoprulu/rtrp/models"
	"google.golang.org/api/idtoken"
)

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// GoogleAuthProvider handles Google OAuth authentication
type GoogleAuthProvider struct {
	userStore UserStore
	clientId  string
}

func NewGoogleAuthProvider(userStore UserStore) *GoogleAuthProvider {
	config := config.GetConfig()
	return &GoogleAuthProvider{userStore: userStore, clientId: config.SocialConfig.GoogleClientID}
}

func (p *GoogleAuthProvider) ValidateRequest(request *models.LoginRequest) error {
	if request.Identifier == "" { // Google ID token
		return fmt.Errorf("google ID token is required")
	}
	return nil
}

func (p *GoogleAuthProvider) Authenticate(ctx context.Context, request *models.LoginRequest) (*models.User, error) {
	payload, err := p.verifyGoogleToken(ctx, request.Identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to verify google token: %w", err)
	}

	user, err := p.userStore.GetUserByIdentifier(ctx, models.Google, payload.Sub)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user = models.NewSocialUser(models.Google, payload.Sub, payload.Email, payload.Name)

		err = p.userStore.CreateUser(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	if user.Player == nil {
		player := models.NewPlayer(user.ID, payload.Name, "avatar_0", 1000000)
		if payload.Picture != "" {
			player.ProfilePicURL = payload.Picture
		}

		player, err = p.userStore.CreatePlayer(ctx, player)
		if err != nil {
			return nil, fmt.Errorf("failed to create player: %w", err)
		}

		user.Player = player
	}

	return user, nil
}

func (p *GoogleAuthProvider) verifyGoogleToken(ctx context.Context, idToken string) (*googleUserInfo, error) {
	validator, err := idtoken.NewValidator(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create token validator: %w", err)
	}

	payload, err := validator.Validate(ctx, idToken, p.clientId)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	info := &googleUserInfo{
		Sub:           payload.Claims["sub"].(string),
		Email:         payload.Claims["email"].(string),
		EmailVerified: payload.Claims["email_verified"].(bool),
		Name:          payload.Claims["name"].(string),
	}

	if picture, ok := payload.Claims["picture"].(string); ok {
		info.Picture = picture
	}

	if !info.EmailVerified {
		return nil, fmt.Errorf("email not verified by Google")
	}

	return info, nil
}

package auth

import (
	"context"
	"fmt"

	"github.com/ahmetkoprulu/rtrp/models"
)

// FacebookAuthProvider handles Facebook OAuth authentication
type FacebookAuthProvider struct {
	userStore UserStore
}

func NewFacebookAuthProvider(userStore UserStore) *FacebookAuthProvider {
	return &FacebookAuthProvider{userStore: userStore}
}

func (p *FacebookAuthProvider) ValidateRequest(request *models.LoginRequest) error {
	if request.Identifier == "" { // Facebook access token
		return fmt.Errorf("facebook access token is required")
	}
	return nil
}

func (p *FacebookAuthProvider) Authenticate(ctx context.Context, request *models.LoginRequest) (*models.User, error) {
	// Verify Facebook access token
	// fbUser, err := verifyFacebookToken(request.Identifier)
	// if err != nil {
	//     return nil, err
	// }

	// For now, just check if user exists
	user, err := p.userStore.GetUserByIdentifier(ctx, models.Facebook, request.Identifier)
	if err != nil {
		return nil, err
	}

	if user == nil {
		// Create new user
		user = models.NewSocialUser(models.Facebook, request.Identifier, request.Identifier, "")
		// Save user
		err = p.userStore.CreateUser(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

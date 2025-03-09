package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ahmetkoprulu/rtrp/game/models"
)

type AuthService struct {
	parent   *ApiService
	endpoint string
	client   *ApiClient
}

func NewAuthService(parent *ApiService, endpoint string) *AuthService {
	service := &AuthService{
		parent:   parent,
		endpoint: parent.config.BaseURL + endpoint,
	}
	service.client = parent.getClient("auth-service", endpoint)

	return service
}

func (s *AuthService) GetUser(authToken string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.client.SetHeader("Authorization", "Bearer "+authToken)

	var response ApiResponse[models.User]

	err := s.client.Get(ctx, s.endpoint+"/user", &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !response.Success {
		return nil, errors.New(response.Message)
	}

	return &response.Data, nil
}

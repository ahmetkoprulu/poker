package services

import (
	"context"
	"fmt"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/common/utils"
	"github.com/ahmetkoprulu/rtrp/internal/services/auth"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db        *data.PgDbContext
	userStore auth.UserStore
	providers map[models.SocialNetwork]auth.AuthProvider
}

func NewAuthService(db *data.PgDbContext) *AuthService {
	userStore := &auth.PgUserStore{Db: db}
	service := &AuthService{
		db:        db,
		userStore: userStore,
		providers: make(map[models.SocialNetwork]auth.AuthProvider),
	}

	service.providers[models.Guest] = auth.NewGuestAuthProvider(userStore)
	service.providers[models.Email] = auth.NewEmailAuthProvider(userStore)
	service.providers[models.Google] = auth.NewGoogleAuthProvider(userStore)
	service.providers[models.Facebook] = auth.NewFacebookAuthProvider(userStore)

	return service
}

func (s *AuthService) Login(ctx context.Context, request *models.LoginRequest) (*models.LoginResponse, error) {
	provider, exists := s.providers[request.Provider]
	if !exists {
		return nil, fmt.Errorf("unsupported authentication provider")
	}

	if err := provider.ValidateRequest(request); err != nil {
		return nil, err
	}

	user, err := provider.Authenticate(ctx, request)
	if err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := utils.GenerateJWTTokenWithClaims(utils.Claims{UserID: user.ID, PlayerID: user.Player.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		User:  *user,
		Token: token,
	}, nil
}

func (s *AuthService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *AuthService) Register(ctx context.Context, request *models.RegisterRequest) error {
	loginReq := &models.LoginRequest{
		Provider:   models.Email,
		Identifier: request.Identifier,
		Secret:     request.Secret,
	}

	if err := s.providers[models.Email].ValidateRequest(loginReq); err != nil {
		return err
	}

	user, err := s.userStore.GetUserByIdentifier(ctx, models.Email, request.Identifier)
	if err != nil {
		return err
	}

	if user != nil {
		return fmt.Errorf("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Secret), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user = models.NewEmailUser(request.Identifier, string(hashedPassword))
	user.ID = uuid.New().String()

	err = s.userStore.CreateUser(ctx, user)
	if err != nil {
		return err
	}

	player := models.NewPlayer(user.ID, "", "", 1000000)
	player, err = s.userStore.CreatePlayer(ctx, player)
	if err != nil {
		return err
	}

	user.Player = player

	return nil
}

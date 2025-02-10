package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/common/utils"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"
)

type AuthService struct {
	db *data.PgDbContext
}

func NewAuthService(db *data.PgDbContext) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Login(ctx context.Context, request *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.getUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		return nil, err
	}

	player, err := s.getPlayerByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	token, err := utils.GenerateJWTTokenWithClaims(utils.Claims{UserID: user.ID, PlayerID: player.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		User: models.UserPlayer{
			ID:     user.ID,
			Player: *player,
		},
		Token: token,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, request *models.RegisterRequest) error {
	user, err := s.getUserByEmail(ctx, request.Email)
	if err != nil {
		return err
	}

	if user != nil {
		return fmt.Errorf("user already exists")
	}

	return s.createUser(ctx, request)
}

func (s *AuthService) createUser(ctx context.Context, request *models.RegisterRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		ID:       uuid.New().String(),
		Email:    request.Email,
		Password: string(hashedPassword),
	}

	var query = `
			INSERT INTO users (id, email, password_hash)
			VALUES ($1, $2, $3)
			RETURNING id
		`

	_, err = s.db.Exec(ctx, query, user.ID, user.Email, user.Password)
	if err != nil {
		return err
	}

	_, err = s.createPlayer(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) createPlayer(ctx context.Context, user *models.User) (*models.Player, error) {
	maxRetries := 10
	var id string
	for i := 0; i < maxRetries; i++ {
		id = generateUniqueID()

		var exists bool
		err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM players WHERE id = $1)", id).Scan(&exists) // Check if ID exists
		if err != nil {
			return nil, fmt.Errorf("failed to check player id existence: %w", err)
		} else if exists {
			continue // Try another ID if this one exists
		}

		break
	}

	if id == "" {
		id = uuid.New().String()
	}

	player := &models.Player{
		ID:     id,
		UserID: user.ID,
		Chips:  1000000,
	}

	query := `
		INSERT INTO players (id, user_id, chips)
		VALUES ($1, $2, $3)
	`

	_, err := s.db.Exec(ctx, query, player.ID, player.UserID, player.Chips)
	if err != nil {
		return nil, err
	}

	return player, nil
}

func (s *AuthService) getUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var query = `
		SELECT id, email, password_hash
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := s.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) getPlayerByUserID(ctx context.Context, userID string) (*models.Player, error) {
	var query = `
		SELECT id, user_id, chips
		FROM players
		WHERE user_id = $1
	`

	var player models.Player
	err := s.db.QueryRow(ctx, query, userID).Scan(&player.ID, &player.UserID, &player.Chips)
	if err != nil {
		return nil, err
	}

	return &player, nil
}

func generateUniqueID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(uint64(time.Now().UnixNano()))
	length := 6 + rand.Intn(4)
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// err = s.db.WithTransaction(ctx, func(tx data.QueryRunner) error {
// 	var query = `
// 		INSERT INTO users (id, email, password_hash)
// 		VALUES ($1, $2, $3)
// 		RETURNING id
// 	`

// 	_, err = tx.Exec(ctx, query, user.ID, user.Email, user.Password)
// 	if err != nil {
// 		return err
// 	}

// 	query = `
// 		INSERT INTO players (id, user_id, chips)
// 		VALUES ($1, $2, $3)
// 	`
// 	_, err = tx.Exec(ctx, query, player.ID, player.UserID, player.Chips)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// })

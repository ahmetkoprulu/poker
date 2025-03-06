package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/rand"
)

// UserStore implementation
type PgUserStore struct {
	Db *data.PgDbContext
}

func (s *PgUserStore) GetUserByIdentifier(ctx context.Context, provider models.SocialNetwork, identifier string) (*models.User, error) {
	var query = `
		SELECT u.id, u.provider, u.identifier, u.password_hash, u.profile,
			   p.id, p.username, p.profile_pic_url, p.chips
		FROM users u
		LEFT JOIN players p ON p.user_id = u.id
		WHERE u.provider = $1 AND u.identifier = $2
	`

	result := &models.User{
		Player: &models.Player{},
	}

	var playerID, playerUsername, playerProfilePicURL *string
	var playerChips *int64
	err := s.Db.QueryRow(ctx, query, provider, identifier).Scan(
		&result.ID,
		&result.Provider,
		&result.Identifier,
		&result.Password,
		&result.Profile,
		&playerID,
		&playerUsername,
		&playerProfilePicURL,
		&playerChips,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	// If player ID is empty (LEFT JOIN returned no player), set Player to nil
	if playerID == nil {
		result.Player = nil
	} else {
		result.Player.ID = *playerID
		result.Player.Username = *playerUsername
		result.Player.ProfilePicURL = *playerProfilePicURL
		result.Player.Chips = *playerChips
	}

	return result, nil
}

func (s *PgUserStore) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var query = `
		SELECT u.id, u.provider, u.identifier, u.password_hash, u.profile,
			   p.id, p.username, p.profile_pic_url, p.chips
		FROM users u
		LEFT JOIN players p ON p.user_id = u.id
		WHERE u.id = $1
	`
	result := &models.User{
		Player: &models.Player{},
	}

	var playerID, playerUsername, playerProfilePicURL *string
	var playerChips *int64
	err := s.Db.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.Provider,
		&result.Identifier,
		&result.Password,
		&result.Profile,
		&playerID,
		&playerUsername,
		&playerProfilePicURL,
		&playerChips,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	// If player ID is empty (LEFT JOIN returned no player), set Player to nil
	if playerID == nil {
		result.Player = nil
	} else {
		result.Player.ID = *playerID
		result.Player.Username = *playerUsername
		result.Player.ProfilePicURL = *playerProfilePicURL
		result.Player.Chips = *playerChips
	}

	return result, nil
}

func (s *PgUserStore) CreateUser(ctx context.Context, user *models.User) error {
	var query = `
		INSERT INTO users (id, provider, identifier, password_hash, profile)
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id
	`

	_, err := s.Db.Exec(ctx, query,
		user.ID,
		user.Provider,
		user.Identifier,
		user.Password,
		user.Profile,
	)

	return err
}

func (s *PgUserStore) CreatePlayer(ctx context.Context, player *models.Player) (*models.Player, error) {
	maxRetries := 10
	var id string
	for i := 0; i < maxRetries; i++ {
		id = generateUniqueID()

		var exists bool
		err := s.Db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM players WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("failed to check player id existence: %w", err)
		} else if exists {
			continue // Try another ID if this one exists
		}

		player.ID = id
		break
	}

	if player.ID == "" {
		player.ID = uuid.New().String()
	}

	query := `
		INSERT INTO players (id, user_id, username, profile_pic_url, chips)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.Db.Exec(ctx, query, player.ID, player.UserID, player.Username, player.ProfilePicURL, player.Chips)
	if err != nil {
		return nil, err
	}

	return player, nil
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

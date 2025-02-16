package services

import (
	"context"
	"fmt"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/jackc/pgx/v5"
)

type PlayerService struct {
	db *data.PgDbContext
}

func NewPlayerService(db *data.PgDbContext) *PlayerService {
	return &PlayerService{db: db}
}

func (s *PlayerService) GetPlayerByID(ctx context.Context, id string) (*models.Player, error) {
	if id == "" {
		return nil, fmt.Errorf("player id is required")
	}

	player, err := s.getPlayerByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return player, nil
}

func (s *PlayerService) IncrementChips(ctx context.Context, id string, amount int) (int64, error) {
	if id == "" {
		return 0, fmt.Errorf("player id is required")
	}

	var chips int64
	err := s.db.QueryRow(ctx, "UPDATE players SET chips = chips + $1 WHERE id = $2 RETURNING chips", amount, id).Scan(&chips)
	if err != nil {
		return 0, err
	}

	return chips, nil
}

func (s *PlayerService) getPlayerByID(ctx context.Context, playerID string) (*models.Player, error) {
	var query = `
		SELECT id, user_id, username, profile_pic_url, chips
		FROM players
		WHERE id = $1
	`

	var player models.Player
	err := s.db.QueryRow(ctx, query, playerID).Scan(&player.ID, &player.UserID, &player.Username, &player.ProfilePicURL, &player.Chips)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &player, nil
}

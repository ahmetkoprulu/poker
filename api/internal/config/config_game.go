package config

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/jackc/pgx/v5"
)

var GameConfig *models.GeneralRemoteConfig

func GetGameConfig() (*models.GeneralRemoteConfig, error) {
	if GameConfig == nil {
		return nil, errors.New("game config not loaded")
	}

	return GameConfig, nil
}

func LoadGameConfig(db *data.PgDbContext) error {
	query := `
		SELECT value
		FROM remote_configs
		WHERE name = 'general' AND version = 0
	`

	var value []byte
	err := db.QueryRow(context.Background(), query).Scan(&value)
	if err == pgx.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}

	err = json.Unmarshal(value, &GameConfig)
	if err != nil {
		return err
	}

	return nil
}

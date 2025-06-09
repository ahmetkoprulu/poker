package services

import (
	"context"
	"time"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/models"
)

type RemoteConfigService struct {
	db *data.PgDbContext
}

func NewRemoteConfigService(db *data.PgDbContext) *RemoteConfigService {
	return &RemoteConfigService{db: db}
}

func (s *RemoteConfigService) GetAllRemoteConfigsByVersion(ctx context.Context, version string) ([]models.RemoteConfig, error) {
	query := `
		SELECT id, name, version, value, created_at, updated_at
		FROM remote_configs
		WHERE version = $1
	`

	rows, err := s.db.Query(ctx, query, version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []models.RemoteConfig
	for rows.Next() {
		var config models.RemoteConfig
		err := rows.Scan(&config.ID, &config.Name, &config.Version, &config.Value, &config.CreatedAt, &config.UpdatedAt)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	return configs, nil
}

func (s *RemoteConfigService) GetRemoteConfig(ctx context.Context, name string) (*models.RemoteConfig, error) {
	query := `
		SELECT id, name, version, value, created_at, updated_at
		FROM remote_configs
		WHERE name = $1
	`

	row := s.db.QueryRow(ctx, query, name)
	var config models.RemoteConfig
	err := row.Scan(
		&config.ID, &config.Name, &config.Version,
		&config.Value, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (s *RemoteConfigService) SaveRemoteConfig(ctx context.Context, config *models.RemoteConfig) error {
	insertQuery := `
		INSERT INTO remote_configs (id, name, version, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	updateQuery := `
		UPDATE remote_configs
		SET name = $2, version = $3, value = $4, updated_at = $5
		WHERE id = $1
	`

	if config.ID == "" {
		err := s.db.QueryRow(ctx, insertQuery, s.db.GenerateNewId(), config.Name, config.Version, config.Value, time.Now(), time.Time{}).Scan(&config.ID)
		if err != nil {
			return err
		}
	} else {
		_, err := s.db.Exec(ctx, updateQuery, config.ID, config.Name, config.Version, config.Value, time.Now())
		if err != nil {
			return err
		}
	}

	return nil
}

package event

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
)

type PgEventStore struct {
	db *data.PgDbContext
}

func NewPgEventStore(db *data.PgDbContext) EventStore {
	return &PgEventStore{db: db}
}

// Event operations
func (s *PgEventStore) CreateEvent(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO events (id, type, name, assets, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	configBytes, err := json.Marshal(event.Config)
	if err != nil {
		return err
	}

	assetsBytes, err := json.Marshal(event.Assets)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, query,
		event.ID,
		event.Type,
		event.Name,
		assetsBytes,
		configBytes,
		event.CreatedAt,
		event.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.New("event already exists")
		}
		return err
	}

	return nil
}

func (s *PgEventStore) UpdateEvent(ctx context.Context, event *models.Event) error {
	query := `
		UPDATE events
		SET type = $1, name = $2, assets = $3, config = $4, updated_at = $5
		WHERE id = $6
	`

	configBytes, err := json.Marshal(event.Config)
	if err != nil {
		return err
	}

	assetsBytes, err := json.Marshal(event.Assets)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, query,
		event.Type,
		event.Name,
		assetsBytes,
		configBytes,
		event.UpdatedAt,
		event.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *PgEventStore) GetEvent(ctx context.Context, id string) (*models.Event, error) {
	query := `
		SELECT id, type, name, assets, config, created_at, updated_at
		FROM events
		WHERE id = $1
	`

	event := &models.Event{}
	var configBytes []byte
	var assetsBytes []byte

	err := s.db.QueryRow(ctx, query, id).Scan(
		&event.ID,
		&event.Type,
		&event.Name,
		&assetsBytes,
		&configBytes,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("event not found")
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &event.Config); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(assetsBytes, &event.Assets); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *PgEventStore) ListEvents(ctx context.Context) ([]*models.Event, error) {
	query := `
		SELECT id, type, name, assets, config, created_at, updated_at
		FROM events
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		var configBytes []byte
		var assetsBytes []byte
		err := rows.Scan(
			&event.ID,
			&event.Type,
			&event.Name,
			&assetsBytes,
			&configBytes,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configBytes, &event.Config); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(assetsBytes, &event.Assets); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// Schedule operations
func (s *PgEventStore) CreateSchedule(ctx context.Context, schedule *models.EventSchedule) error {
	query := `
		INSERT INTO event_schedules (id, event_id, start_time, end_time, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.Exec(ctx, query,
		schedule.ID,
		schedule.EventID,
		schedule.StartTime,
		schedule.EndTime,
		schedule.IsActive,
		schedule.CreatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.New("schedule already exists")
		}
		return err
	}

	return nil
}

func (s *PgEventStore) UpdateSchedule(ctx context.Context, schedule *models.EventSchedule) error {
	query := `
		UPDATE event_schedules
		SET start_time = $1, end_time = $2, is_active = $3
		WHERE id = $4
	`

	_, err := s.db.Exec(ctx, query,
		schedule.StartTime,
		schedule.EndTime,
		schedule.IsActive,
		schedule.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *PgEventStore) GetSchedule(ctx context.Context, id string) (*models.EventSchedule, error) {
	query := `
		SELECT es.id, es.event_id, es.start_time, es.end_time, es.is_active, es.created_at, e.type
		FROM event_schedules es
		JOIN events e ON es.event_id = e.id
		WHERE es.id = $1
	`

	schedule := &models.EventSchedule{
		Event: &models.Event{},
	}
	err := s.db.QueryRow(ctx, query, id).Scan(
		&schedule.ID,
		&schedule.EventID,
		&schedule.StartTime,
		&schedule.EndTime,
		&schedule.IsActive,
		&schedule.CreatedAt,
		&schedule.Event.Type,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("schedule not found")
	}
	if err != nil {
		return nil, err
	}

	return schedule, nil
}

func (s *PgEventStore) ListActiveSchedules(ctx context.Context) ([]*models.ActiveEventSchedule, error) {
	query := `
		SELECT es.id, es.event_id, es.start_time, es.end_time, e.type, e.name, e.assets
		FROM event_schedules es
		JOIN events e ON es.event_id = e.id
		WHERE es.is_active = true AND es.start_time <= NOW() AND es.end_time >= NOW()
		ORDER BY es.start_time ASC
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.ActiveEventSchedule
	for rows.Next() {
		schedule := &models.ActiveEventSchedule{}
		var assetsBytes []byte

		err := rows.Scan(
			&schedule.ID,
			&schedule.EventID,
			&schedule.StartTime,
			&schedule.EndTime,
			&schedule.Type,
			&schedule.Name,
			&assetsBytes,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(assetsBytes, &schedule.Assets); err != nil {
			return nil, err
		}

		schedules = append(schedules, schedule)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}

func (s *PgEventStore) GetSchedulesByEventID(ctx context.Context, eventID string) ([]*models.EventSchedule, error) {
	query := `
		SELECT id, event_id, start_time, end_time, is_active, created_at
		FROM event_schedules
		WHERE event_id = $1
		ORDER BY start_time ASC
	`

	rows, err := s.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.EventSchedule
	for rows.Next() {
		schedule := &models.EventSchedule{}
		err := rows.Scan(
			&schedule.ID,
			&schedule.EventID,
			&schedule.StartTime,
			&schedule.EndTime,
			&schedule.IsActive,
			&schedule.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}

// Player Event operations
func (s *PgEventStore) CreatePlayerEvent(ctx context.Context, playerEvent *models.PlayerEvent) error {
	query := `
		INSERT INTO player_events (id, player_id, schedule_id, score, attempts, last_play, tickets, state, expires_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	stateBytes, err := json.Marshal(playerEvent.State)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, query,
		playerEvent.ID,
		playerEvent.PlayerID,
		playerEvent.ScheduleID,
		playerEvent.Score,
		playerEvent.Attempts,
		playerEvent.LastPlay,
		playerEvent.Tickets,
		stateBytes,
		playerEvent.ExpiresAt,
		playerEvent.CreatedAt,
		playerEvent.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.New("player event already exists")
		}
		return err
	}

	return nil
}

func (s *PgEventStore) UpdatePlayerEvent(ctx context.Context, playerEvent *models.PlayerEvent) error {
	query := `
		UPDATE player_events
		SET score = $1, attempts = $2, last_play = $3, tickets = $4, state = $5, updated_at = $6
		WHERE id = $7
	`

	stateBytes, err := json.Marshal(playerEvent.State)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, query,
		playerEvent.Score,
		playerEvent.Attempts,
		playerEvent.LastPlay,
		playerEvent.Tickets,
		stateBytes,
		playerEvent.UpdatedAt,
		playerEvent.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *PgEventStore) GetPlayerEvent(ctx context.Context, playerID, scheduleID string) (*models.PlayerEvent, error) {
	query := `
		SELECT id, player_id, schedule_id, score, attempts, last_play, tickets, state, expires_at, created_at, updated_at
		FROM player_events
		WHERE player_id = $1 AND schedule_id = $2
	`

	playerEvent := &models.PlayerEvent{}
	var stateBytes []byte

	err := s.db.QueryRow(ctx, query, playerID, scheduleID).Scan(
		&playerEvent.ID,
		&playerEvent.PlayerID,
		&playerEvent.ScheduleID,
		&playerEvent.Score,
		&playerEvent.Attempts,
		&playerEvent.LastPlay,
		&playerEvent.Tickets,
		&stateBytes,
		&playerEvent.ExpiresAt,
		&playerEvent.CreatedAt,
		&playerEvent.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, errors.New("player event not found")
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(stateBytes, &playerEvent.State); err != nil {
		return nil, err
	}

	return playerEvent, nil
}

func (s *PgEventStore) ListPlayerEvents(ctx context.Context, playerID string) ([]*models.PlayerEvent, error) {
	query := `
		SELECT id, player_id, schedule_id, score, attempts, last_play, tickets, state, expires_at, created_at, updated_at
		FROM player_events
		WHERE player_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playerEvents []*models.PlayerEvent
	for rows.Next() {
		playerEvent := &models.PlayerEvent{}
		var stateBytes []byte

		err := rows.Scan(
			&playerEvent.ID,
			&playerEvent.PlayerID,
			&playerEvent.ScheduleID,
			&playerEvent.Score,
			&playerEvent.Attempts,
			&playerEvent.LastPlay,
			&playerEvent.Tickets,
			&stateBytes,
			&playerEvent.ExpiresAt,
			&playerEvent.CreatedAt,
			&playerEvent.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(stateBytes, &playerEvent.State); err != nil {
			return nil, err
		}

		playerEvents = append(playerEvents, playerEvent)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return playerEvents, nil
}

package services

import (
	"context"
	"errors"
	"time"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/internal/services/event"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/google/uuid"
)

var (
	ErrEventNotFound     = errors.New("event not found")
	ErrEventNotActive    = errors.New("event not active")
	ErrScheduleNotFound  = errors.New("schedule not found")
	ErrScheduleNotActive = errors.New("schedule not active")
	ErrEventExpired      = errors.New("event expired")
)

type EventService struct {
	store       event.EventStore
	gameFactory event.IEventGameFactory
}

func NewEventService(db *data.PgDbContext) *EventService {

	return &EventService{
		store:       event.NewPgEventStore(db),
		gameFactory: event.NewEventGameFactory(),
	}
}

func (s *EventService) CreateEvent(ctx context.Context, event *models.Event) error {
	if err := event.Validate(); err != nil {
		return err
	}

	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	return s.store.CreateEvent(ctx, event)
}

func (s *EventService) UpdateEvent(ctx context.Context, event *models.Event) error {
	if err := event.Validate(); err != nil {
		return err
	}

	event.UpdatedAt = time.Now()
	return s.store.UpdateEvent(ctx, event)
}

func (s *EventService) GetEvent(ctx context.Context, id string) (*models.Event, error) {
	return s.store.GetEvent(ctx, id)
}

func (s *EventService) ListEvents(ctx context.Context) ([]*models.Event, error) {
	return s.store.ListEvents(ctx)
}

func (s *EventService) CreateSchedule(ctx context.Context, schedule *models.EventSchedule) error {
	if err := schedule.Validate(); err != nil {
		return err
	}

	now := time.Now()
	active := true

	schedule.ID = uuid.New().String()
	schedule.CreatedAt = now
	schedule.IsActive = active

	return s.store.CreateSchedule(ctx, schedule)
}

func (s *EventService) UpdateSchedule(ctx context.Context, schedule *models.EventSchedule) error {
	if err := schedule.Validate(); err != nil {
		return err
	}

	return s.store.UpdateSchedule(ctx, schedule)
}

func (s *EventService) GetSchedule(ctx context.Context, id string) (*models.EventSchedule, error) {
	return s.store.GetSchedule(ctx, id)
}

func (s *EventService) ListActiveSchedules(ctx context.Context) ([]*models.ActiveEventSchedule, error) {
	return s.store.ListActiveSchedules(ctx)
}

func (s *EventService) GetSchedulesByEvent(ctx context.Context, eventID string) ([]*models.EventSchedule, error) {
	return s.store.GetSchedulesByEventID(ctx, eventID)
}

func (s *EventService) GetOrCreatePlayerEvent(ctx context.Context, playerID, scheduleID string) (*models.PlayerEventSchedule, error) {
	result := &models.PlayerEventSchedule{}
	playerEvent, err := s.store.GetPlayerEvent(ctx, playerID, scheduleID)
	if err == nil {
		result.ScheduleID = playerEvent.ScheduleID
		result.Score = playerEvent.Score
		result.Attempts = playerEvent.Attempts
		result.LastPlay = playerEvent.LastPlay
		result.Tickets = playerEvent.Tickets
		result.State = playerEvent.State

		return result, nil
	}

	schedule, err := s.store.GetSchedule(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	if !schedule.IsActive {
		return nil, ErrScheduleNotActive
	}

	game, err := s.gameFactory.CreateEventGame(schedule.Event.Type)
	if err != nil {
		return nil, err
	}

	event, err := s.store.GetEvent(ctx, schedule.EventID)
	if err != nil {
		return nil, err
	}

	state := game.GetInitialState(*event)
	playerEvent = &models.PlayerEvent{
		ID:         uuid.New().String(),
		PlayerID:   playerID,
		ScheduleID: scheduleID,
		Tickets:    99999,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ExpiresAt:  schedule.EndTime,
		State:      state,
	}

	if err := s.store.CreatePlayerEvent(ctx, playerEvent); err != nil {
		return nil, err
	}

	result.ScheduleID = playerEvent.ScheduleID
	result.Score = playerEvent.Score
	result.Attempts = playerEvent.Attempts
	result.LastPlay = playerEvent.LastPlay
	result.Tickets = playerEvent.Tickets
	result.State = playerEvent.State

	return result, nil
}

func (s *EventService) ListPlayerEvents(ctx context.Context, playerID string) ([]*models.PlayerEvent, error) {
	return s.store.ListPlayerEvents(ctx, playerID)
}

func (s *EventService) UpdatePlayerEvent(ctx context.Context, playerEvent *models.PlayerEvent) error {
	playerEvent.UpdatedAt = time.Now()
	return s.store.UpdatePlayerEvent(ctx, playerEvent)
}

// PlayEvent handles a game play attempt
func (s *EventService) PlayEvent(ctx context.Context, playerID, scheduleID string, playData map[string]interface{}) (*models.EventPlayResult, error) {
	// Get user event
	playerEvent, err := s.store.GetPlayerEvent(ctx, playerID, scheduleID)
	if err != nil {
		return nil, err
	}

	// Get schedule and event
	schedule, err := s.store.GetSchedule(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	if !schedule.IsActive {
		return nil, ErrScheduleNotActive
	}

	event, err := s.store.GetEvent(ctx, schedule.EventID)
	if err != nil {
		return nil, err
	}

	// Create game instance
	gameInstance, err := s.gameFactory.CreateEventGame(event.Type)
	if err != nil {
		return nil, err
	}

	// Initialize game
	if err := gameInstance.Initialize(ctx, *event, *playerEvent); err != nil {
		return nil, err
	}

	// Create play request
	playRequest := &models.EventPlayRequest{
		PlayerEvent: playerEvent,
		PlayData:    playData,
	}

	// Validate play
	if err := gameInstance.ValidatePlay(ctx, playRequest); err != nil {
		return nil, err
	}

	// Process play
	playResult, err := gameInstance.ProcessPlay(ctx, playRequest)
	if err != nil {
		return nil, err
	}

	// Update player event state
	if err := gameInstance.UpdateState(ctx, playerEvent, playResult); err != nil {
		return nil, err
	}

	// Save updated player event
	if err := s.store.UpdatePlayerEvent(ctx, playerEvent); err != nil {
		return nil, err
	}

	return playResult, nil
}

func (s *EventService) RefreshPlayerEventState(ctx context.Context, playerID, scheduleID string) (*models.PlayerEvent, error) {
	playerEvent, err := s.store.GetPlayerEvent(ctx, playerID, scheduleID)
	if err != nil {
		return nil, err
	}

	schedule, err := s.store.GetSchedule(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	if !schedule.IsActive {
		return nil, ErrScheduleNotActive
	}

	game, err := s.gameFactory.CreateEventGame(schedule.Event.Type)
	if err != nil {
		return nil, err
	}

	event, err := s.store.GetEvent(ctx, schedule.EventID)
	if err != nil {
		return nil, err
	}

	state := game.GetInitialState(*event)
	playerEvent.State = state

	if err := s.store.UpdatePlayerEvent(ctx, playerEvent); err != nil {
		return nil, err
	}

	return playerEvent, nil
}

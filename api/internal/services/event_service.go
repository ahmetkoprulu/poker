package services

import (
	"context"
	"errors"
	"time"

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

func NewEventService(store event.EventStore) *EventService {
	return &EventService{
		store:       store,
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

func (s *EventService) ListActiveEvents(ctx context.Context) ([]*models.Event, error) {
	return s.store.ListActiveEvents(ctx)
}

func (s *EventService) CreateSchedule(ctx context.Context, schedule *models.EventSchedule) error {
	if err := schedule.Validate(); err != nil {
		return err
	}

	event, err := s.store.GetEvent(ctx, schedule.EventID)
	if err != nil {
		return err
	}
	if !event.IsActive {
		return ErrEventNotActive
	}

	schedule.ID = uuid.New().String()
	schedule.CreatedAt = time.Now()
	schedule.IsActive = true

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

func (s *EventService) ListActiveSchedules(ctx context.Context) ([]*models.EventSchedule, error) {
	return s.store.ListActiveSchedules(ctx)
}

func (s *EventService) GetSchedulesByEvent(ctx context.Context, eventID string) ([]*models.EventSchedule, error) {
	return s.store.GetSchedulesByEventID(ctx, eventID)
}

func (s *EventService) GetOrCreateUserEvent(ctx context.Context, userID, scheduleID string) (*models.UserEvent, error) {
	userEvent, err := s.store.GetUserEvent(ctx, userID, scheduleID)
	if err == nil {
		return userEvent, nil
	}

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

	if !event.IsActive {
		return nil, ErrEventNotActive
	}

	userEvent = &models.UserEvent{
		ID:         uuid.New().String(),
		UserID:     userID,
		ScheduleID: scheduleID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ExpiresAt:  schedule.EndTime,
		State:      make(map[string]interface{}),
	}

	if err := s.store.CreateUserEvent(ctx, userEvent); err != nil {
		return nil, err
	}

	return userEvent, nil
}

func (s *EventService) ListUserEvents(ctx context.Context, userID string) ([]*models.UserEvent, error) {
	return s.store.ListUserEvents(ctx, userID)
}

func (s *EventService) UpdateUserEvent(ctx context.Context, userEvent *models.UserEvent) error {
	userEvent.UpdatedAt = time.Now()
	return s.store.UpdateUserEvent(ctx, userEvent)
}

// PlayEvent handles a game play attempt
func (s *EventService) PlayEvent(ctx context.Context, userID string, scheduleID string, playData map[string]interface{}) (*models.EventPlayResult, error) {
	// Get user event
	userEvent, err := s.GetOrCreateUserEvent(ctx, userID, scheduleID)
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

	if !event.IsActive {
		return nil, ErrEventNotActive
	}

	// Create game instance
	gameInstance, err := s.gameFactory.CreateEventGame(event.Type)
	if err != nil {
		return nil, err
	}

	// Initialize game
	if err := gameInstance.Initialize(ctx, event, userEvent); err != nil {
		return nil, err
	}

	// Create play request
	playRequest := &models.EventPlayRequest{
		UserEvent: userEvent,
		PlayData:  playData,
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

	// Update user event state
	if err := gameInstance.UpdateState(ctx, userEvent, playResult); err != nil {
		return nil, err
	}

	// Save updated user event
	if err := s.store.UpdateUserEvent(ctx, userEvent); err != nil {
		return nil, err
	}

	return playResult, nil
}

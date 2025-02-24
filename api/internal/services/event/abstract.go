package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ahmetkoprulu/rtrp/models"
)

var (
	ErrGameCompleted    = errors.New("game already completed")
	ErrInvalidConfig    = errors.New("invalid game configuration")
	ErrNotEnoughTickets = errors.New("not enough tickets")
)

type EventGame interface {
	Initialize(ctx context.Context, event models.Event, playerEvent models.PlayerEvent) error
	ValidatePlay(ctx context.Context, req *models.EventPlayRequest) error
	ProcessPlay(ctx context.Context, req *models.EventPlayRequest) (*models.EventPlayResult, error)
	// CalculateRewards(ctx context.Context, req *models.EventPlayRequest, result *models.EventPlayResult) error
	UpdateState(ctx context.Context, playerEvent *models.PlayerEvent, result *models.EventPlayResult) error
	GetInitialState() map[string]interface{}
}

type BaseEventGame struct {
	Event       *models.Event
	PlayerEvent *models.PlayerEvent
}

func (g *BaseEventGame) Initialize(ctx context.Context, event models.Event, playerEvent models.PlayerEvent) error {
	g.Event = &event
	g.PlayerEvent = &playerEvent
	return nil
}

type IEventGameFactory interface {
	CreateEventGame(eventType models.EventType) (EventGame, error)
}

type EventGameFactory struct{}

func NewEventGameFactory() IEventGameFactory {
	return &EventGameFactory{}
}

func (f *EventGameFactory) CreateEventGame(eventType models.EventType) (EventGame, error) {
	switch eventType {
	case models.EventTypePathGame:
		return NewPathGame(), nil
	default:
		return nil, fmt.Errorf("unknown game type: %v", eventType)
	}
}

type EventStore interface {
	CreateEvent(ctx context.Context, event *models.Event) error
	UpdateEvent(ctx context.Context, event *models.Event) error
	GetEvent(ctx context.Context, id string) (*models.Event, error)
	ListEvents(ctx context.Context) ([]*models.Event, error)

	CreateSchedule(ctx context.Context, schedule *models.EventSchedule) error
	UpdateSchedule(ctx context.Context, schedule *models.EventSchedule) error
	GetSchedule(ctx context.Context, id string) (*models.EventSchedule, error)
	ListActiveSchedules(ctx context.Context) ([]*models.ActiveEventSchedule, error)
	GetSchedulesByEventID(ctx context.Context, eventID string) ([]*models.EventSchedule, error)

	CreatePlayerEvent(ctx context.Context, playerEvent *models.PlayerEvent) error
	UpdatePlayerEvent(ctx context.Context, playerEvent *models.PlayerEvent) error
	GetPlayerEvent(ctx context.Context, playerID, scheduleID string) (*models.PlayerEvent, error)
	ListPlayerEvents(ctx context.Context, playerID string) ([]*models.PlayerEvent, error)
}

func parseConfig[T any](input map[string]interface{}, config *T) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, config)
}

func parseState[T any](input map[string]interface{}, state *T) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, state)
}

func stateToMap[T any](state T) map[string]interface{} {
	data, _ := json.Marshal(state)
	result := make(map[string]interface{})
	_ = json.Unmarshal(data, &result)
	return result
}

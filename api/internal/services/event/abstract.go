package event

import (
	"context"
	"fmt"

	"github.com/ahmetkoprulu/rtrp/models"
)

type EventGame interface {
	Initialize(ctx context.Context, event *models.Event, userEvent *models.UserEvent) error
	ValidatePlay(ctx context.Context, req *models.EventPlayRequest) error
	ProcessPlay(ctx context.Context, req *models.EventPlayRequest) (*models.EventPlayResult, error)
	// CalculateRewards(ctx context.Context, req *PlayRequest, result *models.EventPlayResult[any]) error
	UpdateState(ctx context.Context, userEvent *models.UserEvent, result *models.EventPlayResult) error
}

type BaseEventGame struct {
	Event     *models.Event
	UserEvent *models.UserEvent
}

func (g *BaseEventGame) Initialize(ctx context.Context, event *models.Event, userEvent *models.UserEvent) error {
	g.Event = event
	g.UserEvent = userEvent
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

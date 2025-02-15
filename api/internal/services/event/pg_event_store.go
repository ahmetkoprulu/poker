package event

import (
	"context"

	"github.com/ahmetkoprulu/rtrp/models"
)

type EventStore interface {
	CreateEvent(ctx context.Context, event *models.Event) error
	UpdateEvent(ctx context.Context, event *models.Event) error
	GetEvent(ctx context.Context, id string) (*models.Event, error)
	ListActiveEvents(ctx context.Context) ([]*models.Event, error)

	CreateSchedule(ctx context.Context, schedule *models.EventSchedule) error
	UpdateSchedule(ctx context.Context, schedule *models.EventSchedule) error
	GetSchedule(ctx context.Context, id string) (*models.EventSchedule, error)
	ListActiveSchedules(ctx context.Context) ([]*models.EventSchedule, error)
	GetSchedulesByEventID(ctx context.Context, eventID string) ([]*models.EventSchedule, error)

	CreateUserEvent(ctx context.Context, userEvent *models.UserEvent) error
	UpdateUserEvent(ctx context.Context, userEvent *models.UserEvent) error
	GetUserEvent(ctx context.Context, userID, scheduleID string) (*models.UserEvent, error)
	ListUserEvents(ctx context.Context, userID string) ([]*models.UserEvent, error)
}

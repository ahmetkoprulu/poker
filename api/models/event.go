package models

import (
	"errors"
	"time"
)

type EventType int16

const (
	EventTypeSlotMachine EventType = 1
	EventTypeDiceGame    EventType = 2
	EventTypePathGame    EventType = 3
)

const (
	MaxEventLevel     = 100
	DefaultMultiplier = 1.0
	MinEntryFee       = 100
	MaxTickets        = 100
	DefaultOdds       = 100
)

var (
	ErrInvalidLevel    = errors.New("invalid level configuration")
	ErrInvalidPayTable = errors.New("invalid pay table configuration")
	ErrInvalidSchedule = errors.New("invalid schedule configuration")
	ErrInvalidReward   = errors.New("invalid reward configuration")
)

type ProductType int16

const (
	ProductTypeChips ProductType = 1
	ProductTypeGold  ProductType = 2
)

type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Name      string                 `json:"name"`
	Config    map[string]interface{} `json:"config"`
	IsActive  bool                   `json:"is_active"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type EventSchedule struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type UserEvent struct {
	ID         string `json:"id"`
	ScheduleID string `json:"schedule_id"`
	UserID     string `json:"user_id"`

	Score    int64     `json:"score"`
	Attempts int       `json:"attempts"`
	LastPlay time.Time `json:"last_play"`

	Tickets   int                    `json:"tickets_left"`
	State     map[string]interface{} `json:"state"`
	ExpiresAt time.Time              `json:"expires_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EventPlayRequest struct {
	UserEvent *UserEvent
	PlayData  map[string]interface{}
}

type EventPlayResult struct {
	UserEvent UserEvent     `json:"user_event"`
	Rewards   []EventReward `json:"rewards"`
	Data      any           `json:"data"`
}

type RewardValue struct {
	Amount   int64                  `json:"amount"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type EventReward struct {
	Type  ProductType `json:"type"`
	Value RewardValue `json:"value"`
}

func (r *EventReward) GetChipsAmount() int64 {
	if r.Type == ProductTypeChips {
		return r.Value.Amount
	}
	return 0
}

func (r *EventReward) GetGoldAmount() int32 {
	if r.Type == ProductTypeGold {
		return int32(r.Value.Amount)
	}
	return 0
}

func NewChipsReward(amount int64) EventReward {
	return EventReward{
		Type: ProductTypeChips,
		Value: RewardValue{
			Amount: amount,
		},
	}
}

func NewGoldReward(amount int32) EventReward {
	return EventReward{
		Type: ProductTypeGold,
		Value: RewardValue{
			Amount: int64(amount),
		},
	}
}

func (e *Event) Validate() error {
	if e.Type == 0 || e.Name == "" {
		return errors.New("event type and name are required")
	}

	return nil
}

func (s *EventSchedule) Validate() error {
	if s.StartTime.IsZero() || s.EndTime.IsZero() {
		return ErrInvalidSchedule
	}
	if s.EndTime.Before(s.StartTime) {
		return errors.New("end time cannot be before start time")
	}
	return nil
}

func (p *UserEvent) IsActive(schedule EventSchedule) bool {
	now := time.Now()
	return p.Tickets > 0 &&
		now.Before(p.ExpiresAt) &&
		now.Before(schedule.EndTime) &&
		schedule.IsActive
}

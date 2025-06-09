package event

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/ahmetkoprulu/rtrp/models"
)

var (
	ErrInvalidWack         = errors.New("invalid_wack")
	ErrInvalidWackPosition = errors.New("invalid_wack_position")
	ErrAnyRowEnded         = errors.New("any_row_ended")
)

type PushEventConfig struct {
	EntryFee int            `json:"entry_fee"` // Entry fee for the game
	Rows     []PushEventRow `json:"rows"`
}

type PushEventState struct {
	Rows         []PushEventRow `json:"rows"`
	LastPosition int            `json:"last_position"`
	LastReward   *models.Item   `json:"last_reward"`
}

type PushEventRow struct {
	Position int           `json:"position"`
	Rewards  []models.Item `json:"rewards"`
}

type PushEventResult struct {
	Row PushEventRow `json:"row"`
}

type PushEvent struct {
	BaseEventGame
	config PushEventConfig
	state  PushEventState
}

func NewPushEvent() EventGame {
	return &PushEvent{}
}

func (g *PushEvent) Initialize(ctx context.Context, event models.Event, playerEvent models.PlayerEvent) error {
	g.BaseEventGame.Initialize(ctx, event, playerEvent)
	if err := parseConfig(event.Config, &g.config); err != nil {
		return ErrInvalidConfig
	}

	if g.config.Rows == nil {
		return ErrInvalidConfig
	}

	if len(g.config.Rows) == 0 {
		return ErrInvalidConfig
	}

	if g.PlayerEvent.State == nil {
		g.state = PushEventState{
			Rows: g.config.Rows,
		}
		g.state.LastPosition = -1
		g.state.LastReward = nil
		g.PlayerEvent.State = stateToMap(g.state)
	} else {
		if err := parseState(g.PlayerEvent.State, &g.state); err != nil {
			return err
		}
	}

	return nil
}

func (g *PushEvent) ValidatePlay(ctx context.Context, req *models.EventPlayRequest) error {
	if g.PlayerEvent.State == nil {
		return ErrInvalidState
	}

	isAnyRowEnded := false
	for _, row := range g.state.Rows {
		if len(row.Rewards) == 0 {
			isAnyRowEnded = true
			break
		}
	}

	if isAnyRowEnded {
		return ErrAnyRowEnded
	}

	if g.PlayerEvent.Tickets < g.config.EntryFee {
		return ErrNotEnoughTickets
	}
	return nil
}

func (g *PushEvent) ProcessPlay(ctx context.Context, req *models.EventPlayRequest) (*models.EventPlayResult, error) {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	position := rand.Intn(len(g.state.Rows))

	var reward models.Item
	for i, wack := range g.state.Rows {
		if wack.Position == position {
			reward = wack.Rewards[len(wack.Rewards)-1]                   // Get last reward
			g.state.Rows[i].Rewards = wack.Rewards[:len(wack.Rewards)-1] // and pop it
			break
		}
	}

	// Create result
	result := &models.EventPlayResult{
		PlayerEvent: *g.PlayerEvent,
		Rewards:     make([]models.Item, 0),
		Data: PushEventResult{
			Row: PushEventRow{
				Position: position,
				Rewards:  g.state.Rows[position].Rewards,
			},
		},
	}
	result.Rewards = append(result.Rewards, reward)

	// Update state
	g.state.LastPosition = position
	g.state.LastReward = &reward

	// Update user event
	g.PlayerEvent.Score += int64(g.config.EntryFee * 10)

	isAnyRowEnded := false
	for _, row := range g.state.Rows {
		if len(row.Rewards) == 0 {
			isAnyRowEnded = true
			break
		}
	}

	if isAnyRowEnded {
		g.state.Rows = g.config.Rows
	}

	return result, nil
}

func (g *PushEvent) UpdateState(ctx context.Context, playerEvent *models.PlayerEvent, result *models.EventPlayResult) error {
	playerEvent.Score = g.PlayerEvent.Score
	playerEvent.Attempts++
	playerEvent.LastPlay = time.Now()
	playerEvent.Tickets -= g.config.EntryFee
	playerEvent.State = stateToMap(g.state)

	result.PlayerEvent = *playerEvent

	return nil
}

func (g *PushEvent) GetInitialState(event models.Event) map[string]interface{} {
	if err := parseConfig(event.Config, &g.config); err != nil {
		return nil
	}

	return stateToMap(PushEventState{
		Rows:         g.config.Rows,
		LastPosition: -1,
		LastReward:   nil,
	})
}

func (g *PushEvent) GetRewards(eventPlayResult *models.EventPlayResult) []models.Item {
	return eventPlayResult.Rewards
}

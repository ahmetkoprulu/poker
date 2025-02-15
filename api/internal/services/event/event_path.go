package event

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"

	"github.com/ahmetkoprulu/rtrp/models"
)

var (
	ErrInvalidPosition  = errors.New("invalid position")
	ErrGameCompleted    = errors.New("game already completed")
	ErrInvalidConfig    = errors.New("invalid game configuration")
	ErrNotEnoughTickets = errors.New("not enough tickets")
)

type PathGameConfig struct {
	EntryFee  int                    `json:"entry_fee"`  // Entry fee for the game
	Steps     []PathStep             `json:"steps"`      // Array of steps and their rewards
	DiceCount int                    `json:"dice_count"` // Number of dice to roll (default 1)
	DiceSides int                    `json:"dice_sides"` // Number of sides on each die (default 6)
	Features  map[string]interface{} `json:"features"`   // Additional features like special tiles
}

type PathStep struct {
	Reward  models.EventReward     `json:"reward"`
	Special bool                   `json:"special"` // Indicates if this is a special tile
	Effects map[string]interface{} `json:"effects"` // Special effects (move forward/backward, extra roll, etc.)
}

type PathGameState struct {
	TotalSteps     int             `json:"total_steps"`     // Total steps taken
	LastRoll       []int           `json:"last_roll"`       // Last dice roll results
	SpecialEffects []string        `json:"special_effects"` // Active special effects
	Collected      map[string]bool `json:"collected"`       // Steps where rewards were collected
}

type PathGameResult struct {
	TotalSteps     int             `json:"total_steps"`     // Total steps taken
	LastRoll       []int           `json:"last_roll"`       // Last dice roll results
	SpecialEffects []string        `json:"special_effects"` // Active special effects
	Collected      map[string]bool `json:"collected"`       // Steps where rewards were collected
}

type PathGame struct {
	BaseEventGame
	config PathGameConfig
	state  PathGameState
}

func NewPathGame() EventGame {
	return &PathGame{}
}

func (g *PathGame) Initialize(ctx context.Context, event *models.Event, userEvent *models.UserEvent) error {
	g.BaseEventGame.Initialize(ctx, event, userEvent)

	// Parse dynamic config into typed config
	if err := parseConfig(event.Config, &g.config); err != nil {
		return ErrInvalidConfig
	}

	if g.config.DiceCount == 0 {
		g.config.DiceCount = 1
	}
	if g.config.DiceSides == 0 {
		g.config.DiceSides = 6
	}
	if len(g.config.Steps) == 0 {
		return ErrInvalidConfig
	}

	if userEvent.State == nil {
		g.state = PathGameState{
			TotalSteps:     0,
			LastRoll:       make([]int, 0),
			SpecialEffects: make([]string, 0),
			Collected:      make(map[string]bool),
		}
		userEvent.State = stateToMap(g.state)
	} else {
		if err := parseState(userEvent.State, &g.state); err != nil {
			return err
		}
	}

	return nil
}

func (g *PathGame) ValidatePlay(ctx context.Context, req *models.EventPlayRequest) error {
	if g.UserEvent.Tickets < g.config.EntryFee {
		return ErrNotEnoughTickets
	}
	return nil
}

func (g *PathGame) ProcessPlay(ctx context.Context, req *models.EventPlayRequest) (*models.EventPlayResult, error) {
	rolls := g.rollDice()
	totalMove := 0
	for _, roll := range rolls {
		totalMove += roll
	}

	maxPosition := len(g.config.Steps)
	newTotalSteps := g.state.TotalSteps + totalMove
	currentPosition := newTotalSteps % maxPosition
	completedLaps := newTotalSteps / maxPosition

	// Get current step directly from array index
	currentStep := &g.config.Steps[currentPosition]

	result := &models.EventPlayResult{
		UserEvent: *g.UserEvent,
		Rewards:   make([]models.EventReward, 0),
		Data: PathGameResult{
			TotalSteps:     newTotalSteps,
			LastRoll:       rolls,
			SpecialEffects: g.state.SpecialEffects,
			Collected:      g.state.Collected,
		},
	}

	// Update state
	g.state.TotalSteps = newTotalSteps
	g.state.LastRoll = rolls

	// Calculate base score and progress
	baseScore := int64(totalMove * 10) // Base score from movement

	// Apply step rewards if not collected
	posKey := strconv.Itoa(currentPosition)
	if !g.state.Collected[posKey] {
		baseScore += currentStep.Reward.GetChipsAmount()
		g.state.Collected[posKey] = true
		result.Rewards = append(result.Rewards, currentStep.Reward)

		// Apply special effects
		if currentStep.Special {
			// g.applyEffects(currentStep.Effects, result)
		}
	}

	// Add lap completion bonus if we just completed a lap
	prevLaps := g.state.TotalSteps / maxPosition
	if completedLaps > prevLaps {
		lapBonus := models.NewChipsReward(1000 * int64(completedLaps))
		result.Rewards = append(result.Rewards, lapBonus)
	}

	// Update user event
	g.UserEvent.Score += baseScore
	g.UserEvent.State = stateToMap(g.state)

	return result, nil
}

func (g *PathGame) UpdateState(ctx context.Context, userEvent *models.UserEvent, result *models.EventPlayResult) error {
	return nil
}

// Helper methods

func (g *PathGame) rollDice() []int {
	rolls := make([]int, g.config.DiceCount)
	for i := 0; i < g.config.DiceCount; i++ {
		rolls[i] = rand.Intn(g.config.DiceSides) + 1
	}
	return rolls
}

// func (g *PathGame) applyEffects(effects map[string]interface{}, result *models.EventPlayResult) {
// 	if moveForward, ok := effects["move_forward"].(float64); ok {
// 		g.state.TotalSteps += int(moveForward)
// 		result.Data.(PathGameResult).TotalSteps = g.state.TotalSteps
// 		result.Data.(PathGameResult).LastRoll = g.state.LastRoll
// 	}

// 	if moveBackward, ok := effects["move_backward"].(float64); ok {
// 		g.state.TotalSteps -= int(moveBackward)
// 		if g.state.TotalSteps < 0 {
// 			g.state.TotalSteps = 0
// 		}
// 		result.Data.TotalSteps = g.state.TotalSteps
// 		result.Data.LastRoll = g.state.LastRoll
// 	}

// 	if extraRoll, ok := effects["extra_roll"].(bool); ok && extraRoll {
// 		g.state.SpecialEffects = append(g.state.SpecialEffects, "EXTRA_ROLL")
// 	}
// }

func parseConfig(input map[string]interface{}, config *PathGameConfig) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, config)
}

func parseState(input map[string]interface{}, state *PathGameState) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, state)
}

func stateToMap(state PathGameState) map[string]interface{} {
	data, _ := json.Marshal(state)
	result := make(map[string]interface{})
	_ = json.Unmarshal(data, &result)
	return result
}

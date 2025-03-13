package event

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/ahmetkoprulu/rtrp/models"
)

var (
	ErrInvalidPosition = errors.New("invalid_position")
)

type PathGameConfig struct {
	EntryFee  int        `json:"entry_fee"`  // Entry fee for the game
	Steps     []PathStep `json:"steps"`      // Array of steps and their rewards
	DiceCount int        `json:"dice_count"` // Number of dice to roll (default 1)
	DiceSides int        `json:"dice_sides"` // Number of sides on each die (default 6)
	// Features  map[string]interface{} `json:"features"`   // Additional features like special tiles
}

type PathStep struct {
	Reward models.Item `json:"reward"`
}

type PathGameState struct {
	TotalSteps  int   `json:"total_steps"`  // Total steps taken
	LastRoll    []int `json:"last_roll"`    // Last dice roll results
	CurrentStep int   `json:"current_step"` // Current step
}

type PathGameResult struct {
	OldStep       int           `json:"old_step"`        // Old step
	CurrentStep   int           `json:"current_step"`    // Current step
	TotalSteps    int           `json:"total_steps"`     // Total steps taken
	LastRoll      []int         `json:"last_roll"`       // Last dice roll results
	LevelUpReward []models.Item `json:"level_up_reward"` // Level up reward
}

type PathGame struct {
	BaseEventGame
	config PathGameConfig
	state  PathGameState
}

func NewPathGame() EventGame {
	return &PathGame{}
}

func (g *PathGame) Initialize(ctx context.Context, event models.Event, playerEvent models.PlayerEvent) error {
	g.BaseEventGame.Initialize(ctx, event, playerEvent)

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

	if g.PlayerEvent.State == nil {
		g.state = PathGameState{
			TotalSteps:  0,
			CurrentStep: 0,
			LastRoll:    make([]int, 0),
		}
		g.PlayerEvent.State = stateToMap(g.state)
	} else {
		if err := parseState[PathGameState](g.PlayerEvent.State, &g.state); err != nil {
			return err
		}
	}

	return nil
}

func (g *PathGame) ValidatePlay(ctx context.Context, req *models.EventPlayRequest) error {
	if g.PlayerEvent.Tickets < g.config.EntryFee {
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

	prevLaps := g.state.TotalSteps / len(g.config.Steps)
	maxPosition := len(g.config.Steps)
	newTotalSteps := g.state.TotalSteps + totalMove
	currentPosition := newTotalSteps % maxPosition
	completedLaps := newTotalSteps / maxPosition

	// Get current step directly from array index
	currentStep := &g.config.Steps[currentPosition]
	pathResult := &PathGameResult{
		TotalSteps:    newTotalSteps,
		LastRoll:      rolls,
		LevelUpReward: make([]models.Item, 0),
	}

	result := &models.EventPlayResult{
		PlayerEvent: *g.PlayerEvent,
		Rewards:     make([]models.Item, 0),
	}

	// Update state
	g.state.TotalSteps = newTotalSteps
	g.state.CurrentStep = currentPosition
	g.state.LastRoll = rolls

	reward := g.HandleMultiplier(g.PlayerEvent, currentStep.Reward)
	result.Rewards = append(result.Rewards, reward)

	// Add lap completion bonus if we just completed a lap
	if completedLaps > prevLaps {
		lapBonus := models.Item{
			Type:   models.ItemTypeChips,
			Amount: 1000 * completedLaps,
		}
		pathResult.LevelUpReward = append(result.Rewards, lapBonus)
	}

	result.Data = pathResult

	// Update user event
	g.PlayerEvent.Score += int64(totalMove * 10)

	return result, nil
}

func (g *PathGame) UpdateState(ctx context.Context, playerEvent *models.PlayerEvent, result *models.EventPlayResult) error {
	var ticketReward *models.Item
	for _, reward := range result.Rewards {
		if reward.Type == models.ItemTypeEvent {
			ticketReward = &reward
			break
		}
	}

	if ticketReward != nil {
		playerEvent.Tickets += ticketReward.Amount
	}

	playerEvent.Score = g.PlayerEvent.Score
	playerEvent.Multiplier = g.PlayerEvent.Multiplier
	playerEvent.Attempts++
	playerEvent.LastPlay = time.Now()
	playerEvent.Tickets -= g.config.EntryFee
	playerEvent.State = stateToMap(g.state)

	result.PlayerEvent = *playerEvent

	return nil
}

func (g *PathGame) GetInitialState(event models.Event) map[string]interface{} {
	return stateToMap(PathGameState{
		TotalSteps:  0,
		CurrentStep: 0,
		LastRoll:    make([]int, 0),
	})
}

func (g *PathGame) rollDice() []int {
	rolls := make([]int, g.config.DiceCount)
	for i := 0; i < g.config.DiceCount; i++ {
		rolls[i] = rand.Intn(g.config.DiceSides) + 1
	}
	return rolls
}

func (g *PathGame) GetRewards(eventPlayResult *models.EventPlayResult) []models.Item {
	pathEventResult := eventPlayResult.Data.(*PathGameResult)
	rewards := append(eventPlayResult.Rewards, pathEventResult.LevelUpReward...)
	return rewards
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

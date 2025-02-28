package models

import (
	"encoding/json"
	"errors"

	"slices"

	"github.com/google/uuid"
)

type GameStatus string

const (
	GameStatusWaiting  GameStatus = "waiting"
	GameStatusStarted  GameStatus = "started"
	GameStatusFinished GameStatus = "finished"
)

type GamePlayerStatus string

const (
	GamePlayerStatusWaiting  GamePlayerStatus = "waiting"
	GamePlayerStatusActive   GamePlayerStatus = "active"
	GamePlayerStatusInactive GamePlayerStatus = "inactive"
)

type GameError error

var (
	ErrorGameFull            GameError = errors.New("game_full")
	ErrorGamePlayerAlreadyIn GameError = errors.New("game_player_already_in")
	ErrorGamePositionTaken   GameError = errors.New("game_position_taken")
	ErrorGamePlayerNotFound  GameError = errors.New("game_player_not_found")
	ErrorGameNotReady        GameError = errors.New("game_not_ready")
)

type GameType string

const (
	GameTypeHoldem GameType = "holdem"
)

type IPlayable interface {
	Start() error
	End() error
	ProcessAction(action json.RawMessage) error
	CanStart() bool
	DealCards() error
	EvaluateHands() error
	GetGameState() interface{}
}

type GameAction struct {
	PlayerID string          `json:"playerId"`
	Action   string          `json:"action"` // fold, check, call, raise
	Data     json.RawMessage `json:"data"`
}

type Card struct {
	Suit   string `json:"suit"`
	Value  string `json:"value"`
	Hidden bool   `json:"hidden"`
}

type GamePlayer struct {
	Position   int              `json:"position"`
	Balance    int              `json:"balance"`
	LastAction string           `json:"lastAction"`
	Player     Player           `json:"player"`
	Hand       []Card           `json:"hand"`
	Status     GamePlayerStatus `json:"status"`
}

type Game struct {
	ID         string          `json:"id"`
	Status     GameStatus      `json:"status"`
	Players    []*GamePlayer   `json:"players"`
	Playable   IPlayable       `json:"playable"`
	MinBet     int             `json:"minBet"`
	MaxPlayers int             `json:"maxPlayers"`
	ActionChan chan GameAction `json:"-"`
	GameType   GameType        `json:"gameType"`
}

type GameState struct {
	Status  GameStatus    `json:"status"`
	Players []*GamePlayer `json:"players"`
	State   interface{}   `json:"state"`
}

func NewGame(actionChan chan GameAction, maxPlayers int, minBet int, gameType GameType) *Game {
	return &Game{
		ID:         uuid.New().String(),
		Status:     GameStatusWaiting,
		Players:    make([]*GamePlayer, 0),
		MaxPlayers: maxPlayers,
		MinBet:     minBet,
		ActionChan: actionChan,
		GameType:   gameType,
	}
}

func (g *Game) AddPlayer(player *GamePlayer) error {
	if len(g.Players) >= g.MaxPlayers {
		return ErrorGameFull
	}

	for _, p := range g.Players {
		if p.Player.ID == player.Player.ID {
			return ErrorGamePlayerAlreadyIn
		}

		if p.Position == player.Position {
			return ErrorGamePositionTaken
		}
	}

	player.Status = GamePlayerStatusWaiting
	g.Players = append(g.Players, player)
	return nil
}

func (g *Game) RemovePlayer(playerID string) error {
	for i, p := range g.Players {
		if p.Player.ID == playerID {
			g.Players = slices.Delete(g.Players, i, i+1)
			return nil
		}
	}
	return ErrorGamePlayerNotFound
}

func (g *Game) CanStart() bool {
	return len(g.Players) >= 2 && g.Status == GameStatusWaiting
}

func (g *Game) Start() error {
	if !g.CanStart() {
		return ErrorGameNotReady
	}

	err := g.Playable.Start()
	if err != nil {
		return err
	}

	g.Status = GameStatusStarted
	return nil
}

func (g *Game) End() error {
	err := g.Playable.End()
	if err != nil {
		return err
	}

	g.Status = GameStatusFinished
	return nil
}

func (g *Game) GetGameState() GameState {
	return GameState{
		Status:  g.Status,
		Players: g.Players,
		State:   g.Playable.GetGameState(),
	}
}

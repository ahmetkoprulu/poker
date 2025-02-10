package models

import (
	"time"

	"github.com/google/uuid"
)

// GameStatus represents the current state of the game
type GameStatus string

const (
	GameStatusWaiting  GameStatus = "waiting"
	GameStatusStarted  GameStatus = "started"
	GameStatusFinished GameStatus = "finished"
)

// Card represents a playing card
type Card struct {
	Suit   string `json:"suit"`
	Value  string `json:"value"`
	Hidden bool   `json:"hidden"`
}

// Player represents a player in the game
type Player struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Cards      []Card `json:"cards"`
	Chips      int    `json:"chips"`
	Bet        int    `json:"bet"`
	Position   int    `json:"position"`
	Active     bool   `json:"active"`
	Folded     bool   `json:"folded"`
	LastAction string `json:"lastAction"`
}

// Game represents a poker game
type Game struct {
	ID             string     `json:"id"`
	Status         GameStatus `json:"status"`
	Players        []*Player  `json:"players"`
	CommunityCards []Card     `json:"communityCards"`
	Pot            int        `json:"pot"`
	CurrentBet     int        `json:"currentBet"`
	DealerPosition int        `json:"dealerPosition"`
	CurrentTurn    int        `json:"currentTurn"`
	MinBet         int        `json:"minBet"`
	MaxPlayers     int        `json:"maxPlayers"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// GameAction represents a player's action in the game
type GameAction struct {
	PlayerID string `json:"playerId"`
	Action   string `json:"action"` // fold, check, call, raise
	Amount   int    `json:"amount"` // for raise actions
}

// NewGame creates a new poker game instance
func NewGame(maxPlayers int, minBet int) *Game {
	return &Game{
		ID:             uuid.New().String(),
		Status:         GameStatusWaiting,
		Players:        make([]*Player, 0),
		CommunityCards: make([]Card, 0),
		MaxPlayers:     maxPlayers,
		MinBet:         minBet,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// AddPlayer adds a new player to the game
func (g *Game) AddPlayer(player *Player) bool {
	if len(g.Players) >= g.MaxPlayers {
		return false
	}

	player.Position = len(g.Players)
	g.Players = append(g.Players, player)
	return true
}

// RemovePlayer removes a player from the game
func (g *Game) RemovePlayer(playerID string) bool {
	for i, p := range g.Players {
		if p.ID == playerID {
			g.Players = append(g.Players[:i], g.Players[i+1:]...)
			return true
		}
	}
	return false
}

// CanStart checks if the game can be started
func (g *Game) CanStart() bool {
	return len(g.Players) >= 2 && g.Status == GameStatusWaiting
}

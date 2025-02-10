package models

import (
	"time"
)

type RoomStatus string

const (
	RoomStatusActive   RoomStatus = "active"
	RoomStatusInactive RoomStatus = "inactive"
)

// Room represents a game room where players can join and play
type Room struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Status     RoomStatus `json:"status"`
	Game       *Game      `json:"game"`
	MaxPlayers int        `json:"maxPlayers"`
	MinBet     int        `json:"minBet"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// NewRoom creates a new room instance
func NewRoom(name string, maxPlayers int, minBet int) *Room {
	room := &Room{
		ID:         name,
		Name:       name,
		Status:     RoomStatusActive,
		MaxPlayers: maxPlayers,
		MinBet:     minBet,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Initialize with a new game
	room.Game = NewGame(maxPlayers, minBet)
	return room
}

// EnsureGameExists creates a new game if there isn't one or if the current game is finished
func (r *Room) EnsureGameExists() {
	if r.Game == nil || r.Game.Status == GameStatusFinished {
		r.Game = NewGame(r.MaxPlayers, r.MinBet)
	}
}

// StartNewGame creates and starts a new game in the room
func (r *Room) StartNewGame() *Game {
	r.Game = NewGame(r.MaxPlayers, r.MinBet)
	return r.Game
}

// IsGameActive checks if there's an active game in the room
func (r *Room) IsGameActive() bool {
	return r.Game != nil && r.Game.Status == GameStatusStarted
}

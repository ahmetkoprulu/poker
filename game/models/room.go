package models

import (
	"errors"
	"fmt"
)

var (
	ErrorRoomFull = errors.New("room is full")
)

type RoomStatus string

const (
	RoomStatusActive   RoomStatus = "active"
	RoomStatusInactive RoomStatus = "inactive"
)

type Room struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Status        RoomStatus      `json:"status"`
	Game          *Game           `json:"game"`
	MaxPlayers    int             `json:"maxPlayers"`
	MinBet        int             `json:"minBet"`
	Players       []*Player       `json:"players"`
	ActionChannel chan GameAction `json:"-"`
}

type Player struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Picture  string `json:"picture"`
	Chips    int    `json:"chips"`
}

func NewRoom(id, name string, maxPlayers, maxGamePlayers int, minBet int, gameType GameType) *Room {
	room := &Room{
		ID:            id,
		Name:          name,
		Status:        RoomStatusActive,
		MaxPlayers:    maxPlayers,
		MinBet:        minBet,
		Players:       make([]*Player, 0),
		ActionChannel: make(chan GameAction),
	}

	room.Game = NewGame(room.ActionChannel, maxGamePlayers, minBet, gameType)
	return room
}

func (r *Room) EnsureGameExists() {
	if r.Game == nil || r.Game.Status == GameStatusFinished {
		r.Game = NewGame(r.ActionChannel, r.MaxPlayers, r.MinBet, r.Game.GameType)
	}
}

func (r *Room) StartNewGame() *Game {
	r.Game = NewGame(r.ActionChannel, r.MaxPlayers, r.MinBet, r.Game.GameType)
	return r.Game
}

func (r *Room) AddPlayer(player *Player) error {
	if len(r.Players) >= r.MaxPlayers {
		return ErrorRoomFull
	}

	r.Players = append(r.Players, player)
	return nil
}

func (r *Room) RemovePlayer(playerID string) error {
	for i, p := range r.Players {
		if p.ID == playerID {
			r.Game.RemovePlayer(playerID)
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("player not found: %s", playerID)
}

func (r *Room) IsGameActive() bool {
	return r.Game != nil && r.Game.Status == GameStatusStarted
}

func (r *Room) GetRoomState() RoomState {
	gameState := r.Game.GetGameState()

	return RoomState{
		RoomID:    r.ID,
		GameState: gameState,
	}
}

type RoomState struct {
	RoomID    string    `json:"roomId"`
	GameState GameState `json:"gameState"`
}

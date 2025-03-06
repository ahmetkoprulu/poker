package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Status         RoomStatus       `json:"status"`
	Game           *Game            `json:"game"`
	MaxPlayers     int              `json:"maxPlayers"`
	MinBet         int              `json:"minBet"`
	Players        []*Client        `json:"players"`
	ActionChannel  chan GameAction  `json:"-"`
	MessageChannel chan GameMessage `json:"-"`
}

type Player struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Picture  string `json:"picture"`
	Chips    int    `json:"chips"`
}

func NewRoom(id, name string, maxPlayers int, minBet int, gameType GameType) *Room {
	room := &Room{
		ID:             id,
		Name:           name,
		Status:         RoomStatusActive,
		MaxPlayers:     maxPlayers,
		MinBet:         minBet,
		Players:        make([]*Client, 0),
		ActionChannel:  make(chan GameAction),
		MessageChannel: make(chan GameMessage, 100),
	}

	return room
}

func (r *Room) AddPlayer(player *Client) error {
	if len(r.Players) >= r.MaxPlayers {
		return ErrorRoomFull
	}

	r.Players = append(r.Players, player)
	return nil
}

func (r *Room) RemovePlayer(playerID string) error {
	for i, p := range r.Players {
		if p.User.Player.ID == playerID {
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
		RoomID:     r.ID,
		Status:     r.Status,
		Players:    r.Players,
		MaxPlayers: r.MaxPlayers,
		MinBet:     r.MinBet,
		GameState:  gameState,
	}
}

func (r *Room) SendMessage(message GameMessage) {
	r.MessageChannel <- message
}

func (r *Room) SendMessageToRoom(message GameMessage) {
	for _, p := range r.Players {
		msg, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshalling message: %v", err)
			continue
		}

		p.send <- msg
	}
}

type RoomState struct {
	RoomID     string     `json:"roomId"`
	Status     RoomStatus `json:"status"`
	MaxPlayers int        `json:"maxPlayers"`
	MinBet     int        `json:"minBet"`
	Players    []*Client  `json:"players"`
	GameState  GameState  `json:"gameState"`
}

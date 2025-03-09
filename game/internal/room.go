package internal

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
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
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	Status         RoomStatus         `json:"status"`
	Game           *Game              `json:"game"`
	MaxPlayers     int                `json:"max_players"`
	MinBet         int                `json:"min_bet"`
	Players        map[string]*Client `json:"players"`
	ActionChannel  chan GameAction    `json:"-"`
	MessageChannel chan GameMessage   `json:"-"`
	mu             sync.Mutex         `json:"-"`
}

func NewRoom(id, name string, maxPlayers int, minBet int, gameType GameType) *Room {
	room := &Room{
		ID:             id,
		Name:           name,
		Status:         RoomStatusActive,
		MaxPlayers:     maxPlayers,
		MinBet:         minBet,
		Players:        make(map[string]*Client),
		ActionChannel:  make(chan GameAction),
		MessageChannel: make(chan GameMessage, 100),
		mu:             sync.Mutex{},
	}

	return room
}

func (r *Room) AddPlayer(player *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= r.MaxPlayers {
		return ErrorRoomFull
	}

	r.Players[player.User.Player.ID] = player
	return nil
}

func (r *Room) RemovePlayer(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Game.RemovePlayer(playerID)
	delete(r.Players, playerID)
	return nil
}

func (r *Room) IsGameActive() bool {
	return r.Game != nil && r.Game.Status == GameStatusStarted
}

func (r *Room) GetRoomState() RoomState {
	gameState := r.Game.GetGameState()

	r.mu.Lock()
	defer r.mu.Unlock()

	players := make([]*Client, 0, len(r.Players))
	for _, player := range r.Players {
		players = append(players, player)
	}

	return RoomState{
		RoomID:     r.ID,
		Status:     r.Status,
		Players:    players,
		MaxPlayers: r.MaxPlayers,
		MinBet:     r.MinBet,
		GameState:  gameState,
	}
}

func (r *Room) SendMessageToPlayer(playerID string, message GameMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()

	client, ok := r.Players[playerID]
	if !ok {
		return
	}

	msg, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	client.send <- msg
}

func (r *Room) SendMessageToRoom(message GameMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()

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
	RoomID     string     `json:"room_id"`
	Status     RoomStatus `json:"status"`
	MaxPlayers int        `json:"max_players"`
	MinBet     int        `json:"min_bet"`
	Players    []*Client  `json:"players"`
	GameState  GameState  `json:"game_state"`
}

package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/ahmetkoprulu/rtrp/game/models"
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
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	Status         RoomStatus           `json:"status"`
	Game           *Game                `json:"game"`
	MaxPlayers     int                  `json:"max_players"`
	MinBet         int                  `json:"min_bet"`
	Players        map[string]*Client   `json:"players"`
	ActionChannel  chan GameAction      `json:"-"`
	MessageChannel chan models.Response `json:"-"`
	mu             sync.Mutex           `json:"-"`
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
		MessageChannel: make(chan models.Response, 100),
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

	// if player.CurrentRoom.ID == r.ID {
	// 	return nil
	// }

	r.Players[player.User.Player.ID] = player
	player.CurrentRoom = r

	return nil
}

func (r *Room) RemovePlayer(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Game != nil {
		r.Game.RemovePlayer(playerID)
	}

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

	return RoomState{
		RoomID:     r.ID,
		Status:     r.Status,
		Players:    r.GetPlayersState(),
		MaxPlayers: r.MaxPlayers,
		MinBet:     r.MinBet,
		GameType:   r.Game.GameType,
		GameStatus: r.Game.Status,
		GameState:  gameState,
	}
}

func (r *Room) GetPlayersState() []*models.Player {
	players := make([]*models.Player, 0, len(r.Players))
	for _, player := range r.Players {
		players = append(players, player.User.Player)
	}

	return players
}

func (r *Room) GetRoomSummary() *RoomSummary {
	return &RoomSummary{
		Id:             r.ID,
		Status:         r.Status,
		MaxRoomPlayers: r.MaxPlayers,
		PlayersInRoom:  len(r.Players),
		GameStatus:     r.Game.Status,
		GameType:       r.Game.GameType,
		MinBet:         r.MinBet,
		MaxGamePlayers: r.Game.MaxPlayers,
		PlayersInGame:  len(r.Game.Players),
	}
}

func (r *Room) BroadcastToPlayer(playerID string, response models.Response) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	client, ok := r.Players[playerID]
	if !ok {
		return fmt.Errorf("player not found")
	}

	msg, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshalling message: %v", err)
	}

	client.send <- msg
	return nil
}

func (r *Room) BroadcastToOthers(playerID string, response models.Response) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	msg, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshalling message: %v", err)
	}

	for _, p := range r.Players {
		if p.User.Player.ID != playerID {
			p.send <- msg
		}
	}

	return nil
}

func (r *Room) BroadcastToRoom(response models.Response) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	msg, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshalling message: %v", err)
	}

	for _, p := range r.Players {
		p.send <- msg
	}

	return nil
}

// Reset disconnects all clients and resets room/game state
func (r *Room) Reset() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Printf("[ADMIN] Resetting room %s\n", r.ID)

	// Disconnect all clients
	for playerID, client := range r.Players {
		fmt.Printf("[ADMIN] Disconnecting client %s\n", playerID)

		// Close the WebSocket connection
		if client.Conn != nil {
			client.Conn.Close()
		}

		// Close the send channel to stop the client goroutines
		close(client.send)

		// Clear client's room reference
		client.CurrentRoom = nil
		client.CurrentGame = nil
	}

	// Clear all players from room
	r.Players = make(map[string]*Client)

	// Reset game if it exists
	if r.Game != nil {
		err := r.Game.Reset()
		if err != nil {
			fmt.Printf("[ADMIN] Error resetting game: %v\n", err)
		}
	}

	fmt.Printf("[ADMIN] Room %s reset complete\n", r.ID)
	return nil
}

type RoomState struct {
	RoomID         string           `json:"room_id"`
	Status         RoomStatus       `json:"status"`
	MaxPlayers     int              `json:"max_players"`
	MaxGamePlayers int              `json:"max_game_players"`
	Players        []*models.Player `json:"players"`
	MinBet         int              `json:"min_bet"`
	GameType       GameType         `json:"game_type"`
	GameStatus     GameStatus       `json:"game_status"`
	GameState      any              `json:"game_state"`
}

type RoomSummary struct {
	Id             string     `json:"id"`
	Status         RoomStatus `json:"status"`
	MaxRoomPlayers int        `json:"max_room_players"`
	PlayersInRoom  int        `json:"players_in_room"`
	GameStatus     GameStatus `json:"game_status"`
	GameType       GameType   `json:"game_type"`
	MinBet         int        `json:"min_bet"`
	MaxGamePlayers int        `json:"max_game_players"`
	PlayersInGame  int        `json:"players_in_game"`
}

type RoomJoinOkResponse struct {
	RoomID string         `json:"room_id"`
	Player *models.Player `json:"player"`
}

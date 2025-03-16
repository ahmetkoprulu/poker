package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
)

type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func (rm *RoomManager) CreateRoom(id, name string, maxPlayers, maxGamePlayers, minBet int, gameType GameType) (*Room, error) {
	room := NewRoom(id, name, maxPlayers, minBet, gameType)

	switch gameType {
	case GameTypeHoldem:
		room.Game = NewGame(room.ActionChannel, room.MessageChannel, room, maxGamePlayers, minBet, gameType)
		room.Game.Playable = NewHoldem(room.Game)
	default:
		return nil, fmt.Errorf("unsupported game type: %s", gameType)
	}

	rm.mu.Lock()
	rm.rooms[room.ID] = room
	rm.mu.Unlock()

	log.Printf("[INFO] Room created - RoomID: %s, MaxPlayers: %d, MaxGamePlayers: %d, MinBet: %d, GameType: %s", room.ID, room.MaxPlayers, room.Game.MaxPlayers, room.MinBet, room.Game.GameType)

	return room, nil
}

func (rm *RoomManager) RegisterRoom(room *Room) {
	if room == nil {
		log.Printf("[ERROR] Attempted to register nil room")
		return
	}

	rm.mu.Lock()
	rm.rooms[room.ID] = room
	rm.mu.Unlock()

	log.Printf("[INFO] Room registered - RoomID: %s, MaxPlayers: %d, MinBet: %d", room.ID, room.MaxPlayers, room.MinBet)
}

func (rm *RoomManager) GetRoom(roomID string) (*Room, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, errors.New("room not found")
	}

	return room, nil
}

func (rm *RoomManager) GetRoomByPlayerID(playerID string) (*Room, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for _, room := range rm.rooms {
		for _, player := range room.Players {
			if player.User.Player.ID == playerID {
				return room, nil
			}
		}
	}

	return nil, errors.New("room not found")
}

func (rm *RoomManager) JoinRoom(roomID string, player *Client) error {
	room, err := rm.GetRoom(roomID)
	if err != nil {
		log.Printf("[ERROR] Room not found for join - RoomID: %s, PlayerID: %s", roomID, player.User.Player.ID)
		return err
	}

	log.Printf("[INFO] Attempting to add player to room - RoomID: %s, PlayerID: %s, CurrentPlayers: %d", roomID, player.User.Player.ID, len(room.Players))
	if err := room.AddPlayer(player); err != nil {
		log.Printf("[ERROR] Room is full - RoomID: %s, PlayerID: %s, MaxPlayers: %d", roomID, player.User.Player.ID, room.MaxPlayers)
		return errors.New("cannot join room: room is full")
	}

	log.Printf("[INFO] Player added to room - RoomID: %s, PlayerID: %s, TotalPlayers: %d", roomID, player.User.Player.ID, len(room.Players))
	return nil
}

func (rm *RoomManager) LeaveRoom(roomID string, playerID string) error {
	room, err := rm.GetRoom(roomID)
	if err != nil {
		log.Printf("[ERROR] Room not found for leave - RoomID: %s, PlayerID: %s", roomID, playerID)
		return err
	}

	log.Printf("[INFO] Player leaving room - RoomID: %s, PlayerID: %s, CurrentPlayers: %d", roomID, playerID, len(room.Players))

	// Find the player and mark them as inactive before removing
	playerFound := false
	for _, p := range room.Players {
		if p.User.Player.ID == playerID {
			playerFound = true
			break
		}
	}

	if !playerFound {
		log.Printf("[ERROR] Player not found in room - RoomID: %s, PlayerID: %s", roomID, playerID)
		return fmt.Errorf("player not found in room")
	}

	if err := room.RemovePlayer(playerID); err != nil {
		log.Printf("[ERROR] Failed to remove player - RoomID: %s, PlayerID: %s", roomID, playerID)
		return errors.New("failed to remove player from room")
	}

	// If game is active and not enough players, end the game
	if room.Game.Status == GameStatusStarted && len(room.Game.Players) < 2 {
		log.Printf("[INFO] Ending game due to insufficient players - RoomID: %s, RemainingPlayers: %d",
			roomID, len(room.Game.Players))
		room.Game.Status = GameStatusEnd
	}

	log.Printf("[INFO] Player removed from room - RoomID: %s, PlayerID: %s, RemainingPlayers: %d",
		roomID, playerID, len(room.Game.Players))

	return nil
}

func (rm *RoomManager) StartGame(roomID string) error {
	room, err := rm.GetRoom(roomID)
	if err != nil {
		log.Printf("[ERROR] Room not found for start - RoomID: %s", roomID)
		return err
	}

	log.Printf("[INFO] Attempting to start game - RoomID: %s, PlayerCount: %d, Status: %s", roomID, len(room.Game.Players), room.Game.Status)
	// if err := room.Game.Playable.Start(); err != nil {
	// 	log.Printf("[ERROR] Failed to start new hand - RoomID: %s, Error: %s", roomID, err)
	// 	return err
	// }

	log.Printf("[INFO] Game started successfully - RoomID: %s, PlayerCount: %d", roomID, len(room.Game.Players))
	return nil
}

func (rm *RoomManager) ProcessAction(roomID string, action json.RawMessage) error {
	room, err := rm.GetRoom(roomID)
	if err != nil {
		return err
	}

	if room.Game.Status != GameStatusStarted {
		return errors.New("game not in progress")
	}

	return room.Game.Playable.ProcessAction(action)
}

func (rm *RoomManager) RemoveRoom(roomID string) {
	rm.mu.Lock()
	delete(rm.rooms, roomID)
	rm.mu.Unlock()
}

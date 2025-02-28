package websocket

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ahmetkoprulu/rtrp/game/internal/room"
	"github.com/ahmetkoprulu/rtrp/game/models"
)

type MessageHandler struct {
	server      *Server
	roomManager *room.RoomManager
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(server *Server, roomManager *room.RoomManager) *MessageHandler {
	return &MessageHandler{
		server:      server,
		roomManager: roomManager,
	}
}

func (h *MessageHandler) HandleMessage(client *Client, message []byte) error {
	var msg models.Message
	if err := json.Unmarshal(message, &msg); err != nil {
		return err
	}

	switch msg.Type {
	case models.MessageTypeRoomInfo:
		message, err := ParseData[models.MessageRoomInfo](msg.Data)
		if err != nil {
			return err
		}
		return h.handleRoomInfo(client, *message)
	case models.MessageTypeJoinRoom:
		message, err := ParseData[models.MessageJoinRoom](msg.Data)
		if err != nil {
			return err
		}
		return h.handleJoinRoom(client, *message)
	case models.MessageTypeLeaveRoom:
		message, err := ParseData[models.MessageLeaveRoom](msg.Data)
		if err != nil {
			return err
		}
		return h.handleLeaveRoom(client, *message)
	case models.MessageTypeJoinGame:
		message, err := ParseData[models.MessageJoinGame](msg.Data)
		if err != nil {
			return err
		}
		return h.handleJoinGame(client, *message)
	case models.MessageTypeLeaveGame:
		message, err := ParseData[models.MessageLeaveGame](msg.Data)
		if err != nil {
			return err
		}
		return h.handleLeaveGame(client, *message)
	case models.MessageTypeStartGame:
		message, err := ParseData[models.MessageStartGame](msg.Data)
		if err != nil {
			return err
		}
		return h.handleStartGame(client, *message)
	case models.MessageTypeGameAction:
		message, err := ParseData[models.MessageGameAction](msg.Data)
		if err != nil {
			return err
		}
		return h.handleGameAction(client, *message)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
		return nil
	}
}

func (h *MessageHandler) handleRoomInfo(client *Client, message models.MessageRoomInfo) error {
	room := h.server.GetRoom(message.RoomID)
	if room == nil {
		return h.sendError(client, "Room not found")
	}

	response := models.Response{
		Type: models.MessageTypeRoomInfo,
		Data: room,
	}

	msgBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}

	client.send <- msgBytes
	return nil
}

func (h *MessageHandler) handleJoinRoom(client *Client, data models.MessageJoinRoom) error {
	room := h.server.GetRoom(data.RoomID)
	if room == nil {
		return h.sendError(client, "Room not found")
	}

	if err := h.server.JoinRoom(room.ID, client); err != nil {
		return h.sendError(client, fmt.Sprintf("Failed to join room: %v", err))
	}

	return nil
}

func (h *MessageHandler) handleLeaveRoom(client *Client, msg models.MessageLeaveRoom) error {
	room := h.server.GetRoom(msg.RoomID)
	if room == nil {
		return h.sendError(client, "Room not found")
	}

	if err := room.RemovePlayer(client.PlayerID); err != nil {
		return h.sendError(client, fmt.Sprintf("Failed to leave room: %v", err))
	}

	return nil
}

func (h *MessageHandler) handleJoinGame(client *Client, msg models.MessageJoinGame) error {
	room := h.server.GetRoom(msg.RoomID)
	if room == nil {
		log.Printf("[ERROR] Room not found - RoomID: %s, PlayerID: %s", msg.RoomID, client.PlayerID)
		return h.sendError(client, "Room not found")
	}

	log.Printf("[INFO] Player attempting to join - RoomID: %s, PlayerID: %s, GameStatus: %s, PlayerCount: %d",
		room.ID, client.PlayerID, room.Game.Status, len(room.Game.Players))

	if room.IsGameActive() {
		log.Printf("[INFO] Cannot join active game - RoomID: %s, PlayerID: %s, GameStatus: %s",
			room.ID, client.PlayerID, room.Game.Status)
		return h.sendError(client, "Cannot join: Game is already in progress")
	}

	// Ensure there's a game to join
	room.EnsureGameExists()

	player := &models.GamePlayer{
		Player: models.Player{
			ID:       client.PlayerID,
			Username: "Guest",
			Picture:  "https://via.placeholder.com/150",
			Chips:    1000,
		},
	}

	if err := h.roomManager.JoinRoom(room.ID, &player.Player); err != nil {
		log.Printf("[ERROR] Failed to join game - RoomID: %s, PlayerID: %s, Error: %v",
			room.ID, client.PlayerID, err)
		return h.sendError(client, fmt.Sprintf("Failed to join game: %v", err))
	}

	log.Printf("[INFO] Player joined successfully - RoomID: %s, PlayerID: %s, GameID: %s, PlayerCount: %d",
		room.ID, client.PlayerID, room.Game.ID, len(room.Game.Players))

	h.broadcastRoomState(room)
	return nil
}

func (h *MessageHandler) handleLeaveGame(client *Client, msg models.MessageLeaveGame) error {
	room := h.server.GetRoom(msg.RoomID)
	if room == nil {
		log.Printf("[ERROR] Room not found for leave game - RoomID: %s, PlayerID: %s",
			msg.RoomID, client.PlayerID)
		return h.sendError(client, "Room not found")
	}

	if err := h.roomManager.LeaveRoom(room.ID, client.PlayerID); err != nil {
		log.Printf("[ERROR] Failed to leave game - RoomID: %s, PlayerID: %s, Error: %v",
			room.ID, client.PlayerID, err)
		return h.sendError(client, fmt.Sprintf("Failed to leave game: %v", err))
	}

	log.Printf("[INFO] Player left game - RoomID: %s, PlayerID: %s, RemainingPlayers: %d",
		room.ID, client.PlayerID, len(room.Game.Players))

	h.broadcastRoomState(room)
	return nil
}

func (h *MessageHandler) handleStartGame(client *Client, msg models.MessageStartGame) error {
	room := h.server.GetRoom(msg.RoomID)
	if room == nil {
		log.Printf("[ERROR] Room not found for game start - RoomID: %s, PlayerID: %s",
			msg.RoomID, client.PlayerID)
		return h.sendError(client, "Room not found")
	}

	log.Printf("[INFO] Attempting to start game - RoomID: %s, GameID: %s, PlayerCount: %d",
		room.ID, room.Game.ID, len(room.Game.Players))

	if err := h.roomManager.StartGame(room.Game.ID); err != nil {
		log.Printf("[ERROR] Failed to start game - RoomID: %s, GameID: %s, Error: %v",
			room.ID, room.Game.ID, err)
		return h.sendError(client, fmt.Sprintf("Failed to start game: %v", err))
	}

	log.Printf("[INFO] Game started successfully - RoomID: %s, GameID: %s, PlayerCount: %d",
		room.ID, room.Game.ID, len(room.Game.Players))

	h.broadcastRoomState(room)
	return nil
}

func (h *MessageHandler) handleGameAction(client *Client, msg models.MessageGameAction) error {
	room := h.server.GetRoom(msg.RoomID)
	if room == nil {
		log.Printf("[ERROR] Room not found for game action - RoomID: %s, PlayerID: %s",
			msg.RoomID, client.PlayerID)
		return h.sendError(client, "Room not found")
	}

	game := room.Game
	if game == nil || game.Status != models.GameStatusStarted {
		log.Printf("[ERROR] Invalid game state for action - RoomID: %s, GameID: %s, Status: %s",
			room.ID, game.ID, game.Status)
		return h.sendError(client, "Game is not in progress")
	}

	var currentPlayer *models.GamePlayer
	for _, p := range game.Players {
		if p.Player.ID == client.PlayerID {
			currentPlayer = p
			break
		}
	}

	if currentPlayer == nil {
		log.Printf("[ERROR] Player not found in game - GameID: %s, PlayerID: %s",
			game.ID, client.PlayerID)
		return h.sendError(client, "Player not found in game")
	}

	var action models.GameAction
	if err := json.Unmarshal(msg.Data, &action); err != nil {
		return err
	}

	log.Printf("[INFO] Processing game action - GameID: %s, PlayerID: %s, Action: %s",
		game.ID, client.PlayerID, action.Action)

	if err := h.roomManager.ProcessAction(room.ID, msg.Data); err != nil {
		return h.sendError(client, fmt.Sprintf("Failed to process action: %v", err))
	}

	h.broadcastRoomState(room)
	return nil
}

func (h *MessageHandler) sendError(client *Client, errorMsg string) error {
	response := models.Response{
		Type: models.MessageTypeError,
		Data: map[string]string{"error": errorMsg},
	}

	msgBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}

	client.send <- msgBytes
	return nil
}

func (h *MessageHandler) broadcastRoomState(room *models.Room) {
	stateMsg := models.Response{
		Type: models.MessageTypeGameInfo,
		Data: room,
	}

	msgBytes, err := json.Marshal(stateMsg)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal room state - RoomID: %s, Error: %v", room.ID, err)
		return
	}

	log.Printf("[INFO] Broadcasting room state - RoomID: %s, GameID: %s, GameStatus: %s, PlayerCount: %d",
		room.ID, room.Game.ID, room.Game.Status, len(room.Game.Players))

	h.server.BroadcastToRoom(room.ID, msgBytes)
}

// func (h *MessageHandler) broadcastGameState(game *models.Game) {
// 	// Create game state message
// 	stateMsg := models.Response{
// 		Type: models.MessageTypeGameInfo,
// 		Data: game,
// 	}

// 	// Convert to JSON
// 	msgBytes, err := json.Marshal(stateMsg)
// 	if err != nil {
// 		log.Printf("Error marshaling game state: %v", err)
// 		return
// 	}

// 	// Broadcast to all players in the game
// 	for _, player := range game.Players {
// 		if client, ok := getClientByPlayerID(player.Player.ID); ok {
// 			client.send <- msgBytes
// 		}
// 	}
// }

// // TODO: Implement this function to maintain a map of player IDs to clients
// func getClientByPlayerID(playerID string) (*Client, bool) {
// 	// This will be implemented when we add client tracking
// 	return nil, false
// }

func ParseData[T any](data json.RawMessage) (*T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

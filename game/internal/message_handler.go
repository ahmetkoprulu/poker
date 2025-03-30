package internal

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ahmetkoprulu/rtrp/game/models"
)

type MessageHandler struct {
	server      *Server
	roomManager *RoomManager
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(server *Server, roomManager *RoomManager) *MessageHandler {
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
		Data: room.GetRoomState(),
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

	response := models.Response{
		Type: models.MessageTypeJoinRoomOk,
		Data: room.GetRoomState(),
	}

	if err := room.BroadcastToPlayer(client.User.Player.ID, response); err != nil {
		return err
	}

	response = models.Response{
		Type: models.MessageTypeJoinRoom,
		Data: client.User.Player,
	}

	if err := room.BroadcastToOthers(client.User.Player.ID, response); err != nil {
		return err
	}

	return nil
}

func (h *MessageHandler) handleLeaveRoom(client *Client, msg models.MessageLeaveRoom) error {
	room := h.server.GetRoom(msg.RoomID)
	if room == nil {
		return h.sendError(client, "Room not found")
	}

	if err := room.RemovePlayer(client.User.Player.ID); err != nil {
		return h.sendError(client, fmt.Sprintf("Failed to leave room: %v", err))
	}

	return nil
}

func (h *MessageHandler) handleJoinGame(client *Client, msg models.MessageJoinGame) error {
	room := h.server.GetRoom(msg.RoomID)
	if room == nil {
		log.Printf("[ERROR] Room not found - RoomID: %s, PlayerID: %s", msg.RoomID, client.User.Player.ID)
		return h.sendError(client, "Room not found")
	}

	if err := h.roomManager.JoinRoom(room.ID, client); err != nil {
		log.Printf("[ERROR] Failed to join room - RoomID: %s, PlayerID: %s, Error: %v", room.ID, client.User.Player.ID, err)
		return h.sendError(client, fmt.Sprintf("Failed to join game: %v", err))
	}

	if err := room.Game.AddPlayer(msg.Position, client); err != nil {
		log.Printf("[ERROR] Failed to add player to game - RoomID: %s, PlayerID: %s, Error: %v", room.ID, client.User.Player.ID, err)
		return h.sendError(client, fmt.Sprintf("Failed to join game: %v", err))
	}

	log.Printf("[INFO] Player joined successfully - RoomID: %s, PlayerID: %s, GameID: %s, PlayerCount: %d", room.ID, client.User.Player.ID, room.Game.ID, len(room.Game.Players))

	h.broadcastRoomState(room)
	return nil
}

func (h *MessageHandler) handleLeaveGame(client *Client, msg models.MessageLeaveGame) error {
	room := h.server.GetRoom(msg.RoomID)
	if room == nil {
		log.Printf("[ERROR] Room not found for leave game - RoomID: %s, PlayerID: %s", msg.RoomID, client.User.Player.ID)
		return h.sendError(client, "Room not found")
	}

	if err := h.roomManager.LeaveRoom(room.ID, client.User.Player.ID); err != nil {
		log.Printf("[ERROR] Failed to leave game - RoomID: %s, PlayerID: %s, Error: %v", room.ID, client.User.Player.ID, err)
		return h.sendError(client, fmt.Sprintf("Failed to leave game: %v", err))
	}

	log.Printf("[INFO] Player left game - RoomID: %s, PlayerID: %s, RemainingPlayers: %d", room.ID, client.User.Player.ID, len(room.Game.Players))

	h.broadcastRoomState(room)
	return nil
}

func (h *MessageHandler) handleGameAction(client *Client, msg models.MessageGameAction) error {
	room := h.server.GetRoom(msg.RoomID)
	if room == nil {
		log.Printf("[ERROR] Room not found for game action - RoomID: %s, PlayerID: %s", msg.RoomID, client.User.Player.ID)
		return h.sendError(client, "Room not found")
	}

	game := room.Game
	if game == nil || game.Status != GameStatusStarted {
		log.Printf("[ERROR] Invalid game state for action - RoomID: %s, GameID: %s, Status: %s", room.ID, game.ID, game.Status)
		return h.sendError(client, "Game is not in progress")
	}

	var currentPlayer *GamePlayer
	for _, p := range game.Players {
		if p.Client.User.Player.ID == client.User.Player.ID {
			currentPlayer = p
			break
		}
	}

	if currentPlayer == nil {
		log.Printf("[ERROR] Player not found in game - GameID: %s, PlayerID: %s", game.ID, client.User.Player.ID)
		return h.sendError(client, "Player not found in game")
	}

	var action GameAction
	if err := json.Unmarshal(msg.Data, &action); err != nil {
		return err
	}

	log.Printf("[INFO] Processing game action - GameID: %s, PlayerID: %s, Action: %s", game.ID, client.User.Player.ID, action.ActionType)

	if err := h.roomManager.ProcessAction(room.ID, msg.Data); err != nil {
		return h.sendError(client, fmt.Sprintf("Failed to process action: %v", err))
	}

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

// func (h *MessageHandler) broadcastToClient(client *Client, response models.Response) error {
// 	msgBytes, err := json.Marshal(response)
// 	if err != nil {
// 		return err
// 	}

// 	client.send <- msgBytes
// 	return nil
// }

func (h *MessageHandler) broadcastRoomState(room *Room) {
	stateMsg := models.Response{
		Type: models.MessageTypeRoomInfo,
		Data: room.GetRoomState(),
	}

	msgBytes, err := json.Marshal(stateMsg)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal room state - RoomID: %s, Error: %v", room.ID, err)
		return
	}

	log.Printf("[INFO] Broadcasting room state - RoomID: %s, GameID: %s, GameStatus: %s, PlayerCount: %d", room.ID, room.Game.ID, room.Game.Status, len(room.Game.Players))

	h.server.BroadcastToRoom(room.ID, msgBytes)
}

func ParseData[T any](data json.RawMessage) (*T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

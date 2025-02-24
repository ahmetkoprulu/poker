package websocket

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ahmetkoprulu/rtrp/game/internal/game"
	"github.com/ahmetkoprulu/rtrp/game/models"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeJoinGame   MessageType = "join_game"
	MessageTypeLeaveGame  MessageType = "leave_game"
	MessageTypeStartGame  MessageType = "start_game"
	MessageTypeGameAction MessageType = "game_action"
	MessageTypeRoomInfo   MessageType = "room_info"
	MessageTypeGameState  MessageType = "game_state"
	MessageTypePlayerList MessageType = "player_list"
	MessageTypeError      MessageType = "error"
)

// Message represents a WebSocket message
type Message struct {
	Type     MessageType `json:"type"`
	RoomID   string      `json:"roomId"`
	GameID   string      `json:"gameId"`
	PlayerID string      `json:"playerId"`
	Data     interface{} `json:"data"`
}

// MessageHandler processes WebSocket messages
type MessageHandler struct {
	gameManager *game.GameManager
	server      *Server
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(gameManager *game.GameManager) *MessageHandler {
	return &MessageHandler{
		gameManager: gameManager,
	}
}

// SetServer sets the WebSocket server reference
func (h *MessageHandler) SetServer(server *Server) {
	h.server = server
}

func (h *MessageHandler) HandleMessage(client *Client, message []byte) error {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		return err
	}

	switch msg.Type {
	case MessageTypeRoomInfo:
		return h.handleRoomInfo(client)
	case MessageTypeJoinGame:
		return h.handleJoinGame(client, msg)
	case MessageTypeLeaveGame:
		return h.handleLeaveGame(client, msg)
	case MessageTypeStartGame:
		return h.handleStartGame(client, msg)
	case MessageTypeGameAction:
		return h.handleGameAction(client, msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
		return nil
	}
}

func (h *MessageHandler) handleRoomInfo(client *Client) error {
	room := h.server.GetRoom(client.RoomID)
	if room == nil {
		return h.sendError(client, "Room not found")
	}

	response := Message{
		Type:   MessageTypeRoomInfo,
		RoomID: room.ID,
		Data:   room,
	}

	msgBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}

	client.send <- msgBytes
	return nil
}

func (h *MessageHandler) handleJoinGame(client *Client, msg Message) error {
	room := h.server.GetRoom(client.RoomID)
	if room == nil {
		log.Printf("[ERROR] Room not found - RoomID: %s, PlayerID: %s", client.RoomID, client.PlayerID)
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

	player := &models.Player{
		ID:     client.PlayerID,
		Active: true,
		Chips:  1000, // Set initial chips
	}

	if err := h.gameManager.JoinGame(room.Game.ID, player); err != nil {
		log.Printf("[ERROR] Failed to join game - RoomID: %s, PlayerID: %s, Error: %v",
			room.ID, client.PlayerID, err)
		return h.sendError(client, fmt.Sprintf("Failed to join game: %v", err))
	}

	log.Printf("[INFO] Player joined successfully - RoomID: %s, PlayerID: %s, GameID: %s, PlayerCount: %d",
		room.ID, client.PlayerID, room.Game.ID, len(room.Game.Players))

	h.broadcastRoomState(room)
	return nil
}

func (h *MessageHandler) handleLeaveGame(client *Client, msg Message) error {
	room := h.server.GetRoom(client.RoomID)
	if room == nil {
		log.Printf("[ERROR] Room not found for leave game - RoomID: %s, PlayerID: %s",
			client.RoomID, client.PlayerID)
		return h.sendError(client, "Room not found")
	}

	if err := h.gameManager.LeaveGame(room.Game.ID, client.PlayerID); err != nil {
		log.Printf("[ERROR] Failed to leave game - RoomID: %s, PlayerID: %s, Error: %v",
			room.ID, client.PlayerID, err)
		return h.sendError(client, fmt.Sprintf("Failed to leave game: %v", err))
	}

	log.Printf("[INFO] Player left game - RoomID: %s, PlayerID: %s, RemainingPlayers: %d",
		room.ID, client.PlayerID, len(room.Game.Players))

	h.broadcastRoomState(room)
	return nil
}

func (h *MessageHandler) handleStartGame(client *Client, msg Message) error {
	room := h.server.GetRoom(client.RoomID)
	if room == nil {
		log.Printf("[ERROR] Room not found for game start - RoomID: %s, PlayerID: %s",
			client.RoomID, client.PlayerID)
		return h.sendError(client, "Room not found")
	}

	log.Printf("[INFO] Attempting to start game - RoomID: %s, GameID: %s, PlayerCount: %d",
		room.ID, room.Game.ID, len(room.Game.Players))

	if err := h.gameManager.StartGame(room.Game.ID); err != nil {
		log.Printf("[ERROR] Failed to start game - RoomID: %s, GameID: %s, Error: %v",
			room.ID, room.Game.ID, err)
		return h.sendError(client, fmt.Sprintf("Failed to start game: %v", err))
	}

	log.Printf("[INFO] Game started successfully - RoomID: %s, GameID: %s, PlayerCount: %d",
		room.ID, room.Game.ID, len(room.Game.Players))

	h.broadcastRoomState(room)
	return nil
}

func (h *MessageHandler) handleGameAction(client *Client, msg Message) error {
	room := h.server.GetRoom(client.RoomID)
	if room == nil {
		log.Printf("[ERROR] Room not found for game action - RoomID: %s, PlayerID: %s",
			client.RoomID, client.PlayerID)
		return h.sendError(client, "Room not found")
	}

	game := room.Game
	if game == nil || game.Status != models.GameStatusStarted {
		log.Printf("[ERROR] Invalid game state for action - RoomID: %s, GameID: %s, Status: %s",
			room.ID, game.ID, game.Status)
		return h.sendError(client, "Game is not in progress")
	}

	// Find the current player
	var currentPlayer *models.Player
	for _, p := range game.Players {
		if p.ID == client.PlayerID {
			currentPlayer = p
			break
		}
	}

	if currentPlayer == nil {
		log.Printf("[ERROR] Player not found in game - GameID: %s, PlayerID: %s",
			game.ID, client.PlayerID)
		return h.sendError(client, "Player not found in game")
	}

	// Verify it's the player's turn
	if game.CurrentTurn != currentPlayer.Position {
		log.Printf("[ERROR] Not player's turn - GameID: %s, PlayerID: %s, CurrentTurn: %d, PlayerPosition: %d",
			game.ID, client.PlayerID, game.CurrentTurn, currentPlayer.Position)
		return h.sendError(client, "Not your turn")
	}

	// Parse the action
	var action models.GameAction
	actionData, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(actionData, &action); err != nil {
		return err
	}

	log.Printf("[INFO] Processing game action - GameID: %s, PlayerID: %s, Action: %s",
		game.ID, client.PlayerID, action.Action)

	// Process the action
	switch action.Action {
	case "check":
		if game.CurrentBet > currentPlayer.Bet {
			return h.sendError(client, "Cannot check when there's a bet to call")
		}
		currentPlayer.LastAction = "check"

	case "call":
		if game.CurrentBet <= currentPlayer.Bet {
			return h.sendError(client, "No bet to call")
		}
		amountToCall := game.CurrentBet - currentPlayer.Bet
		currentPlayer.Chips -= amountToCall
		currentPlayer.Bet = game.CurrentBet
		game.Pot += amountToCall
		currentPlayer.LastAction = "call"

	default:
		return h.sendError(client, "Invalid action")
	}

	// Move to next player
	nextPosition := (game.CurrentTurn + 1) % len(game.Players)
	game.CurrentTurn = nextPosition

	log.Printf("[INFO] Action processed - GameID: %s, PlayerID: %s, NextTurn: %d",
		game.ID, client.PlayerID, nextPosition)

	// Broadcast updated game state
	h.broadcastRoomState(room)
	return nil
}

func (h *MessageHandler) sendError(client *Client, errorMsg string) error {
	response := Message{
		Type: MessageTypeError,
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
	stateMsg := Message{
		Type:   MessageTypeGameState,
		RoomID: room.ID,
		GameID: room.Game.ID,
		Data:   room,
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

func (h *MessageHandler) broadcastGameState(game *models.Game) {
	// Create game state message
	stateMsg := Message{
		Type:   "game_state",
		GameID: game.ID,
		Data:   game,
	}

	// Convert to JSON
	msgBytes, err := json.Marshal(stateMsg)
	if err != nil {
		log.Printf("Error marshaling game state: %v", err)
		return
	}

	// Broadcast to all players in the game
	for _, player := range game.Players {
		if client, ok := getClientByPlayerID(player.ID); ok {
			client.send <- msgBytes
		}
	}
}

// TODO: Implement this function to maintain a map of player IDs to clients
func getClientByPlayerID(playerID string) (*Client, bool) {
	// This will be implemented when we add client tracking
	return nil, false
}

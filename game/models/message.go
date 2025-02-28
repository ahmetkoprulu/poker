package models

import "encoding/json"

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeRoomInfo   MessageType = "room_info"
	MessageTypeJoinRoom   MessageType = "room_join"
	MessageTypeLeaveRoom  MessageType = "room_leave"
	MessageTypeJoinGame   MessageType = "game_join"
	MessageTypeLeaveGame  MessageType = "game_leave"
	MessageTypeStartGame  MessageType = "game_start"
	MessageTypeGameAction MessageType = "game_action"
	MessageTypeGameInfo   MessageType = "game_info"
	MessageTypePlayerList MessageType = "player_list"
	MessageTypeError      MessageType = "error"
)

// Message represents a WebSocket message
type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Response struct {
	Type MessageType `json:"type"`
	Data interface{} `json:"data"`
}

type MessageRoomInfo struct {
	RoomID string `json:"roomId"`
}

type MessageJoinRoom struct {
	RoomID   string `json:"roomId"`
	PlayerID string `json:"playerId"`
}

type MessageLeaveRoom struct {
	RoomID   string `json:"roomId"`
	PlayerID string `json:"playerId"`
}

type MessageJoinGame struct {
	RoomID   string `json:"roomId"`
	PlayerID string `json:"playerId"`
	Position int    `json:"position"`
}

type MessageLeaveGame struct {
	RoomID   string `json:"roomId"`
	PlayerID string `json:"playerId"`
}

type MessageStartGame struct {
	RoomID   string `json:"roomId"`
	PlayerID string `json:"playerId"`
}

type MessageGameAction struct {
	Action   string          `json:"action"`
	RoomID   string          `json:"roomId"`
	PlayerID string          `json:"playerId"`
	Data     json.RawMessage `json:"data"`
}

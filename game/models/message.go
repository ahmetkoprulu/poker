package models

import "encoding/json"

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeRoomInfo   MessageType = "room_info"
	MessageTypeJoinRoom   MessageType = "room_join"
	MessageTypeJoinRoomOk MessageType = "room_join_ok"
	MessageTypeLeaveRoom  MessageType = "room_leave"
	MessageTypeJoinGame   MessageType = "game_join"
	MessageTypeJoinGameOk MessageType = "game_join_ok"
	MessageTypeLeaveGame  MessageType = "game_leave"
	MessageTypeGameAction MessageType = "game_action"
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
	RoomID   string          `json:"roomId"`
	PlayerID string          `json:"playerId"`
	GameType int             `json:"gameType"`
	Data     json.RawMessage `json:"data"`
}

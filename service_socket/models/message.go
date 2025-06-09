package models

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeRoomInfo         MessageType = "room_info"
	MessageTypeJoinRoom         MessageType = "room_join"
	MessageTypeJoinRoomOk       MessageType = "room_join_ok"
	MessageTypeLeaveRoom        MessageType = "room_leave"
	MessageTypeLeaveRoomOk      MessageType = "room_leave_ok"
	MessageTypeJoinGame         MessageType = "game_join"
	MessageTypeJoinGameOk       MessageType = "game_join_ok"
	MessageTypeLeaveGame        MessageType = "game_leave"
	MessageTypeLeaveGameOk      MessageType = "game_leave_ok"
	MessageTypeGameAction       MessageType = "game_action"
	MessageTypeGameHoldemAction MessageType = "game_holdem_action"
	MessageTypeError            MessageType = "error"
)

// Message represents a WebSocket message
type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Response struct {
	Type      MessageType `json:"type"`
	PlayerID  string      `json:"player_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Message Room
type MessageRoomInfo struct {
	RoomID string `json:"room_id"`
}

type MessageJoinRoom struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
}

type MessageJoinRoomResponse struct {
	RoomID   string      `json:"room_id"`
	PlayerID string      `json:"player_id"`
	Player   *Player     `json:"player"`
	State    interface{} `json:"state"`
}

type MessageLeaveRoom struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
}

type MessageLeaveRoomResponse struct {
	RoomID   string      `json:"room_id"`
	PlayerID string      `json:"player_id"`
	State    interface{} `json:"state"`
}

// Message Game
type MessageJoinGame struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
	Position int    `json:"position"`
}

type MessageJoinGameResponse struct {
	RoomID   string      `json:"room_id"`
	Player   *Player     `json:"player"`
	Position int         `json:"position"`
	State    interface{} `json:"state"`
}

type MessageLeaveGame struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
}

type MessageLeaveGameResponse struct {
	RoomID   string      `json:"room_id"`
	PlayerID string      `json:"player_id"`
	State    interface{} `json:"state"`
}

type MessageGameAction struct {
	RoomID   string          `json:"room_id"`
	PlayerID string          `json:"player_id"`
	GameType int             `json:"game_type"`
	Data     json.RawMessage `json:"data"`
}

type MessageGameActionResponse struct {
	RoomID   string          `json:"room_id"`
	PlayerID string          `json:"player_id"`
	GameType int             `json:"game_type"`
	Data     json.RawMessage `json:"data"`
}

package models

// StringResponse is a concrete type for string responses (for Swagger)
type StringResponse struct {
	Data    string `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// BattlePassResponse is a concrete type for BattlePass responses (for Swagger)
type BattlePassResponse struct {
	Data    *BattlePass `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PlayerBattlePassResponse is a concrete type for PlayerBattlePass responses (for Swagger)
type PlayerBattlePassResponse struct {
	Data    *PlayerBattlePass `json:"data,omitempty"`
	Message string            `json:"message,omitempty"`
	Error   string            `json:"error,omitempty"`
}

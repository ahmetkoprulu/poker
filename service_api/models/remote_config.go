package models

import "time"

type RemoteConfig struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Version   int            `json:"version"`
	Value     map[string]any `json:"value"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
}

type GeneralRemoteConfig struct {
	MiniGames MiniGamesConfig `json:"mini_games"`
}

type MiniGamesConfig struct {
	WheelRewards     []Item `json:"wheel_rewards"`
	GoldWheelRewards []Item `json:"gold_wheel_rewards"`
	SlotRewards      []Item `json:"slot_rewards"`
}

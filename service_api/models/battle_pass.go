package models

import (
	"time"
)

type BattlePassStatus string

const (
	BattlePassStatusActive   BattlePassStatus = "active"
	BattlePassStatusExpired  BattlePassStatus = "expired"
	BattlePassStatusUpcoming BattlePassStatus = "upcoming"
)

// BattlePass represents a season of the battle pass
type BattlePass struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	StartTime   time.Time        `json:"start_time"`
	EndTime     time.Time        `json:"end_time"`
	Status      BattlePassStatus `json:"status"`
	MaxLevel    int              `json:"max_level"`
	CreatedAt   time.Time        `json:"created_at,omitempty"`
	UpdatedAt   time.Time        `json:"updated_at,omitempty"`
}

// BattlePassLevel represents a level in the battle pass with its rewards
type BattlePassLevel struct {
	ID             string    `json:"id"`
	BattlePassID   string    `json:"battle_pass_id"`
	Level          int       `json:"level"`
	RequiredXP     int       `json:"required_xp"`
	FreeRewards    []Item    `json:"free_rewards"`
	PremiumRewards []Item    `json:"premium_rewards"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

// PlayerBattlePass represents a player's progress in a battle pass
type PlayerBattlePass struct {
	ID           string    `json:"id"`
	PlayerID     string    `json:"player_id"`
	BattlePassID string    `json:"battle_pass_id"`
	CurrentLevel int       `json:"current_level"`
	CurrentXP    int       `json:"current_xp"`
	IsPremium    bool      `json:"is_premium"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`

	// Populated from joins
	BattlePass *BattlePass `json:"battle_pass,omitempty"`
}

// PlayerBattlePassReward represents a reward claimed by a player
type PlayerBattlePassReward struct {
	ID                 string    `json:"id"`
	PlayerBattlePassID string    `json:"player_battle_pass_id"`
	Level              int       `json:"level"`
	IsPremium          bool      `json:"is_premium"`
	ClaimedAt          time.Time `json:"claimed_at,omitempty"`
	CreatedAt          time.Time `json:"created_at,omitempty"`
}

// BattlePassXPTransaction represents an XP transaction in the battle pass
type BattlePassXPTransaction struct {
	ID                 string                 `json:"id"`
	PlayerBattlePassID string                 `json:"player_battle_pass_id"`
	Amount             int                    `json:"amount"`
	Source             string                 `json:"source"` // e.g., "challenge", "game_win", "bonus"
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at,omitempty"`
}

type BattlePassProgressDetails struct {
	Progress       PlayerBattlePass         `json:"progress"`
	ClaimedRewards []PlayerBattlePassReward `json:"claimed_rewards"`
	Levels         []BattlePassLevel        `json:"levels"`
}

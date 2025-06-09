package models

import (
	"time"
)

type ChallengeType string
type ChallengeCategory string

const (
	// Challenge Categories
	CategoryDaily       ChallengeCategory = "daily"
	CategoryWeekly      ChallengeCategory = "weekly"
	CategoryAchievement ChallengeCategory = "achievement"
	CategoryEvent       ChallengeCategory = "event"

	// Challenge Types
	ChallengePlayHands     ChallengeType = "play_hands"
	ChallengeWinHands      ChallengeType = "win_hands"
	ChallengeWinWithHand   ChallengeType = "win_with_hand"
	ChallengePlayGames     ChallengeType = "play_games"
	ChallengeWinAllIn      ChallengeType = "win_all_in"
	ChallengeGetRoyalFlush ChallengeType = "get_royal_flush"
)

type Challenge struct {
	ID           string            `json:"id"`
	Type         ChallengeType     `json:"type"`
	Category     ChallengeCategory `json:"category"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Requirements Requirements      `json:"requirements"`
	Rewards      []Item            `json:"rewards"` // Using existing Item type
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	CreatedAt    time.Time         `json:"created_at"`
}

type Requirements struct {
	TargetCount int         `json:"target_count"`
	Conditions  []Condition `json:"conditions,omitempty"`
}

type Condition struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type PlayerChallenge struct {
	ID            string     `json:"id"`
	PlayerID      string     `json:"player_id"`
	ChallengeID   string     `json:"challenge_id"`
	Progress      int        `json:"progress"`
	Completed     bool       `json:"completed"`
	RewardClaimed bool       `json:"reward_claimed"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Challenge *Challenge `json:"challenge,omitempty"`
}

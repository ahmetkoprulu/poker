package models

type ChipUpdateMessage struct {
	PlayerChanges []PlayerChipChange `json:"player_changes"`
}

type PlayerChipChange struct {
	PlayerID string `json:"player_id"`
	Change   int    `json:"change"`
}

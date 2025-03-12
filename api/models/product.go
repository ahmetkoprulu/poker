package models

type ItemType int16

const (
	ItemTypeChips        ItemType = 0
	ItemTypeGold         ItemType = 1
	ItemTypeSpin         ItemType = 2
	ItemTypeEvent        ItemType = 3
	ItemTypeMultiplier   ItemType = 97
	ItemTypeGoldSpin     ItemType = 98
	ItemTypeBattlePassXP ItemType = 99
)

type Item struct {
	Type     ItemType       `json:"type"`
	Amount   int            `json:"amount"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type EventItemMetadata struct {
	EventID string `json:"event_id"`
}

type Product struct {
	ID    string `json:"id"`
	Item  Item   `json:"item"`
	Price int    `json:"price"`
}

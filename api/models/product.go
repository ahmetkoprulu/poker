package models

type ItemType int16

const (
	ItemTypeChips        ItemType = 0
	ItemTypeGold         ItemType = 1
	ItemTypeGoldSpin     ItemType = 98
	ItemTypeBattlePassXP ItemType = 99
)

type Item struct {
	Type     ItemType               `json:"type"`
	Amount   int                    `json:"amount"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type Product struct {
	ID    string `json:"id"`
	Item  Item   `json:"item"`
	Price int    `json:"price"`
}

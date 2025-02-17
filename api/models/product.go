package models

type ItemType int16

const (
	ItemTypeChips ItemType = 1
	ItemTypeGold  ItemType = 2
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

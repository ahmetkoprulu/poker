package mq

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type GameEventPublisher struct {
	client   *MqClient
	exchange string
}

func NewGameEventPublisher() (*GameEventPublisher, error) {
	client, err := NewMqClient()
	if err != nil {
		return nil, err
	}

	// Declare exchange
	err = client.Provider.DeclareExchange(GameExchange, "topic", true)

	return &GameEventPublisher{
		client:   client,
		exchange: GameExchange,
	}, err
}

func (p *GameEventPublisher) PublishChipUpdate(msg *ChipUpdateMessage) error {
	msg.MessageID = uuid.New().String()
	msg.Timestamp = time.Now()
	routingKey := fmt.Sprintf("poker.game.%s.chip_update.%s", msg.GameType, msg.RoomID)

	return p.client.Provider.Publish(p.exchange, routingKey, msg)
}

type ChipUpdateMessage struct {
	MessageID string    `json:"message_id"`
	Timestamp time.Time `json:"timestamp"`
	GameType  string    `json:"game_type"` // "holdem", "tournament"

	RoomID string `json:"room_id"`

	PlayerChanges []PlayerChipChange `json:"player_changes"`
}

type PlayerChipChange struct {
	PlayerID string `json:"player_id"`
	Change   int    `json:"change"`
}

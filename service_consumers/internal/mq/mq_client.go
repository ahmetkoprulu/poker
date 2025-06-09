package mq

import (
	"github.com/ahmetkoprulu/rtrp/consumers/common/mq"
)

const (
	GameExchange string = "poker.game.events" // topic, durable
)

// poker.game.{game_type}.{event_type}.{room_id}
const (
	ChipUpdatesQueue           string = "chip_updates_db"             // queue, durable, routing: poker.game.*.chip_update.*
	GameAnalyticsQueue         string = "game_analytics_db"           // queue, durable, routing: poker.game.*.*.#
	GameAuditsQueue            string = "game_audits_db"              // queue, durable, routing: poker.game.*.hand_complete.*
	ChipUpdatesDeadLetterQueue string = "chip_updates_db_dead_letter" // queue, durable
)

const (
	ChipUpdatesRoutingKey           string = "poker.game.*.chip_update.*" // poker.game.holdem.chip_update.room_123
	GameAnalyticsRoutingKey         string = "poker.game.*.*.#"           // poker.game.holdem.hand_complete.room_123
	GameAuditsRoutingKey            string = "poker.game.*.hand_complete.*"
	ChipUpdatesDeadLetterRoutingKey string = "poker.game.*.chip_update.*"
)

type MqClient struct {
	Config   mq.RabbitMqConfig
	Provider mq.IMqProvider
}

var client *MqClient

func InitMq(url string) error {
	config := mq.RabbitMqConfig{
		URL: url,
	}

	provider, e := mq.NewRabbitmqMqProvider(config)
	if e != nil {
		return e
	}

	client = &MqClient{Config: config, Provider: provider}
	return nil
}

func NewMqClient() (*MqClient, error) {
	if client == nil {
		return nil, nil
	}

	return client, nil
}

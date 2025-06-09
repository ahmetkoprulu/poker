package consumers

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/ahmetkoprulu/rtrp/consumers/common/data"
	"github.com/ahmetkoprulu/rtrp/consumers/internal/mq"
	"github.com/ahmetkoprulu/rtrp/consumers/models"
)

type ChipUpdateConsumer struct {
	key      string
	db       *data.PgDbContext
	mqClient *mq.MqClient
}

func NewChipUpdateConsumer(key string, db *data.PgDbContext) *ChipUpdateConsumer {
	mqClient, err := mq.NewMqClient()
	if err != nil {
		log.Fatal("Failed to initialize MQ: ", err)
	}

	return &ChipUpdateConsumer{key: key, db: db, mqClient: mqClient}
}

func (h *ChipUpdateConsumer) Start(key string, wg *sync.WaitGroup) error {
	// Declare queue (idempotent)
	err := h.mqClient.Provider.DeclareQueue(mq.ChipUpdatesQueue, true, mq.ChipUpdatesRoutingKey, mq.GameExchange)
	if err != nil {
		return err
	}

	log.Printf("Starting %s consumer", key)

	wg.Add(1)
	defer wg.Done()

	if err = h.mqClient.Provider.Subscribe(mq.ChipUpdatesQueue, key, h.Consume); err != nil {
		return err
	}

	return nil
}

func (h *ChipUpdateConsumer) Consume(rawMsg []byte) error {
	var msg models.ChipUpdateMessage
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		return err
	}
	log.Printf("[%s] Consuming message: %+v", h.key, msg)
	return h.db.WithTransaction(context.Background(), func(tx data.QueryRunner) error {
		// Update player chips
		// max 0 or the chip + $1
		for _, change := range msg.PlayerChanges {
			_, err := tx.Exec(context.Background(), `
                UPDATE players 
                SET chips = GREATEST(chips + $1, 0), 
                    updated_at = NOW() 
                WHERE id = $2
            `, change.Change, change.PlayerID)

			if err != nil {
				return err
			}
		}

		return nil
	})
}

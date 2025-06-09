package internal

import (
	"sync"

	"github.com/ahmetkoprulu/rtrp/consumers/common/data"
	"github.com/ahmetkoprulu/rtrp/consumers/internal/consumers"
	"github.com/ahmetkoprulu/rtrp/consumers/internal/mq"
)

type ConsumerManager struct {
	mqClient  *mq.MqClient
	db        *data.PgDbContext
	wg        sync.WaitGroup
	shutdown  chan struct{}
	consumers map[string]consumers.IConsumer
}

func NewConsumerManager(db *data.PgDbContext, consumers map[string]consumers.IConsumer) (*ConsumerManager, error) {
	client, err := mq.NewMqClient()
	if err != nil {
		return nil, err
	}

	return &ConsumerManager{
		mqClient:  client,
		db:        db,
		consumers: consumers,
		shutdown:  make(chan struct{}),
	}, nil
}

func (s *ConsumerManager) Start() {
	// Set QoS - process one message at a time for reliability
	// ch.Qos(1, 0, false)
	for key, consumer := range s.consumers {
		go consumer.Start(key, &s.wg)
	}
}

func (s *ConsumerManager) Shutdown() {
	close(s.shutdown)

	s.wg.Wait()
}

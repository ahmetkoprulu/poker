package mq

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

type RabbitmqMqProvider struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      amqp.Queue
	config     RabbitMqConfig
}

type RabbitMqConfig struct {
	URL          string
	Exchange     string
	ExchangeType string // typically "fanout" for event broadcasting
	QueueName    string
	BindingKey   string
	Durable      bool
	Reliable     bool
}

func NewRabbitmqMqProvider(config RabbitMqConfig) (*RabbitmqMqProvider, error) {
	var provider = &RabbitmqMqProvider{config: config}

	err := provider.Connect(config.URL)
	if err != nil {
		return nil, err
	}

	err = provider.channel.ExchangeDeclare(config.Exchange, config.ExchangeType, config.Durable, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %v", err)
	}

	// if config.Reliable {
	// 	if err := provider.channel.Confirm(false); err != nil {
	// 		return nil, fmt.Errorf("channel could not be put into confirm mode: %v", err)
	// 	}
	// }

	return provider, nil
}

func (r *RabbitmqMqProvider) Connect(connectionString string) error {
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		return err
	}

	r.connection = connection
	r.channel, err = r.connection.Channel()
	if err != nil {
		return err
	}

	return nil
}

func (r *RabbitmqMqProvider) Disconnect() {
	if r.connection == nil {
		return
	}

	r.connection.Close()
}

func (r *RabbitmqMqProvider) Publish(queue string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// headers := amqp.Table{
	//     "event_id":        event.EventID.String(),
	//     "event_type":      event.EventType,
	//     "aggregate_type":  event.AggregateType,
	//     "aggregate_id":    event.AggregateID.String(),
	//     "created_at":      event.CreatedAt.Format(time.RFC3339),
	// }

	err = r.channel.Publish(
		r.config.Exchange,   // exchange
		r.config.BindingKey, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			// Headers:         headers,
			ContentType:     "application/json",
			ContentEncoding: "utf-8",
			Body:            bytes,
			DeliveryMode:    amqp.Persistent,
			Timestamp:       time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	// If publisher confirms are enabled, wait for confirmation
	if r.config.Reliable {
		if confirmed := <-r.channel.NotifyPublish(make(chan amqp.Confirmation, 1)); !confirmed.Ack {
			return fmt.Errorf("failed to receive publish confirmation")
		}
	}

	return nil
}

func (r *RabbitmqMqProvider) Subscribe(queue string, callback func(data []byte) error) error {
	err := r.declareQueue()
	if err != nil {
		return err
	}

	msgs, err := r.channel.Consume(
		r.queue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			_ = callback(msg.Body)
		}
	}()

	return nil
}

func (r *RabbitmqMqProvider) declareQueue() error {
	queue, err := r.channel.QueueDeclare(
		r.config.QueueName, // name
		r.config.Durable,   // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	r.queue = queue
	err = r.channel.QueueBind(
		queue.Name,          // queue name
		r.config.BindingKey, // routing key
		r.config.Exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	return nil
}

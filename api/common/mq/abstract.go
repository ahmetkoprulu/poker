package mq

import (
	"context"
)

type IMqProvider interface {
	Connect(connectionString string) error
	Disconnect()
	Publish(queue string, data interface{}) error
	Subscribe(context context.Context, queue string, callback func(data []byte) error) error
}

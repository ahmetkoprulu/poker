package consumers

import "sync"

type IConsumer interface {
	Start(key string, wg *sync.WaitGroup) error
	Consume(rawMsg []byte) error
}

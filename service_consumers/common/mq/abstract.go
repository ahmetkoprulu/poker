package mq

type IMqProvider interface {
	Connect(connectionString string) error
	Disconnect()
	Publish(exchangeName string, bindingKey string, data interface{}) error
	Subscribe(queueName string, consumerTag string, callback func(data []byte) error) error
	DeclareExchange(exchangeName string, exchangeType string, durable bool) error
	DeclareQueue(queueName string, durable bool, bindingKey, exchange string) error
}

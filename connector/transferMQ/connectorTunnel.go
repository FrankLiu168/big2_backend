package transfermq

import "github.com/rabbitmq/amqp091-go"

type ConnectorTunnel struct {
	ConnectorHandler func(dev *amqp091.Delivery)
}
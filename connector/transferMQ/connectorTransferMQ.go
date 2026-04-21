package transfermq

import (
	"big2backend/infrastructure/rabbitmq"
	"log"
	"strings"

	"big2backend/shared/consts"

	"github.com/rabbitmq/amqp091-go"
)

type ConnectorTransferMQ struct {
	producer         *rabbitmq.Producer
	consumer         *rabbitmq.Consumer
	agentHelper      *ConnectorTransferAgentHelper
	gameHelper       *ConnectorTransferGameHelper
	connectorHandler func(dev *amqp091.Delivery)
	Tunnel           *ConnectorTunnel
}

var transferMQ *ConnectorTransferMQ

func GetTransferMQ() *ConnectorTransferMQ {
	if transferMQ != nil {
		return transferMQ
	}
	transferMQ := &ConnectorTransferMQ{}
	p, err := rabbitmq.NewProducer(consts.GAME_EXCHANGE_NAME)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	c, err := rabbitmq.NewConsumer(consts.GAME_EXCHANGE_NAME)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	transferMQ.producer = p
	transferMQ.consumer = c
	agentHelper := NewConnectorTransferAgentHelper(transferMQ)
	gameHelper := NewConnectorTransferGameHelper(transferMQ)
	transferMQ.agentHelper = agentHelper
	transferMQ.gameHelper = gameHelper
	return transferMQ
}

func (s *ConnectorTransferMQ) SetTunnel(fn func(dev *amqp091.Delivery)) {
	s.Tunnel = &ConnectorTunnel{
		ConnectorHandler: fn,
	}
}

func (s *ConnectorTransferMQ) Start() {
	s.consumer.Listen([]string{
		consts.ROUTING.CONNECTOR.FROM_GAME,
		consts.ROUTING.CONNECTOR.FROM_AGENT,
	}, s.handler)
}

func (s *ConnectorTransferMQ) Close() {
	s.producer.Close()
	s.consumer.Close()
}

func (s *ConnectorTransferMQ) Publish(routingKey string, message string, msgID string, replyID string) {
	s.producer.Publish(routingKey, message, msgID, replyID)
}

func (s *ConnectorTransferMQ) handler(dev *amqp091.Delivery) {
	print("my handloer", string(dev.Body))
	switch true {
	case strings.HasPrefix(consts.ROUTING.CONNECTOR.FROM_GAME, dev.RoutingKey):
		s.gameHelper.HandleConnectMessage(s.Tunnel,dev)

	case strings.HasPrefix(consts.ROUTING.CONNECTOR.FROM_AGENT, dev.RoutingKey):
		s.agentHelper.HandleConnectMessage(dev)
	}
}

// func (s *ConnectorTransferMQ)RegisterHandler(handler func(dev *amqp091.Delivery)) {
// 	s.connectorHandler = handler
// 	s.gameHelper.connectorHandler = handler
// }

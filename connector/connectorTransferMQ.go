package connector

import (
	"big2backend/infrastructure/rabbitmq"
	"log"
	"strings"

	"big2backend/shared/consts"
	"big2backend/shared/data"

	"github.com/rabbitmq/amqp091-go"
)

type ConnectorTransferMQ struct {
	producer    *rabbitmq.Producer
	consumer    *rabbitmq.Consumer
	agentHelper *ConnectorTransferAgentHelper
	gameHelper  *ConnectorTransferGameHelper
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
	data.LogD("Connector", "收到RoutingKey:", dev.RoutingKey)
	switch true {
	case strings.HasPrefix(consts.ROUTING.CONNECTOR.FROM_GAME, dev.RoutingKey):
		s.gameHelper.HandleConnectMessage(dev)

	case strings.HasPrefix(consts.ROUTING.CONNECTOR.FROM_AGENT, dev.RoutingKey):
		data.LogD("Connector", "ROUTING.CONNECTOR.FROM_AGENT", "收到")
		s.agentHelper.HandleConnectMessage(dev)
	}
}

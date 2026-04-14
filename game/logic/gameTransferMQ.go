package logic

import (
	"big2backend/infrastructure/rabbitmq"
	"log"
	"strings"

	"big2backend/shared/consts"
	"big2backend/shared/data"

	"github.com/rabbitmq/amqp091-go"
)

type GameTransferMQ struct {
	producer        *rabbitmq.Producer
	consumer        *rabbitmq.Consumer
	agentHelper     *GameTransferAgentHelper
	connectorHelper *GameTransferConnectorHelper
}

var server *GameTransferMQ

func GetTransferMQ() *GameTransferMQ {
	if server != nil {
		return server
	}
	server := &GameTransferMQ{}
	p, err := rabbitmq.NewProducer(consts.GAME_EXCHANGE_NAME)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	c, err := rabbitmq.NewConsumer(consts.GAME_EXCHANGE_NAME)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	server.producer = p
	server.consumer = c
	agentHelper := NewGameTransferAgentHelper(server)
	connectorHelper := NewGameTransferConnectorHelper(server)
	server.agentHelper = agentHelper
	server.connectorHelper = connectorHelper
	return server
}

func (s *GameTransferMQ) Start() {
	s.consumer.Listen([]string{
		consts.ROUTING.GAME.FROM_AGENT,
		consts.ROUTING.GAME.FROM_CONNECTOR,
	}, s.handler)
}

func (s *GameTransferMQ) Close() {
	s.producer.Close()
	s.consumer.Close()
}

func (s *GameTransferMQ) Publish(routingKey string, message string, msgID string, replyID string) {
	s.producer.Publish(routingKey, message, msgID, replyID)
}

func (s *GameTransferMQ) handler(dev *amqp091.Delivery) {
	data.LogD("game received from connector")
	switch true {
	case strings.HasPrefix(dev.RoutingKey, consts.ROUTING.GAME.FROM_AGENT):
		s.agentHelper.HandleAgentMessage(dev)
	case strings.HasPrefix(dev.RoutingKey, consts.ROUTING.GAME.FROM_CONNECTOR):
		s.connectorHelper.HandleConnectMessage(dev)
		// payload := helper.ConvertToBasePayload(string(dev.Body))
		// switch payload.CommandAction {
		// case data.OnCmdClientPlayerAction:

		// }
	}
}

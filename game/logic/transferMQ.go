package logic

import (
	"big2backend/infrastructure/rabbitmq"
	"big2backend/shared/data"
	"big2backend/shared/helper"
	"log"
	"strings"

	"big2backend/shared/consts"

	"github.com/rabbitmq/amqp091-go"
)

type LogicTransferMQ struct {
	producer *rabbitmq.Producer
	consumer *rabbitmq.Consumer
}

var server *LogicTransferMQ

func GetTransferMQ() *LogicTransferMQ {
	if server != nil {
		return server
	}
	server := &LogicTransferMQ{}
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
	return server
}

func (s *LogicTransferMQ) Start() {
	s.consumer.Listen([]string{
		consts.PLAYER_RECEIVE_FROM_AI_ROUTING_KEY + "*",
		consts.GAME_RECEIVE_FROM_CONNECTOR_ROUTING_KEY,
	}, s.handler)
}

func (s *LogicTransferMQ) Close() {
	s.producer.Close()
	s.consumer.Close()
}

func (s *LogicTransferMQ) Publish(routingKey string, message string, msgID string) {
	s.producer.Publish(routingKey, message, msgID)
}

func (s *LogicTransferMQ) handler(dev *amqp091.Delivery) {
	switch true {
	case strings.HasPrefix(dev.RoutingKey, consts.PLAYER_RECEIVE_FROM_AI_ROUTING_KEY):
		wh := helper.GetWorkHelper()
		wh.Reply(dev.MessageId, string(dev.Body))
	case strings.HasPrefix(dev.RoutingKey, consts.GAME_RECEIVE_FROM_CONNECTOR_ROUTING_KEY):
		payload,_ := helper.ConvertToObject[data.BasePayload](string(dev.Body))
		switch payload.CommandAction {
		case data.OnCmdClientPlayerAction:
			
		}
	}
}

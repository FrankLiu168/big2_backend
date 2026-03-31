package logic

import (
	"big2backend/infrastructure/rabbitmq"
	"big2backend/shared/helper"
	"log"
	"strings"

	"big2backend/shared/consts"

	"github.com/rabbitmq/amqp091-go"
)

type LogicTransferServer struct {
	producer *rabbitmq.Producer
	consumer *rabbitmq.Consumer
}

var server *LogicTransferServer
func GetTransferServer() *LogicTransferServer {
	if server != nil {
		return server
	}
	server := &LogicTransferServer{
	}
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

func (s *LogicTransferServer) Start() {
	s.consumer.Listen([]string{consts.PLAYER_RESPONSE_ROUTING_KEY + "*"}, s.handler)
}

func (s *LogicTransferServer) Close() {
	s.producer.Close()
	s.consumer.Close()
}

func (s *LogicTransferServer) Publish(routingKey string, message string,msgID string) {
	s.producer.Publish(routingKey, message,msgID)
}

func (s *LogicTransferServer) handler(dev amqp091.Delivery) {
	if strings.HasPrefix(dev.RoutingKey,consts.PLAYER_RESPONSE_ROUTING_KEY) {
		wh := helper.GetWorkHelper()
		wh.Reply(dev.MessageId,string(dev.Body))
		
	}
}

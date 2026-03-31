package ai

import (
	"big2backend/infrastructure/rabbitmq"
	"encoding/json"
	"fmt"
	"log"

	"big2backend/shared/consts"
	"big2backend/shared/data"

	"github.com/rabbitmq/amqp091-go"
)

type AITransferServer struct {
	Agent    *Agent
	producer *rabbitmq.Producer
	consumer *rabbitmq.Consumer
}

func NewTransferServer() *AITransferServer {
	server := &AITransferServer{
		Agent: NewAgent(),
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

func (s *AITransferServer) Start() {
	s.consumer.Listen([]string{consts.AGENT_REQUEST_ROUTING_KEY}, s.handler)
}

func (s *AITransferServer) Close() {
	s.producer.Close()
	s.consumer.Close()
}

func (s *AITransferServer) Publish(routingKey string, message string, msgID string) {
	s.producer.Publish(routingKey, message, msgID)
}

func (s *AITransferServer) handler(dev amqp091.Delivery) {
	if dev.RoutingKey == consts.AGENT_REQUEST_ROUTING_KEY {
		payload := &data.BasePayload{}
		_ = json.Unmarshal(dev.Body, payload)
		aiPayload := &data.AIPayloadRequest{}
		_ = json.Unmarshal([]byte(payload.Data), aiPayload)
		action := s.Agent.Strategy(&aiPayload.GameRecord, &aiPayload.Info)
		res := data.AIPayloadResponse{Action: *action}
		resBytes, _ := json.Marshal(res)
		resPayload := data.BasePayload{Command: 1, Data: string(resBytes)}
		routingKey := fmt.Sprintf(consts.PLAYER_RESPONSE_ROUTING_KEY+"%d", action.PlayerID)
		resStr, _ := json.Marshal(resPayload)
		s.Publish(routingKey, string(resStr), dev.MessageId)
	}
}

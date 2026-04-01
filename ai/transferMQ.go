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

type AITransferMQ struct {
	Agent    *Agent
	producer *rabbitmq.Producer
	consumer *rabbitmq.Consumer
}

func NewTransferMQ() *AITransferMQ {
	server := &AITransferMQ{
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

func (s *AITransferMQ) Start() {
	s.consumer.Listen([]string{
		consts.AGENT_RECEIVE_FROM_PLAYER_ROUTING_KEY,
		consts.AGENT_RECEIVE_FROM_CONNECT_ROUTING_KEY,
	}, s.handler)
}

func (s *AITransferMQ) Close() {
	s.producer.Close()
	s.consumer.Close()
}

func (s *AITransferMQ) Publish(routingKey string, message string, msgID string) {
	s.producer.Publish(routingKey, message, msgID)
}

func (s *AITransferMQ) handler(dev *amqp091.Delivery) {
	switch dev.RoutingKey {
	case consts.AGENT_RECEIVE_FROM_CONNECT_ROUTING_KEY:
		handleAgentReceiveFromConnectRoutingKey(s, dev)
	case consts.AGENT_RECEIVE_FROM_PLAYER_ROUTING_KEY:
		handleAgentReceiveFromPlayerRoutingKey(s, dev)
	}
}

func handleAgentReceiveFromPlayerRoutingKey(transfer *AITransferMQ, dev *amqp091.Delivery) {
	payload := &data.BasePayload{}
	_ = json.Unmarshal(dev.Body, payload)
	aiPayload := &data.AIPayloadRequest{}
	_ = json.Unmarshal([]byte(payload.Data), aiPayload)
	action := transfer.Agent.Strategy(&aiPayload.GameRecord, &aiPayload.Info)
	res := data.AIPayloadResponse{Action: *action}
	resBytes, _ := json.Marshal(res)
	resPayload := data.BasePayload{Data: string(resBytes)}
	routingKey := fmt.Sprintf(consts.PLAYER_RECEIVE_FROM_AI_ROUTING_KEY+"%d", action.PlayerID)
	resStr, _ := json.Marshal(resPayload)
	transfer.Publish(routingKey, string(resStr), dev.MessageId)
}

func handleAgentReceiveFromConnectRoutingKey(transfer *AITransferMQ, dev *amqp091.Delivery) {
	payload := &data.BasePayload{}
	_ = json.Unmarshal(dev.Body, payload)
	aiPayload := &data.AIPayloadRequest{}
	_ = json.Unmarshal([]byte(payload.Data), aiPayload)
	action := transfer.Agent.Strategy(&aiPayload.GameRecord, &aiPayload.Info)
	res := data.AIPayloadResponse{Action: *action}
	resBytes, _ := json.Marshal(res)
	resPayload := data.BasePayload{Data: string(resBytes)}
	routingKey := consts.CONNECTOR_RECEIVE_FROM_AGENT_ROUTING_KEY
	resStr, _ := json.Marshal(resPayload)
	transfer.Publish(routingKey, string(resStr), dev.MessageId)
}

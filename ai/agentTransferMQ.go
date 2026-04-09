package ai

import (
	"big2backend/infrastructure/rabbitmq"
	"log"

	"big2backend/shared/consts"

	"github.com/rabbitmq/amqp091-go"
)

type AITransferMQ struct {
	Agent    *Agent
	producer *rabbitmq.Producer
	consumer *rabbitmq.Consumer
	gameHelper *AgentTransferGameHelper
	connectorHelper *AgentTransferConnectorHelper
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
	gameHelper := NewAgentTransferGameHelper(server)
	connectorHelper := NewAgentTransferConnectorHelper(server)
	server.gameHelper = gameHelper
	server.connectorHelper = connectorHelper
	return server
}

func (s *AITransferMQ) Start() {
	s.consumer.Listen([]string{
		consts.ROUTING.AGENT.FROM_GAME,
		consts.ROUTING.AGENT.FROM_CONNECTOR,
	}, s.handler)
}

func (s *AITransferMQ) Close() {
	s.producer.Close()
	s.consumer.Close()
}

func (s *AITransferMQ) Publish(routingKey string, message string, msgID string, replyID string) {
	s.producer.Publish(routingKey, message, msgID, replyID)

}

func (s *AITransferMQ) handler(dev *amqp091.Delivery) {
	switch dev.RoutingKey {
	case consts.ROUTING.AGENT.FROM_CONNECTOR:
		s.connectorHelper.HandleConnectMessage(dev)
	case consts.ROUTING.AGENT.FROM_GAME:
		s.gameHelper.HandleGameMessage(dev)
	}
}


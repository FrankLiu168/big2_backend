package ai

import (
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"

	"github.com/rabbitmq/amqp091-go"
)

type AgentTransferConnectorHelper struct {
	Transfer *AITransferMQ
}

func NewAgentTransferConnectorHelper(transfer *AITransferMQ) *AgentTransferConnectorHelper {
	return &AgentTransferConnectorHelper{
		Transfer: transfer,
	}
}

func (c *AgentTransferConnectorHelper) HandleConnectMessage(dev *amqp091.Delivery) {
	payload := helper.ConvertToBasePayload(string(dev.Body))
	switch payload.CommandAction {
	case data.CommandAction(data.InAIPayloadRequest):
		replyToConnect(c.Transfer, dev)
	}
}
func replyToConnect(transfer *AITransferMQ, dev *amqp091.Delivery) {
	msgID := helper.GetUniqueID()
	payload := helper.ConvertToBasePayload(string(dev.Body))
	aiPayload := helper.ConvertToPayload[data.AIPayloadRequest](payload)
	action := transfer.Agent.Strategy(&aiPayload.GameRecord, &aiPayload.Info)
	resPayload := data.AIPayloadResponse{Action: *action}
	str := helper.PackPayload(data.CommandAction(data.InAIPayloadResponse),99, "", &resPayload)
	routingKey := consts.ROUTING.CONNECTOR.FROM_AGENT
	
	transfer.Publish(routingKey, str, msgID, dev.MessageId)

}

func notify(dev *amqp091.Delivery) {

}

func response(dev *amqp091.Delivery) {

}

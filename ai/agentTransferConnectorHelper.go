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
	data.LogD("Agent","ROUTING.AGENT.FROM_CONNECTOR")
	payload, _ := helper.ConvertToObject[data.BasePayload](string(dev.Body))
	switch payload.CommandAction {
	case data.CommandAction(data.InAIPayloadRequest):
		replyToConnect(c.Transfer, dev)
	}
}
func replyToConnect(transfer *AITransferMQ, dev *amqp091.Delivery) {
	msgID := helper.GetUniqueID()
	payload, _ := helper.ConvertToObject[data.BasePayload](string(dev.Body))
	aiPayload, _ := helper.ConvertToObject[data.AIPayloadRequest](payload.Data)
	action := transfer.Agent.Strategy(&aiPayload.GameRecord, &aiPayload.Info)
	res := data.AIPayloadResponse{Action: *action}
	resStr, _ := helper.ConvertToData(&res)
	resPayload := data.BasePayload{CommandAction: data.CommandAction(data.InAIPayloadResponse),
		Data: resStr}
	routingKey := consts.ROUTING.CONNECTOR.FROM_AGENT
	str, _ := helper.ConvertToData(&resPayload)
	data.LogD("Agent Publish","ROUTING.CONNECTOR.FROM_AGENT",str)
	transfer.Publish(routingKey, str, msgID, dev.MessageId)
}

func notify(dev *amqp091.Delivery) {

}

func response(dev *amqp091.Delivery) {

}

package ai

import (
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"

	"github.com/rabbitmq/amqp091-go"
)

type AgentTransferGameHelper struct {
	Transfer *AITransferMQ
}

func NewAgentTransferGameHelper(transfer *AITransferMQ) *AgentTransferGameHelper {
	return &AgentTransferGameHelper{
		Transfer: transfer,
	}
}

func (g *AgentTransferGameHelper) HandleGameMessage(dev *amqp091.Delivery) {
	payload, _ := helper.ConvertToObject[data.BasePayload](string(dev.Body))
	switch payload.CommandAction {
	case data.CommandAction(data.InAIPayloadRequest):
		replyToPlayer(g.Transfer, dev)
	}
}

func replyToPlayer(transfer *AITransferMQ, dev *amqp091.Delivery) {
	msgID := helper.GetUniqueID()
	payload, _ := helper.ConvertToObject[data.BasePayload](string(dev.Body))
	aiPayload, _ := helper.ConvertToObject[data.AIPayloadRequest](payload.Data)
	action := transfer.Agent.Strategy(&aiPayload.GameRecord, &aiPayload.Info)
	res := data.AIPayloadResponse{Action: *action}
	resStr, _ := helper.ConvertToData(&res)
	resPayload := data.BasePayload{CommandAction: data.CommandAction(data.InAIPayloadResponse),
		Data: resStr}
	routingKey := consts.ROUTING.GAME.FROM_AGENT
	str, _ := helper.ConvertToData(&resPayload)
	data.LogD("Agent Publish", "ROUTING.GAME.FROM_AGENT")
	transfer.Publish(routingKey, str, msgID, dev.MessageId)
}

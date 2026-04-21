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
	payload := helper.ConvertToBasePayload(string(dev.Body))
	switch payload.CommandAction {
	case data.CommandAction(data.InAIPayloadRequest):
		replyToPlayer(g.Transfer, dev)
	}
}

func replyToPlayer(transfer *AITransferMQ, dev *amqp091.Delivery) {
	msgID := helper.GetUniqueID()
	payload := helper.ConvertToBasePayload(string(dev.Body))
	aiPayload := helper.ConvertToPayload[data.AIPayloadRequest](payload)
	action := transfer.Agent.Strategy(&aiPayload.GameRecord, &aiPayload.Info)
	res := data.AIPayloadResponse{Action: *action}
	resStr := helper.PackPayload(data.CommandAction(data.InAIPayloadResponse), 99, "", &res)

	routingKey := consts.ROUTING.GAME.FROM_AGENT
	transfer.Publish(routingKey, resStr, msgID, dev.MessageId)
}

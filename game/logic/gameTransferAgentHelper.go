package logic

import (
	"big2backend/shared/data"
	"big2backend/shared/helper"

	"github.com/rabbitmq/amqp091-go"
)

type GameTransferAgentHelper struct {
	Transfer *GameTransferMQ
}

func NewGameTransferAgentHelper(transfer *GameTransferMQ) *GameTransferAgentHelper {
    return &GameTransferAgentHelper{
		Transfer: transfer,
	}
}

func (c *GameTransferAgentHelper)HandleAgentMessage(dev *amqp091.Delivery) {
	payload := helper.ConvertToBasePayload(string(dev.Body))
	switch payload.CommandAction {
	case data.CommandAction(data.InAIPayloadResponse):
		responseFromAgent(c.Transfer,dev)
	}
}

func responseFromAgent(transfer *GameTransferMQ,dev *amqp091.Delivery) {
	replyID, isExist := dev.Headers["replyID"]
	if isExist && replyID != "" {
		helper.GetGameWork().Reply(replyID.(string), string(dev.Body))
	}
}
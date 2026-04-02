package logic

import (
	"big2backend/shared/data"
	"big2backend/shared/helper"

	"github.com/rabbitmq/amqp091-go"
)

type GameTransferConnectorHelper struct {
	Transfer *GameTransferMQ
}

func NewGameTransferConnectorHelper(transfer *GameTransferMQ) *GameTransferConnectorHelper {
    return &GameTransferConnectorHelper{
		Transfer: transfer,
	}
}

func (c *GameTransferConnectorHelper)HandleConnectMessage(dev *amqp091.Delivery) {
	payload,_ := helper.ConvertToObject[data.BasePayload](string(dev.Body))
	switch payload.CommandAction {
	case data.CommandAction(data.OnCmdClientPlayerAction):
		responseFromConnector(c.Transfer,dev)
	}
}

func responseFromConnector(transfer *GameTransferMQ,dev *amqp091.Delivery) {
	replyID, isExist := dev.Headers["replyID"]
	if isExist && replyID != "" {
		helper.GetGameWork().Reply(replyID.(string), string(dev.Body))
	}
}


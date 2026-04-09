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

func (c *GameTransferConnectorHelper) HandleConnectMessage(dev *amqp091.Delivery) {
	payload := helper.ConvertToBasePayload(string(dev.Body))
	data.LogD("GameTransferConnectorHelper",string(dev.Body))
	switch payload.CommandAction {
	case data.CommandAction(data.OnCmdClientPlayerAction):
		handleConnectorAction(c.Transfer, dev)
	case data.OnCmdClientReady:
		res := helper.ConvertToPayload[data.CmdClientReady](payload)
		data.LogD("---unlock wait----",res.ReplyID)
		helper.GetWaitHelper().Reply(res.ReplyID, string(dev.Body))
	}
}

func handleConnectorAction(transfer *GameTransferMQ, dev *amqp091.Delivery) {
	basePayload := helper.ConvertToBasePayload(string(dev.Body))
	switch basePayload.CommandAction {
	case data.OnCmdClientPlayerAction:
		payload := helper.ConvertToPayload[data.CmdClientPlayerAction](basePayload)
		replyID := payload.ReplyID
		helper.GetWaitHelper().Reply(replyID, string(dev.Body))
	}

}

// func responseFromConnector(transfer *GameTransferMQ,dev *amqp091.Delivery) {
// 	replyID, isExist := dev.Headers["replyID"]
// 	if isExist && replyID != "" {
// 		helper.GetGameWork().Reply(replyID.(string), string(dev.Body))
// 	}
// }

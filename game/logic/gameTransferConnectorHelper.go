package logic

import (
	"big2backend/shared/data"
	"big2backend/shared/helper"

	"github.com/rabbitmq/amqp091-go"
)

type GameTransferConnectorHelper struct {
	Transfer     *GameTransferMQ
	GameListener func(string)
}

func NewGameTransferConnectorHelper(transfer *GameTransferMQ) *GameTransferConnectorHelper {
	return &GameTransferConnectorHelper{
		Transfer: transfer,
	}
}

func (c *GameTransferConnectorHelper) HandleConnectMessage(dev *amqp091.Delivery) {
	cnnPayload := helper.ConvertToConnectorPayload(string(dev.Body))
	data.LogD("GameTransferConnectorHelper", string(dev.Body))
	switch cnnPayload.Data.CommandAction {
	case data.OnCmdClientPlayerAction:
		handleConnectorAction(c.Transfer, dev)
	default:
		c.GameListener(string(dev.Body))
	}
}

func (c *GameTransferConnectorHelper) SetGameListener(listener func(string)) {
	c.GameListener = listener
}

func handleConnectorAction(transfer *GameTransferMQ, dev *amqp091.Delivery) {
	print("handleConnectorAction", string(dev.Body))
	cnnPayload := helper.ConvertToConnectorPayload(string(dev.Body))
	print("\n +++++ handleConnectorAction \n")
	switch cnnPayload.Data.CommandAction {
	case data.OnCmdClientPlayerAction:
		//print("data",basePayload.Data)
		payload := helper.ConvertToClientPayload[data.CmdClientPlayerAction](&cnnPayload.Data)
		replyID := payload.ReplyID
		helper.GetWaitHelper().Reply(replyID, string(dev.Body))
	}

}



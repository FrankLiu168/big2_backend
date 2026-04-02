package connector

import (
	"big2backend/shared/data"
	"big2backend/shared/helper"

	"github.com/rabbitmq/amqp091-go"
)

type ConnectorTransferAgentHelper struct {
	Transfer *ConnectorTransferMQ
}

func NewConnectorTransferAgentHelper(transfer *ConnectorTransferMQ) *ConnectorTransferAgentHelper {
	return &ConnectorTransferAgentHelper{
		Transfer: transfer,
	}
}

func (c *ConnectorTransferAgentHelper) HandleConnectMessage(dev *amqp091.Delivery) {
	data.LogD("Connector", "HandleConnectMessage")
	payload, _ := helper.ConvertToObject[data.BasePayload](string(dev.Body))
	switch payload.CommandAction {
	case data.CommandAction(data.InAIPayloadResponse):
		responseFromAgent(c.Transfer, dev)
	}
}

func responseFromAgent(transfer *ConnectorTransferMQ, dev *amqp091.Delivery) {
	data.LogD("Connector", "responseFromAgent")
	replyID, isExist := dev.Headers["replyID"]
	if isExist && replyID != "" {
		helper.GetConnectorWork().Reply(replyID.(string), string(dev.Body))
	}
}

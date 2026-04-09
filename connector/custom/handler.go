package custom

import (
	transfermq "big2backend/connector/transferMQ"
	"big2backend/connector/common"
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"
)

func HandleMessage(client *common.Client, message []byte) {
	println("Received message from client:", string(message))
	basePayload := helper.ConvertToBasePayload(string(message))
	msgID := helper.GetUniqueID()
	if basePayload.CommandAction == data.OnCmdClientPlayerAction {
		//payload, _ := helper.ConvertToObject[data.CmdClientPlayerAction](basePayload.Data)
		transfer := transfermq.GetTransferMQ()
		transfer.Publish(consts.ROUTING.GAME.FROM_CONNECTOR, string(message), msgID, "")
	}
	if basePayload.CommandAction == data.OnCmdClientReady {
		transfermq := transfermq.GetTransferMQ()
		transfermq.Publish(consts.ROUTING.GAME.FROM_CONNECTOR, string(message), msgID, "")
	}
}

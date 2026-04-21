package custom

import (
	"big2backend/connector/common"
	transfermq "big2backend/connector/transferMQ"
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"
)

func HandleMessage(client *common.Client, message []byte) {
	println("Received message from client:", string(message))
	cliPayload, _ := helper.ConvertToObject[data.ClientBasePayload](string(message))
	cnnPayload := data.ConnectorPayload{
		Identifier: client.ID,
		Data:       *cliPayload,
	}
	str, _ := helper.ConvertToData(&cnnPayload)
	transfer := transfermq.GetTransferMQ()
	msgID := helper.GetUniqueID()
	transfer.Publish(consts.ROUTING.GAME.FROM_CONNECTOR, str, msgID, "")

	// println("basePayload:", basePayload.CommandAction)

	// if basePayload.CommandAction == data.OnCmdClientPlayerAction {
	// 	//payload, _ := helper.ConvertToObject[data.CmdClientPlayerAction](basePayload.Data)
	// 	transfer := transfermq.GetTransferMQ()
	// 	transfer.Publish(consts.ROUTING.GAME.FROM_CONNECTOR, string(message), msgID, "")
	// }
	// if basePayload.CommandAction == data.OnCmdClientReady {
	// 	println("send ready")
	// 	transfermq := transfermq.GetTransferMQ()
	// 	transfermq.Publish(consts.ROUTING.GAME.FROM_CONNECTOR, string(message), msgID, "")
	// }
}


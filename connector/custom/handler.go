package custom

import (
	"big2backend/connector/common"
	"big2backend/connector/custom/command"
)

func HandleMessage(client *common.Client, message []byte) {
	println("Received message from client:", string(message))
	commandName := command.GetCommandName(message)
	switch commandName {
	case "say_all":
		cmd, err := command.GetCommand[command.SayAllCommand](commandName, message)
		if err != nil {
			println("Error parsing say_all command:", err.Error())
			return
		}
		sayAll(client, cmd.Message)
	case "say_to":
		cmd, err := command.GetCommand[command.SayToCommand](commandName, message)
		if err != nil {
			println("Error parsing say_to command:", err.Error())
			return
		}
		sayTo(client, cmd.TargetID, cmd.Message)
	default:
		println("Unknown command:", commandName)
	}
}

func sayAll(client *common.Client, message string) {
	client.ExtendTunnel.BroadcastMessage([]byte("广播消息: " + message))
}

func sayTo(client *common.Client, targetID string, message string) {
	client.ExtendTunnel.SendMessageToClient(targetID, []byte("私聊消息: "+message))
}

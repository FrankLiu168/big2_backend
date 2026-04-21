package connector

import (
	"big2backend/connector/common"
	"big2backend/connector/custom"
	transfermq "big2backend/connector/transferMQ"
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients   = make(map[string]*common.Client)
	clientsMu sync.Mutex
)

func Init() {
	server := transfermq.GetTransferMQ()
	server.Start()
	//server.RegisterHandler(ReceiveFromTransfer)
	server.SetTunnel(ReceiveFromTransfer)
}

func DeleteClient(id string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if client, ok := clients[id]; ok {
		client.Conn.Close()
		delete(clients, id)
		fmt.Printf("用户 %s 已删除\n", id)
		cliPayload := data.ClientBasePayload{
			CommandAction: data.OnCmdConnectOffline,
			Data:          "{}",
			RoomID:        0,
			GameID:        "",
			IsBroadcast:   true,
		}
		cnnPayload := data.ConnectorPayload{
			Data:       cliPayload,
			Identifier: id,
		}
		str, _ := helper.ConvertToData(&cnnPayload)
		msgID := helper.GetUniqueID()
		server := transfermq.GetTransferMQ()
		server.Publish(consts.ROUTING.GAME.FROM_CONNECTOR, str, msgID, "")
	}
}

func BroadcastMessage(message []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(clients, client.ID)
		}
	}
}

func SendMessageToClient(clientID string, message []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if client, ok := clients[clientID]; ok {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(clients, client.ID)
		}
	}
}

func ReceiveFromTransfer(dev *amqp091.Delivery) {
	print("\n +++++ ReceiveFromTransfer \n")
	payload := helper.ConvertToBasePayload(string(dev.Body))
	if payload.Target == "" {
		BroadcastMessage(dev.Body)
	} else {
		print("---- send to client ----")
		SendMessageToClient(payload.Target, dev.Body)
	}
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	if !checkAuth(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	//userID := helper.GetUniqueID()
	userID := "user-1"
	tunnel := &common.ClientTunnel{
		HandleMessage:       custom.HandleMessage,
		BroadcastMessage:    BroadcastMessage,
		SendMessageToClient: SendMessageToClient,
	}

	client := &common.Client{
		Conn:         conn,
		Send:         make(chan []byte, 256),
		ID:           userID,
		DeleteClient: DeleteClient,
		ExtendTunnel: tunnel,
	}

	clientsMu.Lock()
	clients[userID] = client
	clientsMu.Unlock()

	fmt.Printf("用户 %s 已连接\n", userID)

	go client.ReadPump()
	go client.WritePump()
}

func checkAuth(r *http.Request) bool {
	// 在这里实现你的认证逻辑，例如检查请求头中的 token
	return true
}

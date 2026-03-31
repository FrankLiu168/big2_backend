package connector

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"big2backend/connector/common"
	"big2backend/connector/custom"
	"big2backend/shared/helper"
	"github.com/gorilla/websocket"
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

func DeleteClient(id string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if client, ok := clients[id]; ok {
		client.Conn.Close()
		delete(clients, id)
		fmt.Printf("用户 %s 已删除\n", id)
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

	userID := helper.GetUniqueID()

	tunnel := &common.Tunnel{
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

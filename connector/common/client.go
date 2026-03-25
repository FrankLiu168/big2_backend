package common

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn         *websocket.Conn
	Send         chan []byte
	ID           string
	DeleteClient func(id string)
	ExtendTunnel *Tunnel
}

func (c *Client) ReadPump() {
	defer func() {
		c.DeleteClient(c.ID)
	}()

	c.Conn.SetReadLimit(512 * 1024)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		fmt.Printf("用户 %s 收到消息: %s\n", c.ID, string(message))
		c.ExtendTunnel.HandleMessage(c, message)
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			fmt.Printf("发送失败给用户 %s: %v\n", c.ID, err)
			return
		}
		fmt.Printf("发送消息给用户 %s: %s\n", c.ID, string(message))
	}
}

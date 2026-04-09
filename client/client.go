// package client

package client

import (
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// Client 封装 WebSocket 客户端
type Client struct {
	url    string
	conn   *websocket.Conn
	closed bool
	mu     sync.Mutex
}

// NewClient 创建新客户端（不立即连接）
func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}

// Connect 建立 WebSocket 连接
func (c *Client) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	c.conn = conn
	return nil
}

// Start 启动消息收发（非阻塞，需配合 Wait 或 Close 使用）
func (c *Client) Start() error {
	if c.conn == nil {
		print("error")
		return fmt.Errorf("尚未连接，请先调用 Connect()")
	}

	go func() {
		for {
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("[错误] 读取消息: %v", err)
				}
				return
			}
			//log.Printf("[收到] %s", msg)
			c.Handler(string(msg))
		}
	}()

	// 发送消息（从 stdin）
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			if text == "" {
				continue
			}
			c.ParseInput(text)
			// if err := c.WriteMessage(websocket.TextMessage, []byte(text)); err != nil {
			// 	log.Printf("[错误] 发送消息失败: %v", err)
			// 	return
			// }
			// log.Printf("[已发送] %s", text)
		}
	}()

	forever := make(chan struct{})
	<-forever
	return nil
}

var tmpInfo *data.PlayerInfo
var currentReplyID string

func (c *Client) ParseInput(text string) {
	parts := strings.Split(text, "/")
	if len(parts) == 1 {
		c.SendReadyCommand()
	} else {
		isPass := parts[0]
		cardType := parts[1]
		cards := parts[2]
		c.SendActionCommand(isPass, cardType, cards)
	}

}

func (c *Client) Handler(msg string) {
	basePayload := helper.ConvertToBasePayload(msg)
	switch basePayload.CommandAction {
	case data.OnCmdServerDealCards:
		tmpInfo = &data.PlayerInfo{
			HandCards: []int{},
			ID:        1,
			Name:      "",
		}
		payload := helper.ConvertToPayload[data.CmdServerDealCards](basePayload)
		tmpInfo.HandCards = payload.Cards
		cardNameList := consts.GetCardNameList(payload.Cards)
		data.LogD("發牌完成", fmt.Sprintf("我的牌是：%s", strings.Join(cardNameList, ",")))
	case data.OnCmdServerPlayerAction:
		payload := helper.ConvertToPayload[data.CmdServerPlayerAction](basePayload)
		if payload.IsPass {
			data.LogD(fmt.Sprintf("玩家[%d]出牌", payload.PlayerID), "pass")
		} else {
			cardTypeName := consts.GetCardTypeName(payload.CardType)
			cardNameList := consts.GetCardNameList(payload.Cards)
			data.LogD(fmt.Sprintf("玩家[%d]出牌", payload.PlayerID), cardTypeName, strings.Join(cardNameList, ","))
		}

	case data.OnCmdServerCurrentPlayer:
		payload := helper.ConvertToPayload[data.CmdServerCurrentPlayer](basePayload)
		currentReplyID = payload.ReplyID
		if payload.PlayerID == 1 {
			cardNameList := consts.GetCardNameList(tmpInfo.HandCards)
			data.LogD("該你出牌：", strings.Join(cardNameList, ","))
		} else {
			data.LogD("該其它玩家出牌：", "請等候")
		}
	case data.OnCmdServerGameOver:
		data.LogD("OnCmdServerGameOver", "遊戲結束")

	}
}

func (c *Client) SendReadyCommand() {
	payload := data.CmdClientReady{
		PlayerID: 1,
		ReplyID:  "12345",
	}
	str := helper.PackPayload(data.OnCmdClientReady, "", &payload)
	c.WriteMessage(1, []byte(str))
}

func (c *Client) SendActionCommand(isPass string, cardType string, cards string) {
	payload := data.CmdClientPlayerAction{
		ReplyID:  currentReplyID,
		PlayerID: 1,
	}
	if isPass == "1" {
		payload.IsPass = true
	}
	if !payload.IsPass {
		ct, _ := strconv.Atoi(cardType)
		payload.CardType = consts.CardType(ct)
		cs := strings.Split(cards, ",")
		css := []int{}
		for _, v := range cs {
			i, _ := strconv.Atoi(v)
			css = append(css, i)
		}
		payload.Cards = css
	}

	str := helper.PackPayload(data.OnCmdClientPlayerAction, "", &payload)
	data.LogD("send to server", str)
	c.WriteMessage(1, []byte(str))
}

// WriteMessage 是安全的写封装
func (c *Client) WriteMessage(messageType int, msg []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || c.conn == nil {
		return io.ErrClosedPipe
	}
	return c.conn.WriteMessage(messageType, msg)
}

// Close 安全关闭连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || c.conn == nil {
		return nil
	}
	c.closed = true
	err := c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Printf("[警告] 发送关闭帧失败: %v", err)
	}
	return c.conn.Close()
}

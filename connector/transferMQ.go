package connector

import (
	"big2backend/infrastructure/rabbitmq"
	"fmt"
	"log"
	"strings"

	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"

	"github.com/rabbitmq/amqp091-go"
)

type AITransferMQ struct {
	producer *rabbitmq.Producer
	consumer *rabbitmq.Consumer
}

var transferMQ *AITransferMQ

func GetTransferMQ() *AITransferMQ {
	if transferMQ != nil {
		return transferMQ
	}
	transferMQ := &AITransferMQ{}
	p, err := rabbitmq.NewProducer(consts.GAME_EXCHANGE_NAME)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	c, err := rabbitmq.NewConsumer(consts.GAME_EXCHANGE_NAME)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	transferMQ.producer = p
	transferMQ.consumer = c
	return transferMQ
}

func (s *AITransferMQ) Start() {
	s.consumer.Listen([]string{
		consts.CONNECTOR_RECEIVE_FROM_GAME_ROUTING_KEY + "*",
		consts.CONNECTOR_RECEIVE_FROM_AGENT_ROUTING_KEY,
		}, s.handler)
}

func (s *AITransferMQ) Close() {
	s.producer.Close()
	s.consumer.Close()
}

func (s *AITransferMQ) Publish(routingKey string, message string, msgID string) {
	s.producer.Publish(routingKey, message, msgID)
}

func (s *AITransferMQ) handler(dev *amqp091.Delivery) {
	switch true {
		case strings.HasPrefix(consts.CONNECTOR_RECEIVE_FROM_GAME_ROUTING_KEY,dev.RoutingKey):
			payload, _ := helper.ConvertToObject[data.BasePayload](string(dev.Body))
			handleConnectorReceiveFromGame(payload)
		case strings.HasPrefix(consts.CONNECTOR_RECEIVE_FROM_AGENT_ROUTING_KEY,dev.RoutingKey):
			break
	}
}

var tmpPlayerInfo *data.PlayerInfo

func handleConnectorReceiveFromGame(payload *data.BasePayload) {
	switch payload.CommandAction {
	case data.OnCmdServerDealCards:
		tmpPlayerInfo = &data.PlayerInfo{
			ID:        1,
			Name:      "test",
			HandCards: []int{},
		}
		item, _ := helper.ConvertToObject[data.CmdServerDealCards](payload.Data)
		tmpPlayerInfo.HandCards = item.Cards
	case data.OnCmdServerCurrentPlayer:
		item, _ := helper.ConvertToObject[data.CmdServerCurrentPlayer](payload.Data)
		if item.PlayerID == 1 {
			data.LogB("[玩家角度] 當前應由[我]出牌")
			action := DispachAgentHandler(&item.GameRecord,tmpPlayerInfo)
			DispatchGameHandler(action)
		} else {
			data.LogB(fmt.Sprintf("[玩家角度] 當前應由玩家[%d]出牌", item.PlayerID))
		}
	case data.OnCmdServerPlayerAction:
		item, _ := helper.ConvertToObject[data.CmdServerPlayerAction](payload.Data)
		cardTypeName := consts.GetCardTypeName(item.CardType)
		cardNameList := consts.GetCardNameList(item.Cards)
		if item.PlayerID == 1 {
			data.LogB(fmt.Sprintf("[玩家角度] 我出牌 牌型[%s] 牌[%+v]：%+v", cardTypeName, cardNameList))
		} else {
			data.LogB(fmt.Sprintf("[玩家角度] 玩家[%d]出牌 牌型[%s] 牌[%+v]：%+v", item.PlayerID, cardTypeName, cardNameList))
		}
	case data.OnCmdServerGameOver:
		item, _ := helper.ConvertToObject[data.CmdServerGameOver](payload.Data)
		for k, v := range item.Status {
			if k == 1 {
				data.LogB(fmt.Sprintf("[玩家角度] 我手牌剩 [%d]", v))
			} else {
				data.LogB(fmt.Sprintf("[玩家角度] 玩家[%d]手牌剩 [%d]", k, v))
			}
		}
	}
}

func DispachAgentHandler(gameRecord *data.GameRecord, info *data.PlayerInfo) *data.PlayerAction {
	transfer := GetTransferMQ()
	msgID := helper.GetUniqueID()
	reply, _ := helper.GetWorkHelper().MakeRequest(msgID, func() {
		payload := &data.AIPayloadRequest{
			GameRecord: *gameRecord,
			Info:       *info,
		}
		payloadStr, _ := helper.ConvertToData(payload)
		payloadBase := data.BasePayload{
			MainAction:    0,
			CommandAction: 0,
			Data:          payloadStr,
		}
		basePayloadStr, _ := helper.ConvertToData(&payloadBase)
		transfer.Publish(consts.AGENT_RECEIVE_FROM_CONNECT_ROUTING_KEY, basePayloadStr, msgID)
	})

	p1, _ := helper.ConvertToObject[data.BasePayload](reply.Payload)
	p2, _ := helper.ConvertToObject[data.AIPayloadResponse](p1.Data)
	return &p2.Action
}

func DispatchGameHandler(action *data.PlayerAction) {
	transfer := GetTransferMQ()
	msgID := helper.GetUniqueID()
	helper.GetWorkHelper().MakeRequest(msgID, func() {
		payload := &data.CmdClientPlayerAction{
			PlayerID: action.PlayerID,
			IsPass:   action.IsPass,
			CardType: action.CardType,
			Cards:    action.Cards,
			Reason:   action.Reason,
		}
		payloadStr, _ := helper.ConvertToData(payload)
		payloadBase := data.BasePayload{
			MainAction:    0,
			CommandAction: 0,
			Data:          payloadStr,
		}
		basePayloadStr, _ := helper.ConvertToData(&payloadBase)
		transfer.Publish(consts.GAME_RECEIVE_FROM_CONNECTOR_ROUTING_KEY, basePayloadStr, msgID)
	})
}

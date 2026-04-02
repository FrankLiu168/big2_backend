package connector

import (
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type ConnectorTransferGameHelper struct {
	Transfer *ConnectorTransferMQ
}

func NewConnectorTransferGameHelper(transfer *ConnectorTransferMQ) *ConnectorTransferGameHelper {
	return &ConnectorTransferGameHelper{
		Transfer: transfer,
	}
}

func (c *ConnectorTransferGameHelper) HandleConnectMessage(dev *amqp091.Delivery) {
	payload, _ := helper.ConvertToObject[data.BasePayload](string(dev.Body))
	switch payload.CommandAction {
	case data.CommandAction(data.OnCmdServerCurrentPlayer):
		replyToGame(c.Transfer,dev,payload)
	case data.CommandAction(data.OnCmdServerDealCards):
		notifyFromGame(c.Transfer, dev, payload)
	case data.CommandAction(data.OnCmdServerPlayerAction):
		notifyFromGame(c.Transfer, dev, payload)
	case data.CommandAction(data.OnCmdServerGameOver):
		notifyFromGame(c.Transfer, dev, payload)
	}
}

func replyToGame(transfer *ConnectorTransferMQ, dev *amqp091.Delivery, payload *data.BasePayload) {
	data.LogD("Connector","replyToGame")
	switch payload.CommandAction {
	case data.OnCmdServerCurrentPlayer:

		item, _ := helper.ConvertToObject[data.CmdServerCurrentPlayer](payload.Data)
		if item.PlayerID == 1 {
			data.LogB("[玩家角度] 當前應由[我]出牌")
			action := MakeRequestToAgent(&item.GameRecord, tmpPlayerInfo)
			SendReplyToPlayer(action, dev.MessageId)
		} else {
			data.LogB(fmt.Sprintf("[玩家角度] 當前應由玩家[%d]出牌", item.PlayerID))
		}
	}
}

var tmpPlayerInfo *data.PlayerInfo

func notifyFromGame(transfer *ConnectorTransferMQ, dev *amqp091.Delivery, payload *data.BasePayload) {
	switch payload.CommandAction {
	case data.OnCmdServerDealCards:
		tmpPlayerInfo = &data.PlayerInfo{
			ID:        1,
			Name:      "test",
			HandCards: []int{},
		}
		item, _ := helper.ConvertToObject[data.CmdServerDealCards](payload.Data)
		tmpPlayerInfo.HandCards = item.Cards

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

func MakeRequestToAgent(gameRecord *data.GameRecord, info *data.PlayerInfo) *data.PlayerAction {
	data.LogD("Connector","MakeRequestToAgent")
	transfer := GetTransferMQ()
	msgID := helper.GetUniqueID()
	reply, _ := helper.GetConnectorWork().MakeRequest(msgID, func() {
		payload := &data.AIPayloadRequest{
			GameRecord: *gameRecord,
			Info:       *info,
		}
		payloadStr, _ := helper.ConvertToData(payload)
		payloadBase := data.BasePayload{
			MainAction:    0,
			CommandAction: data.CommandAction(data.InAIPayloadRequest),
			Data:          payloadStr,
		}
		basePayloadStr, _ := helper.ConvertToData(&payloadBase)
		data.LogD("Connector Publish","ROUTING.AGENT.FROM_CONNECTOR")
		transfer.Publish(consts.ROUTING.AGENT.FROM_CONNECTOR, basePayloadStr, msgID, msgID)
	})
	data.LogD("Connector","MakeRequestAgent Done")
	p1, _ := helper.ConvertToObject[data.BasePayload](reply.Payload)
	p2, _ := helper.ConvertToObject[data.AIPayloadResponse](p1.Data)
	return &p2.Action
}

func SendReplyToPlayer(action *data.PlayerAction, toID string) {
	data.LogD("Connector","SendReplyToPlayer")
	msgID := helper.GetUniqueID()
	transfer := GetTransferMQ()
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
		CommandAction: data.OnCmdClientPlayerAction,
		Data:          payloadStr,
	}
	basePayloadStr, _ := helper.ConvertToData(&payloadBase)
	data.LogD("Connector Publish","ROUTING.GAME.FROM_CONNECTOR")
	transfer.Publish(consts.ROUTING.GAME.FROM_CONNECTOR, basePayloadStr, msgID, toID)

}

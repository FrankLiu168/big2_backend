package logic

import (
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"
	"encoding/json"
	"log"
	"sort"
)

type Player struct {
	IsAI       bool
	Identifier string
	Info       *data.PlayerInfo
	Transfer   *GameTransferMQ
}

func NewPlayer(id int, name string, isAI bool, transfer *GameTransferMQ) *Player {
	info := data.NewPlayerInfo(id, name)
	return &Player{
		IsAI:       isAI,
		Identifier: "",
		Info:       info,
		Transfer:   transfer,
	}
}

func (p *Player) GetHandCards() []int {
	return p.Info.HandCards
}

func (p *Player) SetHandCards(cards []int) {
	sort.Slice(cards, func(i, j int) bool {
		rankI, rankJ := cards[i]%100, cards[j]%100
		if rankI != rankJ {
			return rankI < rankJ // 點數小的在前
		}
		return cards[i] < cards[j] // 點數相同，比完整 ID（或花色）
	})
	p.Info.HandCards = cards
}

func (p *Player) TakeOutHandCards(cards []int) (bool, []int) {
	handSet := make(map[int]bool, len(p.Info.HandCards))
	for _, card := range p.Info.HandCards {
		handSet[card] = true
	}

	for _, card := range cards {
		if !handSet[card] {
			return false, nil
		}
	}

	toRemove := make(map[int]bool, len(cards))
	for _, card := range cards {
		toRemove[card] = true
	}

	newHand := p.Info.HandCards[:0] // 重用底層陣列
	for _, card := range p.Info.HandCards {
		if !toRemove[card] {
			newHand = append(newHand, card)
		}
	}
	p.Info.HandCards = newHand

	return true, cards
}

func (p *Player) PutInHandCards(cards []int) {
	newCards := append(p.Info.HandCards, cards...)
	p.SetHandCards(newCards)
}

func (p *Player) FindStartCard() bool {
	for _, card := range p.Info.HandCards {
		if card == consts.START_CARD {
			return true
		}
	}
	return false
}

func (p *Player) GetLeftCardCount() int {
	return len(p.Info.HandCards)
}

func (p *Player) DoStrategy(gameRecord *data.GameRecord) *data.PlayerAction {
	payload := data.CmdServerCurrentPlayer{
		PlayerID:   p.Info.ID,
		GameRecord: *gameRecord,
	}
	payloadStr, _ := helper.ConvertToData(&payload)
	payloadBase := data.BasePayload{
		CommandAction: data.OnCmdServerCurrentPlayer,
		Data:          payloadStr,
	}
	baseStr, _ := helper.ConvertToData(&payloadBase)
	msgID := helper.GetUniqueID()
	reply, _ := helper.GetGameWork().MakeRequest(msgID, func() {
		data.LogD("DoStratery Publish","ROUTING.CONNECTOR.FROM_GAME")
		p.Transfer.Publish(consts.ROUTING.CONNECTOR.FROM_GAME, baseStr, msgID, msgID)
	})
	pl, _ := helper.ConvertToObject[data.BasePayload](reply.Payload)
	pa, _ := helper.ConvertToObject[data.CmdClientPlayerAction](pl.Data)
	return &data.PlayerAction{
		PlayerID: pa.PlayerID,
		IsPass:   pa.IsPass,
		CardType: pa.CardType,
		Cards:    pa.Cards,
		Reason:   pa.Reason,
	}
}

func (p *Player) Strategy(gameRecord *data.GameRecord) *data.PlayerAction {
	wh := helper.GetGameWork()
	msgID := helper.GetUniqueID()
	reply, err := wh.MakeRequest(msgID, func() {
		server := GetTransferMQ()
		payload := data.AIPayloadRequest{
			GameRecord: *gameRecord,
			Info:       *p.Info,
		}
		payloadStr, _ := helper.ConvertToData(&payload)
		basePayload := data.BasePayload{
			CommandAction: data.CommandAction(data.InAIPayloadRequest),
			Data: payloadStr,
		}
		str, _ := helper.ConvertToData(&basePayload)
		server.Publish(consts.ROUTING.AGENT.FROM_GAME, str, msgID, msgID)
	})
	if err != nil {
		log.Println("Error:", err)
		return nil
	}
	baseRes := data.BasePayload{}
	_ = json.Unmarshal([]byte(reply.Payload), &baseRes)
	res := data.AIPayloadResponse{}
	_ = json.Unmarshal([]byte(baseRes.Data), &res)

	return &res.Action
}

func (p *Player) SetCommand(command *data.BasePayload) {
	if p.IsAI {
		return
	}
	msgID := helper.GetUniqueID()
	msg, _ := helper.ConvertToData(command)
	routingKey := consts.ROUTING.CONNECTOR.FROM_GAME
	p.Transfer.Publish(routingKey, msg, msgID, "")
}

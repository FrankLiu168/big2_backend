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
}

func NewPlayer(id int, name string, isAI bool) *Player {
	info := data.NewPlayerInfo(id, name)
	return &Player{
		IsAI:       isAI,
		Identifier: "",
		Info:       info,
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

func (p *Player) Strategy(gameRecord *data.GameRecord) *data.PlayerAction {
	wh := helper.GetWorkHelper()
	msgID := helper.GetUniqueID()
	reply,err := wh.MakeRequest(msgID,func() {
		server := GetTransferServer()
		payload := data.AIPayloadRequest{
			GameRecord: *gameRecord,
			Info:       *p.Info,
		}
		payloadBytes, _ := json.Marshal(payload)
		basePayload := data.BasePayload{
			Command: 1,
			Data:    string(payloadBytes),
		}
		b,_ := json.Marshal(basePayload)
		server.Publish(consts.AGENT_REQUEST_ROUTING_KEY, string(b), msgID)
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

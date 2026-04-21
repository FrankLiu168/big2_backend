package logic

import (
	cardrule "big2backend/shared/cardRule"
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"
	"encoding/json"
	"sort"
	"time"
)

type Player struct {
	IsAI       bool
	Identifier string
	Info       *data.PlayerInfo
	Transfer   *GameTransferMQ
	IsOnGame   bool
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

func (p *Player) DoStrategy(isFirst bool, replyID string, sleepTime int) *data.PlayerAction {
	reply, err := helper.GetWaitHelper().WaitWithTimeout(
		replyID, time.Duration(sleepTime)*time.Second)
	if err != nil {
		if isFirst {
			return p.GetMinCardAction()
		} else {
			return p.GetPass()
		}
	}
	cnnPayload := helper.ConvertToConnectorPayload(reply.Payload)
	resPayload := helper.ConvertToClientPayload[data.CmdClientPlayerAction](&cnnPayload.Data)
	action := data.PlayerAction{
		CardType: resPayload.CardType,
		Cards:    resPayload.Cards,
		IsPass:   resPayload.IsPass,
		PlayerID: p.Info.ID,
	}
	return &action
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
		str := helper.PackPayload(data.CommandAction(data.InAIPayloadRequest), 0, "", &payload)
		server.Publish(consts.ROUTING.AGENT.FROM_GAME, str, msgID, msgID)
	})
	if err != nil {
		if gameRecord.CurrentRound.IsFirst {
			return p.GetMinCardAction()
		} else {
			return p.GetPass()
		}
	}
	baseRes := data.BasePayload{}
	_ = json.Unmarshal([]byte(reply.Payload), &baseRes)
	res := data.AIPayloadResponse{}
	_ = json.Unmarshal([]byte(baseRes.Data), &res)

	return &res.Action
}

func (p *Player) GetPass() *data.PlayerAction {
	return &data.PlayerAction{
		PlayerID: p.Info.ID,
		IsPass:   true,
	}
}

func (p *Player) GetMinCardAction() *data.PlayerAction {
	handCards := p.GetHandCards()
	minCard := 0
	for _, card := range handCards {
		if minCard == 0 {
			minCard = card
		} else {
			b := cardrule.CompareSingle([]int{minCard}, []int{card})
			if b {
				minCard = card
			}
		}
	}
	action := data.PlayerAction{
		CardType: consts.CARD_TYPE_SINGLE,
		Cards:    []int{minCard},
		IsPass:   false,
		PlayerID: p.Info.ID,
	}
	return &action
}

func (p *Player) SetCommand(command *data.BasePayload) {
	if p.IsAI {
		return
	}
	command.Target = p.Identifier
	msgID := helper.GetUniqueID()
	msg, _ := helper.ConvertToData(command)
	routingKey := consts.ROUTING.CONNECTOR.FROM_GAME
	p.Transfer.Publish(routingKey, msg, msgID, "")
}

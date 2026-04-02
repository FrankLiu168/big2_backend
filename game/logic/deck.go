package logic

import (
	cardrule "big2backend/shared/cardRule"
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"
	"fmt"
	"strings"

	"math/rand"
	"time"
)

type Deck struct {
	Cards       []int
	Players     []Player
	PlayerChain *PlayerChain
	GameRecord  *data.GameRecord
	Checker     *cardrule.PlayerActionCheck
	Transfer    *GameTransferMQ
}

func (d *Deck) Init(players []Player,transfer *GameTransferMQ) {
	d.Players = players
	d.PlayerChain = NewPlayerChain(players)
	d.InitAndShuffle()
	d.GameRecord = data.NewGameRecord()
	d.Checker = &cardrule.PlayerActionCheck{}
	d.Transfer = transfer
}

func (d *Deck) InitAndShuffle() {
	d.Cards = []int{}
	for i := 1; i <= 13; i++ {
		for j := 1; j <= 4; j++ {
			d.Cards = append(d.Cards, j*100+i)
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

func (d *Deck) StartGame() {
	startPlayerID := 0
	for i, p := range d.Players {
		s := i * 13
		e := s + 13
		p.SetHandCards(d.Cards[s:e])
		isStarter := p.FindStartCard()
		if isStarter {
			startPlayerID = p.Info.ID
		}
		payload := data.CmdServerDealCards{Cards: p.GetHandCards()}
		payloadStr,_ := helper.ConvertToData(&payload)
		basePayload := data.BasePayload{
			CommandAction: data.OnCmdServerDealCards,
			Data:          payloadStr,
		}
		p.SetCommand(&basePayload)
	}
	d.PlayerChain.SetStartPlayer(startPlayerID)
	d.RoundStart()
}

func (d *Deck) RoundStart() {
	isDone := false
	for {
		if isDone {
			data.LogA("遊戲結束")
			data.LogA(d.getAllPlayerHandCards())
			return
		}
		isDone = d.RoundLoop()
		if !isDone {
			d.GameRecord.NewRound()
		}
	}
}

func (d *Deck) getAllPlayerHandCards() string {
	list := []string{}
	for _, player := range d.Players {
		str := "玩家[%d]手牌為%+v"
		handCardNames := consts.GetCardNameList(player.GetHandCards())
		str = fmt.Sprintf(str, player.Info.ID, handCardNames)
		list = append(list, str)
	}
	return strings.Join(list, "\n")
}

func (d *Deck) RoundLoop() bool {
	count := 0
	data.LogA(fmt.Sprintf("當前[%d]輪", d.GameRecord.CurrentRound.RoundNo))
	for {
		player := d.PlayerChain.GetCurrentPlayer()
		if !d.GameRecord.CurrentRound.IsFirst && player.Info.ID == d.GameRecord.DesktopPlayerAction.PlayerID {
			data.LogA("換新的一輪")
			return false
		}
		data.LogA(fmt.Sprintf("當前玩家[%d]", player.Info.ID))
		data.LogA(fmt.Sprintf("手牌 %+v", consts.GetCardNameList(player.GetHandCards())))
		var action *data.PlayerAction
		if player.IsAI {
			action = player.Strategy(d.GameRecord)
			if !action.IsPass {
				isOk, why := d.Checker.IsActionValid(d.GameRecord, action, player.Info)
				if !isOk {
					data.LogA(why)
					continue
				}
			}
		} else {
			data.LogD("不是AI","DoStrategy")
			action = player.DoStrategy(d.GameRecord)
		}

		if action.IsPass {
			data.LogA("策略：Pass")
		} else {
			cards := action.Cards
			cardNames := consts.GetCardNameList((cards))
			data.LogA(fmt.Sprintf("策略：%+v", cardNames))
		}
		data.LogA("策略說明：" + action.Reason)

		d.GameRecord.AddPlayerAction(*action)
		if !action.IsPass {
			d.GameRecord.DesktopPlayerAction = action
			d.GameRecord.CurrentRound.IsFirst = false
			player.TakeOutHandCards(action.Cards)
			count += 1
		}
		if player.GetLeftCardCount() == 0 {
			return true
		}

		d.PlayerChain.Next()
	}

}

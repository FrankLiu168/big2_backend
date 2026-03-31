package logic

import (
	"big2backend/shared/cardrule"
	"big2backend/shared/consts"
	"big2backend/shared/data"
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
}

func (d *Deck) Init(players []Player) {
	d.Players = players
	d.PlayerChain = NewPlayerChain(players)
	d.InitAndShuffle()
	d.GameRecord = data.NewGameRecord()
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
		//data.LogA(fmt.Sprintf("Player %d: %+v\n", i, p.GetHandCards()))
		isStarter := p.FindStartCard()
		if isStarter {
			startPlayerID = p.Info.ID
		}
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
	for _,player := range d.Players {
		str := "玩家[%d]手牌為%+v"
		handCardNames := consts.GetCardNameList(player.GetHandCards())
		str = fmt.Sprintf(str,player.Info.ID,handCardNames)
		list = append(list,str)
	}
	return strings.Join(list,"\n")
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
		data.LogA(fmt.Sprintf("手牌 %+v",consts.GetCardNameList(player.GetHandCards())))
		action := player.Strategy(d.GameRecord)
		if !action.IsPass {
			isOk, why := d.IsActionValid(d.GameRecord, action)
			if !isOk {
				data.LogA(why)
				continue
			}
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

func (d *Deck) IsActionValid(gameRecord *data.GameRecord, action *data.PlayerAction) (bool, string) {
	flag := d.CheckCardType(action)
	if !flag {
		return false, "v未知牌型"
	}
	flag = d.CheckCardTypeWithCards(action)
	if !flag {
		return false, "v不符合牌型"
	}
	if gameRecord.DesktopPlayerAction != nil {
		flag = d.CheckCardTypeWithDesktop(gameRecord, action)
		if !flag {
			return false, "v不符合桌面牌型"
		}
		currCardType := gameRecord.DesktopPlayerAction.CardType
		currCards := gameRecord.DesktopPlayerAction.Cards
		flag = d.CompareCards(currCardType, currCards, action)
	}
	return flag, "比桌面牌面小"
}

func (d *Deck) CheckCardType(action *data.PlayerAction) bool {
	if action.CardType == consts.CARD_TYPE_UNKNOWN {
		return false
	}
	return true
}
func (d *Deck) CheckCardTypeWithDesktop(gameRecord *data.GameRecord, action *data.PlayerAction) bool {
	if gameRecord.CurrentRound.IsFirst {
		return true
	}
	if action.CardType >= consts.CARD_TYPE_FOUR_OF_A_KIND && action.CardType >= gameRecord.DesktopPlayerAction.CardType {
		return true
	}
	if gameRecord.DesktopPlayerAction.CardType != action.CardType {
		return false
	}
	return true
}

func (d *Deck) CheckCardTypeWithCards(action *data.PlayerAction) bool {
	switch action.CardType {
	case consts.CARD_TYPE_STRAIGHT_FLUSH:
		if !cardrule.IsStraightFlush(action.Cards) {
			return false
		}
	case consts.CARD_TYPE_FOUR_OF_A_KIND:
		if !cardrule.IsFourOfAKind(action.Cards) {
			return false
		}
	case consts.CARD_TYPE_FULL_HOUSE:
		if !cardrule.IsFullHouse(action.Cards) {
			return false
		}
	case consts.CARD_TYPE_STRAIGHT:
		if !cardrule.IsStraight(action.Cards) {
			return false
		}
	case consts.CARD_TYPE_ONE_PAIR:
		if !cardrule.IsPair(action.Cards) {
			return false
		}
	case consts.CARD_TYPE_SINGLE:
		if !cardrule.IsSingle(action.Cards) {
			return false
		}
	}
	return true
}

func (d *Deck) CompareCards(desktopCardType consts.CardType, desktopCards []int, action *data.PlayerAction) bool {
	if action.CardType > desktopCardType && action.CardType >= consts.CARD_TYPE_FOUR_OF_A_KIND {
		return true
	}
	switch desktopCardType {
	case consts.CARD_TYPE_STRAIGHT_FLUSH:
		return cardrule.CompareStraightFlush(desktopCards, action.Cards)
	case consts.CARD_TYPE_FOUR_OF_A_KIND:
		return cardrule.CompareFourOfAKind(desktopCards, action.Cards)
	case consts.CARD_TYPE_FULL_HOUSE:
		return cardrule.CompareFullHouse(desktopCards, action.Cards)
	case consts.CARD_TYPE_STRAIGHT:
		return cardrule.CompareStraight(desktopCards, action.Cards)
	case consts.CARD_TYPE_ONE_PAIR:
		return cardrule.ComparePair(desktopCards, action.Cards)
	case consts.CARD_TYPE_SINGLE:
		return cardrule.CompareSingle(desktopCards, action.Cards)
	}
	return false
}

func (d *Deck)CheckHandCards(action *data.PlayerAction, player *Player) bool { 
	for _,card := range action.Cards {
		isMatch := false
		for _,handCard := range player.Info.HandCards {
			if card == handCard {
				isMatch = true
			}
			if !isMatch {
				return false
			}
		}
	}
	return true
}

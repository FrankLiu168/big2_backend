package cardrule

import (
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"strings"
)

type PlayerActionCheck struct {
}

func (p *PlayerActionCheck) IsActionValid(gameRecord *data.GameRecord, action *data.PlayerAction, info *data.PlayerInfo) (bool, string) {
	flag, msg := p.CheckCardType(action)
	if !flag {
		return false, msg
	}
	flag, msg = p.CheckHandCards(action, info)
	if !flag {
		return false, msg
	}
	flag, msg = p.CheckCardTypeWithCards(action)
	if !flag {
		return false, msg
	}
	if gameRecord.DesktopPlayerAction != nil {
		flag, msg = p.CheckCardTypeWithDesktop(gameRecord, action)
		if !flag {
			return false, msg
		}
		currCardType := gameRecord.DesktopPlayerAction.CardType
		currCards := gameRecord.DesktopPlayerAction.Cards
		flag, msg = p.CompareCards(currCardType, currCards, action)
		if !flag {
			return false, msg
		}
	}
	return true, ""
}



func (p *PlayerActionCheck) CheckCardType(action *data.PlayerAction) (bool, string) {
	if action.CardType == consts.CARD_TYPE_UNKNOWN {
		return false, "未知的牌型"
	}
	return true, ""
}
func (p *PlayerActionCheck) CheckCardTypeWithDesktop(gameRecord *data.GameRecord, action *data.PlayerAction) (bool, string) {
	if gameRecord.CurrentRound.IsFirst {
		return true, ""
	}
	if action.CardType >= consts.CARD_TYPE_FOUR_OF_A_KIND && action.CardType >= gameRecord.DesktopPlayerAction.CardType {
		return true, ""
	}
	if gameRecord.DesktopPlayerAction.CardType != action.CardType {
		return false, "與當前桌面牌型不一致"
	}
	return true, ""
}

func (p *PlayerActionCheck) CheckCardTypeWithCards(action *data.PlayerAction) (bool, string) {
	switch action.CardType {
	case consts.CARD_TYPE_STRAIGHT_FLUSH:
		if !IsStraightFlush(action.Cards) {
			return false, "所出的牌不是同花順"
		}
	case consts.CARD_TYPE_FOUR_OF_A_KIND:
		if !IsFourOfAKind(action.Cards) {
			return false, "所出的牌不是鐵支"
		}
	case consts.CARD_TYPE_FULL_HOUSE:
		if !IsFullHouse(action.Cards) {
			return false, "所出的牌不是葫蘆"
		}
	case consts.CARD_TYPE_STRAIGHT:
		if !IsStraight(action.Cards) {
			return false, "所出的牌不是順子"
		}
	case consts.CARD_TYPE_ONE_PAIR:
		if !IsPair(action.Cards) {
			return false, "所出的牌不是對子"
		}
	case consts.CARD_TYPE_SINGLE:
		if !IsSingle(action.Cards) {
			return false, "所出的牌不是單張"
		}
	}
	return true, ""
}

func (p *PlayerActionCheck) CompareCards(desktopCardType consts.CardType, desktopCards []int,
	action *data.PlayerAction) (bool, string) {

	if action.CardType > desktopCardType && action.CardType >= consts.CARD_TYPE_FOUR_OF_A_KIND {
		return true, ""
	}
	flag := false
	switch desktopCardType {
	case consts.CARD_TYPE_STRAIGHT_FLUSH:
		flag = CompareStraightFlush(desktopCards, action.Cards)
	case consts.CARD_TYPE_FOUR_OF_A_KIND:
		flag = CompareFourOfAKind(desktopCards, action.Cards)
	case consts.CARD_TYPE_FULL_HOUSE:
		flag = CompareFullHouse(desktopCards, action.Cards)
	case consts.CARD_TYPE_STRAIGHT:
		flag = CompareStraight(desktopCards, action.Cards)
	case consts.CARD_TYPE_ONE_PAIR:
		flag = ComparePair(desktopCards, action.Cards)
	case consts.CARD_TYPE_SINGLE:
		flag = CompareSingle(desktopCards, action.Cards)
	}
	if flag {
		return true, ""
	} else {
		return false, "所出的牌比桌面的牌小，请重新出牌，或pass"
	}

}

func (p *PlayerActionCheck) CheckHandCards(action *data.PlayerAction, info *data.PlayerInfo) (bool, string) {
	unMatchedCards := []int{}
	for _, card := range action.Cards {
		isMatch := false
		for _, handCard := range info.HandCards {
			if card == handCard {
				isMatch = true
				break
			}
		}
		if !isMatch {
				unMatchedCards = append(unMatchedCards, card)
			}
	}
	if len(unMatchedCards) > 0 {
		cardName := consts.GetCardNameList(unMatchedCards)
		names := strings.Join(cardName, ",")
		return false, "所出的牌" + names + "不在你的手牌里"
	}
	return true, ""
}

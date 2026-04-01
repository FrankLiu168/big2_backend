package ai

import (
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"strconv"


	"strings"
)


func MakeRoundStr(round *data.RoundRecord) string {
	actionStrList := []string{}
	for _, action := range round.PlayerActions {
		actionStr := "玩家[" + strconv.Itoa(action.PlayerID) + "]:"
		if action.IsPass {
			actionStr += "pass"
			actionStrList = append(actionStrList, actionStr)
			continue
		}
		actionStr += consts.GetCardTypeName(action.CardType)
		actionStr += "(" + strings.Join(consts.GetCardNameList(action.Cards), ",") + ")"
		actionStrList = append(actionStrList, actionStr)
	}
	return strings.Join(actionStrList, "\n")
}

func CreateSituation(prompts map[string]string, gameRecord *data.GameRecord, player *data.PlayerInfo) string {
	keys := []string{"{{RoundNO}}", "{{PlayerID}}", "{{HandCards}}", "{{UsedCards}}",
		"{{PlayerHandCardCount}}", "{{DesktopPlayerID}}", "{{DesktopCardType}}",
		"{{DesktopCards}}", "{{PlayerActionInRound}}", "{{HistoryRounds}}"}

	params := map[string]string{}
	params["{{RoundNO}}"] = strconv.Itoa(gameRecord.CurrentRound.RoundNo)
	params["{{PlayerID}}"] = strconv.Itoa(player.ID)
	params["{{HandCards}}"] = strings.Join(consts.GetCardNameList(player.HandCards), ",")
	if gameRecord.CurrentRound.IsFirst && gameRecord.CurrentRound.RoundNo == 1 {
		temp := prompts["SituationForStartPrompt"]
		for _, key := range keys {
			val, isExist := params[key]
			if isExist {
				temp = strings.Replace(temp, key, val, 1)
			}
		}
		return temp
	}
	
	usedCards := strings.Join(consts.GetCardNameList(gameRecord.UsedCards), ",")
	params["{{UsedCards}}"] = usedCards
	playerStrList := []string{}
	for playerID, count := range gameRecord.PlayerHandCardCount {
		playerStr := "玩家[" + strconv.Itoa(playerID) + "]:" + strconv.Itoa(count)
		playerStrList = append(playerStrList, playerStr)
	}
	params["{{PlayerHandCardCount}}"] = strings.Join(playerStrList, "\n")
	if gameRecord.CurrentRound.IsFirst {
		temp := prompts["SituationForFirstPrompt"]
		for _, key := range keys {
			val, isExist := params[key]
			if isExist {
				temp = strings.Replace(temp, key, val, 1)
			}
		}
		return temp
	}
	params["{{DesktopPlayerID}}"] = strconv.Itoa(gameRecord.DesktopPlayerAction.PlayerID)
	cardType := gameRecord.DesktopPlayerAction.CardType
	cardTypeName := consts.GetCardTypeName(cardType)
	params["{{DesktopCardType}}"] = cardTypeName
	cards := gameRecord.DesktopPlayerAction.Cards
	cardNames := consts.GetCardNameList(cards)
	params["{{DesktopCards}}"] = strings.Join(cardNames, ",")
	actionStr := MakeRoundStr(gameRecord.CurrentRound)
	params["{{PlayerActionInRound}}"] = actionStr
	roundStrList := []string{}
	for _, round := range gameRecord.RoundRecords {
		roundStr := MakeRoundStr(&round)
		roundStr = strings.Join([]string{"第" + strconv.Itoa(round.RoundNo) + "輪", roundStr}, "\n")
		roundStrList = append(roundStrList, roundStr)
	}
	params["{{HistoryRounds}}"] = strings.Join(roundStrList, "\n")

	temp := prompts["SituationPrompt"]
	for _, key := range keys {
		val, isExist := params[key]
		if isExist {
			temp = strings.Replace(temp, key, val, 1)
		}
	}
	return temp
}


// func (a *Agent) CheckCardType(gameRecord *data.GameRecord, action map[string]any) string {
// 	cardType := consts.GetCardTypeNo(action["CardType"].(string))
// 	if cardType == consts.CARD_TYPE_UNKNOWN {
// 		return "未知的牌型"
// 	}
// 	return ""
// }
// func (a *Agent) CheckCardTypeWithDesktop(gameRecord *data.GameRecord, action map[string]any) string {
// 	if gameRecord.CurrentRound.IsFirst {
// 		return ""
// 	}
// 	cardType := consts.GetCardTypeNo(action["CardType"].(string))
// 	if cardType >= consts.CARD_TYPE_FOUR_OF_A_KIND && cardType >= gameRecord.DesktopPlayerAction.CardType {
// 		return ""
// 	}
// 	if gameRecord.DesktopPlayerAction.CardType != cardType {
// 		return "與當前桌面牌型不一致"
// 	}
// 	return ""
// }

// func (a *Agent) CheckCardTypeWithCards(cardType consts.CardType, cards []int) string {
// 	switch cardType {
// 	case consts.CARD_TYPE_STRAIGHT_FLUSH:
// 		if !cardrule.IsStraightFlush(cards) {
// 			return "所出的牌不是同花順"
// 		}
// 	case consts.CARD_TYPE_FOUR_OF_A_KIND:
// 		if !cardrule.IsFourOfAKind(cards) {
// 			return "所出的牌不是鐵支"
// 		}
// 	case consts.CARD_TYPE_FULL_HOUSE:
// 		if !cardrule.IsFullHouse(cards) {
// 			return "所出的牌不是葫蘆"
// 		}
// 	case consts.CARD_TYPE_STRAIGHT:
// 		if !cardrule.IsStraight(cards) {
// 			return "所出的牌不是順子"
// 		}
// 	case consts.CARD_TYPE_ONE_PAIR:
// 		if !cardrule.IsPair(cards) {
// 			return "所出的牌不是對子"
// 		}
// 	case consts.CARD_TYPE_SINGLE:
// 		if !cardrule.IsSingle(cards) {
// 			return "所出的牌不是單張"
// 		}
// 	}
// 	return ""
// }

// func (a *Agent) CompareCards(desktopCardType consts.CardType, desktopCards []int,
// 	cardType consts.CardType, cards []int) string {
// 	if cardType > desktopCardType && cardType >= consts.CARD_TYPE_FOUR_OF_A_KIND {
// 		return ""
// 	}
// 	switch desktopCardType {
// 	case consts.CARD_TYPE_STRAIGHT_FLUSH:
// 		if !cardrule.CompareStraightFlush(desktopCards, cards) {
// 			return "所出的牌比桌面的牌小，请重新出牌，或pass"
// 		}
// 	case consts.CARD_TYPE_FOUR_OF_A_KIND:
// 		if !cardrule.CompareFourOfAKind(desktopCards, cards) {
// 			return "所出的牌比桌面的牌小，请重新出牌，或pass"
// 		}
// 	case consts.CARD_TYPE_FULL_HOUSE:
// 		if !cardrule.CompareFullHouse(desktopCards, cards) {
// 			return "所出的牌比桌面的牌小，请重新出牌，或pass"
// 		}
// 	case consts.CARD_TYPE_STRAIGHT:
// 		if !cardrule.CompareStraight(desktopCards, cards) {
// 			return "所出的牌比桌面的牌小，请重新出牌，或pass"
// 		}
// 	case consts.CARD_TYPE_ONE_PAIR:
// 		if !cardrule.ComparePair(desktopCards, cards) {
// 			return "所出的牌比桌面的牌小，请重新出牌，或pass"
// 		}
// 	case consts.CARD_TYPE_SINGLE:
// 		if !cardrule.CompareSingle(desktopCards, cards) {
// 			return "所出的牌比桌面的牌小，请重新出牌，或pass"
// 		}
// 	}
// 	return ""
// }

// func (a *Agent) CheckHandCards(cards []int, player *data.PlayerInfo) string {
// 	for _, card := range cards {
// 		isMatch := false
// 		for _, handCard := range player.HandCards {
// 			if card == handCard {
// 				isMatch = true
// 			}
// 		}
// 		if !isMatch {
// 			cardName := consts.GetCardName(card)
// 			return "手牌裡並沒有" + cardName
// 		}
// 	}
// 	return ""
// }

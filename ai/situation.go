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


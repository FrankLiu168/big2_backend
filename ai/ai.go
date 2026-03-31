package ai

import (
	"big2backend/shared/cardrule"
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"strconv"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Agent struct {
	ApiKey                  string
	ApiUrl                  string
	ApiModel                string
	GameRulePrompt          string
	AIRulePrompt            string
	SituationPrompt         string
	SituationForFirstPrompt string
	SituationForStartPrompt string
}

func NewAgent() *Agent {
	agent := &Agent{
		ApiKey:   os.Getenv("API_KEY"),
		ApiUrl:   os.Getenv("API_URL"),
		ApiModel: os.Getenv("AI_MODEL"),
	}
	agent.GameRulePrompt = agent.LoadRule("gameRule")
	agent.AIRulePrompt = agent.LoadRule("aiRule")
	agent.SituationPrompt = agent.LoadRule("situation")
	agent.SituationForFirstPrompt = agent.LoadRule("situationForFirst")
	agent.SituationForStartPrompt = agent.LoadRule("situationForStart")
	return agent
}

func (a *Agent) LoadRule(ruleType string) string {
	fileName := ""
	switch ruleType {
	case "gameRule":
		fileName = "gameRule.txt"
	case "aiRule":
		fileName = "aiRule.txt"
	case "situation":
		fileName = "situation.txt"
	case "situationForFirst":
		fileName = "situationForFirst.txt"
	case "situationForStart":
		fileName = "situationForStart.txt"
	}
	path := "doc/" + fileName
	data, _ := os.ReadFile(path)
	str := string(data)
	str = strings.Split(str, "*****")[1]
	return str
}

func (a *Agent) Strategy(gameRecord *data.GameRecord, player *data.PlayerInfo) *data.PlayerAction {
	prompt := a.CreateSituation(gameRecord, player)
	replyList := []string{}
	callCount := 0
	for {
		if callCount >= 3{
			return a.SetPlayerAction(player.ID, map[string]any{"IsPass": true,"Reason": "API多次錯誤"})
		}
		//data.LogA(fmt.Sprintf("call AI Agent: %+v",replyList))
		reply, err := a.CallAIApi(prompt, replyList)
		callCount += 1
		reply = strings.ReplaceAll(reply, "`", "")		

		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}
		action := map[string]any{}
		err = json.Unmarshal([]byte(reply), &action)
		if err != nil {
			replyList = append(replyList, reply)
			replyList = append(replyList, "json 格式錯誤")
			continue
		}

		if action["IsPass"].(bool) == true {
			return a.SetPlayerAction(player.ID, action)
		}

		r := a.CheckCardType(gameRecord, action)
		if r != "" {
			replyList = append(replyList, reply)
			replyList = append(replyList, r)
			continue
		}

		r = a.CheckCardTypeWithDesktop(gameRecord, action)
		if r != "" {
			replyList = append(replyList, reply)
			replyList = append(replyList, r)
			continue
		}

		cardType := consts.GetCardTypeNo(action["CardType"].(string))
		cards := []int{}
		for _, c := range action["Cards"].([]any) {
			cards = append(cards, consts.GetCardNumber(c.(string)))
		}
		//cards := consts.GetCardNumberList(action["Cards"].([]string))

		r = a.CheckCardTypeWithCards(cardType, cards)
		if r != "" {
			replyList = append(replyList, reply)
			replyList = append(replyList, r)
			continue
		}

		r = a.CheckHandCards(cards, player)
		if r != "" {
			replyList = append(replyList, reply)
			replyList = append(replyList, r)
			continue
		}
		if gameRecord.DesktopPlayerAction != nil {
			r = a.CheckCardTypeWithDesktop(gameRecord, action)
			if r != "" {
				replyList = append(replyList, reply)
				replyList = append(replyList, r)
				continue
			}

			r = a.CompareCards(gameRecord.DesktopPlayerAction.CardType,
				gameRecord.DesktopPlayerAction.Cards,
				cardType, cards)
			if r != "" {
				replyList = append(replyList, reply)
				replyList = append(replyList, r)
				continue
			}
		}
		playerAction := a.SetPlayerAction(player.ID, action)
		return playerAction
	}
}

func (a *Agent) SetPlayerAction(playerID int, action map[string]any) *data.PlayerAction {
	playerAction := data.PlayerAction{}
	playerAction.PlayerID = playerID
	playerAction.IsPass = action["IsPass"].(bool)
	playerAction.Reason = action["Reason"].(string)
	if playerAction.IsPass {
		return &playerAction
	}
	playerAction.CardType = consts.GetCardTypeNo(action["CardType"].(string))
	cardNames := []string{}
	for _, c := range action["Cards"].([]any) {
		cardNames = append(cardNames, c.(string))
	}
	playerAction.Cards = consts.GetCardNumberList(cardNames)
	playerAction.Reason = action["Reason"].(string)
	return &playerAction
}

func (a *Agent) MakeRoundStr(round *data.RoundRecord) string {
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

func (a *Agent) CreateSituation(gameRecord *data.GameRecord, player *data.PlayerInfo) string {
	keys := []string{"{{RoundNO}}", "{{PlayerID}}", "{{HandCards}}", "{{UsedCards}}",
		"{{PlayerHandCardCount}}", "{{DesktopPlayerID}}", "{{DesktopCardType}}",
		"{{DesktopCards}}", "{{PlayerActionInRound}}", "{{HistoryRounds}}"}

	params := map[string]string{}
	params["{{RoundNO}}"] = strconv.Itoa(gameRecord.CurrentRound.RoundNo)
	params["{{PlayerID}}"] = strconv.Itoa(player.ID)
	params["{{HandCards}}"] = strings.Join(consts.GetCardNameList(player.HandCards), ",")
	if gameRecord.CurrentRound.IsFirst && gameRecord.CurrentRound.RoundNo == 1 {
		temp := a.SituationForStartPrompt
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
		temp := a.SituationForFirstPrompt
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
	actionStr := a.MakeRoundStr(gameRecord.CurrentRound)
	params["{{PlayerActionInRound}}"] = actionStr
	roundStrList := []string{}
	for _, round := range gameRecord.RoundRecords {
		roundStr := a.MakeRoundStr(&round)
		roundStr = strings.Join([]string{"第" + strconv.Itoa(round.RoundNo) + "輪", roundStr}, "\n")
		roundStrList = append(roundStrList, roundStr)
	}
	params["{{HistoryRounds}}"] = strings.Join(roundStrList, "\n")

	temp := a.SituationPrompt
	for _, key := range keys {
		val, isExist := params[key]
		if isExist {
			temp = strings.Replace(temp, key, val, 1)
		}
	}
	return temp
}

func (a *Agent) CallAIApi(prompt string, replyList []string) (string, error) {
	systemContent := strings.Join([]string{a.AIRulePrompt, a.GameRulePrompt}, "\n###") // 或直接寫規則字串
	payload := map[string]interface{}{
		"model": a.ApiModel,
		"messages": []map[string]string{
			{"role": "system", "content": systemContent},
		},
		"temperature": 0.7,
		"max_tokens":  250,
	}
	role := 0
	for _, reply := range replyList {
		if role == 0 {
			payload["messages"] = append(payload["messages"].([]map[string]string), map[string]string{"role": "assistant", "content": reply})
			role = 1
		} else {
			payload["messages"] = append(payload["messages"].([]map[string]string), map[string]string{"role": "user", "content": reply})
			role = 0
		}
	}
	payload["messages"] = append(payload["messages"].([]map[string]string), map[string]string{"role": "user", "content": prompt})

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("序列化 JSON 失敗: %w", err)
	}

	req, err := http.NewRequest("POST", a.ApiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("建立請求失敗: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("發送請求失敗: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("讀取回應失敗: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("API 回傳錯誤: %d, body: %s", resp.StatusCode, string(body))
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析回應失敗: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("回應中無 choices")
	}

	return result.Choices[0].Message.Content, nil
}

func (a *Agent) CheckCardType(gameRecord *data.GameRecord, action map[string]any) string {
	cardType := consts.GetCardTypeNo(action["CardType"].(string))
	if cardType == consts.CARD_TYPE_UNKNOWN {
		return "未知的牌型"
	}
	return ""
}
func (a *Agent) CheckCardTypeWithDesktop(gameRecord *data.GameRecord, action map[string]any) string {
	if gameRecord.CurrentRound.IsFirst {
		return ""
	}
	cardType := consts.GetCardTypeNo(action["CardType"].(string))
	if cardType >= consts.CARD_TYPE_FOUR_OF_A_KIND && cardType >= gameRecord.DesktopPlayerAction.CardType {
		return ""
	}
	if gameRecord.DesktopPlayerAction.CardType != cardType {
		return "與當前桌面牌型不一致"
	}
	return ""
}

func (a *Agent) CheckCardTypeWithCards(cardType consts.CardType, cards []int) string {
	switch cardType {
	case consts.CARD_TYPE_STRAIGHT_FLUSH:
		if !cardrule.IsStraightFlush(cards) {
			return "所出的牌不是同花順"
		}
	case consts.CARD_TYPE_FOUR_OF_A_KIND:
		if !cardrule.IsFourOfAKind(cards) {
			return "所出的牌不是鐵支"
		}
	case consts.CARD_TYPE_FULL_HOUSE:
		if !cardrule.IsFullHouse(cards) {
			return "所出的牌不是葫蘆"
		}
	case consts.CARD_TYPE_STRAIGHT:
		if !cardrule.IsStraight(cards) {
			return "所出的牌不是順子"
		}
	case consts.CARD_TYPE_ONE_PAIR:
		if !cardrule.IsPair(cards) {
			return "所出的牌不是對子"
		}
	case consts.CARD_TYPE_SINGLE:
		if !cardrule.IsSingle(cards) {
			return "所出的牌不是單張"
		}
	}
	return ""
}

func (a *Agent) CompareCards(desktopCardType consts.CardType, desktopCards []int,
	cardType consts.CardType, cards []int) string {
	if cardType > desktopCardType && cardType >= consts.CARD_TYPE_FOUR_OF_A_KIND {
		return ""
	}
	switch desktopCardType {
	case consts.CARD_TYPE_STRAIGHT_FLUSH:
		if !cardrule.CompareStraightFlush(desktopCards, cards) {
			return "所出的牌比桌面的牌小，请重新出牌，或pass"
		}
	case consts.CARD_TYPE_FOUR_OF_A_KIND:
		if !cardrule.CompareFourOfAKind(desktopCards, cards) {
			return "所出的牌比桌面的牌小，请重新出牌，或pass"
		}
	case consts.CARD_TYPE_FULL_HOUSE:
		if !cardrule.CompareFullHouse(desktopCards, cards) {
			return "所出的牌比桌面的牌小，请重新出牌，或pass"
		}
	case consts.CARD_TYPE_STRAIGHT:
		if !cardrule.CompareStraight(desktopCards, cards) {
			return "所出的牌比桌面的牌小，请重新出牌，或pass"
		}
	case consts.CARD_TYPE_ONE_PAIR:
		if !cardrule.ComparePair(desktopCards, cards) {
			return "所出的牌比桌面的牌小，请重新出牌，或pass"
		}
	case consts.CARD_TYPE_SINGLE:
		if !cardrule.CompareSingle(desktopCards, cards) {
			return "所出的牌比桌面的牌小，请重新出牌，或pass"
		}
	}
	return ""
}

func (a *Agent) CheckHandCards(cards []int, player *data.PlayerInfo) string {
	for _, card := range cards {
		isMatch := false
		for _, handCard := range player.HandCards {
			if card == handCard {
				isMatch = true
			}
		}
		if !isMatch {
			cardName := consts.GetCardName(card)
			return "手牌裡並沒有" + cardName
		}
	}
	return ""
}

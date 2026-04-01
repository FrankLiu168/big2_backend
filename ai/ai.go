package ai

import (
	cardrule "big2backend/shared/cardRule"
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"

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
	Checker                 *cardrule.PlayerActionCheck
}

func NewAgent() *Agent {
	agent := &Agent{
		ApiKey:   os.Getenv("API_KEY"),
		ApiUrl:   os.Getenv("API_URL"),
		ApiModel: os.Getenv("AI_MODEL"),
		Checker:  &cardrule.PlayerActionCheck{},
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

func (a *Agent) Strategy(gameRecord *data.GameRecord, info *data.PlayerInfo) *data.PlayerAction {
	promptMap := map[string]string{
		"SituationPrompt":         a.SituationPrompt,
		"SituationForFirstPrompt": a.SituationForFirstPrompt,
		"SituationForStartPrompt": a.SituationForStartPrompt,
	}
	prompt := CreateSituation(promptMap, gameRecord, info)
	replyList := []string{}
	callCount := 0
	for {
		if callCount >= 3 {
			return &data.PlayerAction{
				PlayerID: info.ID,
				IsPass:   true,
				CardType: consts.CARD_TYPE_UNKNOWN,
				Cards:    []int{},
				Reason:   "API錯誤3次",
			}
		} else if len(replyList) > 0 {
			print("\n---replylist----\n")
			for _,s := range replyList {
				print(s + "\n")
			}
		}
		reply, err := a.CallAIApi(prompt, replyList)
		callCount += 1
		reply = helper.GetJsonPart(reply)
		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}
		type AIAction struct {
			IsPass   bool     `json:"isPass"`
			CardType string   `json:"cardType"`
			Cards    []string `json:"cards"`
			Reason   string   `json:"reason"`
		}
		aiAction := AIAction{}
		action := data.PlayerAction{}
		err = json.Unmarshal([]byte(reply), &aiAction)
		if err != nil {
			println("----json-----")
			print(err.Error())
			print(reply)
			replyList = append(replyList, reply)
			replyList = append(replyList, "json 格式錯誤")
			continue
		}
		action.PlayerID = info.ID
		action.CardType = consts.GetCardTypeNo(aiAction.CardType)
		action.Cards = consts.GetCardNumberList(aiAction.Cards)
		action.IsPass = aiAction.IsPass
		action.Reason = aiAction.Reason

		flag, msg := a.Checker.IsActionValid(gameRecord, &action, info)
		if !flag {
			replyList = append(replyList, reply)
			replyList = append(replyList, msg)
			continue
		}

		return &action
	}
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

package data

import (
	"big2backend/shared/consts"
)

type CommandAction int

const (
	InAIPayloadRequest  CommandAction = 101
	InAIPayloadResponse CommandAction = 102

	OnCmdServerCurrentPlayer CommandAction = 201
	OnCmdServerDealCards     CommandAction = 202
	OnCmdServerPlayerAction  CommandAction = 203
	OnCmdServerGameOver      CommandAction = 204
	OnCmdServerNewRound      CommandAction = 205

	OnCmdClientPlayerAction CommandAction = 301
	OnCmdClientReady        CommandAction = 302
)

type BasePayload struct {
	CommandAction CommandAction `json:"commandAction"`
	Target        string        `json:"target"`
	Data          string        `json:"data"`
}

type AIPayloadRequest struct {
	GameRecord GameRecord `json:"gameRecord"`
	Info       PlayerInfo `json:"info"`
}

type AIPayloadResponse struct {
	Action PlayerAction `json:"action"`
}
type CmdServerNewRound struct {
	RoundID  int `json:"roundID"`
	TakeTime int `json:"takeTime"`
}

type CmdClientReady struct {
	PlayerID int    `json:"playerID"`
	ReplyID  string `json:"replyID"`
}

type CmdServerDealCards struct {
	Cards    []int `json:"cards"`
	TakeTime int   `json:"takeTime"`
}

type CmdServerCurrentPlayer struct {
	//GameRecord GameRecord `json:"gameRecord"`
	ReplyID  string `json:"replyID"`
	PlayerID int    `json:"playerID"`
	TakeTime int    `json:"takeTime"`
}

type CmdClientPlayerAction struct {
	ReplyID  string          `json:"replyID"`
	PlayerID int             `json:"playerID"`
	IsPass   bool            `json:"isPass"`
	CardType consts.CardType `json:"cardType"`
	Cards    []int           `json:"cards"`
	Reason   string          `json:"reason"`
}

type CmdServerPlayerAction struct {
	PlayerID int             `json:"playerID"`
	IsPass   bool            `json:"isPass"`
	CardType consts.CardType `json:"cardType"`
	Cards    []int           `json:"cards"`
	TakeTime int             `json:"takeTime"`
}

type CmdServerGameOver struct {
	Status   map[int]int `json:"status"`
	TakeTime int         `json:"takeTime"`
}

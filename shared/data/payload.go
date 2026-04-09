package data

import (
	"big2backend/shared/consts"
)

type MainAction int
type CommandAction int

const (
	InAIPayloadRequest MainAction = iota + 1
	InAIPayloadResponse
)

const (
	OnCmdServerCurrentPlayer CommandAction = iota + 1
	OnCmdServerDealCards
	OnCmdClientPlayerAction
	OnCmdClientReady
	OnCmdServerPlayerAction
	OnCmdServerGameOver
	OnCmdServerNewRound
)

type BasePayload struct {
	MainAction    MainAction    `json : "mainAction"`
	CommandAction CommandAction `json : "command"`
	Target        string        `json : "target"`
	Data          string        `json : "data"`
}

type AIPayloadRequest struct {
	GameRecord GameRecord `json : "gameRecord"`
	Info       PlayerInfo `json : "info"`
}

type AIPayloadResponse struct {
	Action PlayerAction `json : "action"`
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

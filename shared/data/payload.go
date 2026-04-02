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
	OnCmdServerPlayerAction
	OnCmdServerGameOver
)

type BasePayload struct {
	MainAction    MainAction    `json : "mainAction"`
	CommandAction CommandAction `json : "command"`
	Data          string        `json : "data"`
}

type AIPayloadRequest struct {
	GameRecord GameRecord `json : "gameRecord"`
	Info       PlayerInfo `json : "info"`
}

type AIPayloadResponse struct {
	Action PlayerAction `json : "action"`
}

type CmdServerDealCards struct {
	Cards []int `json:"cards"`
}

type CmdServerCurrentPlayer struct {
	GameRecord GameRecord `json:"gameRecord"`
	PlayerID   int        `json:"playerID"`
}

type CmdClientPlayerAction struct {
	PlayerID int             `json:"playerID"`
	IsPass   bool            `json:"isPass"`
	CardType consts.CardType `json:"cardType"`
	Cards    []int           `json:"cards"`
	Reason   string          `json:"reason"`
}


type CmdServerPlayerAction struct {
	PlayerID int             `json:"playerID"`
	CardType consts.CardType `json:"cardType"`
	Cards    []int           `json:"cards"`
}

type CmdServerGameOver struct {
	Status map[int]int `json:"status"`
}

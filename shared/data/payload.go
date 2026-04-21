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
	OnCmdClientOnGame       CommandAction = 302

	OnCmdServerRoomInfo  CommandAction = 401
	OnCmdServerToStart   CommandAction = 402
	OnCmdServerEnterFail CommandAction = 403

	OnCmdClientEnterRoom CommandAction = 501
	OnCmdClientReady     CommandAction = 502
	OnCmdClientCancel    CommandAction = 503
	OnCmdClientLeaveRoom CommandAction = 504

	OnCmdConnectOffline CommandAction = 601
)

const (
	ToGame = 0
	ToRoom = 1
)

type ConnectorPayload struct {
	Identifier string            `json:"identifier"`
	Data       ClientBasePayload `json:"data"`
}

type ClientBasePayload struct {
	CommandAction CommandAction `json:"commandAction"`
	Data          string        `json:"data"`
	RoomID        int           `json:"roomID"`
	GameID        string        `json:"gameID"`
	IsBroadcast   bool          `json:"isBroadcast"`
}

type BasePayload struct {
	CommandAction    CommandAction `json:"commandAction"`
	CommandSubAction int           `json:"commandSubAction"`
	Target           string        `json:"target"`
	Data             string        `json:"data"`
}

type AIPayloadRequest struct {
	GameRecord GameRecord `json:"gameRecord"`
	Info       PlayerInfo `json:"info"`
}

type AIPayloadResponse struct {
	Action PlayerAction `json:"action"`
}

type CmdConnectOffline struct {
}

type CmdClientReady struct {
}

type CmdClientCancel struct {
}

type CmdClientEnterRoom struct {
}

type CmdClientLeaveRoom struct {
}

type CmdServerRoomInfo struct {
	Players []PlayerData `json:"players"`
}

type CmdServerToStart struct {
	GameID string `json:"gameID"`
}

type CmdServerEnterFail struct {
	FailID int `json:"failID"`
}

type CmdServerNewRound struct {
	RoundID  int `json:"roundID"`
	TakeTime int `json:"takeTime"`
}

type CmdServerDealCards struct {
	Players  []PlayerData `json:"players"`
	Cards    []int        `json:"cards"`
	TakeTime int          `json:"takeTime"`
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

type CmdClientOnGame struct {
}

type CmdServerGameOver struct {
	Status   map[string]int `json:"status"`
	TakeTime int            `json:"takeTime"`
}

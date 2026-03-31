package data

type BasePayload struct {
	Command int    `json : "command"`
	Data    string `json : "data"`
}

type AIPayloadRequest struct {
	GameRecord GameRecord   `json : "gameRecord"`
	Info       PlayerInfo   `json : "info"`
}

type AIPayloadResponse struct {
	Action PlayerAction `json : "action"`
}

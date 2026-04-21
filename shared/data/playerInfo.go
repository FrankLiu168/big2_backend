package data

type PlayerInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	HandCards []int  `json:"handCards"`
}

func NewPlayerInfo(id int, name string) *PlayerInfo {
	return &PlayerInfo{
		ID:   id,
		Name: name,
	}
}

type PlayerData struct {
	Identifier string `json:"identifier"`
	PlayerID   int    `json:"playerID"`
	PlayerName string `json:"playerName"`
	IsReady    bool   `json:"isReady"`
	IsOnline   bool   `json:"isOnline"`
}

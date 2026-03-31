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

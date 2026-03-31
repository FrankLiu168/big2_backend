package data

import (
	"big2backend/shared/consts"
)

type GameRecord struct {
	RoundRecords        []RoundRecord `json:"roundRecords"`
	CurrentRound        *RoundRecord  `json:"currentRound"`
	DesktopPlayerAction *PlayerAction `json:"desktopPlayerAction"`
	UsedCards           []int         `json:"usedCards"`
	PlayerHandCardCount map[int]int   `json:"playerHandCardCount"`
}

func NewGameRecord() *GameRecord {
	record := &GameRecord{
		RoundRecords: []RoundRecord{},
		CurrentRound: &RoundRecord{
			RoundNo:       1,
			IsFirst:       true,
			PlayerActions: []PlayerAction{},
		},
	}
	record.PlayerHandCardCount = make(map[int]int)
	record.PlayerHandCardCount[1] = 13
	record.PlayerHandCardCount[2] = 13
	record.PlayerHandCardCount[3] = 13
	record.PlayerHandCardCount[4] = 13
	return record
}

func (g *GameRecord) NewRound() {
	if g.CurrentRound != nil {
		g.RoundRecords = append(g.RoundRecords, *g.CurrentRound)
	}
	g.CurrentRound = &RoundRecord{
		RoundNo:       len(g.RoundRecords) + 1,
		IsFirst:       true,
		PlayerActions: []PlayerAction{},
	}
	g.DesktopPlayerAction = nil
}

func (g *GameRecord) AddPlayerAction(action PlayerAction) {
	g.CurrentRound.PlayerActions = append(g.CurrentRound.PlayerActions, action)
	g.CurrentRound.IsFirst = false
	if !action.IsPass {
		g.DesktopPlayerAction = &action
		g.UsedCards = append(g.UsedCards, action.Cards...)
		g.PlayerHandCardCount[action.PlayerID] -= len(action.Cards)
	}
}

type RoundRecord struct {
	RoundNo       int            `json:"roundNo"`
	IsFirst       bool           `json:"isFirst"`
	PlayerActions []PlayerAction `json:"playerActions"`
}

type PlayerAction struct {
	PlayerID int             `json:"playerID"`
	IsPass   bool            `json:"isPass"`
	CardType consts.CardType `json:"cardType"`
	Cards    []int           `json:"cards"`
	Reason   string          `json:"reason"`
}

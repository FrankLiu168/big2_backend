package logic


type PlayerChain struct {
	Players     map[int]*Player
	CurrentSeat int
}

func NewPlayerChain(players []Player) *PlayerChain {
	playerChain := &PlayerChain{
		Players: map[int]*Player{},
	}
	for _, player := range players {
		playerChain.Players[player.Info.ID] = &player
	}
	return playerChain
}

func (pc *PlayerChain) SetStartPlayer(playerID int) {
	for seat, player := range pc.Players {
		if player.Info.ID == playerID {
			pc.CurrentSeat = seat
			return
		}
	}
}

func (pc *PlayerChain) GetCurrentPlayer() *Player {
	return pc.Players[pc.CurrentSeat]
}

func (pc *PlayerChain) Next() {
	pc.CurrentSeat = pc.CurrentSeat + 1
	if pc.CurrentSeat > len(pc.Players)  {
		pc.CurrentSeat = 1
	}
}

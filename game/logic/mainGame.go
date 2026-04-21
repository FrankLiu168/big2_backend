package logic

import (
	"big2backend/shared/consts"
	"big2backend/shared/data"
	"big2backend/shared/helper"
	"fmt"
)

type MainGame struct {
	deckMap  map[int]*Deck
	roomMap  map[int]*Room
	transfer *GameTransferMQ
}

func NewMainGame(transfer *GameTransferMQ) *MainGame {
	return &MainGame{
		deckMap:  make(map[int]*Deck),
		roomMap:  make(map[int]*Room),
		transfer: transfer,
	}
}

func (m *MainGame) Init() {
	for i := 0; i < 4; i++ {
		deck := NewDeck(i + 1)
		m.deckMap[i+1] = deck
		room := NewRoom(i+1, deck, m)
		m.roomMap[i+1] = room
	}
}

func (m *MainGame) Send(payloadStr string) {
	m.transfer.Publish(consts.ROUTING.CONNECTOR.FROM_GAME, payloadStr, "", "")
}

func (m *MainGame) SendMuti(clientIDList []string, basePayload *data.BasePayload) {
	for _, clientID := range clientIDList {
		basePayload.Target = clientID
		str, _ := helper.ConvertToData(&basePayload)
		m.transfer.Publish(consts.ROUTING.CONNECTOR.FROM_GAME, str, "", "")
	}
}

func (m *MainGame) Start() {
	m.transfer.connectorHelper.SetGameListener(m.Listen)
}

func (m *MainGame) Listen(str string) {
	data.LogD("mainGame Listen", str)
	connectorPayload := helper.ConvertToConnectorPayload(str)
	switch true {
	case connectorPayload.Data.IsBroadcast:
		if connectorPayload.Data.CommandAction == data.OnCmdConnectOffline {
			for _, room := range m.roomMap {
				room.Offline(connectorPayload.Identifier)
			}
		}
		return
	case connectorPayload.Data.GameID != "":
		print("gameID", connectorPayload.Data.GameID)
		for _, deck := range m.deckMap {
			if deck.GetGameID() == connectorPayload.Data.GameID {
				deck.Listen(connectorPayload)
				return
			}
		}
	case connectorPayload.Data.GameID == "":
		room := m.roomMap[connectorPayload.Data.RoomID]
		print("room", fmt.Sprintf("%d %d , %+v", connectorPayload.Data.CommandAction, connectorPayload.Data.RoomID, room))
		if room == nil {
			return
		}
		room.Listen(connectorPayload)
	}
}

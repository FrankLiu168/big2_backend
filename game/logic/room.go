package logic

import (
	"big2backend/shared/data"
	"big2backend/shared/helper"
	"errors"
	"fmt"
)

const (
	MaxPlayerCount = 4

	ErrOverMaxPlayerCount = 1
	ErrPlayerNotExist     = 2
	ErrIdentifierExist    = 3
	ErrRoomIsLock         = 4
)

var (
	ErrPlayerNotFound = errors.New("玩家不存在")
)

var playerNameMap = map[string]string{
	"user-1": "愛玩家1",
	"user-2": "愛玩家2",
}

type Room struct {
	RoomID                int
	PlayerByIDMap         map[int]*data.PlayerData
	PlayerByIdentifierMap map[string]*data.PlayerData
	BindDeck              *Deck
	MainGame              *MainGame
}

func NewRoom(roomId int, deck *Deck, mainGame *MainGame) *Room {
	return &Room{
		RoomID:                roomId,
		PlayerByIDMap:         make(map[int]*data.PlayerData),
		PlayerByIdentifierMap: make(map[string]*data.PlayerData),
		BindDeck:              deck,
		MainGame:              mainGame,
	}
}

// ✅ 改用 PlayerByIdentifierMap，O(1)
func (r *Room) getPlayerByIdentifier(identifier string) (*data.PlayerData, bool) {
	player, ok := r.PlayerByIdentifierMap[identifier]
	return player, ok
}

// ✅ 改用 map 查詢，不再遍歷
func (r *Room) getPlayerIDByIdentifier(identifier string) (int, error) {
	if player, ok := r.getPlayerByIdentifier(identifier); ok {
		return player.PlayerID, nil
	}
	return 0, ErrPlayerNotFound
}

func (r *Room) Listen(connPayload *data.ConnectorPayload) {
	if r.IsLock() {
		r.sendFail(connPayload.Identifier, ErrRoomIsLock)
		return
	}

	identifier := connPayload.Identifier
	var playerID int
	var err error

	// 只有 enterRoom 不需要已有玩家
	if connPayload.Data.CommandAction != data.OnCmdClientEnterRoom {
		playerID, err = r.getPlayerIDByIdentifier(identifier)
		if err != nil {
			r.sendFail(identifier, ErrPlayerNotExist)
			return
		}
	}

	switch connPayload.Data.CommandAction {
	case data.OnCmdClientEnterRoom:
		r.enterRoom(identifier, &connPayload.Data)
	case data.OnCmdClientReady:
		r.ready(playerID, &connPayload.Data)
	case data.OnCmdClientCancel:
		r.cancel(playerID, &connPayload.Data)
	case data.OnCmdClientLeaveRoom:
		r.leaveRoom(playerID, &connPayload.Data)
	}
}

func (r *Room) sendRoomInfo() {
	replyPlayers := make([]data.PlayerData, 0, len(r.PlayerByIDMap))
	for _, item := range r.PlayerByIDMap {
		replyPlayers = append(replyPlayers, *item)
	}
	payload := data.CmdServerRoomInfo{
		Players: replyPlayers,
	}
	str, err := helper.ConvertToData(&payload)
	if err != nil {
		data.LogD("sendRoomInfo marshal failed", err.Error())
		return
	}
	basePayload := data.BasePayload{
		CommandAction:    data.OnCmdServerRoomInfo,
		CommandSubAction: 1,
		Data:             str,
	}
	identifierList := make([]string, 0, len(r.PlayerByIDMap))
	for _, p := range r.PlayerByIDMap {
		identifierList = append(identifierList, p.Identifier)
	}
	r.MainGame.SendMuti(identifierList, &basePayload)
}

func (r *Room) sendFail(identifier string, failID int) {
	if identifier == "" {
		// 防禦：避免空 identifier 發送
		return
	}
	failPayload := data.CmdServerEnterFail{
		FailID: failID,
	}
	str := helper.PackPayload(data.OnCmdServerEnterFail, 1, identifier, &failPayload)
	r.MainGame.Send(str)
}

func (r *Room) startGame() {
	gameID := helper.GetUniqueID()
	payload := data.CmdServerToStart{
		GameID: gameID,
	}
	str, err := helper.ConvertToData(&payload)
	if err != nil {
		data.LogD("startGame marshal failed", err.Error())
		return
	}
	basePayload := data.BasePayload{
		CommandAction:    data.OnCmdServerToStart,
		CommandSubAction: 1,
		Data:             str,
	}
	identifierList := make([]string, 0, len(r.PlayerByIDMap))
	for _, p := range r.PlayerByIDMap {
		identifierList = append(identifierList, p.Identifier)
	}
	r.MainGame.SendMuti(identifierList, &basePayload)
	players := []Player{}
	r.PlayerByIdentifierMap = map[string]*data.PlayerData{}
	transfer := GetTransferMQ()
	for i := 1; i <= 4; i++ {
		p, isExist := r.PlayerByIDMap[i]
		if isExist {
			player := NewPlayer(p.PlayerID, p.PlayerName, false, transfer)
			player.Identifier = p.Identifier
			players = append(players, *player)
			r.PlayerByIdentifierMap[p.Identifier] = &data.PlayerData{
				Identifier: p.Identifier,
				PlayerID:   p.PlayerID,
				PlayerName: p.PlayerName,
				IsOnline:   true,
			}
		} else {
			player := NewPlayer(i, fmt.Sprintf("夏使仁%d", i), true, transfer)
			players = append(players, *player)
			r.PlayerByIdentifierMap[fmt.Sprintf("%d", i)] = &data.PlayerData{
				Identifier: fmt.Sprintf("%d", i),
				PlayerID:   i,
				PlayerName: fmt.Sprintf("夏使仁%d", i),
				IsOnline:   true,
			}
		}
	}
	r.BindDeck.SetGameID(gameID)
	r.BindDeck.Init(players, r.PlayerByIdentifierMap, transfer)
	r.BindDeck.StartGame()
}

func (r *Room) ready(playerID int, cliPayload *data.ClientBasePayload) {
	player, exists := r.PlayerByIDMap[playerID]
	if !exists {
		return // 已經離開？忽略
	}
	player.IsReady = true

	// 檢查是否全部準備
	allReady := true
	for _, p := range r.PlayerByIDMap {
		if !p.IsReady {
			allReady = false
			break
		}
	}

	if allReady {
		r.startGame()
	} else {
		r.sendRoomInfo()
	}
}

func (r *Room) cancel(playerID int, cliPayload *data.ClientBasePayload) {
	if player, exists := r.PlayerByIDMap[playerID]; exists {
		player.IsReady = false
		r.sendRoomInfo()
	}
}

func (r *Room) IsLock() bool {
	return r.BindDeck.GetIsRunning()
}

func (r *Room) enterRoom(identifier string, cliPayload *data.ClientBasePayload) {
	if len(r.PlayerByIDMap) >= MaxPlayerCount {
		r.sendFail(identifier, ErrOverMaxPlayerCount)
		return
	}

	if _, exists := r.PlayerByIdentifierMap[identifier]; exists {
		r.sendFail(identifier, ErrIdentifierExist)
		return
	}

	playerName := "未知玩家"
	if name, ok := playerNameMap[identifier]; ok {
		playerName = name
	}

	// 分配最小可用 PlayerID (1~4)
	var playerID int
	for i := 1; i <= MaxPlayerCount; i++ {
		if _, used := r.PlayerByIDMap[i]; !used {
			playerID = i
			break
		}
	}
	if playerID == 0 {
		// 理論上不會發生，因已檢查人數上限
		r.sendFail(identifier, ErrOverMaxPlayerCount)
		return
	}

	player := &data.PlayerData{
		Identifier: identifier,
		PlayerID:   playerID,
		PlayerName: playerName,
		IsReady:    false,
		IsOnline:   true,
	}

	r.PlayerByIDMap[playerID] = player
	r.PlayerByIdentifierMap[identifier] = player
	r.sendRoomInfo()
}

func (r *Room) leaveRoom(playerID int, cliPayload *data.ClientBasePayload) {
	player, exists := r.PlayerByIDMap[playerID]
	if !exists {
		// 可能重複離開，忽略或記錄
		return
	}
	delete(r.PlayerByIDMap, playerID)
	delete(r.PlayerByIdentifierMap, player.Identifier)
	r.sendRoomInfo() // 通知剩餘玩家更新房間
}

func (r *Room) Offline(identifier string) {
	player, exists := r.PlayerByIdentifierMap[identifier]
	if !exists {
		return
	}
	delete(r.PlayerByIDMap, player.PlayerID)
	delete(r.PlayerByIdentifierMap, identifier)
	r.sendRoomInfo() // 建議也通知其他人
}

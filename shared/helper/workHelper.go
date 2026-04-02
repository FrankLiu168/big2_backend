package helper

import (
	"errors"
	"sync"
	"time"
)

// ====================
// Model（泛型）
// ====================

type Reply struct {
	MsgID   string
	Payload string
}

// ====================
// Work 結構（方便擴展）
// ====================

type work struct {
	ch        chan Reply
	createdAt time.Time
}

// ====================
// WorkHelper
// ====================

type workHelper struct {
	mu      sync.RWMutex
	works   map[string]*work
	timeout time.Duration // 預設 timeout
}

var agentWork *workHelper
var gameWork *workHelper
var connectorWork *workHelper

// 建立 Helper（帶預設 timeout）

func GetAgentWork() *workHelper {
	if agentWork != nil {
		return agentWork
	}
	agentWork = newWorkHelper()
	return agentWork
}

func GetGameWork() *workHelper {
	if gameWork != nil {
		return gameWork
	}
	gameWork = newWorkHelper()
	return gameWork
}

func GetConnectorWork() *workHelper {
	if connectorWork != nil {
		return connectorWork
	}
	connectorWork = newWorkHelper()
	return connectorWork
}

func newWorkHelper() *workHelper {
	timeout := 20 * time.Second
	work := &workHelper{
		works:   make(map[string]*work),
		timeout: timeout,
	}
	return work
}

// ====================
// MakeRequest
// ====================

// MakeRequest 發起請求（使用預設 timeout）
func (w *workHelper) MakeRequest(
	msgID string,
	job func(),
) (Reply, error) {
	return w.MakeRequestWithTimeout(msgID, w.timeout, job)
}

// MakeRequestWithTimeout 可自訂 timeout
func (w *workHelper) MakeRequestWithTimeout(
	msgID string,
	timeout time.Duration,
	job func(),
) (Reply, error) {

	replyCh := make(chan Reply, 1)

	// 註冊
	w.mu.Lock()
	if _, exists := w.works[msgID]; exists {
		w.mu.Unlock()
		return Reply{}, errors.New("duplicate msgID")
	}
	w.works[msgID] = &work{
		ch:        replyCh,
		createdAt: time.Now(),
	}
	w.mu.Unlock()

	// 執行 job（一定要 async）
	go job()

	// timeout 控制（避免 time.After leak）
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {

	case resp := <-replyCh:
		return resp, nil

	case <-timer.C:
		// timeout cleanup
		w.mu.Lock()
		delete(w.works, msgID)
		w.mu.Unlock()

		return Reply{}, errors.New("timeout")
	}
}

// ====================
// Reply
// ====================

// Reply 回傳結果（單一責任）
//
// 特性：
// - delete → send（避免 race）
// - 非阻塞
// - 可重入（安全忽略不存在的 msgID）
func (w *workHelper) Reply(msgID string, payload string) {

	w.mu.Lock()
	wk, exists := w.works[msgID]
	if exists {
		delete(w.works, msgID)
	}
	w.mu.Unlock()

	if !exists {
		return
	}

	select {
	case wk.ch <- Reply{
		MsgID:   msgID,
		Payload: payload,
	}:
	default:
	}
}

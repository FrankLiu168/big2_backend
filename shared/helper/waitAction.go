package helper

import (
	"errors"
	"sync"
	"time"
)

// Reply 是等待方期望收到的响应
type ReplyWaiter struct {
	MsgID   string
	Payload string
}

// waiter 封装等待上下文
type waiter struct {
	ch        chan ReplyWaiter
	createdAt time.Time
}

// WaitHelper 管理所有等待中的请求
type WaitHelper struct {
	mu      sync.RWMutex
	waiters map[string]*waiter
}

var (
	once           sync.Once
	globalInstance *WaitHelper
)

// GetWaitHelper 返回单例实例（线程安全）
func GetWaitHelper() *WaitHelper {
	once.Do(func() {
		globalInstance = &WaitHelper{
			waiters: make(map[string]*waiter),
		}
	})
	return globalInstance
}

// WaitWithTimeout 注册一个等待，直到收到 Reply 或超时
func (h *WaitHelper) WaitWithTimeout(msgID string, timeout time.Duration) (ReplyWaiter, error) {
	if msgID == "" {
		return ReplyWaiter{}, errors.New("msgID cannot be empty")
	}

	replyCh := make(chan ReplyWaiter, 1) // 缓冲通道，避免 sender 阻塞

	h.mu.Lock()
	if _, exists := h.waiters[msgID]; exists {
		h.mu.Unlock()
		return ReplyWaiter{}, errors.New("duplicate msgID: " + msgID)
	}
	h.waiters[msgID] = &waiter{
		ch:        replyCh,
		createdAt: time.Now(),
	}
	h.mu.Unlock()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case resp := <-replyCh:
		return resp, nil
	case <-timer.C:
		// 超时：尝试清理
		h.mu.Lock()
		delete(h.waiters, msgID)
		h.mu.Unlock()
		return ReplyWaiter{}, errors.New("timeout waiting for reply to msgID: " + msgID)
	}
}

// Reply 向指定 msgID 发送响应
func (h *WaitHelper) Reply(msgID string, payload string) {
	h.mu.RLock()
	w, exists := h.waiters[msgID]
	h.mu.RUnlock()

	if !exists {
		// 无等待者，可能已超时或重复回复
		return
	}

	// 尝试发送，不阻塞（因为 channel 有缓冲且只发一次）
	select {
	case w.ch <- ReplyWaiter{MsgID: msgID, Payload: payload}:
		// 成功发送，现在安全地从 map 中移除
		h.mu.Lock()
		delete(h.waiters, msgID)
		h.mu.Unlock()
	default:
		// 理论上不会发生，因为 channel 是缓冲的且只发一次
		// 但保留防御性代码
	}
}

// 可选：定期清理过期 waiter（如果需要长期运行且可能堆积）
func (h *WaitHelper) CleanupExpired(timeout time.Duration) {
	now := time.Now()
	h.mu.Lock()
	for id, w := range h.waiters {
		if now.Sub(w.createdAt) > timeout {
			delete(h.waiters, id)
			close(w.ch) // 可选：通知等待方（但当前 Wait 不监听 close）
		}
	}
	h.mu.Unlock()
}

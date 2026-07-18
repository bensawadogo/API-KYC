package notification

import (
	"sync"
)

type SSEHub struct {
	mu       sync.RWMutex
	subs     map[string]map[chan Notification]struct{}
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		subs: make(map[string]map[chan Notification]struct{}),
	}
}

func (h *SSEHub) Subscribe(phone string) chan Notification {
	ch := make(chan Notification, 16)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs[phone] == nil {
		h.subs[phone] = make(map[chan Notification]struct{})
	}
	h.subs[phone][ch] = struct{}{}
	return ch
}

func (h *SSEHub) Unsubscribe(phone string, ch chan Notification) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if subs, ok := h.subs[phone]; ok {
		delete(subs, ch)
		close(ch)
		if len(subs) == 0 {
			delete(h.subs, phone)
		}
	}
}

func (h *SSEHub) Publish(phone string, notif Notification) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if subs, ok := h.subs[phone]; ok {
		for ch := range subs {
			select {
			case ch <- notif:
			default:
			}
		}
	}
}

func (h *SSEHub) SubscriberCount(phone string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if subs, ok := h.subs[phone]; ok {
		return len(subs)
	}
	return 0
}

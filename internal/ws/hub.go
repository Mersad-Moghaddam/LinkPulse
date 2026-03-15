package ws

import "sync"

type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[chan []byte]struct{}
}

func NewHub() *Hub { return &Hub{subs: map[string]map[chan []byte]struct{}{}} }

func (h *Hub) Subscribe(code string) chan []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan []byte, 8)
	if _, ok := h.subs[code]; !ok {
		h.subs[code] = map[chan []byte]struct{}{}
	}
	h.subs[code][ch] = struct{}{}
	return ch
}

func (h *Hub) Unsubscribe(code string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m, ok := h.subs[code]; ok {
		delete(m, ch)
	}
	close(ch)
}

func (h *Hub) Broadcast(code string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[code] {
		select {
		case ch <- payload:
		default:
		}
	}
}

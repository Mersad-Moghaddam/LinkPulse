package cache

import (
	"context"
	"sync"
	"time"

	"linkpulse/internal/models"
)

type entry struct {
	link    models.Link
	expires time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	links map[string]entry
}

func NewMemory() *MemoryCache { return &MemoryCache{links: map[string]entry{}} }

func (m *MemoryCache) GetLink(_ context.Context, code string) (models.Link, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.links[code]
	if !ok || (!e.expires.IsZero() && time.Now().After(e.expires)) {
		return models.Link{}, false
	}
	return e.link, true
}

func (m *MemoryCache) SetLink(_ context.Context, link models.Link) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var exp time.Time
	if link.ExpiresAt != nil {
		exp = *link.ExpiresAt
	}
	m.links[link.ShortCode] = entry{link: link, expires: exp}
}

func (m *MemoryCache) DeleteLink(_ context.Context, code string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.links, code)
}

func (m *MemoryCache) IncrCounters(context.Context, string) {}

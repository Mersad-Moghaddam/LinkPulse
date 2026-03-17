package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"linkpulse/internal/models"
)

type MemoryRepo struct {
	mu      sync.RWMutex
	links   map[string]models.Link
	clicks  map[string][]models.Click
	visited map[string]map[string]struct{}
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{links: map[string]models.Link{}, clicks: map[string][]models.Click{}, visited: map[string]map[string]struct{}{}}
}

func (m *MemoryRepo) Create(_ context.Context, link models.Link) (models.Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.links[link.ShortCode]; ok {
		return models.Link{}, ErrAliasExists
	}
	now := time.Now().UTC()
	link.CreatedAt, link.UpdatedAt = now, now
	m.links[link.ShortCode] = link
	return link, nil
}

func (m *MemoryRepo) GetByCode(_ context.Context, code string) (models.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.links[code]
	if !ok {
		return models.Link{}, ErrNotFound
	}
	return l, nil
}

func (m *MemoryRepo) List(_ context.Context) ([]models.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]models.Link, 0, len(m.links))
	for _, l := range m.links {
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (m *MemoryRepo) DeleteByCode(_ context.Context, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.links[code]; !ok {
		return ErrNotFound
	}
	delete(m.links, code)
	delete(m.clicks, code)
	delete(m.visited, code)
	return nil
}

func (m *MemoryRepo) Record(_ context.Context, click models.Click) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	code := click.LinkID
	if _, ok := m.visited[code]; !ok {
		m.visited[code] = map[string]struct{}{}
	}
	key := click.IP + "|" + click.UserAgent
	_, seen := m.visited[code][key]
	click.IsUnique = !seen
	m.visited[code][key] = struct{}{}
	m.clicks[code] = append(m.clicks[code], click)
	return nil
}

func (m *MemoryRepo) SummaryByCode(_ context.Context, code string) (models.AnalyticsSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.links[code]; !ok {
		return models.AnalyticsSummary{}, ErrNotFound
	}
	s := models.AnalyticsSummary{ShortCode: code, TopReferrers: map[string]int{}, TopBrowsers: map[string]int{}}
	for _, c := range m.clicks[code] {
		s.TotalClicks++
		if c.IsUnique {
			s.UniqueClicks++
		}
		s.TopReferrers[c.Referrer]++
		s.TopBrowsers[c.Browser]++
	}
	return s, nil
}

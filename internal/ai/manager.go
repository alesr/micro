package ai

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

const (
	maxCacheEntries = 16
	cacheTTL        = 5 * time.Second
)

type cacheEntry struct {
	resp      *Response
	err       error
	expiresAt time.Time
}

type Manager struct {
	mu       sync.Mutex
	provider Provider

	cancelPrev context.CancelFunc
	timer      *time.Timer
	debounce   time.Duration

	cache      map[uint64]*cacheEntry
	cacheOrder []uint64
	cacheMu    sync.Mutex
}

func NewManager(providerName, model, baseURL string, debounceMs float64) (*Manager, error) {
	provider, err := newProvider(providerName, model, baseURL)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, fmt.Errorf("unknown AI provider: %q", providerName)
	}

	return &Manager{
		provider:   provider,
		debounce:   time.Duration(debounceMs) * time.Millisecond,
		cache:      make(map[uint64]*cacheEntry),
		cacheOrder: make([]uint64, 0, maxCacheEntries),
	}, nil
}

func newProvider(name, model, baseURL string) (Provider, error) {
	switch name {
	case "codestral":
		return NewCodestralProvider(model, baseURL)
	default:
		return nil, nil
	}
}

func (m *Manager) Provider() Provider {
	return m.provider
}

func cacheKey(req Request) uint64 {
	h := fnv.New64a()
	h.Write([]byte(req.BeforeCursor))
	h.Write([]byte{0})
	h.Write([]byte(req.AfterCursor))
	h.Write([]byte{0})
	h.Write([]byte(req.FileType))
	return h.Sum64()
}

func (m *Manager) getCache(key uint64) (*cacheEntry, bool) {
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()
	if m.cache == nil {
		return nil, false
	}
	entry, ok := m.cache[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		delete(m.cache, key)
		m.removeCacheOrder(key)
		return nil, false
	}
	m.touchCache(key)
	return entry, true
}

func (m *Manager) touchCache(key uint64) {
	for i, k := range m.cacheOrder {
		if k == key {
			m.cacheOrder = append(m.cacheOrder[:i], m.cacheOrder[i+1:]...)
			m.cacheOrder = append(m.cacheOrder, key)
			return
		}
	}
}

func (m *Manager) removeCacheOrder(key uint64) {
	for i, k := range m.cacheOrder {
		if k == key {
			m.cacheOrder = append(m.cacheOrder[:i], m.cacheOrder[i+1:]...)
			return
		}
	}
}

func (m *Manager) setCache(key uint64, resp *Response, err error) {
	// Don't cache errors (transient network failures)
	if err != nil {
		return
	}
	if m.cache == nil {
		m.cache = make(map[uint64]*cacheEntry)
	}
	if len(m.cache) >= maxCacheEntries {
		delete(m.cache, m.cacheOrder[0])
		m.cacheOrder = m.cacheOrder[1:]
	}
	m.cache[key] = &cacheEntry{
		resp:      resp,
		err:       err,
		expiresAt: time.Now().Add(cacheTTL),
	}
	m.cacheOrder = append(m.cacheOrder, key)
}

func (m *Manager) Cancel() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancelPrev != nil {
		m.cancelPrev()
		m.cancelPrev = nil
	}
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
}

func (m *Manager) RequestDelayed(req Request, onResult func(*Response, error)) {
	key := cacheKey(req)

	if entry, ok := m.getCache(key); ok {
		onResult(entry.resp, entry.err)
		return
	}

	m.mu.Lock()

	if m.cancelPrev != nil {
		m.cancelPrev()
	}
	if m.timer != nil {
		m.timer.Stop()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelPrev = cancel

	m.timer = time.AfterFunc(m.debounce, func() {
		if ctx.Err() != nil {
			return
		}

		if entry, ok := m.getCache(key); ok {
			onResult(entry.resp, entry.err)
			return
		}

		resp, err := m.provider.Complete(ctx, req)
		if ctx.Err() != nil {
			return
		}

		m.cacheMu.Lock()
		m.setCache(key, resp, err)
		m.cacheMu.Unlock()

		onResult(resp, err)
	})

	m.mu.Unlock()
}

func (m *Manager) RequestNow(req Request) (*Response, error) {
	m.Cancel()

	resp, err := m.provider.Complete(context.Background(), req)
	if err == nil && resp != nil {
		key := cacheKey(req)
		m.cacheMu.Lock()
		m.setCache(key, resp, nil)
		m.cacheMu.Unlock()
	}
	return resp, err
}

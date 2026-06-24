package ai

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockProvider struct {
	name     string
	complete func(ctx context.Context, req Request) (*Response, error)
}

func (m *mockProvider) Name() string                                          { return m.name }
func (m *mockProvider) Complete(ctx context.Context, req Request) (*Response, error) {
	if m.complete != nil {
		return m.complete(ctx, req)
	}
	return &Response{Text: "mock"}, nil
}

func TestNewManager(t *testing.T) {
	t.Run("unknown provider", func(t *testing.T) {
		m, err := NewManager("unknown", "", "", 100)
		assert.Nil(t, m)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown AI provider")
	})

	t.Run("codestral missing key", func(t *testing.T) {
		t.Setenv("MISTRAL_API_KEY", "")
		m, err := NewManager("codestral", "", "", 100)
		assert.Nil(t, m)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MISTRAL_API_KEY")
	})
}

func TestManagerRequestNow(t *testing.T) {
	m := &Manager{
		provider: &mockProvider{
			complete: func(ctx context.Context, req Request) (*Response, error) {
				return &Response{Text: "hello"}, nil
			},
		},
		debounce: 10 * time.Millisecond,
	}

	resp, err := m.RequestNow(Request{BeforeCursor: "foo"})
	assert.NoError(t, err)
	assert.Equal(t, "hello", resp.Text)
}

func TestManagerRequestNowError(t *testing.T) {
	m := &Manager{
		provider: &mockProvider{
			complete: func(ctx context.Context, req Request) (*Response, error) {
				return nil, errors.New("api error")
			},
		},
	}

	resp, err := m.RequestNow(Request{BeforeCursor: "foo"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api error")
	assert.Nil(t, resp)
}

func TestManagerRequestDelayedFires(t *testing.T) {
	m := &Manager{
		provider: &mockProvider{
			complete: func(ctx context.Context, req Request) (*Response, error) {
				return &Response{Text: "delayed"}, nil
			},
		},
		debounce: 20 * time.Millisecond,
	}

	var result *Response
	var resultErr error
	done := make(chan struct{})

	m.RequestDelayed(Request{BeforeCursor: "foo"}, func(resp *Response, err error) {
		result = resp
		resultErr = err
		close(done)
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RequestDelayed did not fire")
	}

	assert.NoError(t, resultErr)
	assert.Equal(t, "delayed", result.Text)
}

func TestManagerRequestDelayedDebounce(t *testing.T) {
	var callCount int32
	m := &Manager{
		provider: &mockProvider{
			complete: func(ctx context.Context, req Request) (*Response, error) {
				atomic.AddInt32(&callCount, 1)
				return &Response{Text: "delayed"}, nil
			},
		},
		debounce: 50 * time.Millisecond,
	}

	// Fire multiple requests quickly — only the last one should execute
	m.RequestDelayed(Request{BeforeCursor: "a"}, func(*Response, error) {})
	m.RequestDelayed(Request{BeforeCursor: "b"}, func(*Response, error) {})
	m.RequestDelayed(Request{BeforeCursor: "c"}, func(*Response, error) {})

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
}

func TestManagerCancelStopsPending(t *testing.T) {
	var called bool
	m := &Manager{
		provider: &mockProvider{
			complete: func(ctx context.Context, req Request) (*Response, error) {
				<-ctx.Done()
				return nil, ctx.Err()
			},
		},
		debounce: 10 * time.Millisecond,
	}

	var resultErr error
	m.RequestDelayed(Request{BeforeCursor: "foo"}, func(resp *Response, err error) {
		called = true
		resultErr = err
	})

	time.Sleep(20 * time.Millisecond) // wait for timer to fire
	m.Cancel()
	time.Sleep(20 * time.Millisecond)

	// Completion should not have been called (context was cancelled)
	assert.False(t, called)
	assert.Nil(t, resultErr)
}

func TestManagerCancelBetweenRequests(t *testing.T) {
	var callCount int32
	m := &Manager{
		provider: &mockProvider{
			complete: func(ctx context.Context, req Request) (*Response, error) {
				atomic.AddInt32(&callCount, 1)
				return &Response{Text: req.BeforeCursor}, nil
			},
		},
		debounce: 30 * time.Millisecond,
	}

	// Send request A, cancel immediately via RequestNow
	m.RequestDelayed(Request{BeforeCursor: "a"}, func(*Response, error) {})
	m.RequestNow(Request{BeforeCursor: "now"})

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
}

func TestManagerCacheHit(t *testing.T) {
	var callCount int32
	m := &Manager{
		provider: &mockProvider{
			complete: func(ctx context.Context, req Request) (*Response, error) {
				atomic.AddInt32(&callCount, 1)
				return &Response{Text: "cached"}, nil
			},
		},
		debounce:   10 * time.Millisecond,
		cache:      make(map[uint64]*cacheEntry),
		cacheOrder: make([]uint64, 0, maxCacheEntries),
	}

	// First request populates cache
	done := make(chan struct{})
	m.RequestDelayed(Request{BeforeCursor: "foo", FileType: "go"}, func(resp *Response, err error) {
		close(done)
	})
	<-done
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))

	// Second request with same context should hit cache
	m.RequestDelayed(Request{BeforeCursor: "foo", FileType: "go"}, func(resp *Response, err error) {
		assert.Equal(t, "cached", resp.Text)
	})
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "should not call provider again")
}

func TestManagerCacheMissDifferentContext(t *testing.T) {
	var callCount int32
	m := &Manager{
		provider: &mockProvider{
			complete: func(ctx context.Context, req Request) (*Response, error) {
				atomic.AddInt32(&callCount, 1)
				return &Response{Text: req.BeforeCursor}, nil
			},
		},
		debounce:   10 * time.Millisecond,
		cache:      make(map[uint64]*cacheEntry),
		cacheOrder: make([]uint64, 0, maxCacheEntries),
	}

	done := make(chan struct{}, 2)
	m.RequestDelayed(Request{BeforeCursor: "foo", FileType: "go"}, func(resp *Response, err error) {
		done <- struct{}{}
	})
	<-done

	m.RequestDelayed(Request{BeforeCursor: "bar", FileType: "go"}, func(resp *Response, err error) {
		done <- struct{}{}
	})
	<-done

	assert.Equal(t, int32(2), atomic.LoadInt32(&callCount))
}

func TestManagerCacheExpiry(t *testing.T) {
	var callCount int32
	m := &Manager{
		provider: &mockProvider{
			complete: func(ctx context.Context, req Request) (*Response, error) {
				atomic.AddInt32(&callCount, 1)
				return &Response{Text: "fresh"}, nil
			},
		},
		debounce:   10 * time.Millisecond,
		cache:      make(map[uint64]*cacheEntry),
		cacheOrder: make([]uint64, 0, maxCacheEntries),
	}

	// Force very short TTL via direct cache insertion
	key := cacheKey(Request{BeforeCursor: "foo", FileType: "go"})
	m.cacheMu.Lock()
	m.cache[key] = &cacheEntry{
		resp:      &Response{Text: "stale"},
		expiresAt: time.Now().Add(-time.Second),
	}
	m.cacheOrder = append(m.cacheOrder, key)
	m.cacheMu.Unlock()

	done := make(chan struct{})
	m.RequestDelayed(Request{BeforeCursor: "foo", FileType: "go"}, func(resp *Response, err error) {
		close(done)
	})
	<-done

	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "expired entry should not be used")
}

func TestManagerRequestNowBypassesCache(t *testing.T) {
	var callCount int32
	m := &Manager{
		provider: &mockProvider{
			complete: func(ctx context.Context, req Request) (*Response, error) {
				atomic.AddInt32(&callCount, 1)
				return &Response{Text: "fresh"}, nil
			},
		},
		debounce:   10 * time.Millisecond,
		cache:      make(map[uint64]*cacheEntry),
		cacheOrder: make([]uint64, 0, maxCacheEntries),
	}

	// Pre-populate cache
	key := cacheKey(Request{BeforeCursor: "foo", FileType: "go"})
	m.cacheMu.Lock()
	m.cache[key] = &cacheEntry{
		resp:      &Response{Text: "stale"},
		expiresAt: time.Now().Add(time.Hour),
	}
	m.cacheOrder = append(m.cacheOrder, key)
	m.cacheMu.Unlock()

	// RequestNow always fetches fresh
	resp, err := m.RequestNow(Request{BeforeCursor: "foo", FileType: "go"})
	assert.NoError(t, err)
	assert.Equal(t, "fresh", resp.Text)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
}

func TestManagerCacheEviction(t *testing.T) {
	m := &Manager{
		provider: &mockProvider{
			complete: func(ctx context.Context, req Request) (*Response, error) {
				return &Response{Text: req.BeforeCursor}, nil
			},
		},
		debounce:   5 * time.Millisecond,
		cache:      make(map[uint64]*cacheEntry),
		cacheOrder: make([]uint64, 0, maxCacheEntries),
	}

	// Fill cache to max
	var wg sync.WaitGroup
	for i := 0; i < maxCacheEntries; i++ {
		wg.Add(1)
		key := string(rune('a' + i))
		m.RequestDelayed(Request{BeforeCursor: key, FileType: "go"}, func(resp *Response, err error) {
			wg.Done()
		})
		time.Sleep(20 * time.Millisecond)
	}
	wg.Wait()

	assert.Equal(t, maxCacheEntries, len(m.cache))

	// One more request should evict oldest
	wg.Add(1)
	m.RequestDelayed(Request{BeforeCursor: "zzz", FileType: "go"}, func(resp *Response, err error) {
		wg.Done()
	})
	time.Sleep(50 * time.Millisecond)
	wg.Wait()

	assert.Equal(t, maxCacheEntries, len(m.cache))
	// Oldest entry should be gone
	oldestKey := cacheKey(Request{BeforeCursor: "a", FileType: "go"})
	_, ok := m.cache[oldestKey]
	assert.False(t, ok, "oldest cache entry should have been evicted")
}

func TestManagerConcurrentRequests(t *testing.T) {
	m := &Manager{
		provider: &mockProvider{
			complete: func(ctx context.Context, req Request) (*Response, error) {
				time.Sleep(10 * time.Millisecond)
				return &Response{Text: req.BeforeCursor}, nil
			},
		},
		debounce:   5 * time.Millisecond,
		cache:      make(map[uint64]*cacheEntry),
		cacheOrder: make([]uint64, 0, maxCacheEntries),
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			m.RequestNow(Request{BeforeCursor: string(rune('a' + n))})
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 10, len(m.cache))
}

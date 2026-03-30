package external

import (
	"context"
	"sync"
	"time"

	"github.com/GerindT/GoFlight/internal/apierrors"
)

type memoryEntry struct {
	value     string
	expiresAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	store map[string]memoryEntry
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{store: map[string]memoryEntry{}}
}

func (m *MemoryCache) Get(_ context.Context, key string) (string, error) {
	m.mu.RLock()
	entry, ok := m.store[key]
	m.mu.RUnlock()
	if !ok {
		return "", apierrors.ErrCacheMiss
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		m.mu.Lock()
		delete(m.store, key)
		m.mu.Unlock()
		return "", apierrors.ErrCacheMiss
	}
	return entry.value, nil
}

func (m *MemoryCache) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	expiresAt := time.Time{}
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	m.mu.Lock()
	m.store[key] = memoryEntry{value: value, expiresAt: expiresAt}
	m.mu.Unlock()
	return nil
}

func (m *MemoryCache) Ping(_ context.Context) error {
	return nil
}

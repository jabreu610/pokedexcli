package pokecache

import (
	"context"
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func newCacheEntry(val []byte) cacheEntry {
	return cacheEntry{
		val:       val,
		createdAt: time.Now(),
	}
}

type Cache struct {
	store  map[string]cacheEntry
	mu     sync.RWMutex
	cancel context.CancelFunc
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = newCacheEntry(val)
}

func (c *Cache) Get(key string) (val []byte, ok bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.store[key]
	if !ok {
		return nil, ok
	}
	return entry.val, ok
}

func (c *Cache) Close() {
	c.cancel()
}

func (c *Cache) reapLoop(min time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.store {
		if entry.createdAt.Before(min) {
			delete(c.store, key)
		}
	}
}

func NewCache(interval time.Duration, parentCtx context.Context) *Cache {
	ctx, cancel := context.WithCancel(parentCtx)
	c := Cache{
		store:  map[string]cacheEntry{},
		cancel: cancel,
	}
	ticker := time.NewTicker(interval)

	go func() {
	CacheLoop:
		for {
			select {
			case <-ticker.C:
				c.reapLoop(time.Now().Add(interval * -1))
			case <-ctx.Done():
				break CacheLoop
			}
		}
	}()

	return &c
}

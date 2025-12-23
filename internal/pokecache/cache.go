package pokecache

import (
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
	store map[string]cacheEntry
	mu    sync.Mutex
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = newCacheEntry(val)
}

func (c *Cache) Get(key string) (val []byte, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.store[key]
	if !ok {
		return nil, ok
	}
	return entry.val, ok
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

func NewCache(interval time.Duration) *Cache {
	c := Cache{
		store: map[string]cacheEntry{},
	}
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			c.reapLoop(time.Now().Add(interval * -1))
		}
	}()

	return &c
}

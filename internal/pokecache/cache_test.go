package pokecache_test

import (
	"testing"
	"time"

	"github.com/jabreu610/pokedexcli/internal/pokecache"
)

func TestNewCache(t *testing.T) {
	cache := pokecache.NewCache(5 * time.Second)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}
}

func TestCacheAddAndGet(t *testing.T) {
	cache := pokecache.NewCache(5 * time.Second)
	key := "test-key"
	val := []byte("test-value")

	cache.Add(key, val)

	retrieved, ok := cache.Get(key)
	if !ok {
		t.Fatal("Get returned false for existing key")
	}
	if string(retrieved) != string(val) {
		t.Errorf("Expected value %s, got %s", val, retrieved)
	}
}

func TestCacheGetNonExistent(t *testing.T) {
	cache := pokecache.NewCache(5 * time.Second)

	retrieved, ok := cache.Get("non-existent")
	if ok {
		t.Error("Get returned true for non-existent key")
	}
	if retrieved != nil {
		t.Errorf("Expected nil value, got %v", retrieved)
	}
}

func TestCacheReap(t *testing.T) {
	interval := 100 * time.Millisecond
	cache := pokecache.NewCache(interval)

	// Add an entry
	key := "test-key"
	val := []byte("test-value")
	cache.Add(key, val)

	// Verify it exists
	_, ok := cache.Get(key)
	if !ok {
		t.Fatal("Key should exist immediately after adding")
	}

	// Wait for the reap interval plus a buffer
	time.Sleep(interval + 50*time.Millisecond)

	// Entry should be reaped
	_, ok = cache.Get(key)
	if ok {
		t.Error("Key should have been reaped after interval")
	}
}

func TestCacheReapDoesNotRemoveFreshEntries(t *testing.T) {
	interval := 200 * time.Millisecond
	cache := pokecache.NewCache(interval)

	// Add first entry
	cache.Add("old-key", []byte("old-value"))

	// Wait half the interval
	time.Sleep(interval / 2)

	// Add second entry
	cache.Add("new-key", []byte("new-value"))

	// Wait for just past the first interval
	time.Sleep(interval/2 + 50*time.Millisecond)

	// Old key should be gone
	_, ok := cache.Get("old-key")
	if ok {
		t.Error("Old key should have been reaped")
	}

	// New key should still exist
	_, ok = cache.Get("new-key")
	if !ok {
		t.Error("New key should not have been reaped")
	}
}

func TestCacheAddOverwrite(t *testing.T) {
	cache := pokecache.NewCache(5 * time.Second)
	key := "test-key"
	val1 := []byte("first-value")
	val2 := []byte("second-value")

	cache.Add(key, val1)
	cache.Add(key, val2)

	retrieved, ok := cache.Get(key)
	if !ok {
		t.Fatal("Key should exist")
	}
	if string(retrieved) != string(val2) {
		t.Errorf("Expected value %s, got %s", val2, retrieved)
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := pokecache.NewCache(5 * time.Second)
	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(n int) {
			key := "key"
			val := []byte{byte(n)}
			cache.Add(key, val)
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			cache.Get("key")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// If we get here without a race condition, test passes
}

func TestCacheMultipleKeys(t *testing.T) {
	cache := pokecache.NewCache(5 * time.Second)

	keys := []string{"key1", "key2", "key3"}
	vals := [][]byte{[]byte("val1"), []byte("val2"), []byte("val3")}

	// Add multiple keys
	for i, key := range keys {
		cache.Add(key, vals[i])
	}

	// Verify all keys exist
	for i, key := range keys {
		retrieved, ok := cache.Get(key)
		if !ok {
			t.Errorf("Key %s should exist", key)
		}
		if string(retrieved) != string(vals[i]) {
			t.Errorf("Expected value %s for key %s, got %s", vals[i], key, retrieved)
		}
	}
}

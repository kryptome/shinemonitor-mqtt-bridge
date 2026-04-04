package cache

import (
	"sync"
	"time"
)

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

var (
	store = make(map[string]cacheItem)
	mut   sync.RWMutex
)

// Set stores a value with a specific TTL
func Set(key string, value interface{}, ttl time.Duration) {
	mut.Lock()
	defer mut.Unlock()
	store[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Get retrieves a value, returning false if missing or expired
func Get(key string) (interface{}, bool) {
	mut.RLock()
	defer mut.RUnlock()
	
	item, ok := store[key]
	if !ok {
		return nil, false
	}
	
	// If it has expired
	if time.Now().After(item.expiresAt) {
		return nil, false
	}
	
	return item.value, true
}

// Clear removes all cached items
func Clear() {
	mut.Lock()
	defer mut.Unlock()
	store = make(map[string]cacheItem)
}

package xmanager

import (
	"sync"
	"time"
)

const NoTTL int64 = 0 // Constant to represent no expiration for cache entries

// CacheEntry represents a single cache entry with a value and an optional expiration time
type CacheEntry struct {
	Value      any
	Expiration int64 // Unix timestamp in seconds, 0 means no expiration
}

// GatewayInternalCache represents a thread-safe internal cache
type GatewayInternalCache struct {
	mu       sync.RWMutex
	cache    map[string]CacheEntry
	stopChan chan struct{} // Channel to signal the cleanup goroutine to stop
}

// NewGatewayInternalCache creates a new instance of GatewayInternalCache
func NewGatewayInternalCache(cleanupInterval time.Duration) *GatewayInternalCache {
	cache := &GatewayInternalCache{
		cache:    make(map[string]CacheEntry),
		stopChan: make(chan struct{}),
	}

	// Start the cleanup goroutine
	go cache.startCleanup(cleanupInterval)

	return cache
}

// Set adds a key-value pair to the cache with an optional expiration time (in seconds)
// Use NoTTL (0) for entries that should never expire
func (c *GatewayInternalCache) Set(key string, value any, ttl int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Unix() + ttl
	}

	c.cache[key] = CacheEntry{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves a value from the cache by key
func (c *GatewayInternalCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	// Check if the entry has expired
	if entry.Expiration > 0 && time.Now().Unix() > entry.Expiration {
		// Remove expired entry
		c.mu.RUnlock()
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.cache, key)
		return nil, false
	}

	return entry.Value, true
}

// Delete removes a key-value pair from the cache
func (c *GatewayInternalCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

// Clear removes all entries from the cache
func (c *GatewayInternalCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]CacheEntry)
}

// Keys returns a list of all keys currently in the cache
func (c *GatewayInternalCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.cache))
	for key := range c.cache {
		keys = append(keys, key)
	}
	return keys
}

// Size returns the number of items currently in the cache
func (c *GatewayInternalCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

// StopCleanup stops the background cleanup goroutine
func (c *GatewayInternalCache) StopCleanup() {
	close(c.stopChan)
}

// startCleanup starts a goroutine to periodically remove expired entries
func (c *GatewayInternalCache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpiredEntries()
		case <-c.stopChan:
			return
		}
	}
}

// cleanupExpiredEntries removes expired entries from the cache
func (c *GatewayInternalCache) cleanupExpiredEntries() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().Unix()
	for key, entry := range c.cache {
		if entry.Expiration > 0 && now > entry.Expiration {
			delete(c.cache, key)
		}
	}
}

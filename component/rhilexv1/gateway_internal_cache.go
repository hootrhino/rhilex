// Copyright (C) 2025 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
package rhilex

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
	cache    map[string]map[string]CacheEntry // Nested map for hierarchical keys
	stopChan chan struct{}                    // Channel to signal the cleanup goroutine to stop
}

// NewGatewayInternalCache creates a new instance of GatewayInternalCache
func NewGatewayInternalCache(cleanupInterval time.Duration) *GatewayInternalCache {
	cache := &GatewayInternalCache{
		cache:    make(map[string]map[string]CacheEntry),
		stopChan: make(chan struct{}),
	}

	// Start the cleanup goroutine
	go cache.startCleanup(cleanupInterval)

	return cache
}

// Set adds a key-value pair to the cache with an optional expiration time (in seconds)
// Use NoTTL (0) for entries that should never expire
func (c *GatewayInternalCache) Set(namespace, key string, value any, ttl int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.cache[namespace]; !exists {
		c.cache[namespace] = make(map[string]CacheEntry)
	}

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Unix() + ttl
	}

	c.cache[namespace][key] = CacheEntry{
		Value:      value,
		Expiration: expiration,
	}
}

func (c *GatewayInternalCache) Get(namespace, key string) (any, bool) {
	c.mu.RLock()
	ns, exists := c.cache[namespace]
	if !exists {
		c.mu.RUnlock()
		return nil, false
	}

	entry, exists := ns[key]
	if !exists {
		c.mu.RUnlock()
		return nil, false
	}

	// Check if the entry has expired
	if entry.Expiration > 0 && time.Now().Unix() > entry.Expiration {
		c.mu.RUnlock() // 释放读锁
		c.mu.Lock()    // 获取写锁
		defer c.mu.Unlock()

		// Double-check to ensure the entry still exists and is expired
		if ns, exists := c.cache[namespace]; exists {
			if entry, exists := ns[key]; exists && entry.Expiration > 0 && time.Now().Unix() > entry.Expiration {
				delete(ns, key)
				if len(ns) == 0 {
					delete(c.cache, namespace)
				}
			}
		}
		return nil, false
	}

	c.mu.RUnlock()
	return entry.Value, true
}

// Delete removes a key-value pair from the cache
func (c *GatewayInternalCache) Delete(namespace, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ns, exists := c.cache[namespace]; exists {
		delete(ns, key)
		if len(ns) == 0 {
			delete(c.cache, namespace)
		}
	}
}

// ClearNamespace removes all entries from a specific namespace
func (c *GatewayInternalCache) ClearNamespace(namespace string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, namespace)
}

// Clear removes all entries from the cache
func (c *GatewayInternalCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]map[string]CacheEntry)
}

// Keys returns a list of all keys in a specific namespace
func (c *GatewayInternalCache) Keys(namespace string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ns, exists := c.cache[namespace]
	if !exists {
		return nil
	}

	keys := make([]string, 0, len(ns))
	for key := range ns {
		keys = append(keys, key)
	}
	return keys
}

// Namespaces returns a list of all namespaces currently in the cache
func (c *GatewayInternalCache) Namespaces() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	namespaces := make([]string, 0, len(c.cache))
	for namespace := range c.cache {
		namespaces = append(namespaces, namespace)
	}
	return namespaces
}

// Size returns the number of items in a specific namespace
func (c *GatewayInternalCache) Size(namespace string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ns, exists := c.cache[namespace]
	if !exists {
		return 0
	}

	return len(ns)
}

// TotalSize returns the total number of items in the cache across all namespaces
func (c *GatewayInternalCache) TotalSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := 0
	now := time.Now().Unix()
	for _, ns := range c.cache {
		for key, entry := range ns {
			if entry.Expiration == 0 || entry.Expiration > now {
				total++
			} else {
				delete(ns, key)
			}
		}
	}
	return total
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
	for namespace, ns := range c.cache {
		for key, entry := range ns {
			if entry.Expiration > 0 && now > entry.Expiration {
				delete(ns, key)
			}
		}
		if len(ns) == 0 {
			delete(c.cache, namespace)
		}
	}
}

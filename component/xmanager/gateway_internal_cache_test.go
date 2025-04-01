package xmanager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGatewayInternalCache(t *testing.T) {
	// Create a new cache with a cleanup interval of 1 second
	cache := NewGatewayInternalCache(1 * time.Second)
	defer cache.StopCleanup()

	t.Run("Set and Get", func(t *testing.T) {
		cache.Set("key1", "value1", 10) // Expires in 10 seconds
		value, exists := cache.Get("key1")
		assert.True(t, exists)
		assert.Equal(t, "value1", value)
	})

	t.Run("NoTTL Entry", func(t *testing.T) {
		cache.Set("key2", "value2", NoTTL) // No expiration
		value, exists := cache.Get("key2")
		assert.True(t, exists)
		assert.Equal(t, "value2", value)
	})

	t.Run("Expired Entry", func(t *testing.T) {
		cache.Set("key3", "value3", 1) // Expires in 1 second
		time.Sleep(2 * time.Second)    // Wait for the entry to expire
		value, exists := cache.Get("key3")
		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("Delete Entry", func(t *testing.T) {
		cache.Set("key4", "value4", NoTTL)
		cache.Delete("key4")
		value, exists := cache.Get("key4")
		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("Clear Cache", func(t *testing.T) {
		cache.Set("key5", "value5", NoTTL)
		cache.Set("key6", "value6", NoTTL)
		cache.Clear()
		assert.Equal(t, 0, cache.Size())
	})

	t.Run("Keys and Size", func(t *testing.T) {
		cache.Set("key7", "value7", NoTTL)
		cache.Set("key8", "value8", NoTTL)
		keys := cache.Keys()
		assert.ElementsMatch(t, []string{"key7", "key8"}, keys)
		assert.Equal(t, 2, cache.Size())
	})
	time.Sleep(2 * time.Second)
}

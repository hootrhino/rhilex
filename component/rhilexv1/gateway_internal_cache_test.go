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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGatewayInternalCache(t *testing.T) {
	// Create a new cache with a cleanup interval of 1 second
	cache := NewGatewayInternalCache(1 * time.Second)
	defer cache.StopCleanup()

	t.Run("Set and Get", func(t *testing.T) {
		cache.Set("namespace1", "key1", "value1", 10) // Expires in 10 seconds
		value, exists := cache.Get("namespace1", "key1")
		assert.True(t, exists)
		assert.Equal(t, "value1", value)
		cache.Clear()
	})

	t.Run("NoTTL Entry", func(t *testing.T) {
		cache.Set("namespace1", "key2", "value2", NoTTL) // No expiration
		value, exists := cache.Get("namespace1", "key2")
		assert.True(t, exists)
		assert.Equal(t, "value2", value)
		cache.Clear()
	})

	t.Run("Expired Entry", func(t *testing.T) {
		cache.Set("namespace1", "key3", "value3", 1) // Expires in 1 second
		time.Sleep(2 * time.Second)                  // Wait for the entry to expire
		value, exists := cache.Get("namespace1", "key3")
		assert.False(t, exists)
		assert.Nil(t, value)
		cache.Clear()
	})

	t.Run("Delete Entry", func(t *testing.T) {
		cache.Set("namespace1", "key4", "value4", NoTTL)
		cache.Delete("namespace1", "key4")
		value, exists := cache.Get("namespace1", "key4")
		assert.False(t, exists)
		assert.Nil(t, value)
		cache.Clear()
	})

	t.Run("Clear Namespace", func(t *testing.T) {
		cache.Set("namespace2", "key5", "value5", NoTTL)
		cache.Set("namespace2", "key6", "value6", NoTTL)
		cache.ClearNamespace("namespace2")
		value, exists := cache.Get("namespace2", "key5")
		assert.False(t, exists)
		assert.Nil(t, value)
		assert.Equal(t, 0, cache.Size("namespace2"))
		cache.Clear()
	})

	t.Run("Clear Cache", func(t *testing.T) {
		cache.Set("namespace1", "key7", "value7", NoTTL)
		cache.Set("namespace2", "key8", "value8", NoTTL)
		cache.Clear()
		assert.Equal(t, 0, cache.TotalSize())
	})

	t.Run("Keys and Size", func(t *testing.T) {
		cache.Set("namespace1", "key9", "value9", NoTTL)
		cache.Set("namespace1", "key10", "value10", NoTTL)
		keys := cache.Keys("namespace1")
		assert.ElementsMatch(t, []string{"key9", "key10"}, keys)
		assert.Equal(t, 2, cache.Size("namespace1"))
		cache.Clear()
	})

	t.Run("Namespaces and TotalSize", func(t *testing.T) {
		cache.Set("namespace1", "key11", "value11", NoTTL)
		cache.Set("namespace2", "key12", "value12", NoTTL)
		namespaces := cache.Namespaces()
		assert.ElementsMatch(t, []string{"namespace1", "namespace2"}, namespaces)
		assert.Equal(t, 2, cache.TotalSize())
		cache.Clear()
	})

	t.Run("Cleanup Expired Entries", func(t *testing.T) {
		cache.Set("namespace1", "key13", "value13", 1) // Expires in 1 second
		cache.Set("namespace1", "key14", "value14", NoTTL)
		time.Sleep(2 * time.Second) // Wait for the entry to expire
		cache.cleanupExpiredEntries()
		value, exists := cache.Get("namespace1", "key13")
		assert.False(t, exists)
		assert.Nil(t, value)
		value, exists = cache.Get("namespace1", "key14")
		assert.True(t, exists)
		assert.Equal(t, "value14", value)
		cache.Clear()
	})
}

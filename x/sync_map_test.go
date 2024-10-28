// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncMapStoreAndLoad(t *testing.T) {
	m := NewSyncMap[int, string]()

	m.Store(1, "one")

	// Test Load for an existing key
	val, ok := m.Load(1)
	require.True(t, ok, "Expected key 1 to exist")
	assert.Equal(t, "one", val, "Expected value 'one' for key 1")

	// Test Load for a non-existing key
	_, ok = m.Load(2)
	assert.False(t, ok, "Expected key 2 to be absent")
}

func TestSyncMapLoadOrStore(t *testing.T) {
	m := NewSyncMap[int, string]()

	// Store a new key-value pair
	val, loaded := m.LoadOrStore(1, "one")
	require.False(t, loaded, "Expected key 1 to be newly stored")
	assert.Equal(t, "one", val, "Expected value 'one' for key 1 after LoadOrStore")

	// Attempt to store a new value for an existing key
	val, loaded = m.LoadOrStore(1, "uno")
	require.True(t, loaded, "Expected key 1 to already exist")
	assert.Equal(t, "one", val, "Expected existing value 'one' for key 1")
}

func TestSyncMapDelete(t *testing.T) {
	m := NewSyncMap[int, string]()

	m.Store(1, "one")
	m.Delete(1)

	_, ok := m.Load(1)
	assert.False(t, ok, "Expected key 1 to be deleted")
}

func TestSyncMapRange(t *testing.T) {
	m := NewSyncMap[int, string]()

	m.Store(1, "one")
	m.Store(2, "two")
	m.Store(3, "three")

	expected := map[int]string{
		1: "one",
		2: "two",
		3: "three",
	}

	m.Range(func(key int, value string) bool {
		expectedVal, exists := expected[key]
		require.True(t, exists, "Unexpected key found in map")
		assert.Equal(t, expectedVal, value, "Unexpected value for key %d", key)
		delete(expected, key)
		return true
	})

	assert.Empty(t, expected, "Not all entries were iterated over")
}

func TestSyncMapConcurrentAccess(t *testing.T) {
	m := NewSyncMap[int, int]()
	var wg sync.WaitGroup

	// Run multiple goroutines to test concurrent access
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.Store(i, i)
		}(i)
	}

	wg.Wait()

	// Verify all stored values
	for i := 0; i < 100; i++ {
		val, ok := m.Load(i)
		require.True(t, ok, "Expected key %d to exist", i)
		assert.Equal(t, i, val, "Expected value %d for key %d", i, i)
	}
}

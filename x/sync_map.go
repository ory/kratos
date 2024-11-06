// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"sync"
)

// SyncMap provides a thread-safe map with generic keys and values
type SyncMap[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

// NewSyncMap initializes a new SyncMap instance
func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{
		data: make(map[K]V),
	}
}

// Load retrieves a value for a key. It returns the value and a boolean indicating if the key exists.
func (m *SyncMap[K, V]) Load(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

// Store sets a value for a key, replacing any existing value.
func (m *SyncMap[K, V]) Store(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// LoadOrStore retrieves the existing value for a key if it exists, or stores and returns a given value if it doesn't.
func (m *SyncMap[K, V]) LoadOrStore(key K, value V) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.data[key]; ok {
		return existing, true
	}
	m.data[key] = value
	return value, false
}

// Delete removes a key-value pair from the map.
func (m *SyncMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

// Range iterates over all entries in the map, calling the provided function for each key-value pair.
// If the function returns false, the iteration stops.
func (m *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.data {
		if !f(k, v) {
			break
		}
	}
}

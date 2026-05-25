// Package storage defines the storage interface and implementations for the application.
package storage

import (
	"strings"
	"sync"
)

// MemoryStore is an in-memory implementation of the Store interface.
// It is safe for concurrent use.
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewMemoryStore creates a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]string),
	}
}

// Get retrieves the value for the given key.
func (ms *MemoryStore) Get(key string) (string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	value, ok := ms.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return value, nil
}

// Set stores the value for the given key.
func (ms *MemoryStore) Set(key string, value string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.data[key] = value
	return nil
}

// Delete removes the key and its value from the store.
func (ms *MemoryStore) Delete(key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.data, key)
	return nil
}

// List returns all keys that start with the given prefix.
func (ms *MemoryStore) List(prefix string) ([]string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var keys []string
	for key := range ms.data {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

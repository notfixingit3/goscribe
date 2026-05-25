// Package storage defines the storage interface and implementations for the application.
package storage

import "errors"

// Store defines the interface for key-value storage backends.
type Store interface {
	// Get retrieves the value for the given key.
	// Returns an error if the key does not exist.
	Get(key string) (string, error)

	// Set stores the value for the given key.
	Set(key string, value string) error

	// Delete removes the key and its value from the store.
	Delete(key string) error

	// List returns all keys that start with the given prefix.
	List(prefix string) ([]string, error)
}

// ErrKeyNotFound is returned when a key does not exist in the store.
var ErrKeyNotFound = errors.New("key not found")

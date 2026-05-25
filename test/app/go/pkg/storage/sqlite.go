// Package storage defines the storage interface and implementations for the application.
package storage

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore is a SQLite-backed implementation of the Store interface.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens a SQLite database and initializes the schema.
func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		return nil, err
	}

	return store, nil
}

// migrate creates the necessary tables if they do not exist.
func (s *SQLiteStore) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS store (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	return err
}

// Get retrieves the value for the given key.
func (s *SQLiteStore) Get(key string) (string, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM store WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// Set stores the value for the given key.
func (s *SQLiteStore) Set(key string, value string) error {
	_, err := s.db.Exec(
		"INSERT INTO store (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value,
	)
	return err
}

// Delete removes the key and its value from the store.
func (s *SQLiteStore) Delete(key string) error {
	_, err := s.db.Exec("DELETE FROM store WHERE key = ?", key)
	return err
}

// List returns all keys that start with the given prefix.
func (s *SQLiteStore) List(prefix string) ([]string, error) {
	rows, err := s.db.Query("SELECT key FROM store WHERE key LIKE ?", prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, rows.Err()
}

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

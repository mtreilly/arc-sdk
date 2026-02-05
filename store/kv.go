// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/yourorg/arc-sdk/db"
)

var (
	ErrNotFound = errors.New("store: key not found")
)

// KVStore is a simple key-value store interface.
// Keys are strings. Values are arbitrary byte slices.
// Implementations may persist values to disk or keep them in memory.
type KVStore interface {
	// Get retrieves the value for the given key.
	// Returns ErrNotFound if the key does not exist.
	Get(ctx context.Context, key string) ([]byte, error)
	// Set stores the value for the given key, overwriting any existing value.
	Set(ctx context.Context, key string, value []byte) error
	// Delete removes the key and its value.
	Delete(ctx context.Context, key string) error
	// Close releases any resources held by the store (e.g., database connections).
	Close() error
}

// sqliteStore implements KVStore using a SQLite database table.
type sqliteStore struct {
	db *sql.DB
}

// OpenSQLiteStore opens a SQLite database at the given path and returns a KVStore.
// If path is empty, db.DefaultDBPath() is used.
// The store uses a table named "kv_store". The table is created if it does not exist.
func OpenSQLiteStore(path string) (KVStore, error) {
	if path == "" {
		path = db.DefaultDBPath()
	}
	// Open the database using arc-sdk/db (applies pragmas and migrations)
	d, err := db.Open(path)
	if err != nil {
		return nil, err
	}
	// Ensure kv_store table exists
	if _, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS kv_store (
			key TEXT PRIMARY KEY,
			value TEXT,
			updated_at INTEGER NOT NULL
		)
	`); err != nil {
		d.Close()
		return nil, err
	}
	return &sqliteStore{db: d}, nil
}

func (s *sqliteStore) Get(ctx context.Context, key string) ([]byte, error) {
	var valStr string
	err := s.db.QueryRowContext(ctx, `
		SELECT value FROM kv_store WHERE key = ?
	`, key).Scan(&valStr)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return []byte(valStr), nil
}

func (s *sqliteStore) Set(ctx context.Context, key string, value []byte) error {
	now := time.Now().UnixNano()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO kv_store (key, value, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, string(value), now)
	return err
}

func (s *sqliteStore) Delete(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM kv_store WHERE key = ?`, key)
	return err
}

func (s *sqliteStore) Close() error {
	return s.db.Close()
}

// memoryStore implements KVStore using an in-memory map.
type memoryStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// NewMemoryStore returns a new in-memory KVStore.
// It is suitable for use when persistence is not required or unavailable.
func NewMemoryStore() KVStore {
	return &memoryStore{data: make(map[string][]byte)}
}

func (s *memoryStore) Get(ctx context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *memoryStore) Set(ctx context.Context, key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *memoryStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

func (s *memoryStore) Close() error {
	s.data = nil
	return nil
}

// GetJSON unmarshals JSON from the store into the value pointed to by v.
// If the key does not exist, ErrNotFound is returned.
// The type T is inferred from v.
func GetJSON[T any](ctx context.Context, store KVStore, key string) (T, error) {
	var zero T
	data, err := store.Get(ctx, key)
	if err != nil {
		return zero, err
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return zero, err
	}
	return v, nil
}

// SetJSON marshals value as JSON and stores it under the given key.
func SetJSON[T any](ctx context.Context, store KVStore, key string, value T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return store.Set(ctx, key, data)
}

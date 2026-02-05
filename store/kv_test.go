// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSQLiteStore(t *testing.T) {
	ctx := context.Background()

	// Use a temporary database file
	dbPath := t.TempDir() + "/test.db"
	store, err := OpenSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}
	defer store.Close()

	// Test Set and Get
	key := "test:key"
	value := []byte(`{"counter":42,"name":"example"}`)
	if err := store.Set(ctx, key, value); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != string(value) {
		t.Fatalf("Get returned %q, want %q", got, value)
	}

	// Test overwrite
	newValue := []byte(`{"counter":100}`)
	if err := store.Set(ctx, key, newValue); err != nil {
		t.Fatalf("Set overwrite: %v", err)
	}
	got, err = store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get after overwrite: %v", err)
	}
	if string(got) != string(newValue) {
		t.Fatalf("Get after overwrite returned %q, want %q", got, newValue)
	}

	// Test Delete
	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = store.Get(ctx, key)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after Delete: expected ErrNotFound, got %v", err)
	}

	// Test Get on non-existent key
	_, err = store.Get(ctx, "non-existent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get non-existent: expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore().(*memoryStore)
	defer store.Close()

	// Set and Get
	key := "mem:key"
	value := []byte("hello memory")
	if err := store.Set(ctx, key, value); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != string(value) {
		t.Fatalf("Get returned %q, want %q", got, value)
	}

	// Delete
	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = store.Get(ctx, key)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after Delete: expected ErrNotFound, got %v", err)
	}
}

func TestJSONHelpers(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore().(*memoryStore)

	type State struct {
		Counter int    `json:"counter"`
		LastRun int64  `json:"last_run"`
	}
	state := State{Counter: 5, LastRun: time.Now().Unix()}

	// SetJSON
	key := "json:state"
	if err := SetJSON(ctx, store, key, state); err != nil {
		t.Fatalf("SetJSON: %v", err)
	}

	// GetJSON
	got, err := GetJSON[State](ctx, store, key)
	if err != nil {
		t.Fatalf("GetJSON: %v", err)
	}
	// The got should equal state
	if got.Counter != state.Counter || got.LastRun != state.LastRun {
		t.Fatalf("GetJSON returned %+v, want %+v", got, state)
	}
}

func TestGracefulDegradation(t *testing.T) {
	// Simulate scenario: SQLite fails, fallback to memory
	store, err := OpenSQLiteStore("/invalid/path/that/does/not/exist/test.db")
	if err == nil {
		store.Close()
		t.Fatalf("expected OpenSQLiteStore to fail with invalid path")
	}
	// In real code, you'd fall back: store = NewMemoryStore()
	t.Logf("SQLite open failed as expected: %v", err)

	// Fallback works
	memStore := NewMemoryStore()
	ctx := context.Background()
	if err := memStore.Set(ctx, "fallback:key", []byte("ok")); err != nil {
		t.Fatalf("MemoryStore Set failed: %v", err)
	}
	got, err := memStore.Get(ctx, "fallback:key")
	if err != nil {
		t.Fatalf("MemoryStore Get failed: %v", err)
	}
	if string(got) != "ok" {
		t.Fatalf("MemoryStore Get returned %q, want %q", got, "ok")
	}
}

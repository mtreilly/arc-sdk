// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/yourorg/arc-sdk/db/migrations"
)

// Open opens a SQLite database at the given path and applies embedded migrations.
func Open(path string) (*sql.DB, error) {
	dsn := path
	if dsn == "" {
		dsn = DefaultDBPath()
	}

	// Ensure parent directory exists for file-backed databases
	if dsn != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(dsn), 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	handle, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Connection pool limits suitable for CLI usage
	handle.SetMaxOpenConns(25)
	handle.SetMaxIdleConns(5)
	handle.SetConnMaxLifetime(5 * time.Minute)

	if err := applyPragmas(handle); err != nil {
		_ = handle.Close()
		return nil, err
	}

	if err := migrations.RunMigrations(handle); err != nil {
		_ = handle.Close()
		return nil, err
	}
	return handle, nil
}

// DefaultDBPath resolves the default on-disk database path using XDG_DATA_HOME
// with fallback to ~/.local/share/arc/arc.db
func DefaultDBPath() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "arc", "arc.db")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		// As a final fallback, use current working directory
		return "arc.db"
	}
	return filepath.Join(home, ".local", "share", "arc", "arc.db")
}

func applyPragmas(db *sql.DB) error {
	// WAL mode; select returns the new mode
	var mode string
	if err := db.QueryRow("PRAGMA journal_mode = WAL;").Scan(&mode); err != nil {
		return fmt.Errorf("set WAL: %w", err)
	}
	if mode != "wal" {
		return errors.New("journal_mode not WAL")
	}
	// Recommended durability/perf balance
	if _, err := db.Exec("PRAGMA synchronous = NORMAL;"); err != nil {
		return fmt.Errorf("set synchronous: %w", err)
	}
	// Enforce FKs
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return fmt.Errorf("enable foreign_keys: %w", err)
	}
	// Busy timeout (ms)
	if _, err := db.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
		return fmt.Errorf("set busy_timeout: %w", err)
	}
	// Negative = kibibytes; 64MB
	if _, err := db.Exec("PRAGMA cache_size = -64000;"); err != nil {
		return fmt.Errorf("set cache_size: %w", err)
	}
	// Temp store in memory
	if _, err := db.Exec("PRAGMA temp_store = MEMORY;"); err != nil {
		return fmt.Errorf("set temp_store: %w", err)
	}
	return nil
}

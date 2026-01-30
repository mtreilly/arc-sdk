// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//go:embed sql/*.sql
var migrationsFS embed.FS

type mig struct {
	version int
	name    string
	path    string
}

// RunMigrations applies any pending embedded SQL migrations.
func RunMigrations(db *sql.DB) error {
	if err := ensureSchemaTable(db); err != nil {
		return err
	}

	applied, err := loadApplied(db)
	if err != nil {
		return err
	}

	migs, err := loadEmbeddedMigrations()
	if err != nil {
		return err
	}

	for _, m := range migs {
		if _, ok := applied[m.version]; ok {
			continue // already applied
		}
		if err := applyOne(db, m); err != nil {
			return fmt.Errorf("apply migration %03d_%s: %w", m.version, m.name, err)
		}
	}
	return nil
}

func ensureSchemaTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	return err
}

func loadApplied(db *sql.DB) (map[int]struct{}, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	seen := make(map[int]struct{})
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		seen[v] = struct{}{}
	}
	return seen, rows.Err()
}

func loadEmbeddedMigrations() ([]mig, error) {
	entries, err := fs.ReadDir(migrationsFS, "sql")
	if err != nil {
		return nil, err
	}
	out := make([]mig, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		// Expect prefix NNN_
		base := filepath.Base(name)
		parts := strings.SplitN(base, "_", 2)
		if len(parts) < 2 {
			continue
		}
		v, err := strconv.Atoi(strings.TrimLeft(parts[0], "0"))
		if err != nil {
			continue
		}
		out = append(out, mig{version: v, name: parts[1], path: filepath.Join("sql", name)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}

func applyOne(db *sql.DB, m mig) error {
	sqlBytes, err := migrationsFS.ReadFile(m.path)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(string(sqlBytes)); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations(version, name) VALUES(?, ?)`, m.version, m.name); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// MigrationInfo describes an embedded or applied migration.
type MigrationInfo struct {
	Version int
	Name    string
}

// Embedded returns embedded migration descriptors (version and name) in order.
func Embedded() ([]MigrationInfo, error) {
	list, err := loadEmbeddedMigrations()
	if err != nil {
		return nil, err
	}
	out := make([]MigrationInfo, 0, len(list))
	for _, m := range list {
		out = append(out, MigrationInfo{Version: m.version, Name: m.name})
	}
	return out, nil
}

// Applied returns versions and names recorded in schema_migrations.
func Applied(db *sql.DB) (map[int]string, error) {
	rows, err := db.Query(`SELECT version, name FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[int]string)
	for rows.Next() {
		var v int
		var n string
		if err := rows.Scan(&v, &n); err != nil {
			return nil, err
		}
		m[v] = n
	}
	return m, rows.Err()
}

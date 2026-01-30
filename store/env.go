// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package store

import (
	"context"
	"database/sql"
)

// EnvBackup represents an environment file backup record.
type EnvBackup struct {
	Project string
	Path    string
	Size    int64
	MTime   int64
	Status  string
}

// EnvStore manages environment backup records in the database.
type EnvStore struct {
	DB *sql.DB
}

// NewEnvStore creates a new EnvStore.
func NewEnvStore(db *sql.DB) *EnvStore {
	return &EnvStore{DB: db}
}

// UpsertBackup inserts or updates an environment backup record.
func (s *EnvStore) UpsertBackup(ctx context.Context, b EnvBackup) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO env_backups(project, path, size, mtime, status)
		VALUES(?,?,?,?,?)
		ON CONFLICT(project) DO UPDATE SET
		  path=excluded.path,
		  size=excluded.size,
		  mtime=excluded.mtime,
		  status=excluded.status
	`, b.Project, b.Path, b.Size, b.MTime, b.Status)
	return err
}

// GetBackup retrieves an environment backup by project name.
func (s *EnvStore) GetBackup(ctx context.Context, project string) (*EnvBackup, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT project, path, size, mtime, status
		FROM env_backups
		WHERE project = ?
	`, project)

	var b EnvBackup
	if err := row.Scan(&b.Project, &b.Path, &b.Size, &b.MTime, &b.Status); err != nil {
		return nil, err
	}
	return &b, nil
}

// ListBackups returns all environment backup records.
func (s *EnvStore) ListBackups(ctx context.Context) ([]EnvBackup, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT project, path, size, mtime, status
		FROM env_backups
		ORDER BY project
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []EnvBackup
	for rows.Next() {
		var b EnvBackup
		if err := rows.Scan(&b.Project, &b.Path, &b.Size, &b.MTime, &b.Status); err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, rows.Err()
}

// DeleteBackup removes an environment backup record.
func (s *EnvStore) DeleteBackup(ctx context.Context, project string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM env_backups WHERE project = ?`, project)
	return err
}

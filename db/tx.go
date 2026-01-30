// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package db

import (
	"context"
	"database/sql"
	"time"
)

// WithTx executes fn within a transaction, rolling back on error.
func WithTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// WithTxTimeout executes fn within a transaction with a timeout context.
func WithTxTimeout(db *sql.DB, timeout time.Duration, fn func(context.Context, *sql.Tx) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

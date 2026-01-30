// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package store

import (
	"context"
	"database/sql"
)

// Session represents an agent session.
type Session struct {
	ID              string
	Agent           string
	CWD             string
	Project         string
	Branch          string
	Path            string
	ModTS           int64
	CreateTS        int64
	Lines           int
	LastUser        string
	LastTS          int64
	Tags            string
	Archived        bool
	TmuxSessionID   *string
	TmuxWindowIndex *int
	TmuxPaneIndex   *int
}

// SessionsStore manages session data in the database.
type SessionsStore struct {
	DB *sql.DB
}

// NewSessionsStore creates a new SessionsStore.
func NewSessionsStore(db *sql.DB) *SessionsStore {
	return &SessionsStore{DB: db}
}

// Upsert inserts or updates a session row.
func (s *SessionsStore) Upsert(ctx context.Context, sess Session) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO sessions(
		id, agent, cwd, project, branch, path, mod_ts, create_ts, lines, last_user, last_ts, tags, archived,
		tmux_session_id, tmux_window_index, tmux_pane_index
	) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	ON CONFLICT(id) DO UPDATE SET
	  agent=excluded.agent,
	  cwd=excluded.cwd,
	  project=excluded.project,
	  branch=excluded.branch,
	  path=excluded.path,
	  mod_ts=excluded.mod_ts,
	  create_ts=excluded.create_ts,
	  lines=excluded.lines,
	  last_user=excluded.last_user,
	  last_ts=excluded.last_ts,
	  tags=excluded.tags,
	  archived=excluded.archived,
	  tmux_session_id=excluded.tmux_session_id,
	  tmux_window_index=excluded.tmux_window_index,
	  tmux_pane_index=excluded.tmux_pane_index
	`, sess.ID, sess.Agent, sess.CWD, sess.Project, sess.Branch, sess.Path, sess.ModTS, sess.CreateTS, sess.Lines, sess.LastUser, sess.LastTS, sess.Tags, boolToInt(sess.Archived), stringPtrValue(sess.TmuxSessionID), intPtrValue(sess.TmuxWindowIndex), intPtrValue(sess.TmuxPaneIndex))
	return err
}

// FindLast returns the most recent session for an optional agent.
func (s *SessionsStore) FindLast(ctx context.Context, agent string) (*Session, error) {
	q := `SELECT id, agent, cwd, project, branch, path, mod_ts, create_ts, lines, last_user, last_ts, tags, archived,
	       tmux_session_id, tmux_window_index, tmux_pane_index
          FROM sessions`
	args := []any{}
	if agent != "" {
		q += " WHERE agent = ?"
		args = append(args, agent)
	}
	q += " ORDER BY mod_ts DESC LIMIT 1"
	row := s.DB.QueryRowContext(ctx, q, args...)
	return s.scanSession(row)
}

// FindRecent returns the N most recent sessions.
func (s *SessionsStore) FindRecent(ctx context.Context, limit int) ([]Session, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, agent, cwd, project, branch, path, mod_ts, create_ts, lines, last_user, last_ts, tags, archived,
		       tmux_session_id, tmux_window_index, tmux_pane_index
		FROM sessions
		ORDER BY mod_ts DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanSessions(rows)
}

// Get retrieves a session by ID.
func (s *SessionsStore) Get(ctx context.Context, id string) (*Session, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, agent, cwd, project, branch, path, mod_ts, create_ts, lines, last_user, last_ts, tags, archived,
		       tmux_session_id, tmux_window_index, tmux_pane_index
		FROM sessions
		WHERE id = ?
	`, id)
	return s.scanSession(row)
}

// List returns all sessions.
func (s *SessionsStore) List(ctx context.Context) ([]Session, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, agent, cwd, project, branch, path, mod_ts, create_ts, lines, last_user, last_ts, tags, archived,
		       tmux_session_id, tmux_window_index, tmux_pane_index
		FROM sessions
		ORDER BY mod_ts DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanSessions(rows)
}

// Count returns the total number of sessions.
func (s *SessionsStore) Count(ctx context.Context) (int, error) {
	var count int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions`).Scan(&count)
	return count, err
}

// Delete removes a session by ID.
func (s *SessionsStore) Delete(ctx context.Context, id string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// Archive marks a session as archived.
func (s *SessionsStore) Archive(ctx context.Context, id string) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE sessions SET archived = 1 WHERE id = ?`, id)
	return err
}

func (s *SessionsStore) scanSession(row *sql.Row) (*Session, error) {
	var r Session
	var archivedInt int
	var tmuxSession sql.NullString
	var tmuxWindow sql.NullInt64
	var tmuxPane sql.NullInt64

	if err := row.Scan(&r.ID, &r.Agent, &r.CWD, &r.Project, &r.Branch, &r.Path, &r.ModTS, &r.CreateTS, &r.Lines, &r.LastUser, &r.LastTS, &r.Tags, &archivedInt, &tmuxSession, &tmuxWindow, &tmuxPane); err != nil {
		return nil, err
	}
	r.Archived = archivedInt != 0
	if tmuxSession.Valid {
		val := tmuxSession.String
		r.TmuxSessionID = &val
	}
	if tmuxWindow.Valid {
		val := int(tmuxWindow.Int64)
		r.TmuxWindowIndex = &val
	}
	if tmuxPane.Valid {
		val := int(tmuxPane.Int64)
		r.TmuxPaneIndex = &val
	}
	return &r, nil
}

func (s *SessionsStore) scanSessions(rows *sql.Rows) ([]Session, error) {
	var sessions []Session
	for rows.Next() {
		var r Session
		var archivedInt int
		var tmuxSession sql.NullString
		var tmuxWindow sql.NullInt64
		var tmuxPane sql.NullInt64

		if err := rows.Scan(&r.ID, &r.Agent, &r.CWD, &r.Project, &r.Branch, &r.Path, &r.ModTS, &r.CreateTS, &r.Lines, &r.LastUser, &r.LastTS, &r.Tags, &archivedInt, &tmuxSession, &tmuxWindow, &tmuxPane); err != nil {
			return nil, err
		}
		r.Archived = archivedInt != 0
		if tmuxSession.Valid {
			val := tmuxSession.String
			r.TmuxSessionID = &val
		}
		if tmuxWindow.Valid {
			val := int(tmuxWindow.Int64)
			r.TmuxWindowIndex = &val
		}
		if tmuxPane.Valid {
			val := int(tmuxPane.Int64)
			r.TmuxPaneIndex = &val
		}
		sessions = append(sessions, r)
	}
	return sessions, rows.Err()
}

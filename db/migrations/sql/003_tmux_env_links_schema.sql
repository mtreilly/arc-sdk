-- Tmux runs, env backups, links, settings, events
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS tmux_runs (
    id INTEGER PRIMARY KEY,
    ts INTEGER NOT NULL,
    pane_id TEXT,
    window_key TEXT,
    session TEXT,
    command TEXT,
    status TEXT,
    session_id TEXT NULL REFERENCES sessions(id) ON DELETE SET NULL,
    tags TEXT
);

CREATE INDEX IF NOT EXISTS idx_tmux_runs_ts ON tmux_runs(ts DESC);
CREATE INDEX IF NOT EXISTS idx_tmux_runs_session_id ON tmux_runs(session_id);

CREATE TABLE IF NOT EXISTS env_backups (
    project TEXT PRIMARY KEY,
    path TEXT,
    size INTEGER,
    mtime INTEGER,
    status TEXT
);

CREATE TABLE IF NOT EXISTS links (
    src_type TEXT,
    src_id TEXT,
    dst_type TEXT,
    dst_id TEXT,
    relation TEXT,
    ts INTEGER,
    PRIMARY KEY (src_type, src_id, dst_type, dst_id, relation)
);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT
);

CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY,
    ts INTEGER,
    type TEXT,
    payload TEXT
);


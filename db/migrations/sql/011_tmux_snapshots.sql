-- 011_tmux_snapshots.sql
-- Add tmux session monitoring and snapshot persistence
-- Links tmux sessions to vibe sessions and stores historical snapshots

PRAGMA foreign_keys = ON;

-- ============================================================================
-- 1. Extend sessions table with tmux metadata
-- ============================================================================

-- Link sessions to tmux panes (target for send/capture operations)
ALTER TABLE sessions ADD COLUMN tmux_session_id TEXT;
ALTER TABLE sessions ADD COLUMN tmux_window_index INTEGER;
ALTER TABLE sessions ADD COLUMN tmux_pane_index INTEGER;

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_sessions_tmux ON sessions(tmux_session_id, tmux_window_index, tmux_pane_index);

-- ============================================================================
-- 2. tmux_sessions - Metadata for tmux sessions
-- ============================================================================

CREATE TABLE IF NOT EXISTS tmux_sessions (
    session_id TEXT PRIMARY KEY,      -- tmux session ID ($0, $1, etc.)
    session_name TEXT NOT NULL,       -- User-visible name
    created_at INTEGER NOT NULL,      -- Unix timestamp
    last_activity INTEGER NOT NULL,   -- Unix timestamp
    last_snapshot INTEGER,            -- Unix timestamp of last snapshot
    is_attached INTEGER DEFAULT 0,    -- Boolean: any clients attached?
    is_stale INTEGER DEFAULT 0,       -- Boolean: marked as stale?
    stale_reason TEXT,                -- "inactive_1h" or "all_panes_dead"
    created_ts INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_tmux_sessions_name ON tmux_sessions(session_name);
CREATE INDEX IF NOT EXISTS idx_tmux_sessions_stale ON tmux_sessions(is_stale) WHERE is_stale = 1;
CREATE INDEX IF NOT EXISTS idx_tmux_sessions_activity ON tmux_sessions(last_activity DESC);

-- ============================================================================
-- 3. session_snapshots - Historical pane content
-- ============================================================================

CREATE TABLE IF NOT EXISTS session_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,         -- tmux session ID
    window_id TEXT NOT NULL,          -- tmux window ID (@0, @1, etc.)
    window_index INTEGER NOT NULL,    -- Window index (0, 1, etc.)
    window_name TEXT,                 -- Window name
    pane_id TEXT NOT NULL,            -- tmux pane ID (%0, %1, etc.)
    pane_index INTEGER NOT NULL,      -- Pane index (0, 1, etc.)
    pane_command TEXT,                -- Running command
    pane_title TEXT,                  -- Pane title
    pane_content TEXT,                -- Captured output (last 200 lines)
    pane_active INTEGER DEFAULT 0,    -- Boolean: active pane?
    pane_dead INTEGER DEFAULT 0,      -- Boolean: exited?
    pane_exit_status INTEGER,         -- Exit code if dead
    snapshot_time INTEGER NOT NULL,   -- Unix timestamp
    FOREIGN KEY (session_id) REFERENCES tmux_sessions(session_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_snapshots_session ON session_snapshots(session_id, snapshot_time DESC);
CREATE INDEX IF NOT EXISTS idx_snapshots_time ON session_snapshots(snapshot_time DESC);
CREATE INDEX IF NOT EXISTS idx_snapshots_pane ON session_snapshots(pane_id, snapshot_time DESC);

-- ============================================================================
-- 4. stale_session_log - Analytics for stale sessions
-- ============================================================================

CREATE TABLE IF NOT EXISTS stale_session_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,         -- tmux session ID
    session_name TEXT,                -- Session name (denormalized for history)
    detected_at INTEGER NOT NULL,     -- Unix timestamp when marked stale
    cleaned_at INTEGER,               -- Unix timestamp when killed (if applicable)
    last_activity INTEGER,            -- Last activity before going stale
    reason TEXT,                      -- "inactive_1h", "all_panes_dead", "manual"
    FOREIGN KEY (session_id) REFERENCES tmux_sessions(session_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_stale_log_detected ON stale_session_log(detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_stale_log_session ON stale_session_log(session_id);

-- ============================================================================
-- 5. Triggers for automatic stale detection
-- ============================================================================

-- Update last_snapshot timestamp when snapshots are inserted
CREATE TRIGGER IF NOT EXISTS update_last_snapshot AFTER INSERT ON session_snapshots BEGIN
    UPDATE tmux_sessions
    SET last_snapshot = NEW.snapshot_time
    WHERE session_id = NEW.session_id;
END;

-- Log when a session is marked stale
CREATE TRIGGER IF NOT EXISTS log_stale_detection AFTER UPDATE OF is_stale ON tmux_sessions
WHEN NEW.is_stale = 1 AND OLD.is_stale = 0 BEGIN
    INSERT INTO stale_session_log(session_id, session_name, detected_at, last_activity, reason)
    VALUES (NEW.session_id, NEW.session_name, strftime('%s', 'now'), NEW.last_activity, NEW.stale_reason);
END;

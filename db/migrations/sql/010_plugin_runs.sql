-- Plugin execution telemetry
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS plugin_runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_name TEXT NOT NULL,
    target TEXT NOT NULL,
    command_name TEXT,
    args TEXT,
    success INTEGER DEFAULT 0,
    exit_code INTEGER,
    duration_ms INTEGER,
    error TEXT,
    timestamp INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_plugin_runs_timestamp ON plugin_runs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_plugin_runs_plugin_name ON plugin_runs(plugin_name);
CREATE INDEX IF NOT EXISTS idx_plugin_runs_target ON plugin_runs(target);

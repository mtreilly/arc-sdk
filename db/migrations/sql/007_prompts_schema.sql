-- Prompt usage tracking and metadata
PRAGMA foreign_keys = ON;

-- Track prompt executions
CREATE TABLE IF NOT EXISTS prompt_usage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    prompt_name TEXT NOT NULL,
    model TEXT NOT NULL,
    params TEXT,  -- JSON-encoded parameters
    timestamp INTEGER NOT NULL,
    tokens_used INTEGER DEFAULT 0,
    success INTEGER DEFAULT 1,  -- 0 = failure, 1 = success
    session_id TEXT REFERENCES sessions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_prompt_usage_timestamp ON prompt_usage(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_prompt_usage_prompt_name ON prompt_usage(prompt_name);
CREATE INDEX IF NOT EXISTS idx_prompt_usage_session_id ON prompt_usage(session_id);

-- Optional: Prompt metadata (for cached prompts, templates)
CREATE TABLE IF NOT EXISTS prompt_metadata (
    name TEXT PRIMARY KEY,
    path TEXT NOT NULL,
    checksum TEXT,
    last_modified INTEGER,
    description TEXT
);

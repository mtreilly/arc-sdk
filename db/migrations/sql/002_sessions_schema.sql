-- Sessions and related metadata
PRAGMA foreign_keys = ON;

-- Core sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    agent TEXT NOT NULL,
    cwd TEXT,
    project TEXT,
    branch TEXT,
    path TEXT NOT NULL,
    mod_ts INTEGER,
    create_ts INTEGER,
    lines INTEGER DEFAULT 0,
    last_user TEXT,
    last_ts INTEGER,
    tags TEXT,
    archived INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_sessions_mod_ts ON sessions(mod_ts DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project);
CREATE INDEX IF NOT EXISTS idx_sessions_branch ON sessions(branch);


-- External repos table with FTS5 search
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS external_repos (
    -- Core identification
    name TEXT PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,

    -- Metadata
    description TEXT,
    platform TEXT,      -- 'GitHub', 'GitLab', 'Bitbucket', etc.
    owner TEXT,         -- e.g., 'simonw'
    repo TEXT,          -- e.g., 'llm'

    -- Clone status
    cloned INTEGER DEFAULT 0,         -- boolean: is repo cloned locally
    clone_path TEXT,                  -- e.g., 'external/llm'
    cloned_at INTEGER,                -- unix timestamp
    updated_at INTEGER,               -- unix timestamp of last git pull
    shallow INTEGER DEFAULT 0,        -- boolean: cloned with --depth 1

    -- Rich metadata (from GitHub API, README parsing, etc.)
    language TEXT,                    -- primary language
    topics TEXT,                      -- JSON array: ["llm", "cli", "ai"]
    stars INTEGER,                    -- GitHub stars
    forks INTEGER,
    license TEXT,
    homepage TEXT,

    -- Discovery & usage tracking
    added_at INTEGER NOT NULL,        -- unix timestamp when added to index
    added_by TEXT,                    -- 'user', 'claude', 'codex', 'script'
    last_opened_at INTEGER,           -- last time opened in browser
    access_count INTEGER DEFAULT 0,   -- number of times accessed

    -- Status & tags
    archived INTEGER DEFAULT 0,       -- boolean: archived/inactive
    tags TEXT,                        -- comma-separated or JSON array
    notes TEXT,                       -- user notes

    -- Git metadata
    default_branch TEXT,              -- e.g., 'main', 'master'
    commit_count INTEGER,
    last_commit_at INTEGER,           -- from git log

    -- Constraints
    CHECK(cloned IN (0,1)),
    CHECK(shallow IN (0,1)),
    CHECK(archived IN (0,1))
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_external_repos_platform ON external_repos(platform);
CREATE INDEX IF NOT EXISTS idx_external_repos_language ON external_repos(language);
CREATE INDEX IF NOT EXISTS idx_external_repos_added_at ON external_repos(added_at DESC);
CREATE INDEX IF NOT EXISTS idx_external_repos_cloned ON external_repos(cloned);
CREATE INDEX IF NOT EXISTS idx_external_repos_stars ON external_repos(stars DESC);

-- FTS5 table for full-text search
CREATE VIRTUAL TABLE IF NOT EXISTS external_repos_fts USING fts5(
    name,
    description,
    topics,
    tags,
    notes,
    content=external_repos,
    content_rowid=rowid
);

-- Triggers to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS external_repos_fts_insert AFTER INSERT ON external_repos BEGIN
    INSERT INTO external_repos_fts(rowid, name, description, topics, tags, notes)
    VALUES (new.rowid, new.name, new.description, new.topics, new.tags, new.notes);
END;

CREATE TRIGGER IF NOT EXISTS external_repos_fts_update AFTER UPDATE ON external_repos BEGIN
    UPDATE external_repos_fts SET
        name=new.name,
        description=new.description,
        topics=new.topics,
        tags=new.tags,
        notes=new.notes
    WHERE rowid=new.rowid;
END;

CREATE TRIGGER IF NOT EXISTS external_repos_fts_delete AFTER DELETE ON external_repos BEGIN
    DELETE FROM external_repos_fts WHERE rowid=old.rowid;
END;

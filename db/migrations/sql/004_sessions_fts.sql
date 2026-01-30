-- FTS5 virtual table for sessions text search
PRAGMA foreign_keys = ON;

-- Ensure FTS5 is available (modern SQLite). Create FTS table mirroring columns we search.
CREATE VIRTUAL TABLE IF NOT EXISTS sessions_fts USING fts5(
    id UNINDEXED,
    last_user,
    project,
    branch,
    content='sessions',
    content_rowid='rowid'
);

-- Triggers to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS sessions_ai AFTER INSERT ON sessions BEGIN
  INSERT INTO sessions_fts(rowid, id, last_user, project, branch)
  VALUES (new.rowid, new.id, new.last_user, new.project, new.branch);
END;

CREATE TRIGGER IF NOT EXISTS sessions_ad AFTER DELETE ON sessions BEGIN
  INSERT INTO sessions_fts(sessions_fts, rowid, id, last_user, project, branch)
  VALUES('delete', old.rowid, old.id, old.last_user, old.project, old.branch);
END;

CREATE TRIGGER IF NOT EXISTS sessions_au AFTER UPDATE ON sessions BEGIN
  INSERT INTO sessions_fts(sessions_fts, rowid, id, last_user, project, branch)
  VALUES('delete', old.rowid, old.id, old.last_user, old.project, old.branch);
  INSERT INTO sessions_fts(rowid, id, last_user, project, branch)
  VALUES (new.rowid, new.id, new.last_user, new.project, new.branch);
END;


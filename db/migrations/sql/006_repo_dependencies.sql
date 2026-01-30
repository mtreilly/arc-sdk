-- Repo dependencies table for tracking library dependencies
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS repo_dependencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_name TEXT NOT NULL,
    dependency_name TEXT NOT NULL,
    dependency_version TEXT,
    ecosystem TEXT NOT NULL,      -- 'go', 'npm', 'pypi', 'cargo', etc.
    dependency_type TEXT,         -- 'direct', 'dev', 'build', 'peer', etc.
    detected_at INTEGER NOT NULL, -- unix timestamp when detected

    FOREIGN KEY (repo_name) REFERENCES external_repos(name) ON DELETE CASCADE,
    UNIQUE(repo_name, dependency_name, ecosystem)
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_repo_dependencies_repo ON repo_dependencies(repo_name);
CREATE INDEX IF NOT EXISTS idx_repo_dependencies_dep ON repo_dependencies(dependency_name);
CREATE INDEX IF NOT EXISTS idx_repo_dependencies_ecosystem ON repo_dependencies(ecosystem);
CREATE INDEX IF NOT EXISTS idx_repo_dependencies_type ON repo_dependencies(dependency_type);

-- Composite index for "find repos using dependency X"
CREATE INDEX IF NOT EXISTS idx_repo_dependencies_dep_repo ON repo_dependencies(dependency_name, repo_name);

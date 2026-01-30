-- Schema per docs/plans/repo_analysis_tracking/REPO_ANALYSIS_TRACKING_SCHEMA.md
PRAGMA foreign_keys = ON;

-- Repository analysis history table
CREATE TABLE IF NOT EXISTS repo_analysis (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_name TEXT NOT NULL,
    analysis_type TEXT NOT NULL,
    prompt_template TEXT NOT NULL,
    full_prompt TEXT,
    analyzed_at INTEGER NOT NULL,
    analyzed_by TEXT NOT NULL,
    model TEXT,
    feature_name TEXT,
    context_files TEXT,
    tokens_used INTEGER DEFAULT 0,
    success INTEGER DEFAULT 1,
    output_path TEXT,
    notes TEXT,
    FOREIGN KEY(repo_name) REFERENCES external_repos(name) ON DELETE CASCADE,
    CHECK(success IN (0, 1))
);

-- Indexes to keep lookups fast
CREATE INDEX IF NOT EXISTS idx_repo_analysis_repo_name ON repo_analysis(repo_name);
CREATE INDEX IF NOT EXISTS idx_repo_analysis_analyzed_at ON repo_analysis(analyzed_at DESC);
CREATE INDEX IF NOT EXISTS idx_repo_analysis_type ON repo_analysis(analysis_type);
CREATE INDEX IF NOT EXISTS idx_repo_analysis_analyzed_by ON repo_analysis(analyzed_by);

-- Full text search support for prompts and notes
CREATE VIRTUAL TABLE IF NOT EXISTS repo_analysis_fts USING fts5(
    analysis_type,
    prompt_template,
    full_prompt,
    feature_name,
    notes,
    content=repo_analysis,
    content_rowid=id
);

CREATE TRIGGER IF NOT EXISTS repo_analysis_fts_insert AFTER INSERT ON repo_analysis BEGIN
    INSERT INTO repo_analysis_fts(
        rowid,
        analysis_type,
        prompt_template,
        full_prompt,
        feature_name,
        notes
    ) VALUES (
        new.id,
        new.analysis_type,
        new.prompt_template,
        new.full_prompt,
        new.feature_name,
        new.notes
    );
END;

CREATE TRIGGER IF NOT EXISTS repo_analysis_fts_update AFTER UPDATE ON repo_analysis BEGIN
    UPDATE repo_analysis_fts SET
        analysis_type = new.analysis_type,
        prompt_template = new.prompt_template,
        full_prompt = new.full_prompt,
        feature_name = new.feature_name,
        notes = new.notes
    WHERE rowid = new.id;
END;

CREATE TRIGGER IF NOT EXISTS repo_analysis_fts_delete AFTER DELETE ON repo_analysis BEGIN
    DELETE FROM repo_analysis_fts WHERE rowid = old.id;
END;

-- Summary view for quick lookup of latest analysis metadata
CREATE VIEW IF NOT EXISTS repo_analysis_summary AS
SELECT
    ra.repo_name,
    COUNT(*) AS analysis_count,
    MAX(ra.analyzed_at) AS last_analyzed_at,
    (
        SELECT analysis_type
        FROM repo_analysis rai
        WHERE rai.repo_name = ra.repo_name
        ORDER BY rai.analyzed_at DESC, rai.id DESC
        LIMIT 1
    ) AS last_analysis_type,
    (
        SELECT analyzed_by
        FROM repo_analysis rai
        WHERE rai.repo_name = ra.repo_name
        ORDER BY rai.analyzed_at DESC, rai.id DESC
        LIMIT 1
    ) AS last_analyzed_by
FROM repo_analysis ra
GROUP BY ra.repo_name;

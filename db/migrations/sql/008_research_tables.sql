-- 008_research_tables.sql
-- Research content tracking tables for papers, articles, and reading progress

PRAGMA foreign_keys = ON;

-- Papers table: arxiv-specific metadata extending items
CREATE TABLE IF NOT EXISTS papers (
    item_id TEXT PRIMARY KEY,
    arxiv_id TEXT UNIQUE,
    doi TEXT,
    pdf_url TEXT,
    primary_category TEXT,
    categories_json TEXT,
    abstract TEXT,
    published_date TEXT,
    updated_date TEXT,
    fetched_at TEXT,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_papers_arxiv_id ON papers(arxiv_id) WHERE arxiv_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_papers_published_date ON papers(published_date) WHERE published_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_papers_primary_category ON papers(primary_category) WHERE primary_category IS NOT NULL;

-- Reading status and progress tracking
CREATE TABLE IF NOT EXISTS reading_status (
    item_id TEXT PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'unread',
    priority INTEGER DEFAULT 0,
    progress_pct INTEGER DEFAULT 0,
    started_at TEXT,
    completed_at TEXT,
    last_accessed TEXT,
    notes_word_count INTEGER DEFAULT 0,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reading_status_status ON reading_status(status);
CREATE INDEX IF NOT EXISTS idx_reading_status_priority ON reading_status(priority);

-- Topics: thematic organization (agents, reasoning, retrieval, etc.)
CREATE TABLE IF NOT EXISTS topics (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Item-topic associations (many-to-many)
CREATE TABLE IF NOT EXISTS item_topics (
    item_id TEXT NOT NULL,
    topic_id INTEGER NOT NULL,
    relevance INTEGER DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    PRIMARY KEY (item_id, topic_id)
);

CREATE INDEX IF NOT EXISTS idx_item_topics_topic ON item_topics(topic_id);

-- Full-text search on paper abstracts and titles
CREATE VIRTUAL TABLE IF NOT EXISTS papers_fts USING fts5(
    item_id UNINDEXED,
    title,
    abstract,
    authors,
    categories,
    content='papers',
    tokenize='porter unicode61'
);

-- Trigger to keep FTS in sync with papers table
CREATE TRIGGER IF NOT EXISTS papers_fts_insert AFTER INSERT ON papers BEGIN
    INSERT INTO papers_fts(item_id, title, abstract, authors, categories)
    SELECT
        NEW.item_id,
        items.title,
        NEW.abstract,
        items.authors_json,
        NEW.categories_json
    FROM items WHERE items.id = NEW.item_id;
END;

CREATE TRIGGER IF NOT EXISTS papers_fts_update AFTER UPDATE ON papers BEGIN
    UPDATE papers_fts
    SET
        title = (SELECT title FROM items WHERE id = NEW.item_id),
        abstract = NEW.abstract,
        authors = (SELECT authors_json FROM items WHERE id = NEW.item_id),
        categories = NEW.categories_json
    WHERE item_id = NEW.item_id;
END;

CREATE TRIGGER IF NOT EXISTS papers_fts_delete AFTER DELETE ON papers BEGIN
    DELETE FROM papers_fts WHERE item_id = OLD.item_id;
END;


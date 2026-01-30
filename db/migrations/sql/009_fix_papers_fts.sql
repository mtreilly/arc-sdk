-- 009_fix_papers_fts.sql
-- Fix papers_fts table to remove incorrect content='papers' directive
-- The FTS table references columns (title, authors) that are in items, not papers

-- Drop the old FTS table and triggers
DROP TRIGGER IF EXISTS papers_fts_insert;
DROP TRIGGER IF EXISTS papers_fts_update;
DROP TRIGGER IF EXISTS papers_fts_delete;
DROP TABLE IF EXISTS papers_fts;

-- Recreate FTS table without content directive (regular contentful FTS)
CREATE VIRTUAL TABLE IF NOT EXISTS papers_fts USING fts5(
    item_id UNINDEXED,
    title,
    abstract,
    authors,
    categories,
    tokenize='porter unicode61'
);

-- Recreate triggers to keep FTS in sync
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

-- Repopulate FTS table from existing data
INSERT INTO papers_fts(item_id, title, abstract, authors, categories)
SELECT
    papers.item_id,
    items.title,
    papers.abstract,
    items.authors_json,
    papers.categories_json
FROM papers
JOIN items ON papers.item_id = items.id;


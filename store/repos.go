// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Repo represents an external repository stored in the database.
type Repo struct {
	Name        string
	URL         string
	Description string
	Platform    string
	Owner       string
	Repo        string

	Cloned    bool
	ClonePath string
	ClonedAt  int64
	UpdatedAt int64
	Shallow   bool

	Language string
	Topics   []string
	Stars    int
	Forks    int
	License  string
	Homepage string

	AddedAt      int64
	AddedBy      string
	LastOpenedAt int64
	AccessCount  int

	Archived bool
	Tags     []string
	Notes    string

	DefaultBranch string
	CommitCount   int
	LastCommitAt  int64

	AnalysisCount    int
	LastAnalyzedAt   int64
	LastAnalysisType string
	LastAnalyzedBy   string
}

// RepoAnalysis represents a recorded repository analysis entry.
type RepoAnalysis struct {
	ID             int64
	RepoName       string
	AnalysisType   string
	PromptTemplate string
	FullPrompt     string
	AnalyzedAt     int64
	AnalyzedBy     string
	Model          string
	FeatureName    string
	ContextFiles   []string
	TokensUsed     int
	Success        bool
	OutputPath     string
	Notes          string
}

// SearchOptions provides filters for repo search.
type SearchOptions struct {
	Language   string
	MinStars   int
	ClonedOnly bool
	Limit      int
}

// ReposStore manages repository data in the database.
type ReposStore struct {
	DB *sql.DB
}

// NewReposStore creates a new ReposStore.
func NewReposStore(db *sql.DB) *ReposStore {
	return &ReposStore{DB: db}
}

// Upsert inserts or updates a repo record.
func (s *ReposStore) Upsert(ctx context.Context, repo Repo) error {
	topicsJSON, err := json.Marshal(repo.Topics)
	if err != nil {
		return fmt.Errorf("marshal topics: %w", err)
	}
	tagsJSON, err := json.Marshal(repo.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO external_repos(
			name, url, description, platform, owner, repo,
			cloned, clone_path, cloned_at, updated_at, shallow,
			language, topics, stars, forks, license, homepage,
			added_at, added_by, last_opened_at, access_count,
			archived, tags, notes,
			default_branch, commit_count, last_commit_at
		) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(name) DO UPDATE SET
			url=excluded.url,
			description=excluded.description,
			platform=excluded.platform,
			owner=excluded.owner,
			repo=excluded.repo,
			cloned=excluded.cloned,
			clone_path=excluded.clone_path,
			cloned_at=excluded.cloned_at,
			updated_at=excluded.updated_at,
			shallow=excluded.shallow,
			language=excluded.language,
			topics=excluded.topics,
			stars=excluded.stars,
			forks=excluded.forks,
			license=excluded.license,
			homepage=excluded.homepage,
			added_by=excluded.added_by,
			last_opened_at=excluded.last_opened_at,
			access_count=excluded.access_count,
			archived=excluded.archived,
			tags=excluded.tags,
			notes=excluded.notes,
			default_branch=excluded.default_branch,
			commit_count=excluded.commit_count,
			last_commit_at=excluded.last_commit_at
	`,
		repo.Name, repo.URL, repo.Description, repo.Platform, repo.Owner, repo.Repo,
		boolToInt(repo.Cloned), repo.ClonePath, repo.ClonedAt, repo.UpdatedAt, boolToInt(repo.Shallow),
		repo.Language, string(topicsJSON), repo.Stars, repo.Forks, repo.License, repo.Homepage,
		repo.AddedAt, repo.AddedBy, repo.LastOpenedAt, repo.AccessCount,
		boolToInt(repo.Archived), string(tagsJSON), repo.Notes,
		repo.DefaultBranch, repo.CommitCount, repo.LastCommitAt,
	)
	return err
}

// Get retrieves a repo by name.
func (s *ReposStore) Get(ctx context.Context, name string) (*Repo, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT name, url, description, platform, owner, repo,
		       cloned, clone_path, cloned_at, updated_at, shallow,
		       language, topics, stars, forks, license, homepage,
		       added_at, added_by, last_opened_at, access_count,
		       archived, tags, notes,
		       default_branch, commit_count, last_commit_at
		FROM external_repos
		WHERE name = ?
	`, name)

	return s.scanRepo(row)
}

// GetByURL retrieves a repo by URL.
func (s *ReposStore) GetByURL(ctx context.Context, url string) (*Repo, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT name, url, description, platform, owner, repo,
		       cloned, clone_path, cloned_at, updated_at, shallow,
		       language, topics, stars, forks, license, homepage,
		       added_at, added_by, last_opened_at, access_count,
		       archived, tags, notes,
		       default_branch, commit_count, last_commit_at
		FROM external_repos
		WHERE url = ?
	`, url)

	return s.scanRepo(row)
}

// Delete removes a repo from the database.
func (s *ReposStore) Delete(ctx context.Context, name string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM external_repos WHERE name = ?`, name)
	return err
}

// List returns all repos.
func (s *ReposStore) List(ctx context.Context) ([]Repo, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT name, url, description, platform, owner, repo,
		       cloned, clone_path, cloned_at, updated_at, shallow,
		       language, topics, stars, forks, license, homepage,
		       added_at, added_by, last_opened_at, access_count,
		       archived, tags, notes,
		       default_branch, commit_count, last_commit_at
		FROM external_repos
		ORDER BY added_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanRepos(rows)
}

// ListCloned returns only cloned repos.
func (s *ReposStore) ListCloned(ctx context.Context) ([]Repo, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT name, url, description, platform, owner, repo,
		       cloned, clone_path, cloned_at, updated_at, shallow,
		       language, topics, stars, forks, license, homepage,
		       added_at, added_by, last_opened_at, access_count,
		       archived, tags, notes,
		       default_branch, commit_count, last_commit_at
		FROM external_repos
		WHERE cloned = 1
		ORDER BY added_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanRepos(rows)
}

// Count returns the total number of repos in the database.
func (s *ReposStore) Count(ctx context.Context) (int, error) {
	var count int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM external_repos`).Scan(&count)
	return count, err
}

// CountCloned returns the number of cloned repos.
func (s *ReposStore) CountCloned(ctx context.Context) (int, error) {
	var count int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM external_repos WHERE cloned = 1`).Scan(&count)
	return count, err
}

// MarkCloned updates a repo to mark it as cloned.
func (s *ReposStore) MarkCloned(ctx context.Context, name string, path string) error {
	now := time.Now().Unix()
	_, err := s.DB.ExecContext(ctx, `
		UPDATE external_repos
		SET cloned = 1, clone_path = ?, cloned_at = ?, updated_at = ?
		WHERE name = ?
	`, path, now, now, name)
	return err
}

// MarkUpdated updates the updated_at timestamp for a repo.
func (s *ReposStore) MarkUpdated(ctx context.Context, name string) error {
	now := time.Now().Unix()
	_, err := s.DB.ExecContext(ctx, `
		UPDATE external_repos
		SET updated_at = ?
		WHERE name = ?
	`, now, name)
	return err
}

// RecordOpen increments access count and updates last_opened_at.
func (s *ReposStore) RecordOpen(ctx context.Context, name string) error {
	now := time.Now().Unix()
	_, err := s.DB.ExecContext(ctx, `
		UPDATE external_repos
		SET access_count = access_count + 1, last_opened_at = ?
		WHERE name = ?
	`, now, name)
	return err
}

// Search performs full-text search on repos using FTS5 with optional filters.
func (s *ReposStore) Search(ctx context.Context, query string, opts SearchOptions) ([]Repo, error) {
	sqlQuery := `
		SELECT r.name, r.url, r.description, r.platform, r.owner, r.repo,
		       r.cloned, r.clone_path, r.cloned_at, r.updated_at, r.shallow,
		       r.language, r.topics, r.stars, r.forks, r.license, r.homepage,
		       r.added_at, r.added_by, r.last_opened_at, r.access_count,
		       r.archived, r.tags, r.notes,
		       r.default_branch, r.commit_count, r.last_commit_at
		FROM external_repos r
		JOIN external_repos_fts fts ON fts.rowid = r.rowid
		WHERE external_repos_fts MATCH ?`

	args := []interface{}{query}

	// Add filters
	if opts.Language != "" {
		sqlQuery += " AND r.language = ?"
		args = append(args, opts.Language)
	}
	if opts.MinStars > 0 {
		sqlQuery += " AND r.stars >= ?"
		args = append(args, opts.MinStars)
	}
	if opts.ClonedOnly {
		sqlQuery += " AND r.cloned = 1"
	}

	sqlQuery += " ORDER BY r.stars DESC"

	if opts.Limit > 0 {
		sqlQuery += " LIMIT ?"
		args = append(args, opts.Limit)
	}

	rows, err := s.DB.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanRepos(rows)
}

// AddTag adds a tag to a repo's tags array.
func (s *ReposStore) AddTag(ctx context.Context, name string, tag string) error {
	repo, err := s.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}

	for _, t := range repo.Tags {
		if t == tag {
			return nil
		}
	}

	repo.Tags = append(repo.Tags, tag)
	tagsJSON, err := json.Marshal(repo.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	_, err = s.DB.ExecContext(ctx, `UPDATE external_repos SET tags = ? WHERE name = ?`, string(tagsJSON), name)
	return err
}

// RemoveTag removes a tag from a repo's tags array.
func (s *ReposStore) RemoveTag(ctx context.Context, name string, tag string) error {
	repo, err := s.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}

	newTags := []string{}
	for _, t := range repo.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}

	tagsJSON, err := json.Marshal(newTags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	_, err = s.DB.ExecContext(ctx, `UPDATE external_repos SET tags = ? WHERE name = ?`, string(tagsJSON), name)
	return err
}

// SetTags replaces all tags for a repo.
func (s *ReposStore) SetTags(ctx context.Context, name string, tags []string) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	_, err = s.DB.ExecContext(ctx, `UPDATE external_repos SET tags = ? WHERE name = ?`, string(tagsJSON), name)
	return err
}

// FindByTag returns all repos that have a specific tag.
func (s *ReposStore) FindByTag(ctx context.Context, tag string) ([]Repo, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT name, url, description, platform, owner, repo,
		       cloned, clone_path, cloned_at, updated_at, shallow,
		       language, topics, stars, forks, license, homepage,
		       added_at, added_by, last_opened_at, access_count,
		       archived, tags, notes,
		       default_branch, commit_count, last_commit_at
		FROM external_repos
		WHERE tags LIKE ?
		ORDER BY stars DESC
	`, "%\""+tag+"\"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanRepos(rows)
}

// ListTags returns all unique tags with their usage counts.
func (s *ReposStore) ListTags(ctx context.Context) (map[string]int, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT tags FROM external_repos WHERE tags IS NOT NULL AND tags != '[]' AND tags != 'null'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagCounts := make(map[string]int)
	for rows.Next() {
		var tagsJSON string
		if err := rows.Scan(&tagsJSON); err != nil {
			return nil, err
		}

		var tags []string
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
			continue
		}

		for _, tag := range tags {
			tagCounts[tag]++
		}
	}

	return tagCounts, rows.Err()
}

// LanguageStats represents statistics for a specific language.
type LanguageStats struct {
	Language    string
	Count       int
	AvgStars    float64
	ClonedCount int
	TotalStars  int
}

// GetLanguageBreakdown returns statistics grouped by programming language.
func (s *ReposStore) GetLanguageBreakdown(ctx context.Context) ([]LanguageStats, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT
			COALESCE(language, 'Unknown') as lang,
			COUNT(*) as count,
			AVG(COALESCE(stars, 0)) as avg_stars,
			SUM(CASE WHEN cloned = 1 THEN 1 ELSE 0 END) as cloned_count,
			SUM(COALESCE(stars, 0)) as total_stars
		FROM external_repos
		GROUP BY language
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []LanguageStats
	for rows.Next() {
		var s LanguageStats
		if err := rows.Scan(&s.Language, &s.Count, &s.AvgStars, &s.ClonedCount, &s.TotalStars); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, rows.Err()
}

func (s *ReposStore) scanRepo(row *sql.Row) (*Repo, error) {
	var r Repo
	var clonedInt, shallowInt, archivedInt int
	var topicsJSON, tagsJSON string
	var clonedAt, updatedAt, lastOpenedAt, lastCommitAt sql.NullInt64
	var clonePath, language, license, homepage, addedBy, notes, defaultBranch sql.NullString

	err := row.Scan(
		&r.Name, &r.URL, &r.Description, &r.Platform, &r.Owner, &r.Repo,
		&clonedInt, &clonePath, &clonedAt, &updatedAt, &shallowInt,
		&language, &topicsJSON, &r.Stars, &r.Forks, &license, &homepage,
		&r.AddedAt, &addedBy, &lastOpenedAt, &r.AccessCount,
		&archivedInt, &tagsJSON, &notes,
		&defaultBranch, &r.CommitCount, &lastCommitAt,
	)
	if err != nil {
		return nil, err
	}

	r.Cloned = clonedInt != 0
	r.Shallow = shallowInt != 0
	r.Archived = archivedInt != 0

	if clonePath.Valid {
		r.ClonePath = clonePath.String
	}
	if clonedAt.Valid {
		r.ClonedAt = clonedAt.Int64
	}
	if updatedAt.Valid {
		r.UpdatedAt = updatedAt.Int64
	}
	if language.Valid {
		r.Language = language.String
	}
	if license.Valid {
		r.License = license.String
	}
	if homepage.Valid {
		r.Homepage = homepage.String
	}
	if addedBy.Valid {
		r.AddedBy = addedBy.String
	}
	if lastOpenedAt.Valid {
		r.LastOpenedAt = lastOpenedAt.Int64
	}
	if notes.Valid {
		r.Notes = notes.String
	}
	if defaultBranch.Valid {
		r.DefaultBranch = defaultBranch.String
	}
	if lastCommitAt.Valid {
		r.LastCommitAt = lastCommitAt.Int64
	}

	if topicsJSON != "" && topicsJSON != "null" {
		_ = json.Unmarshal([]byte(topicsJSON), &r.Topics)
	}
	if tagsJSON != "" && tagsJSON != "null" {
		_ = json.Unmarshal([]byte(tagsJSON), &r.Tags)
	}

	return &r, nil
}

func (s *ReposStore) scanRepos(rows *sql.Rows) ([]Repo, error) {
	var repos []Repo
	for rows.Next() {
		var r Repo
		var clonedInt, shallowInt, archivedInt int
		var topicsJSON, tagsJSON string
		var clonedAt, updatedAt, lastOpenedAt, lastCommitAt sql.NullInt64
		var clonePath, language, license, homepage, addedBy, notes, defaultBranch sql.NullString

		err := rows.Scan(
			&r.Name, &r.URL, &r.Description, &r.Platform, &r.Owner, &r.Repo,
			&clonedInt, &clonePath, &clonedAt, &updatedAt, &shallowInt,
			&language, &topicsJSON, &r.Stars, &r.Forks, &license, &homepage,
			&r.AddedAt, &addedBy, &lastOpenedAt, &r.AccessCount,
			&archivedInt, &tagsJSON, &notes,
			&defaultBranch, &r.CommitCount, &lastCommitAt,
		)
		if err != nil {
			return nil, err
		}

		r.Cloned = clonedInt != 0
		r.Shallow = shallowInt != 0
		r.Archived = archivedInt != 0

		if clonePath.Valid {
			r.ClonePath = clonePath.String
		}
		if clonedAt.Valid {
			r.ClonedAt = clonedAt.Int64
		}
		if updatedAt.Valid {
			r.UpdatedAt = updatedAt.Int64
		}
		if language.Valid {
			r.Language = language.String
		}
		if license.Valid {
			r.License = license.String
		}
		if homepage.Valid {
			r.Homepage = homepage.String
		}
		if addedBy.Valid {
			r.AddedBy = addedBy.String
		}
		if lastOpenedAt.Valid {
			r.LastOpenedAt = lastOpenedAt.Int64
		}
		if notes.Valid {
			r.Notes = notes.String
		}
		if defaultBranch.Valid {
			r.DefaultBranch = defaultBranch.String
		}
		if lastCommitAt.Valid {
			r.LastCommitAt = lastCommitAt.Int64
		}

		if topicsJSON != "" && topicsJSON != "null" {
			_ = json.Unmarshal([]byte(topicsJSON), &r.Topics)
		}
		if tagsJSON != "" && tagsJSON != "null" {
			_ = json.Unmarshal([]byte(tagsJSON), &r.Tags)
		}

		repos = append(repos, r)
	}

	return repos, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func stringPtrValue(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

func intPtrValue(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

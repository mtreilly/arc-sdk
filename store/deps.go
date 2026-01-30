// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Dependency represents a repository dependency.
type Dependency struct {
	Name       string
	Version    string
	Ecosystem  string
	Type       string
	DetectedAt time.Time
}

// RepoDependency represents a dependency stored in the database.
type RepoDependency struct {
	ID                int64
	RepoName          string
	DependencyName    string
	DependencyVersion string
	Ecosystem         string
	DependencyType    string
	DetectedAt        int64
}

// DepsStore manages repository dependencies in the database.
type DepsStore struct {
	DB *sql.DB
}

// NewDepsStore creates a new DepsStore.
func NewDepsStore(db *sql.DB) *DepsStore {
	return &DepsStore{DB: db}
}

// UpsertDependencies replaces all dependencies for a repo with the new set.
func (s *DepsStore) UpsertDependencies(ctx context.Context, repoName string, dependencies []Dependency) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing dependencies for this repo
	_, err = tx.ExecContext(ctx, `DELETE FROM repo_dependencies WHERE repo_name = ?`, repoName)
	if err != nil {
		return fmt.Errorf("delete existing dependencies: %w", err)
	}

	// Insert new dependencies
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO repo_dependencies(repo_name, dependency_name, dependency_version, ecosystem, dependency_type, detected_at)
		VALUES(?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, dep := range dependencies {
		_, err := stmt.ExecContext(ctx,
			repoName,
			dep.Name,
			dep.Version,
			dep.Ecosystem,
			dep.Type,
			dep.DetectedAt.Unix(),
		)
		if err != nil {
			return fmt.Errorf("insert dependency %s: %w", dep.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetDependencies returns all dependencies for a repo.
func (s *DepsStore) GetDependencies(ctx context.Context, repoName string) ([]RepoDependency, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, repo_name, dependency_name, dependency_version, ecosystem, dependency_type, detected_at
		FROM repo_dependencies
		WHERE repo_name = ?
		ORDER BY ecosystem, dependency_name
	`, repoName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []RepoDependency
	for rows.Next() {
		var d RepoDependency
		err := rows.Scan(&d.ID, &d.RepoName, &d.DependencyName, &d.DependencyVersion, &d.Ecosystem, &d.DependencyType, &d.DetectedAt)
		if err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}

// FindReposByDependency finds all repos that use a specific dependency.
func (s *DepsStore) FindReposByDependency(ctx context.Context, depName string) ([]RepoDependency, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, repo_name, dependency_name, dependency_version, ecosystem, dependency_type, detected_at
		FROM repo_dependencies
		WHERE dependency_name = ?
		ORDER BY repo_name
	`, depName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []RepoDependency
	for rows.Next() {
		var d RepoDependency
		err := rows.Scan(&d.ID, &d.RepoName, &d.DependencyName, &d.DependencyVersion, &d.Ecosystem, &d.DependencyType, &d.DetectedAt)
		if err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}

// GetEcosystemCounts returns counts of dependencies by ecosystem.
func (s *DepsStore) GetEcosystemCounts(ctx context.Context, repoName string) (map[string]int, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT ecosystem, COUNT(*) as count
		FROM repo_dependencies
		WHERE repo_name = ?
		GROUP BY ecosystem
	`, repoName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var ecosystem string
		var count int
		if err := rows.Scan(&ecosystem, &count); err != nil {
			return nil, err
		}
		counts[ecosystem] = count
	}
	return counts, rows.Err()
}

// GetCommonDependencies finds dependencies shared between two repos.
func (s *DepsStore) GetCommonDependencies(ctx context.Context, repo1, repo2 string) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT DISTINCT d1.dependency_name
		FROM repo_dependencies d1
		JOIN repo_dependencies d2
		  ON d1.dependency_name = d2.dependency_name
		  AND d1.ecosystem = d2.ecosystem
		WHERE d1.repo_name = ? AND d2.repo_name = ?
		ORDER BY d1.dependency_name
	`, repo1, repo2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []string
	for rows.Next() {
		var dep string
		if err := rows.Scan(&dep); err != nil {
			return nil, err
		}
		deps = append(deps, dep)
	}
	return deps, rows.Err()
}

// SimilarRepo represents a repo with similar dependencies.
type SimilarRepo struct {
	RepoName    string
	CommonCount int
}

// GetSimilarRepos finds repos with similar dependencies.
func (s *DepsStore) GetSimilarRepos(ctx context.Context, repoName string, limit int) ([]SimilarRepo, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT d2.repo_name, COUNT(DISTINCT d1.dependency_name) as common_count
		FROM repo_dependencies d1
		JOIN repo_dependencies d2
		  ON d1.dependency_name = d2.dependency_name
		  AND d1.ecosystem = d2.ecosystem
		WHERE d1.repo_name = ? AND d2.repo_name != ?
		GROUP BY d2.repo_name
		ORDER BY common_count DESC
		LIMIT ?
	`, repoName, repoName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SimilarRepo
	for rows.Next() {
		var r SimilarRepo
		if err := rows.Scan(&r.RepoName, &r.CommonCount); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// DeleteDependencies removes all dependencies for a repo.
func (s *DepsStore) DeleteDependencies(ctx context.Context, repoName string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM repo_dependencies WHERE repo_name = ?`, repoName)
	return err
}

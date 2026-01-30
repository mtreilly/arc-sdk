// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Clone clones a repository to the given path.
func Clone(ctx context.Context, url, destPath string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", url, destPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}

	return nil
}

// CloneShallow clones a repository with depth=1.
func CloneShallow(ctx context.Context, url, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", url, destPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone --depth 1: %w", err)
	}

	return nil
}

// Pull performs a git pull in the given directory.
func Pull(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}

	return nil
}

// Fetch performs a git fetch in the given directory.
func Fetch(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "fetch")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch: %w", err)
	}

	return nil
}

// IsGitRepo checks if the given path is a git repository.
func IsGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetDefaultBranch returns the default branch of a repository.
func GetDefaultBranch(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "symbolic-ref", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("get default branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch returns the current branch of a repository.
func GetCurrentBranch(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetRemoteURL returns the URL of the origin remote.
func GetRemoteURL(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("get remote url: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCommitCount returns the number of commits in a repository.
func GetCommitCount(ctx context.Context, repoPath string) (int, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-list", "--count", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("count commits: %w", err)
	}

	var count int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count); err != nil {
		return 0, err
	}
	return count, nil
}

// GetLastCommitTime returns the Unix timestamp of the last commit.
func GetLastCommitTime(ctx context.Context, repoPath string) (int64, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "log", "-1", "--format=%ct")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("get last commit time: %w", err)
	}

	var ts int64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &ts); err != nil {
		return 0, err
	}
	return ts, nil
}

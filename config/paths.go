// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"path/filepath"
)

// DefaultConfigPath returns the default XDG-compliant config file path and
// ensures the parent directory exists (0755).
//
// Order (only returns the global path for backward compatibility):
// 1) $XDG_CONFIG_HOME/arc/config.yaml
// 2) ~/.config/arc/config.yaml
func DefaultConfigPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	dir := filepath.Join(base, "arc")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "config.yaml")
}

// ConfigSearchPaths returns ordered list of paths to search for config files.
// Search order (respects ARC_CONFIG env var and explicit paths):
//
// 1) Explicit path (passed as argument)
// 2) $ARC_CONFIG environment variable
// 3) .arc/config.yaml (project root, searches up to find it)
// 4) $XDG_CONFIG_HOME/arc/config.yaml or ~/.config/arc/config.yaml (global)
func ConfigSearchPaths(explicit string) []string {
	envPath := os.Getenv("ARC_CONFIG")
	home, _ := os.UserHomeDir()

	candidates := []string{}
	seen := make(map[string]struct{})

	push := func(p string) {
		if p == "" {
			return
		}
		normalized := filepath.Clean(p)
		if _, ok := seen[normalized]; ok {
			return
		}
		seen[normalized] = struct{}{}
		candidates = append(candidates, normalized)
	}

	// Priority 1: Explicit path (from flag)
	push(explicit)

	// Priority 2: Environment variable
	push(envPath)

	// Priority 3: Project-local config (.arc/config.yaml)
	push(findProjectConfig())

	// Priority 4: Global config
	xdgBase := os.Getenv("XDG_CONFIG_HOME")
	if xdgBase == "" {
		xdgBase = filepath.Join(home, ".config")
	}
	push(filepath.Join(xdgBase, "arc", "config.yaml"))

	return candidates
}

// findProjectConfig searches for .arc/config.yaml starting from current directory
// and walking up the directory tree until it finds one or reaches root.
func findProjectConfig() string {
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		configPath := filepath.Join(pwd, ".arc", "config.yaml")
		if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
			return configPath
		}

		parent := filepath.Dir(pwd)
		if parent == pwd {
			// Reached root directory
			break
		}
		pwd = parent
	}

	return ""
}

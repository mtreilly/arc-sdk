// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Load reads configuration with precedence: flags (bound elsewhere) > env > file > defaults.
// It returns a fully-populated Config.
//
// Configuration is loaded from the first available source in order:
// 1. CLI flags (bound elsewhere)
// 2. Environment variables (ARC_*)
// 3. .arc/config.yaml (project-local, searched up directory tree)
// 4. ~/.config/arc/config.yaml (global)
// 5. Default values
func Load() (*Config, error) {
	return LoadWithPath("")
}

// LoadWithPath is like Load but accepts an explicit config path.
func LoadWithPath(explicitPath string) (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("research_root", "~/arc-engineering/docs/research-external")
	v.SetDefault("external_root", "~/arc-engineering/external")
	v.SetDefault("taxonomy_path", "~/arc-engineering/docs/research-external/taxonomy.yaml")
	v.SetDefault("claude.bin", "claude")
	v.SetDefault("claude.model", "claude-sonnet-4-5-20250929")
	v.SetDefault("concurrency.fetch", 4)
	v.SetDefault("concurrency.analyze", 2)

	// Env
	v.SetEnvPrefix("ARC")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// File - search in order and load first available
	searchPaths := ConfigSearchPaths(explicitPath)
	for _, path := range searchPaths {
		if path == "" {
			continue
		}
		path = ExpandPath(path)
		if info, err := filepath.Abs(path); err == nil {
			path = info
		}
		if _, err := os.Stat(path); err != nil {
			continue
		}

		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err == nil {
			break
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Expand paths (~ and env vars)
	cfg.ResearchRoot = ExpandPath(cfg.ResearchRoot)
	cfg.ExternalRoot = ExpandPath(cfg.ExternalRoot)
	cfg.TaxonomyPath = ExpandPath(cfg.TaxonomyPath)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

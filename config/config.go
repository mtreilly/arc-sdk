// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the application configuration.
type Config struct {
	ResearchRoot string            `mapstructure:"research_root"`
	ExternalRoot string            `mapstructure:"external_root"`
	TaxonomyPath string            `mapstructure:"taxonomy_path"`
	Claude       ClaudeConfig      `mapstructure:"claude"`
	Concurrency  ConcurrencyConfig `mapstructure:"concurrency"`
	AI           AIConfig          `mapstructure:"ai"`
	Discord      DiscordConfig     `mapstructure:"discord"`
}

// ClaudeConfig holds Claude CLI configuration.
type ClaudeConfig struct {
	Bin   string `mapstructure:"bin"`
	Model string `mapstructure:"model"`
}

// AIConfig holds AI provider configuration.
type AIConfig struct {
	Provider     string  `mapstructure:"provider"`
	APIKey       string  `mapstructure:"api_key"`
	APIKeyEnv    string  `mapstructure:"api_key_env"`
	DefaultModel string  `mapstructure:"default_model"`
	MaxTokens    int     `mapstructure:"max_tokens"`
	Temperature  float64 `mapstructure:"temperature"`
	Timeout      string  `mapstructure:"timeout"`
}

// ConcurrencyConfig holds concurrency limits.
type ConcurrencyConfig struct {
	Fetch   int `mapstructure:"fetch"`
	Analyze int `mapstructure:"analyze"`
}

// DiscordConfig holds Discord integration configuration.
type DiscordConfig struct {
	BotToken       string            `mapstructure:"bot_token"`
	Webhooks       map[string]string `mapstructure:"webhooks"`
	DefaultWebhook string            `mapstructure:"default_webhook"`
	Templates      DiscordTemplates  `mapstructure:"templates"`
}

// DiscordTemplates holds Discord message templates.
type DiscordTemplates struct {
	Update string `mapstructure:"update"`
	Alert  string `mapstructure:"alert"`
	Log    string `mapstructure:"log"`
}

// ExpandPath expands ~ to the user's home directory and performs environment
// variable expansion (e.g., $HOME, $VAR) on the provided path.
func ExpandPath(p string) string {
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			home = os.Getenv("HOME")
		}
		if home != "" {
			p = filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	p = os.ExpandEnv(p)
	return p
}

// Validate checks for sensible values in the configuration.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.ResearchRoot) == "" {
		return errors.New("research_root cannot be empty")
	}
	if strings.TrimSpace(c.ExternalRoot) == "" {
		return errors.New("external_root cannot be empty")
	}
	if c.Concurrency.Fetch <= 0 {
		return errors.New("concurrency.fetch must be > 0")
	}
	if c.Concurrency.Analyze <= 0 {
		return errors.New("concurrency.analyze must be > 0")
	}
	if strings.TrimSpace(c.Claude.Bin) == "" {
		return errors.New("claude.bin cannot be empty")
	}
	return nil
}

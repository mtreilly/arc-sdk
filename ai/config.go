// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds configuration for creating AI clients.
type Config struct {
	// Provider is the AI provider name ("anthropic", "openai", "local")
	Provider string `yaml:"provider"`

	// APIKey is the API key for the provider
	APIKey string `yaml:"api_key"`

	// DefaultModel is the default model to use
	DefaultModel string `yaml:"default_model"`

	// Timeout is the default request timeout
	Timeout time.Duration `yaml:"timeout"`

	// MaxTokens is the default max tokens
	MaxTokens int `yaml:"max_tokens"`

	// Temperature is the default temperature
	Temperature float64 `yaml:"temperature"`

	// CommandDefaults allows configuring provider/model overrides per command
	CommandDefaults map[string]CommandDefaultConfig `yaml:"command_defaults"`
}

// CommandDefaultConfig describes overrides for a specific command.
type CommandDefaultConfig struct {
	Provider string   `yaml:"provider"`
	Model    string   `yaml:"model"`
	CLIArgs  []string `yaml:"cli_args"`
}

// LoadConfig loads AI configuration from the default path.
func LoadConfig() (*Config, error) {
	return LoadConfigFromPath(ConfigPath())
}

// LoadConfigFromPath loads AI configuration from a specific path.
func LoadConfigFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Provider:     "codex",
				DefaultModel: "claude-sonnet-4-5-20250929",
				Timeout:      2 * time.Minute,
				MaxTokens:    4096,
				Temperature:  0.7,
			}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// ConfigPath returns the default AI config path.
func ConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "arc", "ai.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "arc", "ai.yaml")
}

// ValidateConfig checks the config for common errors.
func ValidateConfig(cfg *Config) error {
	if cfg.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if cfg.MaxTokens < 0 {
		return fmt.Errorf("max_tokens must be >= 0")
	}
	if cfg.Temperature < 0 || cfg.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	return nil
}

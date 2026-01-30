// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package ai

import (
	"context"
	"time"
)

// Service wraps an AIClient with additional functionality.
type Service struct {
	client AIClient
	config Config
}

// NewService creates a new AI service with the given client and config.
func NewService(client AIClient, config Config) *Service {
	return &Service{
		client: client,
		config: config,
	}
}

// RunOptions contains options for running an AI request.
type RunOptions struct {
	System      string
	Prompt      string
	Model       string
	MaxTokens   int
	Temperature float64
	Timeout     time.Duration
	CLIArgs     []string
}

// Run executes an AI request with defaults from config.
func (s *Service) Run(ctx context.Context, opts RunOptions) (Response, error) {
	req := Request{
		System:      opts.System,
		Prompt:      opts.Prompt,
		Model:       opts.Model,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		Timeout:     opts.Timeout,
		CLIArgs:     opts.CLIArgs,
	}

	// Apply defaults from config
	if req.Model == "" {
		req.Model = s.config.DefaultModel
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = s.config.MaxTokens
	}
	if req.Temperature == 0 {
		req.Temperature = s.config.Temperature
	}
	if req.Timeout == 0 {
		req.Timeout = s.config.Timeout
	}

	return s.client.Ask(ctx, req)
}

// Client returns the underlying AI client.
func (s *Service) Client() AIClient {
	return s.client
}

// Models returns available models.
func (s *Service) Models() []string {
	return s.client.Models()
}

// ClientFactory creates AI clients based on provider name.
type ClientFactory func(cfg Config) (AIClient, error)

var clientFactories = make(map[string]ClientFactory)

// RegisterClient registers a client factory for a provider.
func RegisterClient(provider string, factory ClientFactory) {
	clientFactories[provider] = factory
}

// NewClient creates a client for the given config.
func NewClient(cfg Config) (AIClient, error) {
	factory, ok := clientFactories[cfg.Provider]
	if !ok {
		return nil, &ClientError{
			Provider: cfg.Provider,
			Message:  "unknown provider",
			Hint:     "Supported providers: anthropic, claude, codex, openrouter, locallm",
		}
	}
	return factory(cfg)
}

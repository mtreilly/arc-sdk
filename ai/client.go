// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package ai

import (
	"context"
	"time"
)

// AIClient is the interface for AI model providers.
type AIClient interface {
	// Ask sends a request to the AI model and returns the response.
	Ask(ctx context.Context, req Request) (Response, error)

	// Models returns the list of available model IDs.
	Models() []string
}

// Request contains the input to the AI model.
type Request struct {
	// System is the system prompt (optional, provider-specific)
	System string

	// Prompt is the user message/prompt
	Prompt string

	// Model is the model ID (e.g., "claude-sonnet-4-5")
	Model string

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int

	// Temperature controls randomness (0.0 to 1.0)
	Temperature float64

	// Timeout is the request timeout duration
	Timeout time.Duration

	// CLIArgs passes extra arguments to CLI-based providers (Codex, etc.)
	CLIArgs []string
}

// Response contains the AI model's output.
type Response struct {
	// Text is the generated text response
	Text string `json:"text"`

	// Model is the actual model used
	Model string `json:"model"`

	// Usage tracks token consumption
	Usage TokenUsage `json:"usage"`

	// Latency is the time taken for the request
	Latency time.Duration `json:"latency_ms"`

	// Metadata contains provider-specific data
	Metadata map[string]any `json:"metadata,omitempty"`
}

// TokenUsage tracks API usage for cost estimation.
type TokenUsage struct {
	Input  int `json:"input_tokens"`
	Output int `json:"output_tokens"`
	Total  int `json:"total_tokens"`
}

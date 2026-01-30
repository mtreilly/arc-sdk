// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package ai

import "fmt"

// ClientError represents an error from the AI client.
type ClientError struct {
	Provider string
	Message  string
	Hint     string
	Err      error
}

func (e *ClientError) Error() string {
	msg := e.Provider + ": " + e.Message
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	if e.Hint != "" {
		msg += "\nHint: " + e.Hint
	}
	return msg
}

func (e *ClientError) Unwrap() error {
	return e.Err
}

// ConfigError represents a configuration error.
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error: %s: %s", e.Field, e.Message)
}

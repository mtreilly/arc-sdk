// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package errors

import "fmt"

// CLIError represents a user-facing CLI error with helpful hints and suggestions.
type CLIError struct {
	Msg         string
	Hint        string
	Suggestions []string
	cause       error
}

// NewCLIError creates a new CLI error with the given message.
func NewCLIError(msg string) *CLIError {
	return &CLIError{Msg: msg}
}

// WithHint adds a hint to the error.
func (e *CLIError) WithHint(hint string) *CLIError {
	e.Hint = hint
	return e
}

// WithSuggestions adds suggestions to the error.
func (e *CLIError) WithSuggestions(suggestions ...string) *CLIError {
	e.Suggestions = suggestions
	return e
}

// WithCause adds an underlying cause to the error.
func (e *CLIError) WithCause(cause error) *CLIError {
	e.cause = cause
	return e
}

// Error implements the error interface.
func (e *CLIError) Error() string {
	msg := fmt.Sprintf("Error: %s", e.Msg)

	if e.cause != nil {
		msg += fmt.Sprintf(": %v", e.cause)
	}

	if e.Hint != "" {
		msg += fmt.Sprintf("\n\nHint: %s", e.Hint)
	}

	if len(e.Suggestions) > 0 {
		msg += "\n\nTry:"
		for _, s := range e.Suggestions {
			msg += fmt.Sprintf("\n  %s", s)
		}
	}

	return msg
}

// Unwrap returns the underlying cause.
func (e *CLIError) Unwrap() error {
	return e.cause
}

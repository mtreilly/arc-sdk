// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package utils

import (
	"os"
	"path/filepath"
	"strings"
)

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

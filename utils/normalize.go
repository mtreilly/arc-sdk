// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package utils

import (
	"regexp"
	"strings"
)

var arxivIDRegex = regexp.MustCompile(`(?:arxiv[:/])?(\d{4}\.\d{4,5})(v\d+)?`)

// NormalizeArxivID extracts and normalizes an arXiv paper ID.
// It handles various formats:
//   - 2301.00001
//   - 2301.00001v1
//   - arxiv:2301.00001
//   - https://arxiv.org/abs/2301.00001
//   - https://arxiv.org/pdf/2301.00001.pdf
func NormalizeArxivID(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)

	// Remove common URL prefixes
	input = strings.TrimPrefix(input, "https://arxiv.org/abs/")
	input = strings.TrimPrefix(input, "http://arxiv.org/abs/")
	input = strings.TrimPrefix(input, "https://arxiv.org/pdf/")
	input = strings.TrimPrefix(input, "http://arxiv.org/pdf/")
	input = strings.TrimSuffix(input, ".pdf")

	// Extract ID using regex
	matches := arxivIDRegex.FindStringSubmatch(input)
	if len(matches) >= 2 {
		id := matches[1]
		if len(matches) >= 3 && matches[2] != "" {
			id += matches[2]
		}
		return id
	}

	return input
}

// NormalizeWhitespace replaces multiple whitespace with single spaces.
func NormalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// TruncateString truncates a string to maxLen characters, adding "..." if truncated.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// SanitizeFilename removes or replaces characters that are invalid in filenames.
func SanitizeFilename(name string) string {
	// Replace problematic characters
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	name = replacer.Replace(name)

	// Trim spaces and dots from ends
	name = strings.Trim(name, " .")

	return name
}

// SlugifyString converts a string to a URL-safe slug.
func SlugifyString(s string) string {
	s = strings.ToLower(s)
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

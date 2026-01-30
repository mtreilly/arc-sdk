// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// JSON writes v as indented JSON to stdout.
func JSON(v any) error {
	return JSONTo(os.Stdout, v)
}

// JSONTo writes v as indented JSON to w.
func JSONTo(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// JSONCompact writes v as compact JSON to stdout.
func JSONCompact(v any) error {
	return JSONCompactTo(os.Stdout, v)
}

// JSONCompactTo writes v as compact JSON to w.
func JSONCompactTo(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

// YAML writes v as YAML to stdout.
func YAML(v any) error {
	return YAMLTo(os.Stdout, v)
}

// YAMLTo writes v as YAML to w.
func YAMLTo(w io.Writer, v any) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(v)
}

// Print prints the value based on the output format.
func Print(format OutputFormat, v any) error {
	switch format {
	case OutputJSON:
		return JSON(v)
	case OutputYAML:
		return YAML(v)
	case OutputQuiet:
		return nil
	default:
		// For table format, caller should handle separately
		return fmt.Errorf("unsupported format for Print: %s", format)
	}
}

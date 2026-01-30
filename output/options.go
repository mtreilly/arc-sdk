// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package output

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourorg/arc-sdk/errors"
)

// OutputFormat defines allowed values for --output.
type OutputFormat string

const (
	OutputTable OutputFormat = "table"
	OutputJSON  OutputFormat = "json"
	OutputYAML  OutputFormat = "yaml"
	OutputQuiet OutputFormat = "quiet"
)

var validOutputFormats = map[OutputFormat]struct{}{
	OutputTable: {},
	OutputJSON:  {},
	OutputYAML:  {},
	OutputQuiet: {},
}

// OutputOptions captures shared output flag behavior across commands.
type OutputOptions struct {
	Format string

	legacyJSON  bool
	legacyQuiet bool

	defaultFormat OutputFormat
}

// AddOutputFlags wires the standard output flags into a command.
func (o *OutputOptions) AddOutputFlags(cmd *cobra.Command, defaultFormat OutputFormat) {
	if defaultFormat == "" {
		defaultFormat = OutputTable
	}
	o.defaultFormat = defaultFormat

	cmd.Flags().StringVarP(&o.Format, "output", "o", string(defaultFormat), "Output format: table|json|yaml|quiet")
	cmd.Flags().BoolVar(&o.legacyJSON, "json", false, "Output JSON (deprecated, use --output json)")
	cmd.Flags().BoolVar(&o.legacyQuiet, "quiet", false, "Quiet mode (deprecated, use --output quiet)")

	_ = cmd.Flags().MarkHidden("json")
	_ = cmd.Flags().MarkHidden("quiet")
}

// Resolve validates the provided output options, applying deprecated aliases when used.
func (o *OutputOptions) Resolve() error {
	format := strings.ToLower(strings.TrimSpace(o.Format))

	switch {
	case o.legacyJSON:
		format = string(OutputJSON)
	case o.legacyQuiet:
		format = string(OutputQuiet)
	case format == "":
		format = string(o.defaultFormat)
	}

	if _, ok := validOutputFormats[OutputFormat(format)]; !ok {
		return errors.NewCLIError(fmt.Sprintf("invalid output format: %s", format)).
			WithHint("Use --output table|json|yaml|quiet").
			WithSuggestions(
				"--output table",
				"--output json",
				"--output yaml",
				"--output quiet",
			)
	}

	o.Format = format
	return nil
}

// Is reports whether the resolved output format matches the provided format.
func (o *OutputOptions) Is(format OutputFormat) bool {
	return o.Format == string(format)
}

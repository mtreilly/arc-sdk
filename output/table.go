// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package output

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Table is a simple table builder for CLI output.
type Table struct {
	headers []string
	rows    [][]string
	w       io.Writer
}

// NewTable creates a new table with the given headers.
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
		rows:    make([][]string, 0),
		w:       os.Stdout,
	}
}

// SetWriter sets the output writer.
func (t *Table) SetWriter(w io.Writer) *Table {
	t.w = w
	return t
}

// AddRow adds a row to the table.
func (t *Table) AddRow(cols ...string) *Table {
	// Ensure row has same number of columns as headers
	row := make([]string, len(t.headers))
	for i := range row {
		if i < len(cols) {
			row[i] = cols[i]
		}
	}
	t.rows = append(t.rows, row)
	return t
}

// Render prints the table.
func (t *Table) Render() {
	if len(t.headers) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
		widths[i] = len(h)
	}
	for _, row := range t.rows {
		for i, col := range row {
			if len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	// Print header
	t.printRow(t.headers, widths)
	t.printSeparator(widths)

	// Print rows
	for _, row := range t.rows {
		t.printRow(row, widths)
	}
}

func (t *Table) printRow(cols []string, widths []int) {
	parts := make([]string, len(cols))
	for i, col := range cols {
		parts[i] = fmt.Sprintf("%-*s", widths[i], col)
	}
	fmt.Fprintln(t.w, strings.Join(parts, "  "))
}

func (t *Table) printSeparator(widths []int) {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("-", w)
	}
	fmt.Fprintln(t.w, strings.Join(parts, "  "))
}

// Count returns the number of rows.
func (t *Table) Count() int {
	return len(t.rows)
}

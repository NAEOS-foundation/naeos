package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// FormatOutput marshals data in the requested format and writes it to w.
func FormatOutput(w io.Writer, data any, format string) error {
	switch format {
	case "json":
		result, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
		_, err = w.Write(append(result, '\n'))
		return err
	case "yaml":
		result, err := yaml.Marshal(data)
		if err != nil {
			return fmt.Errorf("encode yaml: %w", err)
		}
		_, err = w.Write(result)
		return err
	case "table":
		if tf, ok := data.(TableFormatter); ok {
			return tf.FormatTable(w)
		}
		result, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("encode json fallback: %w", err)
		}
		_, err = w.Write(append(result, '\n'))
		return err
	default:
		return fmt.Errorf("unsupported output format %q (use json, yaml, or table)", format)
	}
}

// TableFormatter is implemented by types that can render themselves as a table.
type TableFormatter interface {
	FormatTable(w io.Writer) error
}

// FormatTable writes a simple ASCII table to w with the given headers and rows.
func FormatTable(w io.Writer, headers []string, rows [][]string) error {
	if len(headers) == 0 {
		return nil
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	sep := "+"
	for _, w := range widths {
		sep += strings.Repeat("-", w+2) + "+"
	}

	// Header
	fmt.Fprint(w, sep+"\n")
	fmt.Fprint(w, "|")
	for i, h := range headers {
		fmt.Fprintf(w, " %-*s |", widths[i], h)
	}
	fmt.Fprintln(w)

	// Separator
	fmt.Fprintln(w, sep)

	// Rows
	for _, row := range rows {
		fmt.Fprint(w, "|")
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			fmt.Fprintf(w, " %-*s |", widths[i], cell)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, sep)
	return nil
}

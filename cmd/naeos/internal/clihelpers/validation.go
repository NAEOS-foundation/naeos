// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package clihelpers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// ValidationError represents a single validation error.
type ValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	Line    int    `json:"line,omitempty"`
}

// ValidationResult represents the outcome of a specification validation.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Status   string            `json:"status"`
	Project  string            `json:"project,omitempty"`
	Modules  int               `json:"modules,omitempty"`
	Services int               `json:"services,omitempty"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
	Summary  string            `json:"summary"`
}

// RenderValidation renders a validation result in the requested format.
func RenderValidation(cmd *cobra.Command, vr ValidationResult, format, filePath string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(vr, "", "  ")
		if err != nil {
			return err
		}
		if filePath != "" {
			return os.WriteFile(filePath, data, 0o600)
		}
		_, _ = cmd.OutOrStdout().Write(data)
		_, _ = cmd.OutOrStdout().Write([]byte("\n"))
		return nil
	case "yaml":
		var sb strings.Builder
		fmt.Fprintf(&sb, "status: %s\n", vr.Status)
		fmt.Fprintf(&sb, "valid: %t\n", vr.Valid)
		if vr.Summary != "" {
			fmt.Fprintf(&sb, "summary: %s\n", vr.Summary)
		}
		for _, e := range vr.Errors {
			fmt.Fprintf(&sb, "errors:\n  - code: %s\n    message: %s\n", e.Code, e.Message)
		}
		if filePath != "" {
			return os.WriteFile(filePath, []byte(sb.String()), 0o600)
		}
		_, _ = cmd.OutOrStdout().Write([]byte(sb.String()))
		return nil
	default:
		var sb strings.Builder
		if vr.Valid {
			sb.WriteString("✓ ")
		} else {
			sb.WriteString("✗ ")
		}
		sb.WriteString(vr.Summary)
		sb.WriteString("\n")

		for _, e := range vr.Errors {
			fmt.Fprintf(&sb, "  [%s] %s\n", e.Code, e.Message)
		}
		for _, w := range vr.Warnings {
			fmt.Fprintf(&sb, "  ⚠ %s\n", w)
		}

		if filePath != "" {
			return os.WriteFile(filePath, []byte(sb.String()), 0o600)
		}
		_, _ = cmd.OutOrStdout().Write([]byte(sb.String()))
		return nil
	}
}

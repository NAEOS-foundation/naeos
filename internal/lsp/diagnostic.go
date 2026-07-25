package lsp

import (
	"fmt"
	"strings"
)

type DiagnosticProvider struct{}

func NewDiagnosticProvider() *DiagnosticProvider {
	return &DiagnosticProvider{}
}

func (dp *DiagnosticProvider) Provide(uri string, text string) *PublishDiagnosticsParams {
	diags := []Diagnostic{}

	diags = append(diags, dp.checkProjectRequired(text)...)
	diags = append(diags, dp.checkYAMLStructure(text)...)
	diags = append(diags, dp.checkServicePorts(text)...)

	return &PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diags,
	}
}

func (dp *DiagnosticProvider) checkProjectRequired(text string) []Diagnostic {
	if !strings.Contains(text, "project:") {
		return []Diagnostic{
			{
				Range:    Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 1}},
				Severity: DiagError,
				Message:  "Required field 'project' is missing",
			},
		}
	}
	return nil
}

func (dp *DiagnosticProvider) checkYAMLStructure(text string) []Diagnostic {
	var diags []Diagnostic
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.Contains(trimmed, "\t") {
			diags = append(diags, Diagnostic{
				Range:    Range{Start: Position{Line: i, Character: 0}, End: Position{Line: i, Character: len(line)}},
				Severity: DiagWarning,
				Message:  "Tabs are not allowed in YAML; use spaces",
			})
		}

		if len(trimmed) > 200 {
			diags = append(diags, Diagnostic{
				Range:    Range{Start: Position{Line: i, Character: 200}, End: Position{Line: i, Character: len(line)}},
				Severity: DiagHint,
				Message:  fmt.Sprintf("Line too long (%d characters, max 200)", len(trimmed)),
			})
		}

		if strings.HasSuffix(trimmed, " ") || strings.HasSuffix(trimmed, "\t") {
			diags = append(diags, Diagnostic{
				Range:    Range{Start: Position{Line: i, Character: len(line) - 1}, End: Position{Line: i, Character: len(line)}},
				Severity: DiagHint,
				Message:  "Trailing whitespace",
			})
		}
	}

	return diags
}

func (dp *DiagnosticProvider) checkServicePorts(text string) []Diagnostic {
	var diags []Diagnostic
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "port:") {
			var portStr string
			if idx := strings.Index(trimmed, ":"); idx >= 0 {
				portStr = strings.TrimSpace(trimmed[idx+1:])
			}
			if portStr != "" {
				var port int
				if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil || port < 1 || port > 65535 {
					diags = append(diags, Diagnostic{
						Range:    Range{Start: Position{Line: i, Character: 0}, End: Position{Line: i, Character: len(line)}},
						Severity: DiagError,
						Message:  fmt.Sprintf("Port must be between 1 and 65535, got %q", portStr),
					})
				}
			}
		}
	}

	return diags
}

package lsp

import (
	"fmt"
	"strconv"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

type DiagnosticProvider struct{}

func NewDiagnosticProvider() *DiagnosticProvider {
	return &DiagnosticProvider{}
}

func (dp *DiagnosticProvider) Provide(uri string, text string) *PublishDiagnosticsParams {
	diags := []Diagnostic{}

	diags = append(diags, dp.checkSpecParse(text)...)
	diags = append(diags, dp.checkYAMLStructure(text)...)

	return &PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diags,
	}
}

func (dp *DiagnosticProvider) checkSpecParse(text string) []Diagnostic {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	p := parser.NewParser("")
	doc, err := p.Parse(text)
	if err != nil {
		var nerr *naeoserr.NaeosError
		if naeoserr.As(err, &nerr) {
			diag := diagFromError(nerr, text)
			if diag != nil {
				return []Diagnostic{*diag}
			}
		}
		return []Diagnostic{{
			Range:    Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 1}},
			Severity: DiagError,
			Message:  fmt.Sprintf("Parse error: %s", err.Error()),
		}}
	}

	diags := []Diagnostic{}

	if doc.Project == "" {
		if hasYAMLKey(text, "project") {
			diags = append(diags, Diagnostic{
				Range:    lineRange(text, "project:"),
				Severity: DiagWarning,
				Message:  "'project' value is empty",
			})
		} else {
			diags = append(diags, Diagnostic{
				Range:    Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 1}},
				Severity: DiagError,
				Message:  "Required field 'project' is missing",
			})
		}
	}

	diags = append(diags, dp.checkSpecDoc(doc, text)...)

	return diags
}

func (dp *DiagnosticProvider) checkSpecDoc(doc *parser.SpecDocument, text string) []Diagnostic {
	var diags []Diagnostic
	lines := strings.Split(text, "\n")

	for _, m := range doc.Modules {
		if m.Name == "" {
			continue
		}
		for _, s := range doc.Services {
			if s.Name == m.Name {
				diags = append(diags, Diagnostic{
					Range:    lineRange(text, m.Name+":"),
					Severity: DiagWarning,
					Message:  fmt.Sprintf("Module %q has the same name as service %q", m.Name, s.Name),
				})
			}
		}
		for _, m2 := range doc.Modules {
			if m2.Name != "" && m.Name != m2.Name && m.Name == m2.Name {
				diags = append(diags, Diagnostic{
					Range:    lineRange(text, "  - name: "+m.Name),
					Severity: DiagError,
					Message:  fmt.Sprintf("Duplicate module name: %q", m.Name),
				})
			}
		}
	}

	for _, s := range doc.Services {
		if s.Port != 0 && (s.Port < 1 || s.Port > 65535) {
			diags = append(diags, Diagnostic{
				Range:    lineRange(text, "port: "+strconv.Itoa(s.Port)),
				Severity: DiagError,
				Message:  fmt.Sprintf("Port %d out of range (1-65535)", s.Port),
			})
		}
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "port:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				port, err := strconv.Atoi(val)
				if err != nil || port < 1 || port > 65535 {
					diags = append(diags, Diagnostic{
						Range:    Range{Start: Position{Line: i, Character: 0}, End: Position{Line: i, Character: len(line)}},
						Severity: DiagError,
						Message:  fmt.Sprintf("Port must be between 1 and 65535, got %q", val),
					})
				}
			}
		}
	}

	return diags
}

func (dp *DiagnosticProvider) checkYAMLStructure(text string) []Diagnostic {
	var diags []Diagnostic
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.Contains(line, "\t") {
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

func diagFromError(nerr *naeoserr.NaeosError, text string) *Diagnostic {
	msg := nerr.Message
	if msg == "" {
		msg = nerr.Error()
	}

	sev := DiagError
	switch nerr.Code {
	case naeoserr.ErrValidation:
		sev = DiagError
	case naeoserr.ErrParse:
		sev = DiagError
	case naeoserr.ErrNotFound:
		sev = DiagWarning
	default:
		sev = DiagError
	}

	line := findErrorLine(text, msg)
	return &Diagnostic{
		Range:    Range{Start: Position{Line: line, Character: 0}, End: Position{Line: line, Character: 1}},
		Severity: sev,
		Message:  fmt.Sprintf("[%s] %s", nerr.Code, msg),
	}
}

func findErrorLine(text, msg string) int {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.Contains(line, msg) {
			return i
		}
	}
	return 0
}

func hasYAMLKey(text, key string) bool {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+":") {
			return true
		}
	}
	return false
}

func lineRange(text, substr string) Range {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.Contains(line, substr) {
			return Range{
				Start: Position{Line: i, Character: 0},
				End:   Position{Line: i, Character: len(line)},
			}
		}
	}
	return Range{
		Start: Position{Line: 0, Character: 0},
		End:   Position{Line: 0, Character: 1},
	}
}

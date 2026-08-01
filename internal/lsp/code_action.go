package lsp

import (
	"strings"
	"unicode"
)

type CodeActionProvider struct{}

func NewCodeActionProvider() *CodeActionProvider {
	return &CodeActionProvider{}
}

func (cap *CodeActionProvider) Provide(uri string, diags []Diagnostic) []CodeAction {
	var actions []CodeAction

	hasProject := false
	hasVersion := false

	for _, d := range diags {
		if strings.Contains(d.Message, "'project'") {
			if d.Severity == DiagError {
				actions = append(actions, CodeAction{
					Title:       "Add missing 'project' field",
					Kind:        CodeActionQuickFix,
					Diagnostics: []Diagnostic{d},
					Edit: &WorkspaceEdit{
						Changes: map[string][]TextEdit{
							uri: {
								{Range: Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 0}}, NewText: "project: my-project\n"},
							},
						},
					},
				})
				hasProject = true
			}
		}
		if strings.Contains(d.Message, "'version'") || (!hasProject && strings.Contains(d.Message, "project")) {
			hasProject = true
		}
	}

	if !hasVersion {
		actions = append(actions, CodeAction{
			Title: "Add missing 'version' field",
			Kind:  CodeActionQuickFix,
			Edit: &WorkspaceEdit{
				Changes: map[string][]TextEdit{
					uri: {
						{Range: Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 0}}, NewText: "version: \"0.1.0\"\n"},
					},
				},
			},
		})
	}

	hasTrailingWS := false
	hasTabs := false
	for _, d := range diags {
		if strings.Contains(d.Message, "Trailing whitespace") {
			hasTrailingWS = true
		}
		if strings.Contains(d.Message, "Tabs") {
			hasTabs = true
		}
	}

	if hasTrailingWS {
		actions = append(actions, CodeAction{
			Title: "Fix trailing whitespace",
			Kind:  CodeActionSourceOrganize,
			Edit: &WorkspaceEdit{
				Changes: map[string][]TextEdit{
					uri: {{Range: Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 0}}, NewText: ""}},
				},
			},
		})
	}

	if hasTabs {
		actions = append(actions, CodeAction{
			Title: "Fix tab to space",
			Kind:  CodeActionSourceOrganize,
			Edit: &WorkspaceEdit{
				Changes: map[string][]TextEdit{
					uri: {{Range: Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 0}}, NewText: ""}},
				},
			},
		})
	}

	return actions
}

func (cap *CodeActionProvider) FixTrailingWhitespace(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRightFunc(line, unicode.IsSpace)
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

func (cap *CodeActionProvider) FixTabs(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.ReplaceAll(line, "\t", "  ")
	}
	return strings.Join(lines, "\n")
}

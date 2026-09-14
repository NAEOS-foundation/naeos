// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	"os"

	"gopkg.in/yaml.v3"
)

type LintConfig struct {
	Rules        []CustomLintRule `yaml:"rules"`
	MinSeverity  Severity         `yaml:"min_severity"`
	DisableRules []string         `yaml:"disable_rules"`
}

type CustomLintRule struct {
	ID       string   `yaml:"id"`
	Severity Severity `yaml:"severity"`
	Patterns []string `yaml:"patterns"`
	Message  string   `yaml:"message"`
}

func LoadConfig(path string) (*LintConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config LintConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.MinSeverity == "" {
		config.MinSeverity = SeverityInfo
	}

	return &config, nil
}

func (l *Linter) ApplyConfig(config *LintConfig) {
	if config == nil {
		return
	}

	disabled := make(map[string]bool)
	for _, id := range config.DisableRules {
		disabled[id] = true
	}

	var filtered []LintRule
	for _, rule := range l.rules {
		if !disabled[rule.ID] {
			filtered = append(filtered, rule)
		}
	}
	l.rules = filtered

	for _, cr := range config.Rules {
		rule := LintRule{
			ID:       cr.ID,
			Severity: cr.Severity,
			Check:    buildPatternLintCheck(cr),
		}
		l.rules = append(l.rules, rule)
	}
}

func buildPatternLintCheck(cr CustomLintRule) func(string) []LintIssue {
	return func(content string) []LintIssue {
		var issues []LintIssue
		lines := splitLines(content)
		for i, line := range lines {
			for _, pattern := range cr.Patterns {
				if contains(line, pattern) {
					issues = append(issues, LintIssue{
						Line:     i + 1,
						Severity: cr.Severity,
						Rule:     cr.ID,
						Message:  cr.Message,
					})
					break
				}
			}
		}
		return issues
	}
}

func FilterIssuesBySeverity(issues []LintIssue, minSeverity Severity) []LintIssue {
	levels := map[Severity]int{
		SeverityInfo:    0,
		SeverityWarning: 1,
		SeverityError:   2,
	}

	minLevel := levels[minSeverity]
	var filtered []LintIssue
	for _, issue := range issues {
		if levels[issue.Severity] >= minLevel {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

func (l *Linter) LintWithFilter(path, content string, minSeverity Severity) *LintResult {
	result := l.Lint(path, content)
	result.Issues = FilterIssuesBySeverity(result.Issues, minSeverity)
	return result
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

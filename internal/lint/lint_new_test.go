// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLintLineLength(t *testing.T) {
	linter := NewLinter()
	longLine := "key: " + string(make([]byte, 250))
	for i := range longLine {
		if i < 5 {
			longLine = "key: "
		} else {
			longLine += "x"
		}
	}
	content := longLine + "\n"
	result := linter.Lint("test.yaml", content)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "yaml-line-length" {
			found = true
		}
	}
	if !found {
		t.Error("expected line length warning")
	}
}

func TestLintNestedDuplicateKeys(t *testing.T) {
	linter := NewLinter()
	content := `project:
  name: test
  name: test2
`
	result := linter.Lint("test.yaml", content)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "yaml-nested-duplicate-keys" {
			found = true
		}
	}
	if !found {
		t.Error("expected nested duplicate key warning")
	}
}

func TestLintBoolStringMix(t *testing.T) {
	linter := NewLinter()
	content := `enabled: "true"
disabled: "false"
`
	result := linter.Lint("test.yaml", content)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "yaml-bool-string-mix" {
			found = true
		}
	}
	if !found {
		t.Error("expected bool string mix warning")
	}
}

func TestFilterIssuesBySeverity(t *testing.T) {
	issues := []LintIssue{
		{Severity: SeverityError, Rule: "a"},
		{Severity: SeverityWarning, Rule: "b"},
		{Severity: SeverityInfo, Rule: "c"},
	}

	filtered := FilterIssuesBySeverity(issues, SeverityWarning)
	if len(filtered) != 2 {
		t.Errorf("expected 2 issues, got %d", len(filtered))
	}
}

func TestLintWithFilter(t *testing.T) {
	linter := NewLinter()
	content := `project:
  name: test
`
	result := linter.LintWithFilter("test.yaml", content, SeverityError)
	for _, issue := range result.Issues {
		if issue.Severity != SeverityError {
			t.Errorf("expected only error issues, got %s", issue.Severity)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	configYAML := `
min_severity: warning
disable_rules:
  - yaml-tabs
rules:
  - id: custom-rule
    severity: error
    patterns:
      - "bad_pattern"
    message: "Custom rule triggered"
`
	dir := t.TempDir()
	configPath := filepath.Join(dir, "lint.yaml")
	os.WriteFile(configPath, []byte(configYAML), 0o644)

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.MinSeverity != SeverityWarning {
		t.Errorf("expected min_severity warning, got %s", config.MinSeverity)
	}

	if len(config.DisableRules) != 1 {
		t.Errorf("expected 1 disabled rule, got %d", len(config.DisableRules))
	}

	if len(config.Rules) != 1 {
		t.Errorf("expected 1 custom rule, got %d", len(config.Rules))
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestApplyConfig(t *testing.T) {
	linter := NewLinter()
	config := &LintConfig{
		DisableRules: []string{"yaml-tabs"},
	}

	linter.ApplyConfig(config)

	content := "\tkey: value"
	result := linter.Lint("test.yaml", content)
	for _, issue := range result.Issues {
		if issue.Rule == "yaml-tabs" {
			t.Error("yaml-tabs should be disabled")
		}
	}
}

func TestApplyConfigWithCustomRules(t *testing.T) {
	linter := NewLinter()
	config := &LintConfig{
		Rules: []CustomLintRule{
			{
				ID:       "custom-check",
				Severity: SeverityError,
				Patterns: []string{"dangerous_function"},
				Message:  "Dangerous function detected",
			},
		},
	}

	linter.ApplyConfig(config)

	content := `result = dangerous_function(input)`
	result := linter.Lint("test.py", content)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "custom-check" {
			found = true
		}
	}
	if !found {
		t.Error("expected custom rule to trigger")
	}
}

func TestApplyConfigNil(t *testing.T) {
	linter := NewLinter()
	linter.ApplyConfig(nil)
	if len(linter.rules) == 0 {
		t.Error("linter should have default rules")
	}
}

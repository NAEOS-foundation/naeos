// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	"testing"
)

func TestLintEmptyDocument(t *testing.T) {
	l := NewLinter()
	result := l.Lint("test.yaml", "")
	if len(result.Issues) == 0 {
		t.Fatal("expected issues for empty document")
	}
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "empty-document" {
			found = true
		}
	}
	if !found {
		t.Error("expected empty-document rule to trigger")
	}
}

func TestLintTabs(t *testing.T) {
	l := NewLinter()
	content := "key:\n\tchild:\n"
	result := l.Lint("test.yaml", content)
	if len(result.Issues) == 0 {
		t.Fatal("expected issues for tabs in YAML")
	}
}

func TestLintTrailingWhitespace(t *testing.T) {
	l := NewLinter()
	content := "key: value   \n"
	result := l.Lint("test.yaml", content)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "yaml-trailing-space" {
			found = true
		}
	}
	if !found {
		t.Error("expected trailing whitespace issue")
	}
}

func TestLintProjectNameFormat(t *testing.T) {
	l := NewLinter()
	content := "project: My Project Name\n"
	result := l.Lint("test.yaml", content)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "project-name-format" {
			found = true
		}
	}
	if !found {
		t.Error("expected project-name-format issue")
	}
}

func TestLintPortRange(t *testing.T) {
	l := NewLinter()
	content := "port: 70000\n"
	result := l.Lint("test.yaml", content)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "port-range" {
			found = true
		}
	}
	if !found {
		t.Error("expected port-range issue")
	}
}

func TestLintValidSpec(t *testing.T) {
	l := NewLinter()
	content := "project: my-project\nport: 8080\n"
	result := l.Lint("test.yaml", content)
	for _, issue := range result.Issues {
		if issue.Severity == "error" {
			t.Errorf("unexpected error: %s", issue.Message)
		}
	}
}

func TestFix(t *testing.T) {
	content := "key: value   \nother: test  \n"
	fixed := Fix(content)
	if fixed == content {
		t.Error("expected content to be modified")
	}
}

func TestLintResultStructure(t *testing.T) {
	l := NewLinter()
	result := l.Lint("test.yaml", "project: test\n")
	if result.Path != "test.yaml" {
		t.Errorf("expected path test.yaml, got %s", result.Path)
	}
}

func TestValidateSpecEmpty(t *testing.T) {
	issues := ValidateSpec("")
	if len(issues) == 0 {
		t.Fatal("expected issues for empty spec")
	}
	if issues[0].Rule != "spec-empty" {
		t.Errorf("expected spec-empty rule, got %s", issues[0].Rule)
	}
}

func TestValidateSpecInvalidYAML(t *testing.T) {
	issues := ValidateSpec("{{invalid yaml}}")
	if len(issues) == 0 {
		t.Fatal("expected issues for invalid YAML")
	}
	if issues[0].Rule != "spec-yaml" {
		t.Errorf("expected spec-yaml rule, got %s", issues[0].Rule)
	}
}

func TestValidateSpecMissingProject(t *testing.T) {
	issues := ValidateSpec("modules:\n  - name: core\n    path: ./core\n")
	found := false
	for _, issue := range issues {
		if issue.Rule == "spec-required-project" {
			found = true
		}
	}
	if !found {
		t.Error("expected spec-required-project issue")
	}
}

func TestValidateSpecDuplicateModules(t *testing.T) {
	spec := `project: test
modules:
  - name: core
    path: ./core
  - name: core
    path: ./core2
`
	issues := ValidateSpec(spec)
	found := false
	for _, issue := range issues {
		if issue.Rule == "spec-module-duplicate" {
			found = true
		}
	}
	if !found {
		t.Error("expected spec-module-duplicate issue")
	}
}

func TestValidateSpecInvalidLanguage(t *testing.T) {
	spec := `project: test
modules:
  - name: core
    path: ./core
generation:
  languages:
    - cobol
`
	issues := ValidateSpec(spec)
	found := false
	for _, issue := range issues {
		if issue.Rule == "spec-language-invalid" {
			found = true
		}
	}
	if !found {
		t.Error("expected spec-language-invalid issue")
	}
}

func TestValidateSpecValid(t *testing.T) {
	spec := `project: my-project
modules:
  - name: core
    path: ./internal/core
services:
  - name: api
    port: 8080
generation:
  languages:
    - go
    - typescript
`
	issues := ValidateSpec(spec)
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			t.Errorf("unexpected error: %s — %s", issue.Rule, issue.Message)
		}
	}
}

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const cleanLintSpec = "project: demo-app\nmodules:\n  - name: api\n    path: ./internal/api\n"

const issueLintSpec = "project: MyApp \nport: 99999\n"

func TestLintNoIssues(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", cleanLintSpec)

	root := NewRootCommand()
	output, err := executeCommand(root, "lint", "--input-file", specPath)
	if err != nil {
		t.Fatalf("lint failed: %v", err)
	}
	if !strings.Contains(output, "no issues found") {
		t.Fatalf("expected 'no issues found', got %q", output)
	}
}

func TestLintWithIssues(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", issueLintSpec)

	root := NewRootCommand()
	output, err := executeCommand(root, "lint", "--input-file", specPath)
	if err != nil {
		t.Fatalf("lint failed: %v", err)
	}
	if !strings.Contains(output, "port-range") {
		t.Fatalf("expected port-range issue, got %q", output)
	}
	if !strings.Contains(output, "3 issue(s) found") {
		t.Fatalf("expected 3 issues, got %q", output)
	}
}

func TestLintJSONOutput(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", issueLintSpec)

	root := NewRootCommand()
	output, err := executeCommand(root, "lint", "--input-file", specPath, "--output", "json")
	if err != nil {
		t.Fatalf("lint json failed: %v", err)
	}
	if !strings.Contains(output, `"issue_count": 3`) {
		t.Fatalf("expected issue_count in json output, got %q", output)
	}
}

func TestLintJSONNoIssues(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", cleanLintSpec)

	root := NewRootCommand()
	output, err := executeCommand(root, "lint", "--input-file", specPath, "--output", "json")
	if err != nil {
		t.Fatalf("lint json failed: %v", err)
	}
	if !strings.Contains(output, `"issue_count": 0`) {
		t.Fatalf("expected zero issue count, got %q", output)
	}
}

func TestLintFixRewritesFile(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: demo-app \nmodules:\n  - name: api\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "lint", "--input-file", specPath, "--fix")
	if err != nil {
		t.Fatalf("lint --fix failed: %v", err)
	}
	if !strings.Contains(output, "Applied fixes to") {
		t.Fatalf("expected fix confirmation, got %q", output)
	}
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read fixed file: %v", err)
	}
	if strings.Contains(string(data), " \n") {
		t.Fatalf("expected trailing whitespace to be removed, got %q", string(data))
	}
}

func TestLintMissingInputFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "lint")
	if err == nil {
		t.Fatal("expected error when --input-file is missing")
	}
}

func TestLintNonexistentFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "lint", "--input-file", "/nonexistent/spec.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

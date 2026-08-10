package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocsAPIStdout(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "docs", "api", "--project", "myapp")
	if err != nil {
		t.Fatalf("docs api failed: %v", err)
	}
	if !strings.Contains(output, "/health") || !strings.Contains(output, "/api/v1") {
		t.Fatalf("expected API endpoints in output, got %q", output)
	}
}

func TestDocsAPIOutputDir(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "docs")

	root := NewRootCommand()
	output, err := executeCommand(root, "docs", "api", "--project", "myapp", "-o", outDir)
	if err != nil {
		t.Fatalf("docs api with output failed: %v", err)
	}
	if !strings.Contains(output, "Generated api.md") {
		t.Fatalf("expected 'Generated api.md', got %q", output)
	}
	if _, err := os.Stat(filepath.Join(outDir, "api.md")); err != nil {
		t.Fatalf("expected api.md: %v", err)
	}
}

func TestDocsArchitectureStdout(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "docs", "architecture", "--project", "myapp")
	if err != nil {
		t.Fatalf("docs architecture failed: %v", err)
	}
	if len(strings.TrimSpace(output)) == 0 {
		t.Fatal("expected architecture diagram output")
	}
}

func TestDocsArchitectureOutputDir(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "docs")

	root := NewRootCommand()
	output, err := executeCommand(root, "docs", "architecture", "--project", "myapp", "-o", outDir)
	if err != nil {
		t.Fatalf("docs architecture with output failed: %v", err)
	}
	if !strings.Contains(output, "Generated architecture.md") {
		t.Fatalf("expected 'Generated architecture.md', got %q", output)
	}
	if _, err := os.Stat(filepath.Join(outDir, "architecture.md")); err != nil {
		t.Fatalf("expected architecture.md: %v", err)
	}
}

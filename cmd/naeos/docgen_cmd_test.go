package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocGenFull(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "docgen", "--input", "project: myapp")
	if err != nil {
		t.Fatalf("docgen failed: %v", err)
	}
	if !strings.Contains(output, "Auto-generated documentation") {
		t.Fatalf("expected full docs output, got %q", output)
	}
}

func TestDocGenAPI(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "docgen", "--input", "project: myapp", "--output", "api")
	if err != nil {
		t.Fatalf("docgen api failed: %v", err)
	}
	if !strings.Contains(output, "API Reference") {
		t.Fatalf("expected API docs output, got %q", output)
	}
}

func TestDocGenModules(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "docgen", "--input", "project: myapp", "--output", "modules")
	if err != nil {
		t.Fatalf("docgen modules failed: %v", err)
	}
	if !strings.Contains(output, "Module Documentation") {
		t.Fatalf("expected module docs output, got %q", output)
	}
}

func TestDocGenOutputFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "docs.md")

	root := NewRootCommand()
	output, err := executeCommand(root, "docgen", "--input", "project: myapp", "--output-file", out)
	if err != nil {
		t.Fatalf("docgen output-file failed: %v", err)
	}
	if len(output) != 0 {
		t.Fatalf("expected no stdout when writing to file, got %q", output)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read docs: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty docs file")
	}
}

func TestDocGenInputFile(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: from-file\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "docgen", "--input-file", spec)
	if err != nil {
		t.Fatalf("docgen input-file failed: %v", err)
	}
	if !strings.Contains(output, "from-file") {
		t.Fatalf("expected project from file, got %q", output)
	}
}

func TestDocGenMissingInput(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "docgen")
	if err == nil {
		t.Fatal("expected error when both --input and --input-file are missing")
	}
}

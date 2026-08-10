package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContextMarkdown(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "context", "--input", "project: myapp")
	if err != nil {
		t.Fatalf("context failed: %v", err)
	}
	if !strings.Contains(output, "myapp") {
		t.Fatalf("expected project name in markdown output, got %q", output)
	}
}

func TestContextOutputJSON(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "context", "--input", "project: myapp", "--output", "json")
	if err != nil {
		t.Fatalf("context json failed: %v", err)
	}
	if !strings.Contains(output, `"project": "myapp"`) {
		t.Fatalf("expected json output, got %q", output)
	}
}

func TestContextOutputYAML(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "context", "--input", "project: myapp", "--output", "yaml")
	if err != nil {
		t.Fatalf("context yaml failed: %v", err)
	}
	if !strings.Contains(output, "project: myapp") {
		t.Fatalf("expected yaml output, got %q", output)
	}
}

func TestContextOutputFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "bundle.md")

	root := NewRootCommand()
	output, err := executeCommand(root, "context", "--input", "project: myapp", "--output-file", out)
	if err != nil {
		t.Fatalf("context with output-file failed: %v", err)
	}
	if len(output) != 0 {
		t.Fatalf("expected no stdout when writing to file, got %q", output)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	if !strings.Contains(string(data), "myapp") {
		t.Fatalf("expected bundle content, got %q", string(data))
	}
}

func TestContextInputFile(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: from-file\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "context", "--input-file", spec)
	if err != nil {
		t.Fatalf("context input-file failed: %v", err)
	}
	if !strings.Contains(output, "from-file") {
		t.Fatalf("expected project from file, got %q", output)
	}
}

func TestContextMissingInput(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "context")
	if err == nil {
		t.Fatal("expected error when both --input and --input-file are missing")
	}
}

func TestContextInvalidSpec(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "context", "--input", "{unclosed")
	if err == nil {
		t.Fatal("expected error for invalid spec")
	}
}

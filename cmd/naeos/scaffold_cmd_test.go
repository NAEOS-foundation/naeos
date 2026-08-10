package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldMissingName(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "scaffold", "--output", filepath.Join(t.TempDir(), "app"))
	if err == nil || !strings.Contains(err.Error(), "missing required --name") {
		t.Fatalf("expected missing name error, got %v", err)
	}
}

func TestScaffoldWithLanguage(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "app")

	root := NewRootCommand()
	output, err := executeCommand(root, "scaffold", "--name", "app", "--output", outDir, "--language", "go")
	if err != nil {
		t.Fatalf("scaffold with language failed: %v", err)
	}
	if !strings.Contains(output, "scaffolded "+outDir) {
		t.Fatalf("expected scaffolded message, got %q", output)
	}
	for _, name := range []string{"README.md", "spec.yaml", "Makefile", ".gitignore", "config.yaml", "Dockerfile", "internal/core/package.go"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("expected %s to exist: %v", name, err)
		}
	}
}

func TestScaffoldUnknownLanguage(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "app")

	root := NewRootCommand()
	output, err := executeCommand(root, "scaffold", "--name", "app", "--output", outDir, "--language", "not-a-language")
	if err != nil {
		t.Fatalf("scaffold with unknown language should not error: %v", err)
	}
	if !strings.Contains(output, "not-a-language") {
		t.Fatalf("expected language name in output, got %q", output)
	}
	if _, err := os.Stat(filepath.Join(outDir, "spec.yaml")); err != nil {
		t.Fatalf("expected base files even for unknown language: %v", err)
	}
}

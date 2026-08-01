package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocsGenOutput(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "cli-docs")

	root := NewRootCommand()
	_, err := executeCommand(root, "docsgen", "--output", outDir)
	if err != nil {
		t.Fatalf("docsgen failed: %v", err)
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("read docs dir: %v", err)
	}
	found := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected markdown files in %s", outDir)
	}
	if _, err := os.Stat(filepath.Join(outDir, "naeos.md")); err != nil {
		t.Fatalf("expected naeos.md: %v", err)
	}
}

func TestDocsGenInvalidOutputDir(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}

	root := NewRootCommand()
	_, err := executeCommand(root, "docsgen", "--output", filepath.Join(blocked, "sub"))
	if err == nil {
		t.Fatal("expected error when output dir cannot be created")
	}
}

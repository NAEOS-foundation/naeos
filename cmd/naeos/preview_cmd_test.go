package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreviewHappyPath(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: "+filepath.Join(dir, "out")+"\n")
	writeTestFile(t, dir, "spec.yaml", "project: preview-demo\nmodules:\n  - name: api\n    path: ./internal/api\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "preview", "--config", configPath, "--input", specPath)
	if err != nil {
		t.Fatalf("preview failed: %v", err)
	}
	if !strings.Contains(output, "Preview for demo") || !strings.Contains(output, "Total:") {
		t.Fatalf("expected preview summary, got %q", output)
	}
	if _, err := os.Stat(filepath.Join(dir, "out")); !os.IsNotExist(err) {
		t.Fatalf("preview should not write artifacts, out dir should not exist: %v", err)
	}
}

func TestPreviewMissingInput(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n")

	root := NewRootCommand()
	_, err := executeCommand(root, "preview", "--config", configPath)
	if err == nil || !strings.Contains(err.Error(), "missing required --input") {
		t.Fatalf("expected missing input error, got %v", err)
	}
}

func TestPreviewBadConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "preview", "--config", filepath.Join(t.TempDir(), "nope.yaml"), "--input", "project: x")
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

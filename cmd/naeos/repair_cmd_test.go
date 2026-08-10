package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepairHappyPath(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: "+outDir+"\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "repair", "--config", configPath)
	if err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	if !strings.Contains(output, "repaired "+outDir) {
		t.Fatalf("expected repaired message, got %q", output)
	}
	if _, err := os.Stat(filepath.Join(outDir, "README.md")); err != nil {
		t.Fatalf("expected README.md to exist: %v", err)
	}
}

func TestRepairExistingReadme(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: "+outDir+"\n")
	writeTestFile(t, dir, "out/README.md", "# keep me\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "repair", "--config", configPath)
	if err != nil {
		t.Fatalf("repair with existing README failed: %v", err)
	}
	if !strings.Contains(output, "repaired") {
		t.Fatalf("expected repaired message, got %q", output)
	}
	data, err := os.ReadFile(filepath.Join(outDir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "keep me") {
		t.Fatalf("expected existing README to be untouched, got %q", string(data))
	}
}

func TestRepairMissingConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "repair", "--config", filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestRepairNoOutputDir(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n")

	root := NewRootCommand()
	_, err := executeCommand(root, "repair", "--config", configPath)
	if err == nil || !strings.Contains(err.Error(), "output_dir is not configured") {
		t.Fatalf("expected output_dir error, got %v", err)
	}
}

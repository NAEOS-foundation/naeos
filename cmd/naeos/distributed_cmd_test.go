package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDistributedHappy(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: dist-test\n  mode: development\n  output_dir: ./out\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "distributed", "--config", configPath, "--workers", "2")
	if err != nil {
		t.Fatalf("distributed failed: %v", err)
	}
	if !strings.Contains(output, "Distributed pipeline: 2 workers, 8 tasks") {
		t.Fatalf("expected worker/task summary, got %q", output)
	}
	if !strings.Contains(output, "Pipeline: dist-test") {
		t.Fatalf("expected pipeline name, got %q", output)
	}
}

func TestDistributedSingleWorker(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: dist-single\n  mode: development\n  output_dir: ./out\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "distributed", "--config", configPath, "-w", "1")
	if err != nil {
		t.Fatalf("distributed single worker failed: %v", err)
	}
	if !strings.Contains(output, "Distributed pipeline: 1 workers, 8 tasks") {
		t.Fatalf("expected single worker summary, got %q", output)
	}
}

func TestDistributedMissingConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "distributed", "--config", "/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

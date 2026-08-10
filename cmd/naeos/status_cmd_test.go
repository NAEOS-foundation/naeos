package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatusText(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "status", "--config", configPath)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(output, "NAEOS Status") || !strings.Contains(output, "Pipeline:      demo") {
		t.Fatalf("expected status output, got %q", output)
	}
	if !strings.Contains(output, "Version:") || !strings.Contains(output, "Cache:") {
		t.Fatalf("expected system sections, got %q", output)
	}
}

func TestStatusJSON(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "status", "--config", configPath, "--output-format", "json")
	if err != nil {
		t.Fatalf("status json failed: %v", err)
	}
	if !strings.Contains(output, `"pipeline": "demo"`) || !strings.Contains(output, `"mode": "development"`) {
		t.Fatalf("expected json status output, got %q", output)
	}
}

func TestStatusYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "status", "--config", configPath, "--output-format", "yaml")
	if err != nil {
		t.Fatalf("status yaml failed: %v", err)
	}
	if !strings.Contains(output, "pipeline: demo") || !strings.Contains(output, "mode: development") {
		t.Fatalf("expected yaml status output, got %q", output)
	}
}

func TestStatusMissingConfig(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	root := NewRootCommand()
	_, err := executeCommand(root, "status")
	if err == nil {
		t.Fatal("expected error when no config found")
	}
}

func TestStatusBadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: [unclosed")

	root := NewRootCommand()
	_, err := executeCommand(root, "status", "--config", configPath)
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestJoinStrings(t *testing.T) {
	if got := joinStrings(nil); got != "(none)" {
		t.Fatalf("expected (none), got %q", got)
	}
	if got := joinStrings([]string{"go", "ts"}); got != "go, ts" {
		t.Fatalf("expected 'go, ts', got %q", got)
	}
}

func TestStatusCacheLine(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	writeTestFile(t, dir, "config.yaml", "pipeline:\n  name: demo\n  mode: development\n  output_dir: "+filepath.Join(dir, "cache")+"\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "status", "--config", configPath)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(output, "Cache:") {
		t.Fatalf("expected cache line in output, got %q", output)
	}
}

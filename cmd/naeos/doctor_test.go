package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckSpecValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(path, []byte("project: valid-project\nmodules:\n  - name: api\n    path: ./internal/api\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	result := checkSpec(path)
	if result.Status != "pass" {
		t.Fatalf("expected pass status, got %q: %s", result.Status, result.Detail)
	}
}

func TestCheckSpecNonexistentFile(t *testing.T) {
	result := checkSpec("/nonexistent/spec.yaml")
	if result.Status != "fail" {
		t.Fatalf("expected fail status for nonexistent file, got %q", result.Status)
	}
}

func TestCheckSpecMinimalValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(path, []byte("project: minimal\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	result := checkSpec(path)
	if result.Status != "pass" && result.Status != "warn" {
		t.Fatalf("expected pass or warn status, got %q: %s", result.Status, result.Detail)
	}
}

func TestDoctorQuickFlag(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: quick-test\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := newRootCommand()
	output, err := executeCommand(root, "doctor", "--config", configPath, "--quick")
	if err != nil {
		t.Fatalf("execute doctor --quick failed: %v", err)
	}

	if !strings.Contains(output, "NAEOS Doctor") {
		t.Fatalf("expected doctor header, got %q", output)
	}
}

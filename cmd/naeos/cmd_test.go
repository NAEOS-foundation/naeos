package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "version")
	if err != nil {
		t.Fatalf("execute version failed: %v", err)
	}
	if !strings.Contains(output, "naeos ") {
		t.Fatalf("expected version output, got %q", output)
	}
}

func TestStatusCmd(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "status")
	if err != nil {
		t.Fatalf("execute status failed: %v", err)
	}
}

func TestHealthCmd(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "health")
	if err != nil {
		t.Fatalf("execute health failed: %v", err)
	}
	if !strings.Contains(output, "Status:") {
		t.Fatalf("expected health report, got %q", output)
	}
}

func writeConfig(path string) {
	os.WriteFile(path, []byte("pipeline:\n  name: demo\n  mode: development\n  verbose: true\n  output_dir: ./out\n"), 0o644)
}

func TestMigratePlanCmd(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "migrate", "plan")
	if err != nil {
		t.Fatalf("execute migrate plan failed: %v", err)
	}
	if !strings.Contains(output, "Migration plan") && !strings.Contains(output, "No migrations needed") {
		t.Fatalf("expected migration plan output, got %q", output)
	}
}

func TestMigrateVersionsCmd(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "migrate", "versions")
	if err != nil {
		t.Fatalf("execute migrate versions failed: %v", err)
	}
	if !strings.Contains(output, "Supported versions") {
		t.Fatalf("expected versions output, got %q", output)
	}
}

func TestMigrateRunCmd(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "spec.yaml")
	os.WriteFile(specFile, []byte("project: test\n"), 0o644)

	root := NewRootCommand()
	output, err := executeCommand(root, "migrate", "run", specFile)
	if err != nil {
		t.Fatalf("execute migrate run failed: %v", err)
	}
	if !strings.Contains(output, "Migrated") {
		t.Fatalf("expected migration confirmation, got %q", output)
	}
}

func TestMigrateRunDryRunCmd(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "spec.yaml")
	os.WriteFile(specFile, []byte("project: test\n"), 0o644)

	root := NewRootCommand()
	output, err := executeCommand(root, "migrate", "run", specFile, "--dry-run")
	if err != nil {
		t.Fatalf("execute migrate run --dry-run failed: %v", err)
	}
	if !strings.Contains(output, "DRY RUN") {
		t.Fatalf("expected dry run output, got %q", output)
	}
}

func TestMigrateRunOutputCmd(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "spec.yaml")
	outFile := filepath.Join(dir, "out.yaml")
	os.WriteFile(specFile, []byte("project: test\n"), 0o644)

	root := NewRootCommand()
	output, err := executeCommand(root, "migrate", "run", specFile, "--output", outFile)
	if err != nil {
		t.Fatalf("execute migrate run --output failed: %v", err)
	}
	if !strings.Contains(output, "Migrated") {
		t.Fatalf("expected migration confirmation, got %q", output)
	}
}

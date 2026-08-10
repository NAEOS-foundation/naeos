package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateVersions(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "migrate", "versions")
	if err != nil {
		t.Fatalf("migrate versions failed: %v", err)
	}
	if !strings.Contains(output, "Supported versions:") || !strings.Contains(output, "Current: 0.1.0") {
		t.Fatalf("expected version list, got %q", output)
	}
}

func TestMigratePlanDefault(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "migrate", "plan")
	if err != nil {
		t.Fatalf("migrate plan failed: %v", err)
	}
	if !strings.Contains(output, "Migration plan: 0.1.0 -> 0.3.0") {
		t.Fatalf("expected default migration plan, got %q", output)
	}
	if !strings.Contains(output, "0.1.0 -> 0.2.0") || !strings.Contains(output, "0.2.0 -> 0.3.0") {
		t.Fatalf("expected both migration steps, got %q", output)
	}
}

func TestMigratePlanPartial(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "migrate", "plan", "--from", "0.1.0", "--to", "0.2.0")
	if err != nil {
		t.Fatalf("migrate plan failed: %v", err)
	}
	if !strings.Contains(output, "0.1.0 -> 0.2.0") {
		t.Fatalf("expected single step plan, got %q", output)
	}
	if strings.Contains(output, "0.2.0 -> 0.3.0") {
		t.Fatalf("expected only the 0.2.0 step, got %q", output)
	}
}

func TestMigratePlanInvalidVersion(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "migrate", "plan", "--from", "not-a-version")
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestMigratePlanSameVersion(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "migrate", "plan", "--from", "0.3.0", "--to", "0.3.0")
	if err == nil {
		t.Fatal("expected error when target is not newer than source")
	}
}

func TestMigrateRunDryRun(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: demo\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "migrate", "run", specPath, "--dry-run")
	if err != nil {
		t.Fatalf("migrate run --dry-run failed: %v", err)
	}
	if !strings.Contains(output, "DRY RUN OUTPUT") {
		t.Fatalf("expected dry run output, got %q", output)
	}
	if !strings.Contains(output, "generation:") {
		t.Fatalf("expected migrated content in dry run, got %q", output)
	}
}

func TestMigrateRunWritesOutputFile(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: demo\n")
	outputPath := filepath.Join(dir, "migrated.yaml")

	root := NewRootCommand()
	output, err := executeCommand(root, "migrate", "run", specPath, "--output", outputPath)
	if err != nil {
		t.Fatalf("migrate run failed: %v", err)
	}
	if !strings.Contains(output, "Migrated") {
		t.Fatalf("expected migration summary, got %q", output)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read migrated file: %v", err)
	}
	if !strings.Contains(string(data), "testing:") {
		t.Fatalf("expected migrated content, got %q", string(data))
	}
}

func TestMigrateRunInPlace(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	writeTestFile(t, dir, "spec.yaml", "project: demo\n")

	root := NewRootCommand()
	if _, err := executeCommand(root, "migrate", "run", specPath); err != nil {
		t.Fatalf("migrate run in place failed: %v", err)
	}
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read migrated file: %v", err)
	}
	if !strings.Contains(string(data), "generation:") || !strings.Contains(string(data), "testing:") {
		t.Fatalf("expected in-place migration, got %q", string(data))
	}
}

func TestMigrateRunMissingFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "migrate", "run", "/nonexistent/spec.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent spec file")
	}
}

func TestMigrateRunNoArgs(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "migrate", "run")
	if err == nil {
		t.Fatal("expected error when spec file arg is missing")
	}
}

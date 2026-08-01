package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHealthText(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "health")
	if err != nil {
		t.Fatalf("health failed: %v", err)
	}
	if !strings.Contains(output, "NAEOS Health Report") {
		t.Fatalf("expected health report header, got %q", output)
	}
	if !strings.Contains(output, "Status: degraded") {
		t.Fatalf("expected degraded status with isolated HOME, got %q", output)
	}
}

func TestHealthJSON(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "health", "--output-format", "json")
	if err != nil {
		t.Fatalf("health json failed: %v", err)
	}
	if !strings.Contains(output, `"status": "degraded"`) {
		t.Fatalf("expected json status, got %q", output)
	}
	if !strings.Contains(output, `"checks"`) {
		t.Fatalf("expected checks in json output, got %q", output)
	}
}

func TestHealthYAML(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "health", "--output-format", "yaml")
	if err != nil {
		t.Fatalf("health yaml failed: %v", err)
	}
	if !strings.Contains(output, "status: degraded") {
		t.Fatalf("expected yaml status, got %q", output)
	}
}

func TestHealthUnknownFormatFallsBackToText(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "health", "--output-format", "xml")
	if err != nil {
		t.Fatalf("health with unknown format failed: %v", err)
	}
	if !strings.Contains(output, "NAEOS Health Report") {
		t.Fatalf("expected text fallback, got %q", output)
	}
}

func TestRunHealthChecks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	report := runHealthChecks()
	if report.Status != "degraded" {
		t.Fatalf("expected degraded status, got %q", report.Status)
	}
	if len(report.Checks) != 4 {
		t.Fatalf("expected 4 checks, got %d", len(report.Checks))
	}
	if report.Version == "" || report.Go == "" || report.Platform == "" {
		t.Fatal("expected version, go and platform fields to be set")
	}
}

func TestCheckConfigDirHealthy(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".config", "naeos"), 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	check := checkConfigDir()
	if check.Status != "healthy" {
		t.Fatalf("expected healthy config dir, got %q", check.Status)
	}
}

func TestCheckConfigDirMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	check := checkConfigDir()
	if check.Status != "degraded" {
		t.Fatalf("expected degraded config dir, got %q", check.Status)
	}
}

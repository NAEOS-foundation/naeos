// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"os"
	"path/filepath"
	"testing"
)

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

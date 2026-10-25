// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
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

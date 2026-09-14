// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileNotFound(t *testing.T) {
	_, err := LoadFile("/nonexistent/path/config.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadFileEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadFile(path)
	if err == nil {
		t.Error("expected error for empty config")
	}
}

func TestLoadFileInvalidContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadFile(path)
	if err == nil {
		t.Error("expected error for invalid content")
	}
}

func TestParseUnsupportedFormat(t *testing.T) {
	var f File
	err := parse([]byte("not json nor yaml {"), &f)
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestLoadFilePipelinePolicies(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `pipeline:
  name: demo
  policies:
    - rule_id: policy-1
      condition: exists:project
      enabled: true
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}
	if len(cfg.Pipeline.Policies) != 1 {
		t.Fatalf("expected 1 policy loaded, got %d", len(cfg.Pipeline.Policies))
	}
	if cfg.Pipeline.Policies[0].RuleID != "policy-1" {
		t.Fatalf("expected policy rule id policy-1, got %q", cfg.Pipeline.Policies[0].RuleID)
	}
}

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/cloud"
)

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create dir for %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestLoadCloudConfigFromSpec(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "cloud.yaml")
	writeTestFile(t, dir, "cloud.yaml", `cloud:
  provider: aws
  region: us-east-1
  project: demo
  environment: dev
  resources:
    - name: app
      kind: compute
      spec:
        size: small
    - name: db
      type: database
`)

	cfg, err := loadCloudConfigFromSpec(specPath)
	if err != nil {
		t.Fatalf("loadCloudConfigFromSpec failed: %v", err)
	}
	if cfg.Provider != cloud.AWS {
		t.Errorf("expected aws provider, got %v", cfg.Provider)
	}
	if cfg.Region != "us-east-1" {
		t.Errorf("expected us-east-1, got %q", cfg.Region)
	}
	if len(cfg.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(cfg.Resources))
	}
	if cfg.Resources[0].Type != "compute" {
		t.Errorf("expected kind fallback for type, got %q", cfg.Resources[0].Type)
	}
	if cfg.Resources[1].Type != "database" {
		t.Errorf("expected explicit type, got %q", cfg.Resources[1].Type)
	}
	if cfg.Resources[0].Spec["size"] != "small" {
		t.Errorf("expected spec size small, got %v", cfg.Resources[0].Spec["size"])
	}
}

func TestLoadCloudConfigFromSpecMissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := loadCloudConfigFromSpec(filepath.Join(dir, "missing.yaml"))
	if err == nil {
		t.Fatal("expected error for missing spec file")
	}
}

func TestLoadCloudConfigFromSpecInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	writeTestFile(t, dir, "bad.yaml", "{{{{invalid")
	_, err := loadCloudConfigFromSpec(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadCloudConfigFromSpecNoProvider(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noprovider.yaml")
	writeTestFile(t, dir, "noprovider.yaml", "cloud:\n  region: us-east-1\n")
	_, err := loadCloudConfigFromSpec(path)
	if err == nil {
		t.Fatal("expected error when provider missing")
	}
}

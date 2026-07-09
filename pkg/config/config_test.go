package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"pipeline":{"name":"demo"}}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}
	if cfg.Pipeline.Name != "demo" {
		t.Fatalf("expected pipeline name demo, got %q", cfg.Pipeline.Name)
	}
}

func TestLoadFileYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("pipeline:\n  name: demo\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}
	if cfg.Pipeline.Name != "demo" {
		t.Fatalf("expected pipeline name demo, got %q", cfg.Pipeline.Name)
	}
}

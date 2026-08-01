package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConnectionsFile(t *testing.T, home string, conns []map[string]any) {
	t.Helper()
	data, err := json.Marshal(conns)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".naeos", "db", "connections.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationStatusNoConnections(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "migration", "status")
	if err != nil {
		t.Fatalf("migration status failed: %v", err)
	}
	if !strings.Contains(output, "No database connections configured.") {
		t.Fatalf("expected empty-state message, got %q", output)
	}
}

func TestMigrationStatusUnsupportedDriverText(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConnectionsFile(t, home, []map[string]any{
		{"name": "legacy", "driver": "not-a-driver", "config": map[string]any{"Database": "app"}},
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "migration", "status")
	if err != nil {
		t.Fatalf("migration status failed: %v", err)
	}
	if !strings.Contains(output, "legacy") || !strings.Contains(output, "unsupported driver") {
		t.Fatalf("expected row with unsupported driver, got %q", output)
	}
	if !strings.Contains(output, "CONNECTION") || !strings.Contains(output, "DRIVER") {
		t.Fatalf("expected table header, got %q", output)
	}
}

func TestMigrationStatusJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConnectionsFile(t, home, []map[string]any{
		{"name": "legacy", "driver": "not-a-driver", "config": map[string]any{"Database": "app"}},
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "migration", "status", "--output", "json")
	if err != nil {
		t.Fatalf("migration status json failed: %v", err)
	}
	if !strings.Contains(output, `"status": "unsupported driver"`) {
		t.Fatalf("expected json status field, got %q", output)
	}
}

func TestMigrationStatusYAML(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConnectionsFile(t, home, []map[string]any{
		{"name": "legacy", "driver": "not-a-driver", "config": map[string]any{"Database": "app"}},
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "migration", "status", "--output", "yaml")
	if err != nil {
		t.Fatalf("migration status yaml failed: %v", err)
	}
	if !strings.Contains(output, "status: unsupported driver") {
		t.Fatalf("expected yaml status field, got %q", output)
	}
}

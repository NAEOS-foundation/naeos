package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestHistoryNonexistentStoreDir(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "history", "--store-dir", filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("expected error for nonexistent store dir")
	}
}

func TestHistoryCorruptEventFile(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "run1.json", "not valid json")

	root := NewRootCommand()
	output, err := executeCommand(root, "history", "--store-dir", dir)
	if err != nil {
		t.Fatalf("history failed: %v", err)
	}
	if !strings.Contains(output, "run1") {
		t.Fatalf("expected corrupt run to be listed, got %q", output)
	}
	if !strings.Contains(output, "error:") {
		t.Fatalf("expected error detail for corrupt run, got %q", output)
	}
}

func TestHistoryDurationComputed(t *testing.T) {
	dir := t.TempDir()
	events := `[
		{"id":"e1","stream_id":"run1","type":"pipeline.started","data":{"name":"dur-run"},"version":1,"timestamp":"2025-01-01T00:00:00Z"},
		{"id":"e2","stream_id":"run1","type":"pipeline.completed","data":{"artifacts":1},"version":2,"timestamp":"2025-01-01T00:00:05Z"}
	]`
	writeTestFile(t, dir, "run1.json", events)

	root := NewRootCommand()
	output, err := executeCommand(root, "history", "--store-dir", dir)
	if err != nil {
		t.Fatalf("history failed: %v", err)
	}
	if !strings.Contains(output, "dur-run") {
		t.Fatalf("expected run name, got %q", output)
	}
	if !strings.Contains(output, "5s") {
		t.Fatalf("expected computed duration, got %q", output)
	}
}

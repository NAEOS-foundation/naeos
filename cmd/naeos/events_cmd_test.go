package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testEventsJSON = `[
	{"id":"e1","stream_id":"run1","type":"pipeline.started","data":{"name":"demo"},"version":1,"timestamp":"2025-01-01T00:00:00Z"},
	{"id":"e2","stream_id":"run1","type":"pipeline.completed","data":{"artifacts":3},"version":2,"timestamp":"2025-01-01T00:00:05Z"}
]`

func TestEventsReplayToStdout(t *testing.T) {
	dir := t.TempDir()
	eventsFile := filepath.Join(dir, "events.json")
	writeTestFile(t, dir, "events.json", testEventsJSON)

	root := NewRootCommand()
	output, err := executeCommand(root, "events", "replay", "--input", eventsFile)
	if err != nil {
		t.Fatalf("events replay failed: %v", err)
	}
	if !strings.Contains(output, `"name": "demo"`) {
		t.Fatalf("expected snapshot with project name, got %q", output)
	}
	if !strings.Contains(output, "total_events") {
		t.Fatalf("expected total_events in snapshot, got %q", output)
	}
}

func TestEventsReplayToFile(t *testing.T) {
	dir := t.TempDir()
	eventsFile := filepath.Join(dir, "events.json")
	outFile := filepath.Join(dir, "snapshot.json")
	writeTestFile(t, dir, "events.json", testEventsJSON)

	root := NewRootCommand()
	output, err := executeCommand(root, "events", "replay", "--input", eventsFile, "--output", outFile)
	if err != nil {
		t.Fatalf("events replay to file failed: %v", err)
	}
	if !strings.Contains(output, "Replayed 2 events") {
		t.Fatalf("expected replay summary, got %q", output)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read snapshot file: %v", err)
	}
	if !strings.Contains(string(data), "demo") {
		t.Fatalf("expected snapshot content, got %q", string(data))
	}
}

func TestEventsReplayEmptyEvents(t *testing.T) {
	dir := t.TempDir()
	eventsFile := filepath.Join(dir, "events.json")
	writeTestFile(t, dir, "events.json", "[]")

	root := NewRootCommand()
	output, err := executeCommand(root, "events", "replay", "--input", eventsFile)
	if err != nil {
		t.Fatalf("events replay empty failed: %v", err)
	}
	if !strings.Contains(output, "No events to replay.") {
		t.Fatalf("expected 'No events to replay.', got %q", output)
	}
}

func TestEventsReplayMissingInputFlag(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "events", "replay")
	if err == nil {
		t.Fatal("expected error when --input is missing")
	}
}

func TestEventsReplayNonexistentFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "events", "replay", "--input", "/nonexistent/events.json")
	if err == nil {
		t.Fatal("expected error for nonexistent events file")
	}
}

func TestEventsReplayInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	eventsFile := filepath.Join(dir, "events.json")
	writeTestFile(t, dir, "events.json", "not json")

	root := NewRootCommand()
	_, err := executeCommand(root, "events", "replay", "--input", eventsFile)
	if err == nil {
		t.Fatal("expected error for invalid events JSON")
	}
}

func TestEventsList(t *testing.T) {
	dir := t.TempDir()
	eventsFile := filepath.Join(dir, "events.json")
	writeTestFile(t, dir, "events.json", testEventsJSON)

	root := NewRootCommand()
	output, err := executeCommand(root, "events", "list", "--input", eventsFile)
	if err != nil {
		t.Fatalf("events list failed: %v", err)
	}
	if !strings.Contains(output, "[1] pipeline.started") {
		t.Fatalf("expected formatted event line, got %q", output)
	}
	if !strings.Contains(output, "[2] pipeline.completed") {
		t.Fatalf("expected second formatted event line, got %q", output)
	}
}

func TestEventsListNonexistentFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "events", "list", "--input", "/nonexistent/events.json")
	if err == nil {
		t.Fatal("expected error for nonexistent events file")
	}
}

func TestEventsListInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	eventsFile := filepath.Join(dir, "events.json")
	writeTestFile(t, dir, "events.json", "{broken")

	root := NewRootCommand()
	_, err := executeCommand(root, "events", "list", "--input", eventsFile)
	if err == nil {
		t.Fatal("expected error for invalid events JSON")
	}
}

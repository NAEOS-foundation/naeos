package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSearchLifecycle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "search", "index", "--name", "idx1", "--id", "doc1", "--title", "Hello World", "--content", "this is a search test", "--tag", "go", "--tag", "cli")
	if err != nil {
		t.Fatalf("search index failed: %v", err)
	}
	if !strings.Contains(output, "Indexed document 'doc1' in 'idx1'") {
		t.Fatalf("expected index message, got %q", output)
	}

	output, err = executeCommand(root, "search", "count", "--name", "idx1")
	if err != nil {
		t.Fatalf("search count failed: %v", err)
	}
	if !strings.Contains(output, "Documents in 'idx1': 1") {
		t.Fatalf("expected count 1, got %q", output)
	}

	output, err = executeCommand(root, "search", "query", "--name", "idx1", "--term", "search")
	if err != nil {
		t.Fatalf("search query failed: %v", err)
	}
	if !strings.Contains(output, "Found 1 results") || !strings.Contains(output, "Hello World") {
		t.Fatalf("expected search hit, got %q", output)
	}

	output, err = executeCommand(root, "search", "query", "--name", "idx1", "--term", "search", "--output", "json")
	if err != nil {
		t.Fatalf("search query json failed: %v", err)
	}
	if !strings.Contains(output, `"total": 1`) || !strings.Contains(output, `"id": "doc1"`) {
		t.Fatalf("expected json search output, got %q", output)
	}

	output, err = executeCommand(root, "search", "delete", "--name", "idx1", "--id", "doc1")
	if err != nil {
		t.Fatalf("search delete failed: %v", err)
	}
	if !strings.Contains(output, "Deleted document 'doc1' from 'idx1'") {
		t.Fatalf("expected delete message, got %q", output)
	}

	output, err = executeCommand(root, "search", "count", "--name", "idx1")
	if err != nil {
		t.Fatalf("search count after delete failed: %v", err)
	}
	if !strings.Contains(output, "Documents in 'idx1': 0") {
		t.Fatalf("expected count 0 after delete, got %q", output)
	}
}

func TestSearchQueryNoMatches(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "search", "index", "--name", "idx1", "--id", "doc1", "--title", "Hello", "--content", "world")
	if err != nil {
		t.Fatalf("search index failed: %v", err)
	}

	output, err = executeCommand(root, "search", "query", "--name", "idx1", "--term", "missingterm")
	if err != nil {
		t.Fatalf("search query failed: %v", err)
	}
	if !strings.Contains(output, "Found 0 results") {
		t.Fatalf("expected zero results, got %q", output)
	}
}

func TestSearchDeleteNotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "search", "delete", "--name", "idx1", "--id", "ghost")
	if err == nil {
		t.Fatal("expected error deleting unknown document")
	}
}

func TestSearchIndexMissingID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "search", "index", "--name", "idx1")
	if err == nil {
		t.Fatal("expected error for missing required --id")
	}
}

func TestSearchQueryMissingTerm(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "search", "query", "--name", "idx1")
	if err == nil {
		t.Fatal("expected error for missing required --term")
	}
}

func TestSearchListNone(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "search", "list")
	if err != nil {
		t.Fatalf("search list failed: %v", err)
	}
	if !strings.Contains(output, "Search indexes: none") {
		t.Fatalf("expected none message, got %q", output)
	}
}

func TestSearchListWithIndexes(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := NewRootCommand()
	if _, err := executeCommand(root, "search", "index", "--name", "idx1", "--id", "doc1", "--title", "Hello", "--content", "world"); err != nil {
		t.Fatalf("search index failed: %v", err)
	}

	output, err := executeCommand(root, "search", "list")
	if err != nil {
		t.Fatalf("search list failed: %v", err)
	}
	if !strings.Contains(output, "Search indexes: idx1") {
		t.Fatalf("expected index name in list, got %q", output)
	}
	if _, err := os.Stat(filepath.Join(home, ".naeos", "search", "idx1")); err != nil {
		t.Fatalf("expected index dir to exist: %v", err)
	}
}

func TestSearchBareHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "search")
	if err != nil {
		t.Fatalf("search bare failed: %v", err)
	}
	if !strings.Contains(output, "Manage search indexes") {
		t.Fatalf("expected help output, got %q", output)
	}
}

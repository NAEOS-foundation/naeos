package main

import (
	"os"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/rollback"
)

func TestRollbackListEmpty(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	root := NewRootCommand()
	output, err := executeCommand(root, "rollback", "list")
	if err != nil {
		t.Fatalf("rollback list failed: %v", err)
	}
	if !strings.Contains(output, "No snapshots found") {
		t.Fatalf("expected empty-state message, got %q", output)
	}
}

func TestRollbackListWithSnapshot(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	store := rollback.NewStore(".")
	snap, err := store.Create("out", []rollback.SnapshotArtifact{
		{Path: "README.md", Content: []byte("hello world")},
	})
	if err != nil {
		t.Fatalf("seed snapshot: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "rollback", "list")
	if err != nil {
		t.Fatalf("rollback list failed: %v", err)
	}
	if !strings.Contains(output, snap.ID) {
		t.Fatalf("expected snapshot id in list, got %q", output)
	}
}

func TestRollbackRestoreNotFound(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	root := NewRootCommand()
	_, err := executeCommand(root, "rollback", "restore", "snap-does-not-exist")
	if err == nil {
		t.Fatal("expected error restoring unknown snapshot")
	}
}

func TestRollbackRestoreBadID(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	root := NewRootCommand()
	_, err := executeCommand(root, "rollback", "restore", "../escape")
	if err == nil {
		t.Fatal("expected error for invalid snapshot id")
	}
}

func TestRollbackRestoreIntoSubdir(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	store := rollback.NewStore(".")
	snap, err := store.Create("out", []rollback.SnapshotArtifact{
		{Path: "src/main.go", Content: []byte("package main")},
	})
	if err != nil {
		t.Fatalf("seed snapshot: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "rollback", "restore", snap.ID, "--output-dir", "target")
	if err != nil {
		t.Fatalf("rollback restore failed: %v", err)
	}
	if !strings.Contains(output, "Restored snapshot") {
		t.Fatalf("expected restore confirmation, got %q", output)
	}

	restored, err := os.ReadFile("target/src/main.go")
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if string(restored) != "package main" {
		t.Fatalf("unexpected restored content: %q", string(restored))
	}
}

func TestRollbackRestoreIntoCurrentDir(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	store := rollback.NewStore(".")
	snap, err := store.Create("out", []rollback.SnapshotArtifact{
		{Path: "README.md", Content: []byte("restored content")},
	})
	if err != nil {
		t.Fatalf("seed snapshot: %v", err)
	}

	root := NewRootCommand()
	if _, err := executeCommand(root, "rollback", "restore", snap.ID); err != nil {
		t.Fatalf("rollback restore to current dir failed: %v", err)
	}

	restored, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read restored file in current dir: %v", err)
	}
	if string(restored) != "restored content" {
		t.Fatalf("unexpected restored content: %q", string(restored))
	}
}

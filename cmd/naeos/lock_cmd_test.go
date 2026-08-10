package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(orig)
	})
}

func TestLockGenerateWritesLockFile(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)
	writeTestFile(t, dir, "app.txt", "hello")

	root := NewRootCommand()
	output, err := executeCommand(root, "lock", "generate", "app.txt")
	if err != nil {
		t.Fatalf("lock generate failed: %v", err)
	}
	if !strings.Contains(output, "Generated naeos.lock with 1 artifacts") {
		t.Fatalf("expected generate summary, got %q", output)
	}
	if _, err := os.Stat(filepath.Join(dir, "naeos.lock")); err != nil {
		t.Fatalf("expected lock file to be written: %v", err)
	}
}

func TestLockVerifyNoChanges(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)
	writeTestFile(t, dir, "app.txt", "hello")

	root := NewRootCommand()
	if _, err := executeCommand(root, "lock", "generate", "app.txt"); err != nil {
		t.Fatalf("lock generate failed: %v", err)
	}

	lockFilePath := filepath.Join(dir, "naeos.lock")
	output, err := executeCommand(root, "lock", "verify", "app.txt", "--lock-file", lockFilePath)
	if err != nil {
		t.Fatalf("lock verify failed: %v", err)
	}
	if !strings.Contains(output, "no changes detected") {
		t.Fatalf("expected 'no changes detected', got %q", output)
	}
}

func TestLockVerifyMissingArtifact(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)
	writeTestFile(t, dir, "app.txt", "hello")
	writeTestFile(t, dir, "other.txt", "world")

	root := NewRootCommand()
	if _, err := executeCommand(root, "lock", "generate", "app.txt", "other.txt"); err != nil {
		t.Fatalf("lock generate failed: %v", err)
	}

	lockFilePath := filepath.Join(dir, "naeos.lock")
	output, err := executeCommand(root, "lock", "verify", "app.txt", "--lock-file", lockFilePath)
	if err != nil {
		t.Fatalf("lock verify should succeed with changes: %v", err)
	}
	if !strings.Contains(output, "Changes detected") {
		t.Fatalf("expected 'Changes detected', got %q", output)
	}
}

func TestLockGenerateNonexistentFile(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)

	root := NewRootCommand()
	_, err := executeCommand(root, "lock", "generate", "/nonexistent/artifact.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent artifact")
	}
}

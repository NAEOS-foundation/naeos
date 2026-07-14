package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCpDirCopiesFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir() + "/target"

	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatalf("create sub dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "root.txt"), []byte("root content"), 0o644); err != nil {
		t.Fatalf("write root file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "nested.txt"), []byte("nested content"), 0o644); err != nil {
		t.Fatalf("write nested file: %v", err)
	}

	if err := cpDir(src, dst); err != nil {
		t.Fatalf("cpDir returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "root.txt"))
	if err != nil {
		t.Fatalf("read copied root file: %v", err)
	}
	if string(data) != "root content" {
		t.Fatalf("expected 'root content', got %q", string(data))
	}

	data, err = os.ReadFile(filepath.Join(dst, "sub", "nested.txt"))
	if err != nil {
		t.Fatalf("read copied nested file: %v", err)
	}
	if string(data) != "nested content" {
		t.Fatalf("expected 'nested content', got %q", string(data))
	}
}

func TestCpDirEmptySource(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir() + "/target"

	if err := cpDir(src, dst); err != nil {
		t.Fatalf("cpDir on empty source returned error: %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("expected destination to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected destination to be a directory")
	}
}

func TestCpDirPreservesFileMode(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir() + "/target"

	path := filepath.Join(src, "executable.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho hi"), 0o755); err != nil {
		t.Fatalf("write executable file: %v", err)
	}

	if err := cpDir(src, dst); err != nil {
		t.Fatalf("cpDir returned error: %v", err)
	}

	info, err := os.Stat(filepath.Join(dst, "executable.sh"))
	if err != nil {
		t.Fatalf("read copied file: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("expected mode 0755, got %o", info.Mode().Perm())
	}
}

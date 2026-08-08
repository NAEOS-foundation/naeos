package artifacts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetByHash(t *testing.T) {
	s := NewStore(t.TempDir())
	a, err := s.Add("main.go", []byte("package main"), KindCode)
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	got, ok := s.GetByHash(a.ContentHash)
	if !ok {
		t.Fatal("expected artifact by hash")
	}
	if got.Path != "main.go" {
		t.Errorf("expected main.go, got %s", got.Path)
	}

	if _, ok := s.GetByHash("deadbeef"); ok {
		t.Error("expected no artifact for unknown hash")
	}
}

func TestSetNEIRVersion(t *testing.T) {
	s := NewStore(t.TempDir())
	_, err := s.Add("a.go", []byte("a"), KindCode)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	_, err = s.Add("b.md", []byte("b"), KindDocs)
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	s.SetNEIRVersion("v2.0.0")
	for _, a := range s.manifest.Artifacts {
		if a.NEIRVersion != "v2.0.0" {
			t.Errorf("expected NEIRVersion v2.0.0 for %s, got %q", a.Path, a.NEIRVersion)
		}
	}
}

func TestSetNEIRVersionEmptyStore(t *testing.T) {
	s := NewStore(t.TempDir())
	s.SetNEIRVersion("v1.0.0") // must not panic
	if len(s.manifest.Artifacts) != 0 {
		t.Error("expected empty store to stay empty")
	}
}

func TestWriteToDiskInvalidPath(t *testing.T) {
	root := t.TempDir()
	s := NewStore(root)

	// Inject an artifact with a path-traversal name directly into the
	// manifest, bypassing Add()'s path validation.
	s.manifest.Artifacts = append(s.manifest.Artifacts, Artifact{
		ID:      "x",
		Path:    "../escape.txt",
		Content: []byte("evil"),
	})

	if err := s.WriteToDisk(); err == nil {
		t.Fatal("expected error for path traversal artifact")
	}
	if _, err := os.Stat(filepath.Join(root, "..", "escape.txt")); !os.IsNotExist(err) {
		t.Error("expected no file written outside root")
	}
}

func TestWriteToDiskUnwritableRoot(t *testing.T) {
	// A file (not a directory) as root makes MkdirAll fail.
	root := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(root, []byte("x"), 0o600); err != nil {
		t.Fatalf("create blocker: %v", err)
	}
	s := NewStore(root)
	if err := s.WriteToDisk(); err == nil {
		t.Fatal("expected error when root cannot be created as directory")
	}
}

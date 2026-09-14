// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package artifacts

import (
	"os"
	"testing"
)

func TestStoreAddAndGet(t *testing.T) {
	store := NewStore(t.TempDir())

	a, err := store.Add("test.go", []byte("package main"), KindCode, WithLanguage("go"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Path != "test.go" {
		t.Errorf("expected path test.go, got %s", a.Path)
	}
	if a.Kind != KindCode {
		t.Errorf("expected kind code, got %s", a.Kind)
	}
	if a.Size != 12 {
		t.Errorf("expected size 12, got %d", a.Size)
	}

	got, ok := store.Get("test.go")
	if !ok {
		t.Fatal("expected to find artifact")
	}
	if got.ContentHash != a.ContentHash {
		t.Error("hash mismatch")
	}
}

func TestStoreDeduplicate(t *testing.T) {
	store := NewStore(t.TempDir())

	// Add different content (won't be deduped by Add)
	store.Add("a.go", []byte("content a"), KindCode)
	store.Add("b.go", []byte("content b"), KindCode)

	// No duplicates to remove
	removed := store.Deduplicate()
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}

func TestStoreDeduplicateByHash(t *testing.T) {
	store := NewStore(t.TempDir())
	store.Add("a.go", []byte("content a"), KindCode)
	store.Add("b.go", []byte("content b"), KindCode)

	removed := store.Deduplicate()
	if removed != 0 {
		t.Errorf("expected 0 removed for different content, got %d", removed)
	}
}

func TestStoreRemove(t *testing.T) {
	store := NewStore(t.TempDir())
	store.Add("test.go", []byte("content"), KindCode)

	if !store.Remove("test.go") {
		t.Error("expected remove to return true")
	}
	if store.Remove("test.go") {
		t.Error("expected second remove to return false")
	}

	_, ok := store.Get("test.go")
	if ok {
		t.Error("expected artifact to be removed")
	}
}

func TestStoreList(t *testing.T) {
	store := NewStore(t.TempDir())
	store.Add("a.go", []byte("a"), KindCode)
	store.Add("b.yaml", []byte("b"), KindConfig)

	list := store.List()
	if len(list) != 2 {
		t.Errorf("expected 2 artifacts, got %d", len(list))
	}
}

func TestStoreGetByKind(t *testing.T) {
	store := NewStore(t.TempDir())
	store.Add("a.go", []byte("a"), KindCode)
	store.Add("b.yaml", []byte("b"), KindConfig)

	codes := store.GetByKind(KindCode)
	if len(codes) != 1 {
		t.Errorf("expected 1 code artifact, got %d", len(codes))
	}
}

func TestStoreSummary(t *testing.T) {
	store := NewStore(t.TempDir())
	store.Add("a.go", []byte("a"), KindCode)
	store.Add("b.yaml", []byte("b"), KindConfig)
	store.Add("c.md", []byte("c"), KindDocs)

	summary := store.Summary()
	if summary["total"] != 3 {
		t.Errorf("expected 3 total, got %d", summary["total"])
	}
	if summary["code"] != 1 {
		t.Errorf("expected 1 code, got %d", summary["code"])
	}
}

func TestStoreWriteAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	store.SetProject("test-proj")
	store.Add("a.go", []byte("package main"), KindCode, WithLanguage("go"))
	store.Add("b.yaml", []byte("key: value"), KindConfig)

	if err := store.WriteToDisk(); err != nil {
		t.Fatalf("write error: %v", err)
	}

	store2 := NewStore(dir)
	if err := store2.LoadFromDisk(); err != nil {
		t.Fatalf("load error: %v", err)
	}

	list := store2.List()
	if len(list) != 2 {
		t.Errorf("expected 2 artifacts after load, got %d", len(list))
	}

	manifest, _ := os.ReadFile(dir + "/.artifacts.json")
	if manifest == nil {
		t.Error("expected manifest file")
	}
}

func TestStoreLoadNonexistent(t *testing.T) {
	store := NewStore(t.TempDir())
	err := store.LoadFromDisk()
	if err != nil {
		t.Fatalf("expected nil error for nonexistent manifest: %v", err)
	}
}

func TestStoreOptions(t *testing.T) {
	store := NewStore(t.TempDir())
	a, err := store.Add("test.go", []byte("content"), KindCode,
		WithLanguage("go"),
		WithNEIRVersion("1.0"),
		WithSource("compiler"),
		WithMetadata("key", "value"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Language != "go" {
		t.Errorf("expected language go, got %s", a.Language)
	}
	if a.NEIRVersion != "1.0" {
		t.Errorf("expected neir version 1.0, got %s", a.NEIRVersion)
	}
	if a.Source != "compiler" {
		t.Errorf("expected source compiler, got %s", a.Source)
	}
	if a.Metadata["key"] != "value" {
		t.Errorf("expected metadata key=value")
	}
}

func TestDetectKind(t *testing.T) {
	tests := []struct {
		path string
		want ArtifactKind
	}{
		{"main.go", KindCode},
		{"app.ts", KindCode},
		{"config.yaml", KindConfig},
		{"README.md", KindDocs},
		{"Dockerfile", KindDocker},
		{"docker-compose.yml", KindDocker},
		{".github/workflows/ci.yml", KindCI},
		{".github/copilot-instructions.md", KindAI},
		{"CLAUDE.md", KindAI},
		{".cursorrules", KindAI},
		{".gemini/CONFIG.md", KindAI},
		{".opencode/context.md", KindAI},
		{"AGENTS.md", KindAI},
	}

	for _, tt := range tests {
		got := DetectKind(tt.path)
		if got != tt.want {
			t.Errorf("DetectKind(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestComputeHash(t *testing.T) {
	h1 := computeHash([]byte("hello"))
	h2 := computeHash([]byte("hello"))
	h3 := computeHash([]byte("world"))

	if h1 != h2 {
		t.Error("expected same hash for same content")
	}
	if h1 == h3 {
		t.Error("expected different hash for different content")
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID("test.go")
	id2 := generateID("test.go")
	id3 := generateID("other.go")

	if id1 != id2 {
		t.Error("expected same ID for same path")
	}
	if id1 == id3 {
		t.Error("expected different ID for different path")
	}
}

// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package lock

import (
	"sync"
	"testing"
)

func TestGenerateWithAlgorithm(t *testing.T) {
	artifacts := []ArtifactInfo{
		{Path: "a.txt", Content: []byte("hello")},
		{Path: "b.txt", Content: []byte("world")},
	}

	lock, err := GenerateWithAlgorithm(artifacts, HashSHA512)
	if err != nil {
		t.Fatal(err)
	}

	if lock.Algorithm != HashSHA512 {
		t.Errorf("expected sha512 algorithm, got %s", lock.Algorithm)
	}
	if lock.Checksum == "" {
		t.Error("expected non-empty checksum")
	}
	if len(lock.Artifacts) != 2 {
		t.Errorf("expected 2 artifacts, got %d", len(lock.Artifacts))
	}
}

func TestVerifyConcurrent(t *testing.T) {
	artifacts := []ArtifactInfo{
		{Path: "a.txt", Content: []byte("hello")},
		{Path: "b.txt", Content: []byte("world")},
	}

	lock, _ := Generate(artifacts)

	result := VerifyConcurrent(lock, artifacts, 4)
	if len(result.Changes) != 0 {
		t.Errorf("expected 0 changes, got %d: %v", len(result.Changes), result.Changes)
	}
	if result.Duration == 0 {
		t.Error("expected non-zero duration")
	}
}

func TestVerifyConcurrentModified(t *testing.T) {
	original := []ArtifactInfo{
		{Path: "a.txt", Content: []byte("hello")},
	}
	lock, _ := Generate(original)

	current := []ArtifactInfo{
		{Path: "a.txt", Content: []byte("modified")},
	}

	result := VerifyConcurrent(lock, current, 4)
	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	if result.Changes[0] != "modified: a.txt" {
		t.Errorf("expected 'modified: a.txt', got %s", result.Changes[0])
	}
}

func TestVerifyConcurrentAdded(t *testing.T) {
	lock, _ := Generate([]ArtifactInfo{
		{Path: "a.txt", Content: []byte("hello")},
	})

	current := []ArtifactInfo{
		{Path: "a.txt", Content: []byte("hello")},
		{Path: "b.txt", Content: []byte("new")},
	}

	result := VerifyConcurrent(lock, current, 4)
	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	if result.Changes[0] != "added: b.txt" {
		t.Errorf("expected 'added: b.txt', got %s", result.Changes[0])
	}
}

func TestVerifyConcurrentRemoved(t *testing.T) {
	lock, _ := Generate([]ArtifactInfo{
		{Path: "a.txt", Content: []byte("hello")},
		{Path: "b.txt", Content: []byte("world")},
	})

	current := []ArtifactInfo{
		{Path: "a.txt", Content: []byte("hello")},
	}

	result := VerifyConcurrent(lock, current, 4)
	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	if result.Changes[0] != "removed: b.txt" {
		t.Errorf("expected 'removed: b.txt', got %s", result.Changes[0])
	}
}

func TestVerifyDryRun(t *testing.T) {
	lock, _ := Generate([]ArtifactInfo{
		{Path: "a.txt", Content: []byte("hello")},
	})

	current := []ArtifactInfo{
		{Path: "a.txt", Content: []byte("modified")},
	}

	result := VerifyDryRun(lock, current)
	if len(result.Changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(result.Changes))
	}
}

func TestAuditLog(t *testing.T) {
	log := NewAuditLog()

	log.Add("created", "a.txt", "first file")
	log.Add("modified", "a.txt", "updated")

	entries := log.GetEntries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Action != "created" {
		t.Errorf("expected 'created', got %s", entries[0].Action)
	}
	if entries[1].Details != "updated" {
		t.Errorf("expected 'updated', got %s", entries[1].Details)
	}

	log.Clear()
	entries = log.GetEntries()
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after clear, got %d", len(entries))
	}
}

func TestAuditLogConcurrency(t *testing.T) {
	log := NewAuditLog()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			log.Add("test", "file.txt", "concurrent")
		}(i)
	}
	wg.Wait()

	entries := log.GetEntries()
	if len(entries) != 100 {
		t.Errorf("expected 100 entries, got %d", len(entries))
	}
}

func TestHashAlgorithms(t *testing.T) {
	content := []byte("test content")

	sha256Hash := hashContent(content, HashSHA256)
	sha512Hash := hashContent(content, HashSHA512)
	md5Hash := hashContent(content, HashMD5)

	if sha256Hash == sha512Hash {
		t.Error("SHA-256 and SHA-512 should produce different hashes")
	}
	if sha256Hash == md5Hash {
		t.Error("SHA-256 and MD5 should produce different hashes")
	}
	if len(sha256Hash) != 64 {
		t.Errorf("expected SHA-256 hash length 64, got %d", len(sha256Hash))
	}
	if len(sha512Hash) != 128 {
		t.Errorf("expected SHA-512 hash length 128, got %d", len(sha512Hash))
	}
	if len(md5Hash) != 32 {
		t.Errorf("expected MD5 hash length 32, got %d", len(md5Hash))
	}
}

func TestVerifyNilLock(t *testing.T) {
	result := VerifyConcurrent(nil, []ArtifactInfo{}, 4)
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestVerifyEmptyArtifacts(t *testing.T) {
	lock, _ := Generate([]ArtifactInfo{})
	result := VerifyConcurrent(lock, []ArtifactInfo{}, 4)
	if len(result.Changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(result.Changes))
	}
}

func TestWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/lock.json"

	lock, _ := Generate([]ArtifactInfo{
		{Path: "a.txt", Content: []byte("test")},
	})

	if err := WriteToFile(lock, path); err != nil {
		t.Fatal(err)
	}

	readLock, err := ReadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if readLock.Version != lock.Version {
		t.Errorf("version mismatch: %s vs %s", readLock.Version, lock.Version)
	}
	if readLock.Checksum != lock.Checksum {
		t.Errorf("checksum mismatch: %s vs %s", readLock.Checksum, lock.Checksum)
	}
}

func TestGenerateSortedArtifacts(t *testing.T) {
	lock, _ := Generate([]ArtifactInfo{
		{Path: "z.txt", Content: []byte("z")},
		{Path: "a.txt", Content: []byte("a")},
		{Path: "m.txt", Content: []byte("m")},
	})

	for i := 1; i < len(lock.Artifacts); i++ {
		if lock.Artifacts[i].Path < lock.Artifacts[i-1].Path {
			t.Errorf("artifacts not sorted: %s > %s", lock.Artifacts[i-1].Path, lock.Artifacts[i].Path)
		}
	}
}

func TestVerifyDuration(t *testing.T) {
	artifacts := make([]ArtifactInfo, 1000)
	for i := range artifacts {
		artifacts[i] = ArtifactInfo{
			Path:    string(rune('a'+i%26)) + ".txt",
			Content: []byte("content"),
		}
	}

	lock, _ := Generate(artifacts)
	result := VerifyConcurrent(lock, artifacts, 4)

	if result.Duration <= 0 {
		t.Error("expected positive duration")
	}
	if len(result.Changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(result.Changes))
	}
}

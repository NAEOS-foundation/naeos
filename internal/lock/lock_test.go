// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package lock

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	artifacts := []ArtifactInfo{
		{Path: "file1.txt", Content: []byte("content1")},
		{Path: "file2.txt", Content: []byte("content2")},
	}

	lockFile, err := Generate(artifacts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lockFile.Version != "1" {
		t.Errorf("expected version 1, got %s", lockFile.Version)
	}
	if len(lockFile.Artifacts) != 2 {
		t.Errorf("expected 2 artifacts, got %d", len(lockFile.Artifacts))
	}
	if lockFile.Checksum == "" {
		t.Error("expected non-empty checksum")
	}
}

func TestGenerateSortedPaths(t *testing.T) {
	artifacts := []ArtifactInfo{
		{Path: "z.txt", Content: []byte("z")},
		{Path: "a.txt", Content: []byte("a")},
	}

	lockFile, err := Generate(artifacts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lockFile.Artifacts[0].Path != "a.txt" {
		t.Errorf("expected first path to be a.txt, got %s", lockFile.Artifacts[0].Path)
	}
}

func TestVerifyNoChanges(t *testing.T) {
	artifacts := []ArtifactInfo{
		{Path: "file1.txt", Content: []byte("content1")},
	}

	lockFile, err := Generate(artifacts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	changes, err := Verify(lockFile, artifacts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 0 {
		t.Errorf("expected no changes, got %d", len(changes))
	}
}

func TestVerifyWithModification(t *testing.T) {
	artifacts := []ArtifactInfo{
		{Path: "file1.txt", Content: []byte("content1")},
	}

	lockFile, err := Generate(artifacts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	current := []ArtifactInfo{
		{Path: "file1.txt", Content: []byte("modified")},
	}

	changes, err := Verify(lockFile, current)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(changes))
	}
	if changes[0] != "modified: file1.txt" {
		t.Errorf("expected modification message, got %s", changes[0])
	}
}

func TestVerifyWithAddition(t *testing.T) {
	lockFile := &LockFile{
		Artifacts: []LockArtifact{
			{Path: "file1.txt", Checksum: "abc"},
		},
	}

	current := []ArtifactInfo{
		{Path: "file1.txt", Content: []byte("same")},
		{Path: "file2.txt", Content: []byte("new")},
	}

	changes, err := Verify(lockFile, current)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, c := range changes {
		if c == "added: file2.txt" {
			found = true
		}
	}
	if !found {
		t.Error("expected to find addition of file2.txt")
	}
}

func TestVerifyWithRemoval(t *testing.T) {
	lockFile := &LockFile{
		Artifacts: []LockArtifact{
			{Path: "file1.txt", Checksum: "abc"},
			{Path: "file2.txt", Checksum: "def"},
		},
	}

	current := []ArtifactInfo{
		{Path: "file1.txt", Content: []byte("same")},
	}

	changes, err := Verify(lockFile, current)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, c := range changes {
		if c == "removed: file2.txt" {
			found = true
		}
	}
	if !found {
		t.Error("expected to find removal of file2.txt")
	}
}

func TestVerifyNilLockFile(t *testing.T) {
	_, err := Verify(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil lock file")
	}
}

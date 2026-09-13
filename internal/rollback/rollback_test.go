// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package rollback

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSnapshotStoreCreate(t *testing.T) {
	store := NewStore(t.TempDir())
	artifacts := []SnapshotArtifact{
		{Path: "file1.txt", Content: []byte("content1")},
	}

	snap, err := store.Create("/output", artifacts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestSnapshotStoreListEmpty(t *testing.T) {
	store := NewStore(t.TempDir())
	snaps, err := store.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestSnapshotStoreList(t *testing.T) {
	store := NewStore(t.TempDir())
	store.Create("/output", []SnapshotArtifact{{Path: "f.txt", Content: []byte("c")}})
	store.Create("/output", []SnapshotArtifact{{Path: "f.txt", Content: []byte("c")}})

	snaps, err := store.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snaps) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(snaps))
	}
}

func TestSnapshotStoreRestore(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, _ := store.Create("/output", []SnapshotArtifact{
		{Path: "file1.txt", Content: []byte("restored content")},
	})

	restoreDir := t.TempDir()
	err := store.Restore(snap.ID, restoreDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(restoreDir, "file1.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "restored content" {
		t.Errorf("expected 'restored content', got %s", string(data))
	}
}

func TestSnapshotStoreRestoreNotFound(t *testing.T) {
	store := NewStore(t.TempDir())
	err := store.Restore("nonexistent", t.TempDir())
	if err == nil {
		t.Fatal("expected error for nonexistent snapshot")
	}
}

func TestSnapshotStoreDelete(t *testing.T) {
	store := NewStore(t.TempDir())
	snap, _ := store.Create("/output", []SnapshotArtifact{})

	err := store.Delete(snap.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snaps, _ := store.List()
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots after delete, got %d", len(snaps))
	}
}

func TestSnapshotStoreDeleteNotFound(t *testing.T) {
	store := NewStore(t.TempDir())
	err := store.Delete("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent snapshot")
	}
}

func TestSnapshotStoreLatest(t *testing.T) {
	store := NewStore(t.TempDir())
	store.Create("/output", []SnapshotArtifact{{Path: "old.txt", Content: []byte("old")}})
	time.Sleep(10 * time.Millisecond)
	store.Create("/output", []SnapshotArtifact{{Path: "new.txt", Content: []byte("new")}})

	latest, err := store.Latest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestSnapshotStoreLatestEmpty(t *testing.T) {
	store := NewStore(t.TempDir())
	_, err := store.Latest()
	if err == nil {
		t.Fatal("expected error for empty store")
	}
}

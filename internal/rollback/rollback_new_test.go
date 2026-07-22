package rollback

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateAndRestore(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	artifacts := []SnapshotArtifact{
		{Path: "file1.txt", Content: []byte("hello")},
		{Path: "sub/file2.txt", Content: []byte("world")},
	}

	snap, err := store.Create("/output", artifacts)
	if err != nil {
		t.Fatal(err)
	}

	if snap.ID == "" {
		t.Error("expected non-empty ID")
	}
	if snap.Manifest == nil {
		t.Error("expected manifest")
	}
	if len(snap.Manifest.Files) != 2 {
		t.Errorf("expected 2 files in manifest, got %d", len(snap.Manifest.Files))
	}

	target := filepath.Join(dir, "restore-target")
	if err := store.Restore(snap.ID, target); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(target, "file1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("expected 'hello', got %q", string(data))
	}
}

func TestVerifyIntegrity(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", []SnapshotArtifact{
		{Path: "a.txt", Content: []byte("data")},
	})
	if err != nil {
		t.Fatal(err)
	}

	snapDir := filepath.Join(store.snapshotDir(), snap.ID)
	if err := os.Remove(filepath.Join(snapDir, "a.txt")); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(dir, "restore-target")
	err = store.Restore(snap.ID, target)
	if err == nil {
		t.Error("expected integrity check to fail")
	}
}

func TestExportImport(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", []SnapshotArtifact{
		{Path: "test.txt", Content: []byte("exported")},
	})
	if err != nil {
		t.Fatal(err)
	}

	exportPath := filepath.Join(dir, "export.tar.gz")
	if err := store.Export(snap.ID, exportPath); err != nil {
		t.Fatal(err)
	}

	store2 := NewStore(filepath.Join(dir, "import"))
	imported, err := store2.Import(exportPath)
	if err != nil {
		t.Fatal(err)
	}

	if imported.ID == "" {
		t.Error("expected non-empty ID on import")
	}
}

func TestAtomicRestoreFailure(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", []SnapshotArtifact{
		{Path: "a.txt", Content: []byte("ok")},
	})
	if err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(dir, "target")
	os.MkdirAll(target, 0o755)
	os.WriteFile(filepath.Join(target, "existing.txt"), []byte("keep"), 0o600)

	if err := store.Restore(snap.ID, target); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(target, "existing.txt")); err == nil {
		t.Error("expected existing file to be replaced")
	}

	if _, err := os.Stat(filepath.Join(target, "a.txt")); err != nil {
		t.Error("expected restored file to exist")
	}
}

func TestDeleteSnapshot(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", []SnapshotArtifact{
		{Path: "a.txt", Content: []byte("data")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Delete(snap.ID); err != nil {
		t.Fatal(err)
	}

	if err := store.Delete("nonexistent"); err == nil {
		t.Error("expected error for nonexistent snapshot")
	}
}

func TestLatestSnapshot(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	store.Create("/output", []SnapshotArtifact{
		{Path: "a.txt", Content: []byte("first")},
	})
	time.Sleep(10 * time.Millisecond)
	store.Create("/output", []SnapshotArtifact{
		{Path: "b.txt", Content: []byte("second")},
	})

	latest, err := store.Latest()
	if err != nil {
		t.Fatal(err)
	}

	if latest.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestListEmpty(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snaps, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestRestoreNotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.Restore("nonexistent", filepath.Join(dir, "target"))
	if err == nil {
		t.Error("expected error for nonexistent snapshot")
	}
}

func TestImportEmptyArchive(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	exportPath := filepath.Join(dir, "empty.tar.gz")

	emptyFile, err := os.Create(exportPath)
	if err != nil {
		t.Fatal(err)
	}
	emptyFile.Close()

	_, err = store.Import(exportPath)
	if err == nil {
		t.Error("expected error for empty/corrupt archive")
	}
}

func TestImportCorruptGzip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	exportPath := filepath.Join(dir, "corrupt.tar.gz")

	if err := os.WriteFile(exportPath, []byte("not-a-gzip-file"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := store.Import(exportPath)
	if err == nil {
		t.Error("expected error for corrupt gzip")
	}
}

func TestCreateEmptyArtifacts(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", nil)
	if err != nil {
		t.Fatal(err)
	}
	if snap.ID == "" {
		t.Error("expected non-empty ID")
	}
	if snap.Manifest == nil {
		t.Error("expected manifest for empty artifact list")
	}
	if len(snap.Manifest.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(snap.Manifest.Files))
	}
}

func TestManifestChecksum(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", []SnapshotArtifact{
		{Path: "a.txt", Content: []byte("test")},
		{Path: "b.txt", Content: []byte("data")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if snap.Manifest.Checksum == "" {
		t.Error("expected non-empty checksum")
	}
	if snap.Manifest.TotalSize != 8 {
		t.Errorf("expected total size 8, got %d", snap.Manifest.TotalSize)
	}
}

func TestValidateSnapshotIDEmpty(t *testing.T) {
	err := validateSnapshotID("")
	if err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestValidateSnapshotIDPathSeparator(t *testing.T) {
	err := validateSnapshotID("foo/bar")
	if err == nil {
		t.Error("expected error for ID with slash")
	}
}

func TestValidateSnapshotIDBackslash(t *testing.T) {
	err := validateSnapshotID("foo\\bar")
	if err == nil {
		t.Error("expected error for ID with backslash")
	}
}

func TestValidateSnapshotIDDotDot(t *testing.T) {
	err := validateSnapshotID("foo..bar")
	if err == nil {
		t.Error("expected error for ID with ..")
	}
}

func TestDeleteInvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	err := store.Delete("")
	if err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestLatestEmpty(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	_, err := store.Latest()
	if err == nil {
		t.Error("expected error when no snapshots")
	}
}

func TestExportInvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	err := store.Export("", filepath.Join(dir, "out.tar.gz"))
	if err == nil {
		t.Error("expected error for invalid ID")
	}
}

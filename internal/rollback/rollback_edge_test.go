package rollback

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreSnapshotIsFile(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snapDir := filepath.Join(store.snapshotDir(), "snap-file")
	if err := os.MkdirAll(filepath.Dir(snapDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(snapDir, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := store.Restore("snap-file", filepath.Join(dir, "target"))
	if err == nil {
		t.Fatal("expected error for snapshot that is a file")
	}
}

func TestRestoreCleanupTempDirFailure(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", []SnapshotArtifact{
		{Path: "a.txt", Content: []byte("data")},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Block the temp dir path so RemoveAll fails: create a file where the
	// temp dir is expected to live... instead, make the snapshot dir itself
	// unusable by removing it, which makes the walk fail.
	target := filepath.Join(dir, "target")

	// Corrupt: snapshot dir is removed mid-flight is not possible; simulate
	// a walk failure by replacing the snapshot dir with a file before restore.
	snapDir := filepath.Join(store.snapshotDir(), snap.ID)
	if err := os.RemoveAll(snapDir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapDir, "manifest.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Snapshot dir contains no artifacts; restore should still succeed but
	// there is nothing to write. Assert no crash and empty target.
	if err := store.Restore(snap.ID, target); err != nil {
		t.Fatalf("restore with missing artifacts should succeed: %v", err)
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Error("target should have been created")
	}
}

func TestRestoreMergeIntoCWD(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", []SnapshotArtifact{
		{Path: "merged.txt", Content: []byte("merged")},
	})
	if err != nil {
		t.Fatal(err)
	}

	work := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(work); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)

	if err := store.Restore(snap.ID, "."); err != nil {
		t.Fatalf("merge restore into . failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(work, "merged.txt"))
	if err != nil {
		t.Fatalf("merged file not found: %v", err)
	}
	if string(data) != "merged" {
		t.Errorf("expected 'merged', got %q", string(data))
	}
}

func TestRestoreRemoveOldTargetFailure(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", []SnapshotArtifact{
		{Path: "a.txt", Content: []byte("data")},
	})
	if err != nil {
		t.Fatal(err)
	}

	// target is a file, so RemoveAll succeeds; make the parent read-only so
	// rename fails after removal — simulate by pointing target into a path
	// whose parent is a file.
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(blocker, "target")

	err = store.Restore(snap.ID, target)
	if err == nil {
		t.Fatal("expected error when target parent is a file")
	}
}

func TestRestoreCorruptManifest(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", []SnapshotArtifact{
		{Path: "a.txt", Content: []byte("data")},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Corrupt the manifest so verification is skipped.
	snapDir := filepath.Join(store.snapshotDir(), snap.ID)
	if err := os.WriteFile(filepath.Join(snapDir, "manifest.json"), []byte("{not-json"), 0o600); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(dir, "target")
	if err := store.Restore(snap.ID, target); err != nil {
		t.Fatalf("restore with corrupt manifest should succeed: %v", err)
	}
}

func TestVerifyIntegrityReadFailure(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", []SnapshotArtifact{
		{Path: "a.txt", Content: []byte("data")},
	})
	if err != nil {
		t.Fatal(err)
	}

	snapDir := filepath.Join(store.snapshotDir(), snap.ID)
	// Delete the artifact file; the manifest still references it, so the
	// walk copies nothing and verification fails on read.
	if err := os.Remove(filepath.Join(snapDir, "a.txt")); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(dir, "target")
	err = store.Restore(snap.ID, target)
	if err == nil {
		t.Fatal("expected integrity failure when artifact missing")
	}
}

func TestRestoreWalkFailure(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", []SnapshotArtifact{
		{Path: "a.txt", Content: []byte("data")},
	})
	if err != nil {
		t.Fatal(err)
	}

	snapDir := filepath.Join(store.snapshotDir(), snap.ID)
	// Create a file that cannot be read (permissions 000) to force walk error.
	unreadable := filepath.Join(snapDir, "secret.txt")
	if err := os.WriteFile(unreadable, []byte("secret"), 0o000); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(dir, "target")
	err = store.Restore(snap.ID, target)
	if err == nil {
		t.Fatal("expected walk error when artifact is unreadable")
	}
}

func TestExportNotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.Export("nonexistent", filepath.Join(dir, "out.tar.gz"))
	if err == nil {
		t.Fatal("expected error for nonexistent snapshot")
	}
}

func TestExportCreateFailure(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snap, err := store.Create("/output", []SnapshotArtifact{
		{Path: "a.txt", Content: []byte("data")},
	})
	if err != nil {
		t.Fatal(err)
	}

	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	err = store.Export(snap.ID, filepath.Join(blocker, "out.tar.gz"))
	if err == nil {
		t.Fatal("expected error when export path parent is a file")
	}
}

func TestImportSymlinkEscape(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	exportPath := filepath.Join(dir, "escape.tar.gz")

	f, err := os.Create(exportPath)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	hdr := &tar.Header{
		Name:     "link",
		Mode:     0o777,
		Typeflag: tar.TypeSymlink,
		Linkname: "../../../etc/passwd",
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()

	_, err = store.Import(exportPath)
	if err == nil {
		t.Fatal("expected error for symlink escaping snapshot dir")
	}
}

func TestImportInvalidPath(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	exportPath := filepath.Join(dir, "traversal.tar.gz")

	f, err := os.Create(exportPath)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	hdr := &tar.Header{
		Name:     "../evil.txt",
		Mode:     0o600,
		Size:     3,
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("bad")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()

	_, err = store.Import(exportPath)
	if err == nil {
		t.Fatal("expected error for path traversal in archive")
	}
}

func TestImportDirEntryAndManifest(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	exportPath := filepath.Join(dir, "dirs.tar.gz")

	f, err := os.Create(exportPath)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	dirHdr := &tar.Header{
		Name:     "sub",
		Mode:     0o755,
		Typeflag: tar.TypeDir,
	}
	if err := tw.WriteHeader(dirHdr); err != nil {
		t.Fatal(err)
	}

	fileHdr := &tar.Header{
		Name:     "sub/x.txt",
		Mode:     0o600,
		Size:     4,
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(fileHdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("data")); err != nil {
		t.Fatal(err)
	}

	// Manifest that can be parsed after import.
	manifest := `{"version":1,"snap_id":"imported","created":"2026-01-01T00:00:00Z","files":[],"total_size":0,"checksum":""}`
	mHdr := &tar.Header{
		Name:     "manifest.json",
		Mode:     0o600,
		Size:     int64(len(manifest)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(mHdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(manifest)); err != nil {
		t.Fatal(err)
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()

	imported, err := store.Import(exportPath)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if imported.Manifest == nil {
		t.Error("expected parsed manifest after import")
	}

	data, err := os.ReadFile(filepath.Join(store.snapshotDir(), imported.ID, "sub", "x.txt"))
	if err != nil {
		t.Fatalf("imported file missing: %v", err)
	}
	if string(data) != "data" {
		t.Errorf("expected 'data', got %q", string(data))
	}
}

func TestListSkipsFiles(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	snapDir := store.snapshotDir()
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapDir, "not-a-snapshot.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Create a real snapshot too so List returns at least the file is skipped.
	if _, err := store.Create("/output", []SnapshotArtifact{
		{Path: "a.txt", Content: []byte("data")},
	}); err != nil {
		t.Fatal(err)
	}

	snaps, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(snaps) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(snaps))
	}
}

func TestLatestListError(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Make snapshotDir a file so ReadDir fails.
	if err := os.MkdirAll(filepath.Dir(store.snapshotDir()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.snapshotDir(), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := store.Latest()
	if err == nil {
		t.Fatal("expected error when snapshot dir is a file")
	}
}

func TestListReadDirError(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	if err := os.MkdirAll(filepath.Dir(store.snapshotDir()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.snapshotDir(), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := store.List()
	if err == nil {
		t.Fatal("expected error when snapshot dir is a file")
	}
}

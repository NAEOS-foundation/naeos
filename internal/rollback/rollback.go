package rollback

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Snapshot struct {
	ID        string           `json:"id"`
	Timestamp time.Time        `json:"timestamp"`
	OutputDir string           `json:"output_dir"`
	Artifacts []SnapshotArtifact `json:"artifacts"`
	Manifest  *Manifest        `json:"manifest,omitempty"`
}

type SnapshotArtifact struct {
	Path    string `json:"path"`
	Content []byte `json:"-"`
}

type Manifest struct {
	Version   int               `json:"version"`
	SnapID    string            `json:"snap_id"`
	Created   time.Time         `json:"created"`
	Files     []ManifestFile    `json:"files"`
	TotalSize int64             `json:"total_size"`
	Checksum  string            `json:"checksum"`
}

type ManifestFile struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
}

type SnapshotStore struct {
	baseDir string
}

func NewStore(baseDir string) *SnapshotStore {
	return &SnapshotStore{baseDir: baseDir}
}

func (s *SnapshotStore) snapshotDir() string {
	return filepath.Join(s.baseDir, ".naeos", "snapshots")
}

func (s *SnapshotStore) Create(outputDir string, artifacts []SnapshotArtifact) (*Snapshot, error) {
	createdAt := time.Now()
	id := fmt.Sprintf("snap-%d", createdAt.UnixNano())
	snapDir := filepath.Join(s.snapshotDir(), id)

	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		return nil, fmt.Errorf("create snapshot dir: %w", err)
	}

	var files []ManifestFile
	var totalSize int64
	hasher := sha256.New()

	for _, a := range artifacts {
		path := filepath.Join(snapDir, a.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, a.Content, 0o600); err != nil {
			return nil, err
		}

		fHasher := sha256.New()
		fHasher.Write(a.Content)
		checksum := fmt.Sprintf("%x", fHasher.Sum(nil))

		files = append(files, ManifestFile{
			Path:     a.Path,
			Size:     int64(len(a.Content)),
			Checksum: checksum,
		})
		totalSize += int64(len(a.Content))
		hasher.Write(a.Content)
	}

	manifest := &Manifest{
		Version:   1,
		SnapID:    id,
		Created:   createdAt,
		Files:     files,
		TotalSize: totalSize,
		Checksum:  fmt.Sprintf("%x", hasher.Sum(nil)),
	}

	manifestPath := filepath.Join(snapDir, "manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, data, 0o600); err != nil {
		return nil, fmt.Errorf("write manifest: %w", err)
	}

	snap := &Snapshot{
		ID:        id,
		Timestamp: createdAt,
		OutputDir: outputDir,
		Artifacts: artifacts,
		Manifest:  manifest,
	}

	return snap, nil
}

func (s *SnapshotStore) List() ([]Snapshot, error) {
	snapDir := s.snapshotDir()
	entries, err := os.ReadDir(snapDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var snapshots []Snapshot
	for _, entry := range entries {
		if entry.IsDir() {
			snap := Snapshot{
				ID: entry.Name(),
			}
			info, err := entry.Info()
			if err == nil {
				snap.Timestamp = info.ModTime()
			}

			manifestPath := filepath.Join(snapDir, entry.Name(), "manifest.json")
			if data, err := os.ReadFile(manifestPath); err == nil {
				var m Manifest
				if json.Unmarshal(data, &m) == nil {
					snap.Manifest = &m
					snap.OutputDir = m.Files[0].Path
				}
			}

			snapshots = append(snapshots, snap)
		}
	}
	return snapshots, nil
}

func (s *SnapshotStore) Restore(snapshotID, targetDir string) error {
	snapDir := filepath.Join(s.snapshotDir(), snapshotID)
	info, err := os.Stat(snapDir)
	if err != nil {
		return fmt.Errorf("snapshot %s not found: %w", snapshotID, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("snapshot %s is not a directory", snapshotID)
	}

	tmpDir := targetDir + ".tmp-rollback"
	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("cleanup temp dir: %w", err)
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	err = filepath.Walk(snapDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(snapDir, path)
		if err != nil {
			return err
		}
		if relPath == "manifest.json" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		target := filepath.Join(tmpDir, relPath)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600)
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("restore to temp: %w", err)
	}

	manifestPath := filepath.Join(snapDir, "manifest.json")
	if data, err := os.ReadFile(manifestPath); err == nil {
		var m Manifest
		if json.Unmarshal(data, &m) == nil {
			if err := s.verifyIntegrity(tmpDir, &m); err != nil {
				os.RemoveAll(tmpDir)
				return fmt.Errorf("integrity check failed: %w", err)
			}
		}
	}

	if err := os.RemoveAll(targetDir); err != nil && !os.IsNotExist(err) {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("remove old target: %w", err)
	}

	if err := os.Rename(tmpDir, targetDir); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("atomic rename: %w", err)
	}

	return nil
}

func (s *SnapshotStore) verifyIntegrity(dir string, manifest *Manifest) error {
	for _, f := range manifest.Files {
		path := filepath.Join(dir, f.Path)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", f.Path, err)
		}
		hasher := sha256.New()
		hasher.Write(data)
		checksum := fmt.Sprintf("%x", hasher.Sum(nil))
		if checksum != f.Checksum {
			return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", f.Path, f.Checksum, checksum)
		}
	}
	return nil
}

func (s *SnapshotStore) Delete(snapshotID string) error {
	snapDir := filepath.Join(s.snapshotDir(), snapshotID)
	if _, err := os.Stat(snapDir); os.IsNotExist(err) {
		return fmt.Errorf("snapshot %s not found", snapshotID)
	}
	return os.RemoveAll(snapDir)
}

func (s *SnapshotStore) Latest() (*Snapshot, error) {
	snaps, err := s.List()
	if err != nil {
		return nil, err
	}
	if len(snaps) == 0 {
		return nil, fmt.Errorf("no snapshots found")
	}
	latest := snaps[0]
	for _, snap := range snaps[1:] {
		if snap.Timestamp.After(latest.Timestamp) {
			latest = snap
		}
	}
	return &latest, nil
}

func (s *SnapshotStore) Export(snapshotID, destPath string) error {
	snapDir := filepath.Join(s.snapshotDir(), snapshotID)
	if _, err := os.Stat(snapDir); os.IsNotExist(err) {
		return fmt.Errorf("snapshot %s not found", snapshotID)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create export file: %w", err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	return filepath.Walk(snapDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(snapDir, path)
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = tw.Write(data)
		return err
	})
}

func (s *SnapshotStore) Import(srcPath string) (*Snapshot, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("open import file: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	id := fmt.Sprintf("snap-%d", time.Now().UnixMilli())
	snapDir := filepath.Join(s.snapshotDir(), id)
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		return nil, fmt.Errorf("create snapshot dir: %w", err)
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		target := filepath.Join(snapDir, header.Name)
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return nil, err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil, err
		}

		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}

		if err := os.WriteFile(target, data, 0o600); err != nil {
			return nil, err
		}
	}

	snap := &Snapshot{
		ID:        id,
		Timestamp: time.Now(),
	}

	manifestPath := filepath.Join(snapDir, "manifest.json")
	if data, err := os.ReadFile(manifestPath); err == nil {
		var m Manifest
		if json.Unmarshal(data, &m) == nil {
			snap.Manifest = &m
		}
	}

	return snap, nil
}

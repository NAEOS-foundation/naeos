package rollback

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/securityext"
)

type Snapshot struct {
	ID        string             `json:"id"`
	Timestamp time.Time          `json:"timestamp"`
	OutputDir string             `json:"output_dir"`
	Artifacts []SnapshotArtifact `json:"artifacts"`
	Manifest  *Manifest          `json:"manifest,omitempty"`
}

type SnapshotArtifact struct {
	Path    string `json:"path"`
	Content []byte `json:"-"`
}

type Manifest struct {
	Version   int            `json:"version"`
	SnapID    string         `json:"snap_id"`
	Created   time.Time      `json:"created"`
	Files     []ManifestFile `json:"files"`
	TotalSize int64          `json:"total_size"`
	Checksum  string         `json:"checksum"`
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
		return nil, naeoserr.Wrapf(err, naeoserr.ErrPipeline, "create snapshot dir")
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
		return nil, naeoserr.Wrapf(err, naeoserr.ErrPipeline, "marshal manifest")
	}
	if err := os.WriteFile(manifestPath, data, 0o600); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrPipeline, "write manifest")
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
	if err := validateSnapshotID(snapshotID); err != nil {
		return err
	}
	snapDir := filepath.Join(s.snapshotDir(), snapshotID)
	info, err := os.Stat(snapDir)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrNotFound, "snapshot %s not found", snapshotID)
	}
	if !info.IsDir() {
		return naeoserr.New(naeoserr.ErrPipeline, fmt.Sprintf("snapshot %s is not a directory", snapshotID))
	}

	tmpDir := targetDir + ".tmp-rollback"
	if err := os.RemoveAll(tmpDir); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrPipeline, "cleanup temp dir")
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrPipeline, "create temp dir")
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
		data, err := os.ReadFile(path) //nolint:gosec // G122: path is from filepath.Walk under known snapDir
		if err != nil {
			return err
		}
		target := filepath.Join(tmpDir, relPath)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600) //nolint:gosec // G703: target is rooted under snapDir
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		return naeoserr.Wrapf(err, naeoserr.ErrPipeline, "restore to temp")
	}

	manifestPath := filepath.Join(snapDir, "manifest.json")
	if data, err := os.ReadFile(manifestPath); err == nil {
		var m Manifest
		if json.Unmarshal(data, &m) == nil {
			if err := s.verifyIntegrity(tmpDir, &m); err != nil {
				os.RemoveAll(tmpDir)
				return naeoserr.Wrapf(err, naeoserr.ErrPipeline, "integrity check failed")
			}
		}
	}

	if targetDir == "." || targetDir == string(filepath.Separator) {
		// Merge restore: write snapshot files into the existing target instead of
		// replacing it, since RemoveAll/Rename cannot operate on "." or the root.
		err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(tmpDir, path)
			if err != nil {
				return err
			}
			dst := filepath.Join(targetDir, rel)
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(dst, data, 0o644)
		})
		os.RemoveAll(tmpDir)
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrPipeline, "merge restore into %s", targetDir)
		}
		return nil
	}

	if err := os.RemoveAll(targetDir); err != nil && !os.IsNotExist(err) {
		os.RemoveAll(tmpDir)
		return naeoserr.Wrapf(err, naeoserr.ErrPipeline, "remove old target")
	}

	if err := os.Rename(tmpDir, targetDir); err != nil {
		os.RemoveAll(tmpDir)
		return naeoserr.Wrapf(err, naeoserr.ErrPipeline, "atomic rename")
	}

	return nil
}

func (s *SnapshotStore) verifyIntegrity(dir string, manifest *Manifest) error {
	for _, f := range manifest.Files {
		path := filepath.Join(dir, f.Path)
		data, err := os.ReadFile(path)
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrPipeline, "read %s", f.Path)
		}
		hasher := sha256.New()
		hasher.Write(data)
		checksum := fmt.Sprintf("%x", hasher.Sum(nil))
		if checksum != f.Checksum {
			return naeoserr.New(naeoserr.ErrPipeline, fmt.Sprintf("checksum mismatch for %s: expected %s, got %s", f.Path, f.Checksum, checksum))
		}
	}
	return nil
}

func (s *SnapshotStore) Delete(snapshotID string) error {
	if err := validateSnapshotID(snapshotID); err != nil {
		return err
	}
	snapDir := filepath.Join(s.snapshotDir(), snapshotID)
	if _, err := os.Stat(snapDir); os.IsNotExist(err) {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("snapshot %s not found", snapshotID))
	}
	return os.RemoveAll(snapDir)
}

func (s *SnapshotStore) Latest() (*Snapshot, error) {
	snaps, err := s.List()
	if err != nil {
		return nil, err
	}
	if len(snaps) == 0 {
		return nil, naeoserr.New(naeoserr.ErrNotFound, "no snapshots found")
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
	if err := validateSnapshotID(snapshotID); err != nil {
		return err
	}
	snapDir := filepath.Join(s.snapshotDir(), snapshotID)
	if _, err := os.Stat(snapDir); os.IsNotExist(err) {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("snapshot %s not found", snapshotID))
	}

	f, err := os.Create(destPath)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrPipeline, "create export file")
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
		data, err := os.ReadFile(path) //nolint:gosec // G122: path is from filepath.Walk under known snapDir
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
		return nil, naeoserr.Wrapf(err, naeoserr.ErrPipeline, "open import file")
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrPipeline, "gzip reader")
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	id := fmt.Sprintf("snap-%d", time.Now().UnixMilli())
	snapDir := filepath.Join(s.snapshotDir(), id)
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrPipeline, "create snapshot dir")
	}

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		target, err := securityext.ValidateFilePath(filepath.Join(snapDir, header.Name), snapDir) //nolint:gosec // G305: validated by ValidateFilePath
		if err != nil {
			return nil, naeoserr.New(naeoserr.ErrPipeline, fmt.Sprintf("invalid path in archive: %s", header.Name))
		}
		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
			linkTarget, err := securityext.ValidateFilePath(filepath.Join(snapDir, header.Linkname), snapDir) //nolint:gosec // G305: validated by ValidateFilePath
			if err != nil {
				return nil, naeoserr.New(naeoserr.ErrPipeline, fmt.Sprintf("symlink target escapes snapshot directory: %s", header.Linkname))
			}
			_ = linkTarget
		}
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

func validateSnapshotID(id string) error {
	if id == "" {
		return naeoserr.New(naeoserr.ErrPipeline, "snapshot ID must not be empty")
	}
	if strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return naeoserr.New(naeoserr.ErrPipeline, "snapshot ID must not contain path separators")
	}
	if strings.Contains(id, "..") {
		return naeoserr.New(naeoserr.ErrPipeline, "snapshot ID must not contain relative path components")
	}
	return nil
}

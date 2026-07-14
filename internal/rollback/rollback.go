package rollback

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Snapshot struct {
	ID        string
	Timestamp time.Time
	OutputDir string
	Artifacts []SnapshotArtifact
}

type SnapshotArtifact struct {
	Path    string
	Content []byte
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
	id := fmt.Sprintf("snap-%d", time.Now().UnixMilli())
	snapDir := filepath.Join(s.snapshotDir(), id)

	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		return nil, fmt.Errorf("create snapshot dir: %w", err)
	}

	for _, a := range artifacts {
		path := filepath.Join(snapDir, a.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, a.Content, 0o600); err != nil {
			return nil, err
		}
	}

	snap := &Snapshot{
		ID:        id,
		Timestamp: time.Now(),
		OutputDir: outputDir,
		Artifacts: artifacts,
	}

	return snap, nil
}

func (s *SnapshotStore) List() ([]Snapshot, error) {
	snapDir := s.snapshotDir()
	entries, err := os.ReadDir(snapDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Snapshot directory does not exist yet — empty list
		}
		return nil, err
	}

	var snapshots []Snapshot
	for _, entry := range entries {
		if entry.IsDir() {
			snap := Snapshot{
				ID:        entry.Name(),
				OutputDir: filepath.Join(snapDir, entry.Name()),
			}
			info, err := entry.Info()
			if err == nil {
				snap.Timestamp = info.ModTime()
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

	return filepath.Walk(snapDir, func(path string, info os.FileInfo, err error) error {
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
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		target := filepath.Join(targetDir, relPath)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600)
	})
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

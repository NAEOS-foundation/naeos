package watch

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWatcher(t *testing.T) {
	t.Parallel()
	w := NewWatcher(time.Second, func(path string) {})
	if w == nil {
		t.Fatal("expected non-nil watcher")
	}
	if w.interval != time.Second {
		t.Errorf("expected interval 1s, got %v", w.interval)
	}
}

func TestNewWatcherDefaultInterval(t *testing.T) {
	t.Parallel()
	w := NewWatcher(0, func(path string) {})
	if w.interval != 500*time.Millisecond {
		t.Errorf("expected default interval 500ms, got %v", w.interval)
	}
}

func TestAddDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	w := NewWatcher(time.Second, func(path string) {})
	if err := w.AddDirectory(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(w.directories) != 1 {
		t.Errorf("expected 1 directory, got %d", len(w.directories))
	}
}

func TestAddDirectoryNotFound(t *testing.T) {
	t.Parallel()
	w := NewWatcher(time.Second, func(path string) {})
	err := w.AddDirectory("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestAddFileAsDirectory(t *testing.T) {
	t.Parallel()
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	os.WriteFile(tmpFile, []byte("test"), 0o600)
	w := NewWatcher(time.Second, func(path string) {})
	err := w.AddDirectory(tmpFile)
	if err == nil {
		t.Fatal("expected error for file instead of directory")
	}
}

func TestStartStop(t *testing.T) {
	t.Parallel()
	w := NewWatcher(time.Second, func(path string) {})
	if w.IsRunning() {
		t.Fatal("expected not running initially")
	}
	if err := w.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !w.IsRunning() {
		t.Fatal("expected running after start")
	}
	w.Stop()
	if w.IsRunning() {
		t.Fatal("expected not running after stop")
	}
}

func TestStartTwice(t *testing.T) {
	t.Parallel()
	w := NewWatcher(time.Second, func(path string) {})
	w.Start()
	err := w.Start()
	if err == nil {
		t.Fatal("expected error on double start")
	}
	w.Stop()
}

func TestSnapshot(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content1"), 0o600)
	os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("content2"), 0o600)

	w := NewWatcher(time.Second, func(path string) {})
	w.AddDirectory(dir)

	snap, err := w.Snapshot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snap) < 2 {
		t.Errorf("expected at least 2 entries, got %d", len(snap))
	}
}

func TestDetectChanges(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content1"), 0o600)

	w := NewWatcher(time.Second, func(path string) {})
	w.AddDirectory(dir)

	snap, _ := w.Snapshot()

	os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("content2"), 0o600)

	changes := w.DetectChanges(snap)
	if len(changes) == 0 {
		t.Fatal("expected at least 1 change")
	}

	found := false
	for _, c := range changes {
		if c.EventType == "created" {
			found = true
		}
	}
	if !found {
		t.Error("expected to find a 'created' event")
	}
}

func TestPipelineWatcherShouldProcess(t *testing.T) {
	t.Parallel()
	pw := NewPipelineWatcher("spec.yaml", "out", func(ctx context.Context, input string) error {
		return nil
	})

	tests := []struct {
		path string
		want bool
	}{
		{"config.yaml", true},
		{"config.yml", true},
		{"config.json", true},
		{"readme.txt", false},
		{"data.csv", false},
		{"noext", false},
	}

	for _, tt := range tests {
		got := pw.shouldProcess(tt.path)
		if got != tt.want {
			t.Errorf("shouldProcess(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestPipelineWatcherNew(t *testing.T) {
	t.Parallel()
	called := false
	pw := NewPipelineWatcher("spec.yaml", "output", func(ctx context.Context, input string) error {
		called = true
		return nil
	})

	if pw.specPath != "spec.yaml" {
		t.Errorf("expected specPath 'spec.yaml', got %q", pw.specPath)
	}
	if pw.outputDir != "output" {
		t.Errorf("expected outputDir 'output', got %q", pw.outputDir)
	}
	if pw.debounceMs != 500 {
		t.Errorf("expected debounceMs 500, got %d", pw.debounceMs)
	}
	if pw.ctx == nil {
		t.Error("expected non-nil context")
	}
	if pw.cancel == nil {
		t.Error("expected non-nil cancel function")
	}
	if pw.pipeline == nil {
		t.Error("expected non-nil pipeline function")
	}
	// Verify pipeline function is wired correctly
	err := pw.pipeline(pw.ctx, "test")
	if err != nil {
		t.Fatalf("unexpected error calling pipeline: %v", err)
	}
	if !called {
		t.Error("expected pipeline function to be called")
	}
}

func TestDetectChangesModified(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	targetFile := filepath.Join(dir, "file.txt")
	os.WriteFile(targetFile, []byte("original"), 0o600)

	w := NewWatcher(time.Second, func(path string) {})
	w.AddDirectory(dir)

	snap, _ := w.Snapshot()

	// Modify the existing file
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(targetFile, []byte("modified content"), 0o600)

	changes := w.DetectChanges(snap)
	if len(changes) == 0 {
		t.Fatal("expected at least 1 change event")
	}

	found := false
	for _, c := range changes {
		if c.EventType == "modified" && c.Path == targetFile {
			found = true
		}
	}
	if !found {
		t.Error("expected a 'modified' event for the changed file")
	}
}

func TestDetectChangesEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0o600)

	w := NewWatcher(time.Second, func(path string) {})
	w.AddDirectory(dir)

	snap, _ := w.Snapshot()

	// Detect changes without modifying anything
	changes := w.DetectChanges(snap)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}
}

func TestWatcherRunWithDebounce(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	onChangeCalled := make(chan string, 1)

	w := NewWatcher(50*time.Millisecond, func(path string) {
		select {
		case onChangeCalled <- path:
		default:
		}
	})
	w.AddDirectory(dir)

	errCh := make(chan error, 1)
	go func() {
		errCh <- w.Run(func() error {
			return nil
		})
	}()

	// Wait briefly for the watcher to start
	time.Sleep(100 * time.Millisecond)

	// Write a .yaml file to trigger the onChange callback
	yamlFile := filepath.Join(dir, "test.yaml")
	os.WriteFile(yamlFile, []byte("key: value"), 0o600)

	// Wait for onChange to be called (with timeout)
	select {
	case path := <-onChangeCalled:
		if filepath.Ext(path) != ".yaml" {
			t.Errorf("expected .yaml extension, got %q", filepath.Ext(path))
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for onChange callback")
	}

	// Stop the watcher
	w.Stop()

	// Drain any error (expected when stopped)
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
	}
}

func TestPipelineWatcherStop(t *testing.T) {
	t.Parallel()
	pw := NewPipelineWatcher("spec.yaml", "", func(ctx context.Context, input string) error {
		return nil
	})

	// Stop should not panic even if watcher was never started
	pw.Stop()

	// Context should be cancelled after Stop
	select {
	case <-pw.ctx.Done():
		// expected
	case <-time.After(1 * time.Second):
		t.Fatal("expected context to be cancelled after Stop")
	}
}

func TestPipelineWatcherStartAndStop(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	specFile := filepath.Join(dir, "spec.yaml")
	os.WriteFile(specFile, []byte("key: value"), 0o600)

	ran := make(chan struct{}, 1)
	pw := NewPipelineWatcher(specFile, "", func(ctx context.Context, input string) error {
		select {
		case ran <- struct{}{}:
		default:
		}
		return nil
	})

	// Start runs in a goroutine because it blocks on fsnotify
	go func() {
		_ = pw.Start()
	}()

	// Give the watcher a moment to initialize
	time.Sleep(200 * time.Millisecond)

	// Write a .yaml file to trigger the pipeline
	os.WriteFile(specFile, []byte("key: updated"), 0o600)

	// Wait for pipeline to be invoked (with timeout)
	select {
	case <-ran:
		// pipeline was called as expected
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for pipeline to run")
	}

	// Stop the watcher (Run loop won't exit because fsnotify isn't closed,
	// but Stop cancels context and sets running=false)
	pw.Stop()
}

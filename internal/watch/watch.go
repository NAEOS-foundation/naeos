package watch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	naeoslog "github.com/NAEOS-foundation/naeos/internal/shared/log"
)

type Watcher struct {
	directories []string
	interval    time.Duration
	onChange    func(path string)
	running     bool
	mu          sync.Mutex
}

type WatchEvent struct {
	Path      string
	Timestamp time.Time
	EventType string
}

type PipelineWatcher struct {
	watcher    *Watcher
	pipeline   func(ctx context.Context, input string) error
	specPath   string
	outputDir  string
	ctx        context.Context
	cancel     context.CancelFunc
	debounceMs int
}

func NewPipelineWatcher(specPath, outputDir string, pipelineFn func(ctx context.Context, input string) error) *PipelineWatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &PipelineWatcher{
		specPath:   specPath,
		outputDir:  outputDir,
		pipeline:   pipelineFn,
		ctx:        ctx,
		cancel:     cancel,
		debounceMs: 500,
	}
}

func (pw *PipelineWatcher) Start() error {
	specDir := filepath.Dir(pw.specPath)

	pw.watcher = NewWatcher(time.Duration(pw.debounceMs)*time.Millisecond, func(path string) {
		if pw.shouldProcess(path) {
			naeoslog.Info("spec change detected, re-running pipeline", "path", path)
			if err := pw.runPipeline(); err != nil {
				naeoslog.Error("pipeline re-run failed", "error", err)
			}
		}
	})

	if err := pw.watcher.AddDirectory(specDir); err != nil {
		return fmt.Errorf("watch spec dir: %w", err)
	}

	if pw.outputDir != "" {
		if _, err := os.Stat(pw.outputDir); err == nil {
			_ = pw.watcher.AddDirectory(pw.outputDir)
		}
	}

	return pw.watcher.Run(func() error {
		return nil
	})
}

func (pw *PipelineWatcher) Stop() {
	pw.cancel()
	if pw.watcher != nil {
		pw.watcher.Stop()
	}
}

func (pw *PipelineWatcher) shouldProcess(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml", ".json":
		return true
	}
	return false
}

func (pw *PipelineWatcher) runPipeline() error {
	data, err := os.ReadFile(pw.specPath)
	if err != nil {
		return fmt.Errorf("read spec: %w", err)
	}

	return pw.pipeline(pw.ctx, string(data))
}

func NewWatcher(interval time.Duration, onChange func(path string)) *Watcher {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	return &Watcher{
		interval: interval,
		onChange: onChange,
	}
}

func (w *Watcher) AddDirectory(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("cannot watch %s: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	w.directories = append(w.directories, dir)
	return nil
}

func (w *Watcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		return fmt.Errorf("watcher already running")
	}
	w.running = true
	return nil
}

func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.running = false
}

func (w *Watcher) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

func (w *Watcher) Snapshot() (map[string]int64, error) {
	snapshot := make(map[string]int64)
	for _, dir := range w.directories {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			snapshot[path] = info.ModTime().UnixMilli()
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return snapshot, nil
}

func (w *Watcher) DetectChanges(prev map[string]int64) []WatchEvent {
	var events []WatchEvent
	for _, dir := range w.directories {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			currentMod := info.ModTime().UnixMilli()
			prevMod, exists := prev[path]
			if !exists || currentMod > prevMod {
				eventType := "modified"
				if !exists {
					eventType = "created"
				}
				events = append(events, WatchEvent{
					Path:      path,
					Timestamp: time.Now(),
					EventType: eventType,
				})
			}
			return nil
		})
	}
	return events
}

func (w *Watcher) Run(fn func() error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create fsnotify watcher: %w", err)
	}
	defer watcher.Close()

	for _, dir := range w.directories {
		if err := watcher.Add(dir); err != nil {
			return fmt.Errorf("watch directory %s: %w", dir, err)
		}
	}

	if err := w.Start(); err != nil {
		return err
	}
	defer w.Stop()

	naeoslog.Info("watching for changes (fsnotify)")

	debounce := time.NewTimer(0)
	if !debounce.Stop() {
		<-debounce.C
	}
	defer debounce.Stop()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				ext := filepath.Ext(event.Name)
				switch ext {
				case ".yaml", ".yml", ".json":
					if w.onChange != nil {
						w.onChange(event.Name)
					}
					debounce.Reset(w.interval)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			naeoslog.Error("watcher error", "error", err)
		case <-debounce.C:
			naeoslog.Info("change detected, re-running pipeline")
			if err := fn(); err != nil {
				naeoslog.Error("pipeline error", "error", err)
			}
		}
	}
}

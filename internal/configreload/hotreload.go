package configreload

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	naeoslog "github.com/NAEOS-foundation/naeos/internal/shared/log"
)

type HotReloader struct {
	config   *Config
	watcher  *fsnotify.Watcher
	interval time.Duration
	stopCh   chan struct{}
	running  bool
	mu       sync.RWMutex
}

func NewHotReloader(config *Config) *HotReloader {
	return &HotReloader{
		config: config,
		stopCh: make(chan struct{}),
	}
}

func (hr *HotReloader) Start() error {
	hr.mu.Lock()
	if hr.running {
		hr.mu.Unlock()
		return fmt.Errorf("hot reloader already running")
	}
	hr.mu.Unlock()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	hr.watcher = watcher

	dir := filepath.Dir(hr.config.filePath)
	if dir == "" {
		dir = "."
	}
	if err := watcher.Add(dir); err != nil {
		watcher.Close()
		return fmt.Errorf("watch config directory %s: %w", dir, err)
	}

	hr.mu.Lock()
	hr.running = true
	hr.stopCh = make(chan struct{})
	hr.mu.Unlock()

	go hr.loop()

	naeoslog.Info("hot reload started", "path", hr.config.filePath)
	return nil
}

func (hr *HotReloader) Stop() {
	hr.mu.Lock()
	if !hr.running {
		hr.mu.Unlock()
		return
	}
	hr.running = false
	hr.mu.Unlock()

	close(hr.stopCh)
	if hr.watcher != nil {
		hr.watcher.Close()
	}
	naeoslog.Info("hot reload stopped")
}

func (hr *HotReloader) IsRunning() bool {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return hr.running
}

func (hr *HotReloader) loop() {
	debounce := time.NewTimer(0)
	if !debounce.Stop() {
		<-debounce.C
	}
	defer debounce.Stop()

	for {
		select {
		case <-hr.stopCh:
			return
		case event, ok := <-hr.watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				if hr.matchesConfigFile(event.Name) {
					debounce.Reset(300 * time.Millisecond)
				}
			}
		case err, ok := <-hr.watcher.Errors:
			if !ok {
				return
			}
			naeoslog.Error("config watcher error", "error", err)
		case <-debounce.C:
			if err := hr.config.Load(); err != nil {
				naeoslog.Error("config reload failed", "error", err)
			} else {
				naeoslog.Info("config reloaded", "version", hr.config.Version())
			}
		}
	}
}

func (hr *HotReloader) matchesConfigFile(path string) bool {
	base := filepath.Base(path)
	configBase := filepath.Base(hr.config.filePath)
	return base == configBase
}

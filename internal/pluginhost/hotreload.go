package pluginhost

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type PluginWatcher struct {
	pluginDir string
	manager   *Manager
	watcher   *fsnotify.Watcher
	debounce  time.Duration
	stopCh    chan struct{}
	doneCh    chan struct{}
	running   bool
	mu        sync.Mutex
}

func NewPluginWatcher(pluginDir string, manager *Manager) *PluginWatcher {
	return &PluginWatcher{
		pluginDir: pluginDir,
		manager:   manager,
		debounce:  500 * time.Millisecond,
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

func (pw *PluginWatcher) Start(ctx context.Context) error {
	pw.mu.Lock()
	if pw.running {
		pw.mu.Unlock()
		return nil
	}
	pw.mu.Unlock()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	pw.watcher = watcher

	if err := watcher.Add(pw.pluginDir); err != nil {
		watcher.Close()
		return err
	}

	pw.mu.Lock()
	pw.running = true
	pw.stopCh = make(chan struct{})
	pw.mu.Unlock()

	go pw.loop(ctx)

	return nil
}

func (pw *PluginWatcher) Stop() {
	pw.mu.Lock()
	if !pw.running {
		pw.mu.Unlock()
		return
	}
	pw.running = false
	pw.mu.Unlock()

	close(pw.stopCh)
	if pw.watcher != nil {
		pw.watcher.Close()
	}
	<-pw.doneCh
}

func (pw *PluginWatcher) IsRunning() bool {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	return pw.running
}

func (pw *PluginWatcher) loop(ctx context.Context) {
	defer close(pw.doneCh)
	defer pw.stopRunning()

	debounce := time.NewTimer(0)
	if !debounce.Stop() {
		<-debounce.C
	}
	defer debounce.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pw.stopCh:
			return
		case event, ok := <-pw.watcher.Events:
			if !ok {
				return
			}
			if pw.matchesPluginFile(event.Name) {
				if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
					debounce.Reset(pw.debounce)
				}
			}
		case _, ok := <-pw.watcher.Errors:
			if !ok {
				return
			}
		case <-debounce.C:
			pw.reloadPlugins()
		}
	}
}

func (pw *PluginWatcher) stopRunning() {
	pw.mu.Lock()
	pw.running = false
	pw.mu.Unlock()
}

func (pw *PluginWatcher) matchesPluginFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".so" || ext == ".wasm"
}

func (pw *PluginWatcher) reloadPlugins() {
	pluginFiles, err := filepath.Glob(filepath.Join(pw.pluginDir, "*.so"))
	if err != nil {
		return
	}
	wasmFiles, err := filepath.Glob(filepath.Join(pw.pluginDir, "*.wasm"))
	if err != nil {
		return
	}

	onDisk := make(map[string]bool, len(pluginFiles)+len(wasmFiles))
	for _, f := range pluginFiles {
		onDisk[f] = true
	}
	for _, f := range wasmFiles {
		onDisk[f] = true
	}

	for _, p := range pw.manager.List() {
		if !onDisk[p.Path] {
			_ = pw.manager.Uninstall(p.Name)
		}
	}

	for _, f := range pluginFiles {
		if !pw.isRegistered(f) {
			_, _ = pw.manager.Install(f)
		}
	}
}

func (pw *PluginWatcher) isRegistered(path string) bool {
	for _, p := range pw.manager.List() {
		if p.Path == path {
			return true
		}
	}
	return false
}

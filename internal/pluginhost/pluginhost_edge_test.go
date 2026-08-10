package pluginhost

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestManagerInstallOpenFailure(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	_, err := m.Install(filepath.Join(dir, "missing.so"))
	if err == nil {
		t.Fatal("expected error opening nonexistent plugin")
	}
}

func TestManagerLoadAllGoPluginLoadError(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	fakePath := filepath.Join(dir, "nonexistent.so")
	m.config.Plugins = []PluginInfo{
		{Name: "broken", Enabled: true, Path: fakePath},
	}

	err := m.LoadAll(&PluginContext{})
	if err == nil {
		t.Fatal("expected error from loadGoPlugin")
	}
	if len(m.plugins) != 0 {
		t.Error("expected no plugins registered after load failure")
	}
}

func TestManagerLoadAllMixedErrors(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	m.config.Plugins = []PluginInfo{
		{Name: "outside", Enabled: true, Path: "/outside/allowed/plugin.so"},
		{Name: "missing", Enabled: true, Path: filepath.Join(dir, "missing.so")},
		{Name: "skipped", Enabled: false, Path: filepath.Join(dir, "whatever.so")},
	}
	m.sandbox = NewSandbox(SandboxConfig{
		AllowedDirs: []string{dir},
		MaxCalls:    100,
	})

	err := m.LoadAll(&PluginContext{})
	if err == nil {
		t.Fatal("expected grouped error")
	}
}

func TestPluginWatcherLoopErrorEvent(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Skip("fsnotify not available")
	}
	w.watcher = watcher
	defer watcher.Close()

	go w.loop(context.Background())

	// A real (non-closed) error must be consumed without exiting the loop.
	watcher.Errors <- os.ErrPermission

	time.Sleep(30 * time.Millisecond)
	select {
	case <-w.doneCh:
		t.Fatal("loop exited after receiving a watcher error")
	default:
	}
	close(w.stopCh)
	<-w.doneCh
}

func TestPluginWatcherLoopDebounceReload(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)
	w.debounce = 20 * time.Millisecond

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Skip("fsnotify not available")
	}
	w.watcher = watcher
	defer watcher.Close()
	if err := watcher.Add(dir); err != nil {
		t.Fatalf("watch dir: %v", err)
	}

	go w.loop(context.Background())

	// Non-plugin files must not schedule a reload.
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write txt: %v", err)
	}

	// A .so create event schedules a debounced reload; reloadPlugins runs
	// and attempts (and fails silently on) Install of the bogus .so.
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(filepath.Join(dir, "plugin.so"), []byte("not a plugin"), 0o644); err != nil {
		t.Fatalf("write so: %v", err)
	}

	time.Sleep(120 * time.Millisecond)

	select {
	case <-w.doneCh:
		t.Fatal("loop exited unexpectedly")
	default:
	}
	close(w.stopCh)
	<-w.doneCh
}

func TestPluginWatcherStopNotRunning(t *testing.T) {
	dir := t.TempDir()
	w := NewPluginWatcher(dir, NewManager(dir))
	w.Stop() // must not block or panic
}

package pluginhost

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestPluginWatcherIsRegistered(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	// Add a plugin to the manager's config so List returns it
	m.config.Plugins = append(m.config.Plugins, PluginInfo{
		Name: "test-plugin",
		Path: filepath.Join(dir, "test.so"),
	})

	if !w.isRegistered(filepath.Join(dir, "test.so")) {
		t.Error("expected isRegistered to return true for registered plugin")
	}
	if w.isRegistered(filepath.Join(dir, "other.so")) {
		t.Error("expected isRegistered to return false for unregistered plugin")
	}
}

func TestPluginWatcherReloadPluginsUninstall(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	// Pre-register a plugin in the manager that doesn't exist on disk
	pInfo := PluginInfo{
		Name:    "old-plugin",
		Path:    filepath.Join(dir, "old.so"),
		Enabled: true,
	}
	m.config.Plugins = append(m.config.Plugins, pInfo)
	m.info["old-plugin"] = &pInfo

	w := NewPluginWatcher(dir, m)
	w.reloadPlugins()
	// Should attempt to uninstall old-plugin since old.so doesn't exist on disk
}

func TestPluginWatcherReloadPluginsInstallNew(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	// Create a fake .so file (won't actually load, but tests the path)
	os.WriteFile(filepath.Join(dir, "newplugin.so"), []byte("fake binary"), 0o644)

	w := NewPluginWatcher(dir, m)
	w.reloadPlugins()
	// Install will fail because it's not a real plugin, but the code path is exercised
}

func TestPluginWatcherReloadPluginsWasmFiles(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	os.WriteFile(filepath.Join(dir, "plugin.wasm"), []byte("fake wasm"), 0o644)

	w := NewPluginWatcher(dir, m)
	w.reloadPlugins()
}

func TestPluginWatcherLoopContextCancel(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Skip("fsnotify not available")
	}
	w.watcher = watcher
	defer watcher.Close()

	watcher.Add(dir)

	ctx, cancel := context.WithCancel(context.Background())
	go w.loop(ctx)
	cancel()
	<-w.doneCh
}

func TestPluginWatcherLoopStopCh(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Skip("fsnotify not available")
	}
	w.watcher = watcher
	defer watcher.Close()

	watcher.Add(dir)

	go w.loop(context.Background())
	close(w.stopCh)
	<-w.doneCh
}

func TestPluginWatcherLoopWatcherEventsClosed(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Skip("fsnotify not available")
	}
	w.watcher = watcher

	go w.loop(context.Background())
	watcher.Close()
	<-w.doneCh
}

func TestPluginWatcherLoopWatcherErrorsClosed(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Skip("fsnotify not available")
	}
	w.watcher = watcher

	go w.loop(context.Background())
	// Close the errors channel by closing the watcher
	watcher.Close()
	<-w.doneCh
}

func TestManagerLoadAllDisabledPlugin(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	m.config.Plugins = []PluginInfo{
		{Name: "disabled", Enabled: false, Path: "/fake/path.so"},
		{Name: "nopath", Enabled: true, Path: ""},
	}
	ctx := &PluginContext{}
	err := m.LoadAll(ctx)
	if err != nil {
		t.Fatalf("expected no error for disabled/no-path plugins, got %v", err)
	}
}

func TestManagerLoadAllSandboxValidationFail(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	m.config.Plugins = []PluginInfo{
		{Name: "bad-path", Enabled: true, Path: "/outside/allowed/plugin.so"},
	}
	m.sandbox = NewSandbox(SandboxConfig{
		AllowedDirs: []string{dir},
		MaxCalls:    100,
	})
	ctx := &PluginContext{}
	err := m.LoadAll(ctx)
	if err == nil {
		t.Fatal("expected error from sandbox validation")
	}
}

func TestManagerLazyLoadNotFound(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	ctx := &PluginContext{}
	err := m.lazyLoad("nonexistent", ctx)
	if err == nil {
		t.Fatal("expected error for nonexistent plugin")
	}
}

func TestManagerLazyLoadDisabled(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	m.config.Plugins = []PluginInfo{
		{Name: "disabled", Enabled: false, Path: "/fake/path.so"},
	}
	ctx := &PluginContext{}
	err := m.lazyLoad("disabled", ctx)
	if err == nil {
		t.Fatal("expected error for disabled plugin")
	}
}

func TestManagerLazyLoadAlreadyLoaded(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	m.config.Plugins = []PluginInfo{
		{Name: "loaded", Enabled: true, Path: "/fake/path.so", Loaded: true},
	}
	ctx := &PluginContext{}
	err := m.lazyLoad("loaded", ctx)
	if err != nil {
		t.Fatalf("expected no error for already-loaded plugin, got %v", err)
	}
}

func TestManagerLazyLoadBadSandboxPath(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	m.config.Plugins = []PluginInfo{
		{Name: "bad", Enabled: true, Path: "/outside/plugin.so"},
	}
	m.sandbox = NewSandbox(SandboxConfig{
		AllowedDirs: []string{dir},
		MaxCalls:    100,
	})
	ctx := &PluginContext{}
	err := m.lazyLoad("bad", ctx)
	if err == nil {
		t.Fatal("expected error from sandbox validation")
	}
}

func TestManagerLazyLoadLoadGoPluginFail(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	fakePath := filepath.Join(dir, "nonexistent.so")
	m.config.Plugins = []PluginInfo{
		{Name: "bad", Enabled: true, Path: fakePath},
	}
	ctx := &PluginContext{}
	err := m.lazyLoad("bad", ctx)
	if err == nil {
		t.Fatal("expected error from loadGoPlugin")
	}
}

func TestManagerExecuteLazyLoadPath(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	m.config.Plugins = []PluginInfo{
		{Name: "test", Enabled: true, Path: filepath.Join(dir, "fake.so")},
	}
	_, err := m.Execute(context.Background(), "test", "action", nil)
	if err == nil {
		t.Fatal("expected error from lazy load")
	}
}

func TestManagerExecuteNotFound(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	_, err := m.Execute(context.Background(), "nonexistent", "action", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent plugin")
	}
}

func TestManagerExecuteSuccess(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	p := newStubPlugin("stub")
	m.plugins["stub"] = p
	m.info["stub"] = &PluginInfo{Name: "stub"}
	m.config.Plugins = []PluginInfo{{Name: "stub"}}

	result, err := m.Execute(context.Background(), "stub", "test-action", map[string]any{"key": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Errorf("expected 'ok', got %v", result)
	}
	if !p.executed {
		t.Error("expected plugin to be executed")
	}
	if p.action != "test-action" {
		t.Errorf("expected action 'test-action', got %s", p.action)
	}
}

func TestManagerExecutePluginError(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	p := newStubPlugin("stub")
	p.execErr = os.ErrNotExist
	m.plugins["stub"] = p
	m.info["stub"] = &PluginInfo{Name: "stub"}
	m.config.Plugins = []PluginInfo{{Name: "stub"}}

	_, err := m.Execute(context.Background(), "stub", "act", nil)
	if err == nil {
		t.Fatal("expected error from plugin execution")
	}
}

func TestManagerInitializeAllSuccess(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	p := newStubPlugin("p1")
	m.plugins["p1"] = p
	m.info["p1"] = &PluginInfo{Name: "p1"}

	ctx := &PluginContext{}
	err := m.InitializeAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerInitializeAllError(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	p := &failingPlugin{}
	p.NameVal = "fail"
	m.plugins["fail"] = p
	m.info["fail"] = &PluginInfo{Name: "fail"}

	ctx := &PluginContext{}
	err := m.InitializeAll(ctx)
	if err == nil {
		t.Fatal("expected error from InitializeAll")
	}
}

func TestManagerCleanupSuccess(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	p := newStubPlugin("p1")
	m.plugins["p1"] = p

	err := m.Cleanup()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerCleanupError(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	p := newStubPlugin("p1")
	p.shutErr = os.ErrPermission
	m.plugins["p1"] = p

	err := m.Cleanup()
	if err == nil {
		t.Fatal("expected error from Cleanup")
	}
}

func TestSandboxValidatePathInsideAllowed(t *testing.T) {
	dir := t.TempDir()
	s := NewSandbox(SandboxConfig{
		AllowedDirs: []string{dir},
		MaxCalls:    100,
	})
	err := s.ValidatePath(filepath.Join(dir, "plugin.so"))
	if err != nil {
		t.Fatalf("expected no error for path inside allowed dir, got %v", err)
	}
}

func TestSandboxValidatePathExactMatchExt(t *testing.T) {
	dir := t.TempDir()
	s := NewSandbox(SandboxConfig{
		AllowedDirs: []string{dir},
		MaxCalls:    100,
	})
	err := s.ValidatePath(dir)
	if err != nil {
		t.Fatalf("expected no error for exact dir match, got %v", err)
	}
}

func TestSandboxValidatePathOutsideAllowed(t *testing.T) {
	dir := t.TempDir()
	otherDir := t.TempDir()
	s := NewSandbox(SandboxConfig{
		AllowedDirs: []string{dir},
		MaxCalls:    100,
	})
	err := s.ValidatePath(filepath.Join(otherDir, "plugin.so"))
	if err == nil {
		t.Fatal("expected error for path outside allowed dir")
	}
}

func TestSandboxCheckRateLimit(t *testing.T) {
	s := NewSandbox(SandboxConfig{MaxCalls: 2})
	if err := s.CheckRateLimit("p1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.CheckRateLimit("p1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.CheckRateLimit("p1"); err == nil {
		t.Fatal("expected error after exceeding rate limit")
	}
}

func TestSandboxCheckRateLimitDifferentPlugin(t *testing.T) {
	s := NewSandbox(SandboxConfig{MaxCalls: 1})
	s.CheckRateLimit("p1")
	if err := s.CheckRateLimit("p2"); err != nil {
		t.Fatalf("expected no error for different plugin, got %v", err)
	}
}

func TestSandboxExecuteWithTimeoutCtx(t *testing.T) {
	s := NewSandbox(SandboxConfig{MaxCalls: 100})
	result, err := s.ExecuteWithTimeout(context.Background(), func() (any, error) {
		return "done", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "done" {
		t.Errorf("expected 'done', got %v", result)
	}
}

func TestSandboxExecuteWithTimeoutErrorCtx(t *testing.T) {
	s := NewSandbox(SandboxConfig{MaxCalls: 100})
	_, err := s.ExecuteWithTimeout(context.Background(), func() (any, error) {
		return nil, os.ErrNotExist
	})
	if err == nil {
		t.Fatal("expected error from function")
	}
}

func TestManagerLoadConfig(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	m.config.Plugins = []PluginInfo{
		{Name: "p1", Version: "1.0.0", Enabled: true},
	}
	err := m.SaveConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m2 := NewManager(dir)
	err = m2.LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m2.config.Plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(m2.config.Plugins))
	}
}

func TestManagerLoadConfigBadJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "plugins.json"), []byte("bad json"), 0o600)
	m := NewManager(dir)
	err := m.LoadConfig()
	if err == nil {
		t.Fatal("expected error from bad JSON")
	}
}

func TestManagerSaveConfigMkdirError(t *testing.T) {
	m := NewManager("/proc/fake/path/that/cant/be/created")
	err := m.SaveConfig()
	if err == nil {
		t.Fatal("expected error from save with bad dir")
	}
}

func TestManagerInstallBadPath(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	_, err := m.Install(filepath.Join(dir, "nonexistent.so"))
	if err == nil {
		t.Fatal("expected error from nonexistent plugin file")
	}
}

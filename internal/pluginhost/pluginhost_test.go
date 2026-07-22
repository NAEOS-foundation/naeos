package pluginhost

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --- Test helpers ---

type stubPlugin struct {
	BasePlugin
	initErr  error
	execErr  error
	shutErr  error
	executed bool
	action   string
	params   map[string]any
}

func newStubPlugin(name string) *stubPlugin {
	return &stubPlugin{
		BasePlugin: BasePlugin{
			NameVal:        name,
			VersionVal:     "1.0.0",
			DescriptionVal: "test plugin",
		},
	}
}

func (p *stubPlugin) Execute(action string, params map[string]any) (any, error) {
	p.executed = true
	p.action = action
	p.params = params
	if p.execErr != nil {
		return nil, p.execErr
	}
	return "ok", nil
}

func (p *stubPlugin) Initialize(_ *PluginContext) error {
	return p.initErr
}

func (p *stubPlugin) Shutdown() error {
	return p.shutErr
}

type failingPlugin struct {
	BasePlugin
}

func (p *failingPlugin) Initialize(_ *PluginContext) error {
	return os.ErrPermission
}

func (p *failingPlugin) Execute(_ string, _ map[string]any) (any, error) {
	return nil, os.ErrPermission
}

func (p *failingPlugin) Shutdown() error {
	return os.ErrPermission
}

func tmpDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

// --- Manager Tests ---

func TestNewManager(t *testing.T) {
	dir := tmpDir(t)
	m := NewManager(dir)
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
	if m.pluginDir != dir {
		t.Errorf("expected dir %s, got %s", dir, m.pluginDir)
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	dir := tmpDir(t)
	m := NewManager(filepath.Join(dir, "nonexistent"))
	if err := m.LoadConfig(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.config.Plugins) != 0 {
		t.Error("expected empty config")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	dir := tmpDir(t)
	if err := os.WriteFile(filepath.Join(dir, "plugins.json"), []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	m := NewManager(dir)
	if err := m.LoadConfig(); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := tmpDir(t)
	m := NewManager(dir)

	m.config.Plugins = []PluginInfo{
		{Name: "test", Version: "1.0.0", Enabled: true},
	}
	if err := m.SaveConfig(); err != nil {
		t.Fatalf("save error: %v", err)
	}

	m2 := NewManager(dir)
	if err := m2.LoadConfig(); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(m2.config.Plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(m2.config.Plugins))
	}
	if m2.config.Plugins[0].Name != "test" {
		t.Errorf("expected name 'test', got %s", m2.config.Plugins[0].Name)
	}
}

func TestSaveAndLoadConfigWithSandbox(t *testing.T) {
	dir := tmpDir(t)
	m := NewManager(dir)

	m.config = PluginConfig{
		Plugins: []PluginInfo{{Name: "p1", Version: "1.0.0", Enabled: true}},
		Sandbox: SandboxConfig{
			AllowedDirs: []string{"/tmp"},
			ExecTimeout: 10 * time.Second,
			MaxCalls:    500,
		},
	}
	if err := m.SaveConfig(); err != nil {
		t.Fatal(err)
	}

	m2 := NewManager(dir)
	if err := m2.LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if m2.config.Sandbox.MaxCalls != 500 {
		t.Errorf("expected MaxCalls 500, got %d", m2.config.Sandbox.MaxCalls)
	}
}

func TestList(t *testing.T) {
	m := NewManager(tmpDir(t))
	m.config.Plugins = []PluginInfo{
		{Name: "a"}, {Name: "b"},
	}
	list := m.List()
	if len(list) != 2 {
		t.Errorf("expected 2, got %d", len(list))
	}
}

func TestGetNotLoaded(t *testing.T) {
	m := NewManager(tmpDir(t))
	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

func TestGetInfo(t *testing.T) {
	m := NewManager(tmpDir(t))
	m.config.Plugins = []PluginInfo{{Name: "test", Version: "2.0.0"}}
	info, ok := m.GetInfo("test")
	if !ok {
		t.Fatal("expected found")
	}
	if info.Version != "2.0.0" {
		t.Errorf("expected 2.0.0, got %s", info.Version)
	}
}

func TestGetInfoNotFound(t *testing.T) {
	m := NewManager(tmpDir(t))
	_, ok := m.GetInfo("nope")
	if ok {
		t.Error("expected not found")
	}
}

func TestRegisterAndUnregister(t *testing.T) {
	m := NewManager(tmpDir(t))
	p := newStubPlugin("my-plugin")

	if err := m.Register(p); err != nil {
		t.Fatalf("register error: %v", err)
	}

	got, ok := m.Get("my-plugin")
	if !ok {
		t.Fatal("expected found")
	}
	if got.Name() != "my-plugin" {
		t.Errorf("expected my-plugin, got %s", got.Name())
	}

	if err := m.Unregister("my-plugin"); err != nil {
		t.Fatalf("unregister error: %v", err)
	}
	_, ok = m.Get("my-plugin")
	if ok {
		t.Error("expected not found after unregister")
	}
}

func TestRegisterDuplicate(t *testing.T) {
	m := NewManager(tmpDir(t))
	p := newStubPlugin("dup")
	if err := m.Register(p); err != nil {
		t.Fatal(err)
	}
	if err := m.Register(p); err == nil {
		t.Error("expected error for duplicate")
	}
}

func TestUnregisterNotFound(t *testing.T) {
	m := NewManager(tmpDir(t))
	if err := m.Unregister("nope"); err == nil {
		t.Error("expected error for not found")
	}
}

func TestEnableDisable(t *testing.T) {
	dir := tmpDir(t)
	m := NewManager(dir)
	m.config.Plugins = []PluginInfo{
		{Name: "p1", Enabled: true},
	}
	if err := m.SaveConfig(); err != nil {
		t.Fatal(err)
	}

	if err := m.Disable("p1"); err != nil {
		t.Fatal(err)
	}
	m2 := NewManager(dir)
	if err := m2.LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if m2.config.Plugins[0].Enabled {
		t.Error("expected disabled")
	}

	if err := m2.Enable("p1"); err != nil {
		t.Fatal(err)
	}
	if err := m2.SaveConfig(); err != nil {
		t.Fatal(err)
	}
	m3 := NewManager(dir)
	if err := m3.LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if !m3.config.Plugins[0].Enabled {
		t.Error("expected enabled")
	}
}

func TestEnableNotFound(t *testing.T) {
	m := NewManager(tmpDir(t))
	if err := m.Enable("nope"); err == nil {
		t.Error("expected error")
	}
}

func TestDisableNotFound(t *testing.T) {
	m := NewManager(tmpDir(t))
	if err := m.Disable("nope"); err == nil {
		t.Error("expected error")
	}
}

func TestUninstallNotFound(t *testing.T) {
	m := NewManager(tmpDir(t))
	if err := m.Uninstall("nope"); err == nil {
		t.Error("expected error")
	}
}

func TestUninstallFound(t *testing.T) {
	dir := tmpDir(t)
	m := NewManager(dir)
	m.config.Plugins = []PluginInfo{{Name: "p1", Enabled: true}}
	if err := m.SaveConfig(); err != nil {
		t.Fatal(err)
	}
	p := newStubPlugin("p1")
	if err := m.Register(p); err != nil {
		t.Fatal(err)
	}

	if err := m.Uninstall("p1"); err != nil {
		t.Fatal(err)
	}
	_, ok := m.Get("p1")
	if ok {
		t.Error("expected not found")
	}
}

func TestExecuteNotLoaded(t *testing.T) {
	m := NewManager(tmpDir(t))
	_, err := m.Execute(context.Background(), "nope", "act", nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestExecuteSuccess(t *testing.T) {
	m := NewManager(tmpDir(t))
	p := newStubPlugin("my-plugin")
	if err := m.Register(p); err != nil {
		t.Fatal(err)
	}

	result, err := m.Execute(context.Background(), "my-plugin", "run", map[string]any{"key": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Errorf("expected 'ok', got %v", result)
	}
	if !p.executed {
		t.Error("expected plugin to be executed")
	}
	if p.action != "run" {
		t.Errorf("expected action 'run', got %s", p.action)
	}
	if p.params["key"] != "val" {
		t.Errorf("expected param key=val, got %v", p.params)
	}
}

func TestExecuteRateLimit(t *testing.T) {
	m := NewManager(tmpDir(t))
	m.sandbox = NewSandbox(SandboxConfig{MaxCalls: 2})
	p := newStubPlugin("rl")
	if err := m.Register(p); err != nil {
		t.Fatal(err)
	}

	if _, err := m.Execute(context.Background(), "rl", "a", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Execute(context.Background(), "rl", "a", nil); err != nil {
		t.Fatal(err)
	}
	_, err := m.Execute(context.Background(), "rl", "a", nil)
	if err == nil {
		t.Error("expected rate limit error")
	}
}

func TestExecuteTimeout(t *testing.T) {
	m := NewManager(tmpDir(t))
	m.sandbox = NewSandbox(SandboxConfig{ExecTimeout: 50 * time.Millisecond})

	slow := &slowPlugin{
		BasePlugin: BasePlugin{NameVal: "slow", VersionVal: "1.0.0"},
		delay:      200 * time.Millisecond,
	}
	if err := m.Register(slow); err != nil {
		t.Fatal(err)
	}

	_, err := m.Execute(context.Background(), "slow", "a", nil)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestExecuteContextCancel(t *testing.T) {
	m := NewManager(tmpDir(t))
	m.sandbox = NewSandbox(SandboxConfig{ExecTimeout: 5 * time.Second})

	slow := &slowPlugin{
		BasePlugin: BasePlugin{NameVal: "slow", VersionVal: "1.0.0"},
		delay:      5 * time.Second,
	}
	if err := m.Register(slow); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := m.Execute(ctx, "slow", "a", nil)
	if err == nil {
		t.Error("expected context error")
	}
}

func TestExecutePluginError(t *testing.T) {
	m := NewManager(tmpDir(t))
	fp := &failingPlugin{
		BasePlugin: BasePlugin{NameVal: "fail", VersionVal: "1.0.0"},
	}
	if err := m.Register(fp); err != nil {
		t.Fatal(err)
	}

	_, err := m.Execute(context.Background(), "fail", "a", nil)
	if err == nil {
		t.Error("expected plugin error")
	}
}

func TestInitializeAll(t *testing.T) {
	m := NewManager(tmpDir(t))
	p1 := newStubPlugin("p1")
	p2 := newStubPlugin("p2")
	if err := m.Register(p1); err != nil {
		t.Fatal(err)
	}
	if err := m.Register(p2); err != nil {
		t.Fatal(err)
	}

	ctx := &PluginContext{
		ConfigDir: "/tmp",
		OutputDir: "/tmp/out",
		Logger:    NewSimpleLogger("test"),
	}
	if err := m.InitializeAll(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitializeAllError(t *testing.T) {
	m := NewManager(tmpDir(t))
	fp := &failingPlugin{
		BasePlugin: BasePlugin{NameVal: "fail", VersionVal: "1.0.0"},
	}
	if err := m.Register(fp); err != nil {
		t.Fatal(err)
	}

	ctx := &PluginContext{}
	if err := m.InitializeAll(ctx); err == nil {
		t.Error("expected error")
	}
}

func TestShutdownAll(t *testing.T) {
	m := NewManager(tmpDir(t))
	p := newStubPlugin("p1")
	if err := m.Register(p); err != nil {
		t.Fatal(err)
	}
	if err := m.ShutdownAll(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShutdownAllError(t *testing.T) {
	m := NewManager(tmpDir(t))
	fp := &failingPlugin{
		BasePlugin: BasePlugin{NameVal: "fail", VersionVal: "1.0.0"},
	}
	if err := m.Register(fp); err != nil {
		t.Fatal(err)
	}
	if err := m.ShutdownAll(); err == nil {
		t.Error("expected error")
	}
}

func TestCleanupEmpty(t *testing.T) {
	m := NewManager(tmpDir(t))
	if err := m.Cleanup(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCleanupError(t *testing.T) {
	m := NewManager(tmpDir(t))
	fp := &failingPlugin{
		BasePlugin: BasePlugin{NameVal: "fail", VersionVal: "1.0.0"},
	}
	if err := m.Register(fp); err != nil {
		t.Fatal(err)
	}
	if err := m.Cleanup(); err == nil {
		t.Error("expected error")
	}
}

func TestLoadAllNoPlugins(t *testing.T) {
	m := NewManager(tmpDir(t))
	if err := m.LoadAll(&PluginContext{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadAllDisabledPlugin(t *testing.T) {
	dir := tmpDir(t)
	m := NewManager(dir)
	m.config.Plugins = []PluginInfo{
		{Name: "disabled", Enabled: false, Path: "/fake/path.so"},
	}
	if err := m.LoadAll(&PluginContext{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadAllEmptyPath(t *testing.T) {
	dir := tmpDir(t)
	m := NewManager(dir)
	m.config.Plugins = []PluginInfo{
		{Name: "no-path", Enabled: true, Path: ""},
	}
	if err := m.LoadAll(&PluginContext{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Sandbox Tests ---

func TestSandboxDefaults(t *testing.T) {
	s := NewSandbox(SandboxConfig{})
	if s.config.ExecTimeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %s", s.config.ExecTimeout)
	}
	if s.config.MaxCalls != 1000 {
		t.Errorf("expected 1000 max calls, got %d", s.config.MaxCalls)
	}
}

func TestSandboxValidatePathNoRestrictions(t *testing.T) {
	s := NewSandbox(SandboxConfig{})
	if err := s.ValidatePath("/any/path"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSandboxValidatePathAllowed(t *testing.T) {
	dir := t.TempDir()
	s := NewSandbox(SandboxConfig{AllowedDirs: []string{dir}})
	if err := s.ValidatePath(filepath.Join(dir, "file.so")); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSandboxValidatePathDenied(t *testing.T) {
	dir := t.TempDir()
	s := NewSandbox(SandboxConfig{AllowedDirs: []string{dir}})
	if err := s.ValidatePath("/etc/passwd"); err == nil {
		t.Error("expected error for path outside allowed dirs")
	}
}

func TestSandboxValidatePathExactMatch(t *testing.T) {
	dir := t.TempDir()
	s := NewSandbox(SandboxConfig{AllowedDirs: []string{dir}})
	if err := s.ValidatePath(dir); err != nil {
		t.Errorf("unexpected error for exact match: %v", err)
	}
}

func TestSandboxRateLimit(t *testing.T) {
	s := NewSandbox(SandboxConfig{MaxCalls: 2})
	if err := s.CheckRateLimit("p1"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckRateLimit("p1"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckRateLimit("p1"); err == nil {
		t.Error("expected rate limit error")
	}
}

func TestSandboxRateLimitSeparatePlugins(t *testing.T) {
	s := NewSandbox(SandboxConfig{MaxCalls: 1})
	if err := s.CheckRateLimit("p1"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckRateLimit("p2"); err != nil {
		t.Fatal(err)
	}
}

func TestSandboxExecuteWithTimeout(t *testing.T) {
	s := NewSandbox(SandboxConfig{ExecTimeout: 1 * time.Second})
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

func TestSandboxExecuteWithTimeoutError(t *testing.T) {
	s := NewSandbox(SandboxConfig{ExecTimeout: 1 * time.Second})
	_, err := s.ExecuteWithTimeout(context.Background(), func() (any, error) {
		return nil, os.ErrPermission
	})
	if err == nil {
		t.Error("expected error")
	}
}

func TestSandboxExecuteWithTimeoutCancellation(t *testing.T) {
	s := NewSandbox(SandboxConfig{ExecTimeout: 5 * time.Second})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := s.ExecuteWithTimeout(ctx, func() (any, error) {
		return "done", nil
	})
	if err == nil {
		t.Error("expected cancellation error")
	}
}

// --- Logger Tests ---

func TestSimpleLogger(t *testing.T) {
	l := NewSimpleLogger("test")
	if l == nil || l.logger == nil {
		t.Error("expected non-nil logger")
	}
	l.Info("hello %s", "world")
	l.Warn("warn %d", 42)
	l.Error("err")
	l.Debug("debug %v", true)
}

func TestSimpleMetrics(t *testing.T) {
	m := NewSimpleMetrics()
	m.CounterInc("c", nil)
	m.GaugeSet("g", 1.0, nil)
	m.HistogramObserve("h", 0.5, nil)
}

func TestSimpleEventEmitter(t *testing.T) {
	e := NewSimpleEventEmitter()
	received := false
	e.On("test", func(data any) {
		received = true
	})
	e.Emit("test", nil)
	if !received {
		t.Error("expected handler to be called")
	}
}

func TestSimpleEventEmitterNoHandler(t *testing.T) {
	e := NewSimpleEventEmitter()
	e.Emit("nonexistent", nil)
}

// --- BasePlugin Tests ---

func TestBasePlugin(t *testing.T) {
	b := &BasePlugin{NameVal: "n", VersionVal: "v", DescriptionVal: "d"}
	if b.Name() != "n" {
		t.Error("wrong name")
	}
	if b.Version() != "v" {
		t.Error("wrong version")
	}
	if b.Description() != "d" {
		t.Error("wrong description")
	}
	if err := b.Initialize(nil); err != nil {
		t.Error("expected nil error")
	}
	if err := b.Shutdown(); err != nil {
		t.Error("expected nil error")
	}
}

// --- State Tracking Tests ---

func TestStateTrackingOnExecute(t *testing.T) {
	m := NewManager(tmpDir(t))
	p := newStubPlugin("st")
	if err := m.Register(p); err != nil {
		t.Fatal(err)
	}

	m.mu.RLock()
	info, ok := m.info["st"]
	m.mu.RUnlock()
	if !ok {
		t.Fatal("expected info")
	}
	if info.State != StateCreated {
		t.Errorf("expected state created, got %s", info.State)
	}

	_, _ = m.Execute(context.Background(), "st", "act", nil)

	m.mu.RLock()
	if info.State != StateInitialized {
		t.Errorf("expected state initialized after execute, got %s", info.State)
	}
	m.mu.RUnlock()
}

func TestStateTrackingOnError(t *testing.T) {
	m := NewManager(tmpDir(t))
	fp := &failingPlugin{
		BasePlugin: BasePlugin{NameVal: "fail", VersionVal: "1.0.0"},
	}
	if err := m.Register(fp); err != nil {
		t.Fatal(err)
	}

	_, _ = m.Execute(context.Background(), "fail", "act", nil)

	m.mu.RLock()
	info, ok := m.info["fail"]
	m.mu.RUnlock()
	if !ok {
		t.Fatal("expected info")
	}
	if info.State != StateError {
		t.Errorf("expected state error, got %s", info.State)
	}
	if info.Error == nil {
		t.Error("expected error to be set")
	}
}

func TestStateTrackingOnInitializeAll(t *testing.T) {
	m := NewManager(tmpDir(t))
	p := newStubPlugin("p1")
	if err := m.Register(p); err != nil {
		t.Fatal(err)
	}

	if err := m.InitializeAll(&PluginContext{}); err != nil {
		t.Fatal(err)
	}

	m.mu.RLock()
	info, ok := m.info["p1"]
	m.mu.RUnlock()
	if !ok {
		t.Fatal("expected info")
	}
	if info.State != StateInitialized {
		t.Errorf("expected state initialized, got %s", info.State)
	}
	if info.StartedAt.IsZero() {
		t.Error("expected StartedAt to be set")
	}
}

// --- slowPlugin for timeout tests ---

type slowPlugin struct {
	BasePlugin
	delay time.Duration
}

func (p *slowPlugin) Execute(_ string, _ map[string]any) (any, error) {
	time.Sleep(p.delay)
	return "done", nil
}

// --- Config Persistence Tests ---

func TestConfigPersistenceRoundTrip(t *testing.T) {
	dir := tmpDir(t)
	m := NewManager(dir)
	m.config = PluginConfig{
		Plugins: []PluginInfo{
			{Name: "a", Version: "1.0.0", Enabled: true, Path: "/a.so"},
			{Name: "b", Version: "2.0.0", Enabled: false},
		},
		Sandbox: SandboxConfig{
			AllowedDirs: []string{"/tmp", "/var"},
			ExecTimeout: 15 * time.Second,
			MaxCalls:    500,
		},
	}
	if err := m.SaveConfig(); err != nil {
		t.Fatal(err)
	}

	// Read raw JSON to verify
	data, err := os.ReadFile(filepath.Join(dir, "plugins.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw PluginConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw.Plugins) != 2 {
		t.Errorf("expected 2 plugins in JSON, got %d", len(raw.Plugins))
	}
	if raw.Sandbox.MaxCalls != 500 {
		t.Errorf("expected MaxCalls 500 in JSON, got %d", raw.Sandbox.MaxCalls)
	}

	// Load in new manager
	m2 := NewManager(dir)
	if err := m2.LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if m2.config.Plugins[0].Name != "a" || m2.config.Plugins[1].Name != "b" {
		t.Error("plugin names don't match")
	}
	if m2.config.Sandbox.ExecTimeout != 15*time.Second {
		t.Error("sandbox config doesn't match")
	}
}

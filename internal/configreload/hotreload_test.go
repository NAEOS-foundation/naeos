package configreload

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHotReloaderStartStop(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("key: value1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := New(cfgPath)
	if err := cfg.Load(); err != nil {
		t.Fatal(err)
	}

	hr := NewHotReloader(cfg)
	if err := hr.Start(); err != nil {
		t.Fatal(err)
	}
	defer hr.Stop()

	if !hr.IsRunning() {
		t.Error("expected hot reloader to be running")
	}

	hr.Stop()
	if hr.IsRunning() {
		t.Error("expected hot reloader to be stopped")
	}
}

func TestHotReloaderDetectsChange(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("key: value1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := New(cfgPath)
	if err := cfg.Load(); err != nil {
		t.Fatal(err)
	}

	reloaded := make(chan struct{}, 1)
	cfg.OnChange(func(old, new map[string]interface{}) {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})

	hr := NewHotReloader(cfg)
	if err := hr.Start(); err != nil {
		t.Fatal(err)
	}
	defer hr.Stop()

	time.Sleep(100 * time.Millisecond)

	if err := os.WriteFile(cfgPath, []byte("key: value2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-reloaded:
	case <-time.After(2 * time.Second):
		t.Error("config was not reloaded after file change")
	}

	val, ok := cfg.Get("key")
	if !ok || val != "value2" {
		t.Errorf("expected key=value2 after reload, got %v", val)
	}
}

func TestHotReloaderIgnoresUnrelatedFiles(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("key: value1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := New(cfgPath)
	if err := cfg.Load(); err != nil {
		t.Fatal(err)
	}

	reloaded := make(chan struct{}, 1)
	cfg.OnChange(func(old, new map[string]interface{}) {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})

	hr := NewHotReloader(cfg)
	if err := hr.Start(); err != nil {
		t.Fatal(err)
	}
	defer hr.Stop()

	time.Sleep(100 * time.Millisecond)

	otherFile := filepath.Join(dir, "other.txt")
	if err := os.WriteFile(otherFile, []byte("ignore me\n"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-reloaded:
		t.Error("config should not reload for unrelated file changes")
	case <-time.After(500 * time.Millisecond):
		// expected
	}
}

func TestConfigLoadYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := "name: test-project\nport: 8080\ndebug: true\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := New(cfgPath)
	if err := cfg.Load(); err != nil {
		t.Fatal(err)
	}

	name, _ := cfg.Get("name")
	if name != "test-project" {
		t.Errorf("expected name=test-project, got %v", name)
	}

	port := cfg.GetInt("port", 0)
	if port != 8080 {
		t.Errorf("expected port=8080, got %d", port)
	}

	debug := cfg.GetBool("debug", false)
	if !debug {
		t.Error("expected debug=true")
	}
}

func TestConfigLoadJSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	content := `{"name": "test", "count": 42}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := New(cfgPath)
	if err := cfg.Load(); err != nil {
		t.Fatal(err)
	}

	name, _ := cfg.Get("name")
	if name != "test" {
		t.Errorf("expected name=test, got %v", name)
	}
}

func TestConfigDiff(t *testing.T) {
	old := map[string]interface{}{"a": 1, "b": 2, "c": 3}
	new := map[string]interface{}{"a": 1, "b": 99, "d": 4}

	diff := Diff(old, new)

	if len(diff.Added) != 1 || diff.Added["d"] != 4 {
		t.Errorf("expected 1 added key (d=4), got %v", diff.Added)
	}
	if len(diff.Removed) != 1 {
		t.Errorf("expected 1 removed key (c), got %v", diff.Removed)
	}
	if len(diff.Modified) != 1 {
		t.Errorf("expected 1 modified key (b), got %v", diff.Modified)
	}
}

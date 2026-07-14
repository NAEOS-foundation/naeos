package configreload

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := New("config.json")
	if c == nil {
		t.Fatal("expected config to be created")
	}
}

func TestSetAndGet(t *testing.T) {
	c := New("config.json")

	c.Set("key1", "value1")
	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}
}

func TestGetDefault(t *testing.T) {
	c := New("config.json")

	val := c.GetString("nonexistent", "default")
	if val != "default" {
		t.Errorf("expected 'default', got %s", val)
	}

	c.Set("number", 42)
	num := c.GetInt("number", 0)
	if num != 42 {
		t.Errorf("expected 42, got %d", num)
	}

	c.Set("flag", true)
	flag := c.GetBool("flag", false)
	if !flag {
		t.Error("expected true")
	}
}

func TestKeys(t *testing.T) {
	c := New("config.json")

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	keys := c.Keys()
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}

func TestVersion(t *testing.T) {
	c := New("config.json")

	if c.Version() != 0 {
		t.Errorf("expected version 0, got %d", c.Version())
	}

	c.Set("key", "value")
	if c.Version() != 1 {
		t.Errorf("expected version 1, got %d", c.Version())
	}

	c.Set("key2", "value2")
	if c.Version() != 2 {
		t.Errorf("expected version 2, got %d", c.Version())
	}
}

func TestOnChange(t *testing.T) {
	c := New("config.json")

	var called bool
	c.OnChange(func(old, new map[string]any) {
		called = true
	})

	c.Set("key", "value")

	if !called {
		t.Error("expected watcher to be called")
	}
}

func TestSnapshot(t *testing.T) {
	c := New("config.json")

	c.Set("a", 1)
	c.Set("b", 2)

	snap := c.Snapshot()
	if snap["a"] != 1 || snap["b"] != 2 {
		t.Error("expected snapshot to contain values")
	}

	// Modify original
	c.Set("a", 100)
	if snap["a"] != 1 {
		t.Error("expected snapshot to be independent")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	os.WriteFile(configPath, []byte(`{"name":"test","port":8080}`), 0644)

	c := New(configPath)
	err := c.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	name, _ := c.Get("name")
	if name != "test" {
		t.Errorf("expected 'test', got %v", name)
	}
}

func TestSaveToFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	c := New(configPath)
	c.Set("name", "test")
	c.Set("port", 8080)

	err := c.Save()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file exists
	_, err = os.Stat(configPath)
	if err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}

func TestSetAll(t *testing.T) {
	c := New("config.json")

	c.Set("old", "value")

	newData := map[string]any{
		"new": "value",
	}

	c.SetAll(newData)

	keys := c.Keys()
	if len(keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(keys))
	}

	_, ok := c.Get("old")
	if ok {
		t.Error("expected 'old' to be removed")
	}
}

func TestFileWatcher(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	os.WriteFile(configPath, []byte(`{"key":"initial"}`), 0644)

	c := New(configPath)
	c.Load()

	fw := NewFileWatcher(c, 10*time.Millisecond)
	fw.Start()

	if !fw.IsRunning() {
		t.Error("expected watcher to be running")
	}

	fw.Stop()

	if fw.IsRunning() {
		t.Error("expected watcher to be stopped")
	}
}

func TestManager(t *testing.T) {
	m := NewManager()

	c1 := New("config1.json")
	c2 := New("config2.json")

	m.Add("app1", c1)
	m.Add("app2", c2)

	list := m.List()
	if len(list) != 2 {
		t.Errorf("expected 2 configs, got %d", len(list))
	}

	got, ok := m.Get("app1")
	if !ok {
		t.Fatal("expected config to be found")
	}
	if got != c1 {
		t.Error("expected same config instance")
	}

	m.Remove("app1")
	_, ok = m.Get("app1")
	if ok {
		t.Error("expected config to be removed")
	}
}

func TestManagerReloadAll(t *testing.T) {
	dir := t.TempDir()
	configPath1 := filepath.Join(dir, "config1.json")
	configPath2 := filepath.Join(dir, "config2.json")

	os.WriteFile(configPath1, []byte(`{"key1":"value1"}`), 0644)
	os.WriteFile(configPath2, []byte(`{"key2":"value2"}`), 0644)

	m := NewManager()
	m.Add("app1", New(configPath1))
	m.Add("app2", New(configPath2))

	err := m.ReloadAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDiff(t *testing.T) {
	old := map[string]any{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	new := map[string]any{
		"b": 20,
		"c": 3,
		"d": 4,
	}

	diff := Diff(old, new)

	if len(diff.Added) != 1 || diff.Added["d"] != 4 {
		t.Errorf("expected 1 added, got %d", len(diff.Added))
	}

	if len(diff.Removed) != 1 || diff.Removed["a"] != 1 {
		t.Errorf("expected 1 removed, got %d", len(diff.Removed))
	}

	if len(diff.Modified) != 1 {
		t.Errorf("expected 1 modified, got %d", len(diff.Modified))
	}
}

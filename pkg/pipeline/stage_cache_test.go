package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStageCache(t *testing.T) {
	dir := t.TempDir()
	sc := NewStageCache(filepath.Join(dir, "cache"))
	if sc == nil {
		t.Fatal("expected non-nil StageCache")
	}
}

func TestStageCacheSetGet(t *testing.T) {
	dir := t.TempDir()
	sc := NewStageCache(filepath.Join(dir, "cache"))
	sc.SetVersion("parse", "v1")

	data := []byte("input-spec")
	output := []byte("parsed-result")

	sc.Set("parse", data, output)

	got, ok := sc.Get("parse", data)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if string(got) != string(output) {
		t.Errorf("expected %q, got %q", string(output), string(got))
	}
}

func TestStageCacheMiss(t *testing.T) {
	dir := t.TempDir()
	sc := NewStageCache(filepath.Join(dir, "cache"))
	sc.SetVersion("parse", "v1")

	_, ok := sc.Get("parse", []byte("never-cached"))
	if ok {
		t.Error("expected cache miss")
	}
}

func TestStageCacheVersionInvalidation(t *testing.T) {
	dir := t.TempDir()
	sc := NewStageCache(filepath.Join(dir, "cache"))

	sc.SetVersion("parse", "v1")
	sc.Set("parse", []byte("input"), []byte("output-v1"))

	sc.SetVersion("parse", "v2")
	_, ok := sc.Get("parse", []byte("input"))
	if ok {
		t.Error("expected cache miss after version bump")
	}
}

func TestStageCacheInvalidateStage(t *testing.T) {
	dir := t.TempDir()
	sc := NewStageCache(filepath.Join(dir, "cache"))
	sc.SetVersion("parse", "v1")

	sc.Set("parse", []byte("in1"), []byte("out1"))
	sc.Set("generate", []byte("in2"), []byte("out2"))

	sc.Invalidate("parse")

	_, ok := sc.Get("parse", []byte("in1"))
	if ok {
		t.Error("expected cache miss after invalidation")
	}
	_, ok = sc.Get("generate", []byte("in2"))
	if !ok {
		t.Error("expected cache hit for non-invalidated stage")
	}
}

func TestStageCacheClear(t *testing.T) {
	dir := t.TempDir()
	sc := NewStageCache(filepath.Join(dir, "cache"))
	sc.SetVersion("parse", "v1")

	sc.Set("parse", []byte("in1"), []byte("out1"))
	sc.Set("generate", []byte("in2"), []byte("out2"))

	sc.Clear()

	_, ok := sc.Get("parse", []byte("in1"))
	if ok {
		t.Error("expected miss after clear")
	}
	if len(sc.entries) != 0 {
		t.Error("expected empty entries after clear")
	}
}

func TestStageCacheDiskPersistence(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")

	sc := NewStageCache(cacheDir)
	sc.SetVersion("validate", "v1")
	sc.Set("validate", []byte("spec"), []byte("valid"))

	sc2 := NewStageCache(cacheDir)
	sc2.SetVersion("validate", "v1")

	got, ok := sc2.Get("validate", []byte("spec"))
	if !ok {
		t.Error("expected cache hit from disk")
	}
	if string(got) != "valid" {
		t.Errorf("expected 'valid', got %q", string(got))
	}
}

func TestStageCacheVersion(t *testing.T) {
	sc := NewStageCache("")
	sc.SetVersion("parse", "v3")
	if v := sc.Version("parse"); v != "v3" {
		t.Errorf("expected 'v3', got %q", v)
	}
	if v := sc.Version("unknown"); v != "" {
		t.Errorf("expected empty, got %q", v)
	}
}

func TestStageCacheStats(t *testing.T) {
	dir := t.TempDir()
	sc := NewStageCache(filepath.Join(dir, "cache"))
	sc.SetVersion("a", "v1")
	sc.SetVersion("b", "v1")

	sc.Set("a", []byte("x"), []byte("1"))
	sc.Set("b", []byte("y"), []byte("2"))
	sc.Set("b", []byte("z"), []byte("3"))

	stats := sc.Stats()
	if stats["a"] != 1 {
		t.Errorf("expected 1 entry for stage 'a', got %d", stats["a"])
	}
	if stats["b"] != 2 {
		t.Errorf("expected 2 entries for stage 'b', got %d", stats["b"])
	}
}

func TestStageCacheDefaultDir(t *testing.T) {
	sc := NewStageCache("")
	defer os.RemoveAll(sc.dir)
	if sc.dir == "" {
		t.Error("expected default dir")
	}
}

func TestStageCacheSameInputOverride(t *testing.T) {
	dir := t.TempDir()
	sc := NewStageCache(filepath.Join(dir, "cache"))
	sc.SetVersion("normalize", "v1")

	data := []byte("same-input")
	sc.Set("normalize", data, []byte("first"))
	sc.Set("normalize", data, []byte("second"))

	got, ok := sc.Get("normalize", data)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if string(got) != "second" {
		t.Errorf("expected 'second', got %q", string(got))
	}
}

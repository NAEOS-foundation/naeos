package pipelinecache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func TestCacheSetGet(t *testing.T) {
	c := New(t.TempDir(), 10)

	result := &pipeline.Result{Source: "test"}
	hash := c.HashSpec("project: test")

	c.Set(hash, result)

	got, ok := c.Get(hash)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Source != "test" {
		t.Errorf("expected 'test', got %q", got.Source)
	}
}

func TestCacheMiss(t *testing.T) {
	c := New(t.TempDir(), 10)

	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected cache miss")
	}
}

func TestCacheEviction(t *testing.T) {
	c := New(t.TempDir(), 3)

	for i := 0; i < 5; i++ {
		result := &pipeline.Result{Source: "test"}
		hash := c.HashSpec("spec" + string(rune('0'+i)))
		c.Set(hash, result)
		time.Sleep(5 * time.Millisecond)
	}

	if c.Size() > 3 {
		t.Errorf("expected eviction to keep size <= 3, got %d", c.Size())
	}
}

func TestCacheInvalidate(t *testing.T) {
	c := New(t.TempDir(), 10)

	result := &pipeline.Result{Source: "test"}
	hash := c.HashSpec("project: test")
	c.Set(hash, result)

	c.Invalidate(hash)

	_, ok := c.Get(hash)
	if ok {
		t.Error("expected cache miss after invalidation")
	}
}

func TestCacheClear(t *testing.T) {
	c := New(t.TempDir(), 10)

	for i := 0; i < 5; i++ {
		result := &pipeline.Result{Source: "test"}
		hash := c.HashSpec("spec" + string(rune('0'+i)))
		c.Set(hash, result)
	}

	c.Clear()

	if c.Size() != 0 {
		t.Errorf("expected 0 after clear, got %d", c.Size())
	}
}

func TestCacheHashDeterministic(t *testing.T) {
	c := New(t.TempDir(), 10)

	h1 := c.HashSpec("project: test")
	h2 := c.HashSpec("project: test")

	if h1 != h2 {
		t.Error("expected deterministic hash")
	}
}

func TestCacheHashDifferent(t *testing.T) {
	c := New(t.TempDir(), 10)

	h1 := c.HashSpec("project: test1")
	h2 := c.HashSpec("project: test2")

	if h1 == h2 {
		t.Error("expected different hashes for different input")
	}
}

func TestCacheHitCount(t *testing.T) {
	c := New(t.TempDir(), 10)

	result := &pipeline.Result{Source: "test"}
	hash := c.HashSpec("project: test")
	c.Set(hash, result)

	for i := 0; i < 5; i++ {
		c.Get(hash)
	}

	entry := c.entries[hash]
	if entry.HitCount != 5 {
		t.Errorf("expected hit count 5, got %d", entry.HitCount)
	}
}

func TestCacheNoDir(t *testing.T) {
	c := New("", 10)

	result := &pipeline.Result{Source: "test"}
	hash := c.HashSpec("project: test")
	c.Set(hash, result)

	got, ok := c.Get(hash)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Source != "test" {
		t.Errorf("expected 'test', got %q", got.Source)
	}
}

func TestCacheTTLExpiration(t *testing.T) {
	c := New(t.TempDir(), 10)
	c.SetMaxAge(50 * time.Millisecond)

	result := &pipeline.Result{Source: "ttl-test"}
	hash := c.HashSpec("project: ttl")
	c.Set(hash, result)

	got, ok := c.Get(hash)
	if !ok {
		t.Fatal("expected cache hit before TTL")
	}
	if got.Source != "ttl-test" {
		t.Errorf("expected 'ttl-test', got %q", got.Source)
	}

	time.Sleep(100 * time.Millisecond)

	_, ok = c.Get(hash)
	if ok {
		t.Error("expected cache miss after TTL expiration")
	}

	if c.Size() != 0 {
		t.Errorf("expected 0 entries after TTL eviction, got %d", c.Size())
	}
}

func TestCacheTTLNotSet(t *testing.T) {
	c := New(t.TempDir(), 10)

	result := &pipeline.Result{Source: "no-ttl"}
	hash := c.HashSpec("project: no-ttl")
	c.Set(hash, result)

	time.Sleep(10 * time.Millisecond)

	got, ok := c.Get(hash)
	if !ok {
		t.Fatal("expected cache hit when MaxAge is zero")
	}
	if got.Source != "no-ttl" {
		t.Errorf("expected 'no-ttl', got %q", got.Source)
	}
}

func TestCacheLRUEvictionByHitCount(t *testing.T) {
	c := New(t.TempDir(), 3)

	hashLow := c.HashSpec("low-usage")
	hashMid := c.HashSpec("mid-usage")
	hashHigh := c.HashSpec("high-usage")
	hashNew := c.HashSpec("new-entry")

	c.Set(hashLow, &pipeline.Result{Source: "low"})
	time.Sleep(1 * time.Millisecond)
	c.Set(hashMid, &pipeline.Result{Source: "mid"})
	time.Sleep(1 * time.Millisecond)
	c.Set(hashHigh, &pipeline.Result{Source: "high"})

	for i := 0; i < 10; i++ {
		c.Get(hashHigh)
	}
	for i := 0; i < 5; i++ {
		c.Get(hashMid)
	}

	c.Set(hashNew, &pipeline.Result{Source: "new"})

	if c.Size() > 3 {
		t.Errorf("expected size <= 3, got %d", c.Size())
	}

	if _, ok := c.Get(hashHigh); !ok {
		t.Error("expected high-hit entry to survive eviction")
	}
	if _, ok := c.Get(hashMid); !ok {
		t.Error("expected mid-hit entry to survive eviction")
	}
}

func TestCacheSetMaxAgeZero(t *testing.T) {
	c := New(t.TempDir(), 10)
	c.SetMaxAge(50 * time.Millisecond)
	c.SetMaxAge(0)

	result := &pipeline.Result{Source: "reset"}
	hash := c.HashSpec("project: reset")
	c.Set(hash, result)

	time.Sleep(10 * time.Millisecond)

	got, ok := c.Get(hash)
	if !ok {
		t.Fatal("expected cache hit after resetting MaxAge to zero")
	}
	if got.Source != "reset" {
		t.Errorf("expected 'reset', got %q", got.Source)
	}
}

func TestCacheStats(t *testing.T) {
	c := New(t.TempDir(), 10)
	result := &pipeline.Result{Source: "test"}
	hash := c.HashSpec("project: test")
	c.Set(hash, result)

	stats := c.Stats()
	if stats.Size != 1 {
		t.Errorf("expected size 1, got %d", stats.Size)
	}
	if stats.MaxSize != 10 {
		t.Errorf("expected maxsize 10, got %d", stats.MaxSize)
	}
}

func TestCacheDefaultSize(t *testing.T) {
	c := New(t.TempDir(), 0)
	if c.maxSize != 100 {
		t.Errorf("expected default size 100, got %d", c.maxSize)
	}
}

func TestCacheModuleKey(t *testing.T) {
	t.Parallel()

	c := New("", 10)

	tests := []struct {
		name       string
		specHash   string
		moduleName string
		stage      string
		want       string
	}{
		{
			name:       "basic",
			specHash:   "abc123",
			moduleName: "mymodule",
			stage:      "build",
			want:       "abc123_mymodule_build",
		},
		{
			name:       "empty stage",
			specHash:   "xyz",
			moduleName: "mod",
			stage:      "",
			want:       "xyz_mod_",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := c.ModuleKey(tt.specHash, tt.moduleName, tt.stage)
			if got != tt.want {
				t.Errorf("ModuleKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCacheGetModuleStage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(*Cache) string
		want  bool
	}{
		{
			name: "miss",
			setup: func(c *Cache) string {
				return "nonexistent"
			},
			want: false,
		},
		{
			name: "hit with data",
			setup: func(c *Cache) string {
				key := "hit_key"
				c.SetModuleStage(key, []byte(`{"result":"ok"}`))
				return key
			},
			want: true,
		},
		{
			name: "nil data",
			setup: func(c *Cache) string {
				key := "nil_key"
				c.SetModuleStage(key, nil)
				return key
			},
			want: false,
		},
		{
			name: "expired",
			setup: func(c *Cache) string {
				c.SetMaxAge(1 * time.Nanosecond)
				key := "expired_key"
				c.SetModuleStage(key, []byte(`{"result":"ok"}`))
				return key
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := New(t.TempDir(), 10)
			key := tt.setup(c)
			if tt.name == "expired" {
				time.Sleep(5 * time.Millisecond)
			}
			_, ok := c.GetModuleStage(key)
			if ok != tt.want {
				t.Errorf("GetModuleStage() ok = %v, want %v", ok, tt.want)
			}
			if tt.name == "expired" && c.Size() != 0 {
				t.Errorf("expected empty cache after expired GetModuleStage, got size %d", c.Size())
			}
		})
	}
}

func TestCacheSetModuleStage(t *testing.T) {
	t.Parallel()

	c := New(t.TempDir(), 10)
	key := "stage_key"
	data := []byte(`{"data":"value"}`)
	c.SetModuleStage(key, data)

	got, ok := c.GetModuleStage(key)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if string(got) != string(data) {
		t.Errorf("got %s, want %s", string(got), string(data))
	}
}

func TestCacheSetModuleStageEviction(t *testing.T) {
	t.Parallel()

	c := New(t.TempDir(), 2)

	c.SetModuleStage("key1", []byte(`1`))
	time.Sleep(time.Millisecond)
	c.SetModuleStage("key2", []byte(`2`))
	c.SetModuleStage("key3", []byte(`3`))

	if c.Size() > 2 {
		t.Errorf("expected size <= 2 after eviction, got %d", c.Size())
	}
}

func TestCacheInvalidateModule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		specHash   string
		moduleName string
		setup      func(*Cache)
		wantSize   int
	}{
		{
			name:       "invalidate matching entry",
			specHash:   "spec1",
			moduleName: "mod1",
			setup: func(c *Cache) {
				c.SetModuleStage("spec1_mod1_build", []byte(`{}`))
			},
			wantSize: 0,
		},
		{
			name:       "no matching entries",
			specHash:   "spec1",
			moduleName: "mod1",
			setup: func(c *Cache) {
				c.SetModuleStage("other_other_build", []byte(`{}`))
			},
			wantSize: 1,
		},
		{
			name:       "invalidate only matching modules",
			specHash:   "spec1",
			moduleName: "mod1",
			setup: func(c *Cache) {
				c.SetModuleStage("spec1_mod1_build", []byte(`{}`))
				c.SetModuleStage("spec1_mod2_build", []byte(`{}`))
			},
			wantSize: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := New(t.TempDir(), 10)
			tt.setup(c)
			c.InvalidateModule(tt.specHash, tt.moduleName)
			if c.Size() != tt.wantSize {
				t.Errorf("after InvalidateModule size = %d, want %d", c.Size(), tt.wantSize)
			}
		})
	}
}

func TestCacheUnchangedModules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		specHash      string
		setup         func(*Cache) map[string]string
		wantUnchanged int
		wantSize      int
	}{
		{
			name:     "all unchanged",
			specHash: "spec1",
			setup: func(c *Cache) map[string]string {
				c.SetModuleStage("spec1_mod1_content", []byte(`{}`))
				c.SetModuleStage("spec1_mod2_content", []byte(`{}`))
				return map[string]string{"mod1": "hash1", "mod2": "hash2"}
			},
			wantUnchanged: 2,
			wantSize:      2,
		},
		{
			name:     "some unchanged, some new",
			specHash: "spec1",
			setup: func(c *Cache) map[string]string {
				c.SetModuleStage("spec1_mod1_content", []byte(`{}`))
				return map[string]string{"mod1": "hash1", "mod2": "hash2"}
			},
			wantUnchanged: 1,
			wantSize:      2,
		},
		{
			name:     "empty module hashes",
			specHash: "spec1",
			setup: func(c *Cache) map[string]string {
				return map[string]string{}
			},
			wantUnchanged: 0,
			wantSize:      0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := New(t.TempDir(), 10)
			hashes := tt.setup(c)
			unchanged := c.UnchangedModules(tt.specHash, hashes)
			if len(unchanged) != tt.wantUnchanged {
				t.Errorf("UnchangedModules() returned %d, want %d", len(unchanged), tt.wantUnchanged)
			}
			if c.Size() != tt.wantSize {
				t.Errorf("cache size = %d, want %d", c.Size(), tt.wantSize)
			}
		})
	}
}

func TestCacheDiskPersistence(t *testing.T) {
	dir := t.TempDir()

	c1 := New(dir, 10)
	result := &pipeline.Result{Source: "disk-test"}
	hash := c1.HashSpec("persist-spec")
	c1.Set(hash, result)

	c2 := New(dir, 10)
	got, ok := c2.Get(hash)
	if !ok {
		t.Fatal("expected cache hit after reload from disk")
	}
	if got.Source != "disk-test" {
		t.Errorf("expected 'disk-test', got %q", got.Source)
	}
}

func TestCacheLoadFromDiskInvalidFile(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "valid.json"), []byte(`{"key":"valid","timestamp":"2024-01-01T00:00:00Z"}`), 0o600)
	os.WriteFile(filepath.Join(dir, "invalid.json"), []byte(`not json`), 0o600)

	c := New(dir, 10)
	if c.Size() != 1 {
		t.Errorf("expected 1 loaded entry, got %d", c.Size())
	}
}

func TestCacheStatsEntries(t *testing.T) {
	t.Parallel()

	c := New(t.TempDir(), 10)
	result := &pipeline.Result{Source: "test"}
	hash := c.HashSpec("project: test")
	c.Set(hash, result)

	stats := c.Stats()
	if stats.Size != 1 {
		t.Errorf("expected size 1, got %d", stats.Size)
	}
	if stats.MaxSize != 10 {
		t.Errorf("expected maxsize 10, got %d", stats.MaxSize)
	}
	if len(stats.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(stats.Entries))
	}
	if stats.Entries[0].Key != hash {
		t.Errorf("expected key %q, got %q", hash, stats.Entries[0].Key)
	}
}

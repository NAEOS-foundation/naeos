package pipelinecache

import (
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

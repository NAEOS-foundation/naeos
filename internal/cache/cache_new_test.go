package cache

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCleanupExpired(t *testing.T) {
	c := New(10, 50*time.Millisecond)

	c.Set("a", "1")
	c.Set("b", "2")
	c.Set("c", "3")

	time.Sleep(60 * time.Millisecond)

	removed := c.Cleanup()
	if removed != 3 {
		t.Errorf("expected 3 removed, got %d", removed)
	}

	if c.Size() != 0 {
		t.Errorf("expected 0 size, got %d", c.Size())
	}
}

func TestCleanupPartialExpired(t *testing.T) {
	c := New(10, 30*time.Millisecond)

	c.Set("a", "1")
	time.Sleep(50 * time.Millisecond)
	c.SetWithTTL("b", "2", 5*time.Second)

	removed := c.Cleanup()
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	if c.Size() != 1 {
		t.Errorf("expected 1 size, got %d", c.Size())
	}

	if _, ok := c.Get("b"); !ok {
		t.Error("expected b to still exist")
	}
}

func TestEvictionCallback(t *testing.T) {
	var evictedKeys []string
	var mu sync.Mutex

	c := New(2, time.Hour)
	c.SetEvictionCallback(func(key string, value any) {
		mu.Lock()
		evictedKeys = append(evictedKeys, key)
		mu.Unlock()
	})

	c.Set("a", "1")
	c.Set("b", "2")
	c.Set("c", "3")

	mu.Lock()
	defer mu.Unlock()
	if len(evictedKeys) != 1 {
		t.Errorf("expected 1 eviction callback, got %d", len(evictedKeys))
	}
	if evictedKeys[0] != "a" {
		t.Errorf("expected evicted key a, got %s", evictedKeys[0])
	}
}

func TestStartStopEviction(t *testing.T) {
	c := New(10, 50*time.Millisecond)
	c.StartEviction(30 * time.Millisecond)
	c.StartEviction(30 * time.Millisecond)

	c.Set("a", "1")

	deadline := time.Now().Add(2 * time.Second)
	for c.Size() != 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if c.Size() != 0 {
		t.Errorf("expected 0 size after eviction, got %d", c.Size())
	}

	c.StopEviction()
	c.StopEviction()
}

func TestStatsEvicted(t *testing.T) {
	c := New(2, time.Hour)

	c.Set("a", "1")
	c.Set("b", "2")
	c.Set("c", "3")

	stats := c.Stats()
	if stats.Evicted != 1 {
		t.Errorf("expected 1 eviction, got %d", stats.Evicted)
	}
}

func TestCacheConcurrency(t *testing.T) {
	c := New(100, time.Hour)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key"
			c.Set(key, i)
			c.Get(key)
			c.Exists(key)
			c.Keys()
		}(i)
	}
	wg.Wait()

	if c.Size() != 1 {
		t.Errorf("expected 1 size, got %d", c.Size())
	}
}

func TestDeleteExisting(t *testing.T) {
	c := New(10, time.Hour)
	c.Set("a", "1")

	if !c.Delete("a") {
		t.Error("expected true for deleting existing key")
	}

	if c.Delete("a") {
		t.Error("expected false for deleting non-existing key")
	}
}

func TestManagerCreate(t *testing.T) {
	m := NewManager()
	c1 := m.Create("test", 10, time.Hour)
	c2 := m.Create("test", 20, time.Hour)

	if c1 != c2 {
		t.Error("expected same cache instance")
	}
}

func TestManagerList(t *testing.T) {
	m := NewManager()
	m.Create("a", 10, time.Hour)
	m.Create("b", 10, time.Hour)

	names := m.List()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestHTTPCacheInvalidate(t *testing.T) {
	h := NewHTTPCache(time.Hour)
	h.Set("user:1", "data1")
	h.Set("user:2", "data2")
	h.Set("post:1", "data3")

	h.Invalidate("user:*")

	if _, ok := h.Get("user:1"); ok {
		t.Error("expected user:1 to be invalidated")
	}
	if _, ok := h.Get("user:2"); ok {
		t.Error("expected user:2 to be invalidated")
	}
	if _, ok := h.Get("post:1"); !ok {
		t.Error("expected post:1 to still exist")
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		key     string
		pattern string
		want    bool
	}{
		{"user:1", "*", true},
		{"user:1", "user:*", true},
		{"user:1", "post:*", false},
		{"user:1", "user:1", true},
		{"user:1", "user:2", false},
	}

	for _, tt := range tests {
		got := matchPattern(tt.key, tt.pattern)
		if got != tt.want {
			t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.key, tt.pattern, got, tt.want)
		}
	}
}

func TestCacheStats(t *testing.T) {
	c := New(10, time.Hour)
	c.Set("a", "1")
	c.Get("a")
	c.Get("nonexistent")

	stats := c.Stats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
}

func TestEvictionCallbackConcurrency(t *testing.T) {
	var count atomic.Int32
	c := New(2, time.Hour)
	c.SetEvictionCallback(func(key string, value any) {
		count.Add(1)
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Set(fmt.Sprintf("key-%d", i), i)
		}(i)
	}
	wg.Wait()

	if c.Size() != 2 {
		t.Errorf("expected 2 size, got %d", c.Size())
	}

	if count.Load() < 95 {
		t.Errorf("expected at least 95 evictions, got %d", count.Load())
	}
}

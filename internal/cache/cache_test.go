package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := New(100, time.Minute)
	if c == nil {
		t.Fatal("expected cache to be created")
	}
}

func TestSetAndGet(t *testing.T) {
	c := New(100, time.Minute)

	c.Set("key1", "value1")
	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}
}

func TestGetMiss(t *testing.T) {
	c := New(100, time.Minute)

	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected key not found")
	}
}

func TestDelete(t *testing.T) {
	c := New(100, time.Minute)

	c.Set("key1", "value1")
	deleted := c.Delete("key1")
	if !deleted {
		t.Error("expected key to be deleted")
	}

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key not found after delete")
	}
}

func TestDeleteNotFound(t *testing.T) {
	c := New(100, time.Minute)

	deleted := c.Delete("nonexistent")
	if deleted {
		t.Error("expected false for nonexistent key")
	}
}

func TestExists(t *testing.T) {
	c := New(100, time.Minute)

	c.Set("key1", "value1")
	if !c.Exists("key1") {
		t.Error("expected key to exist")
	}

	if c.Exists("nonexistent") {
		t.Error("expected nonexistent key")
	}
}

func TestLRUEviction(t *testing.T) {
	c := New(3, time.Minute)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.Set("d", 4) // Should evict "a"

	if c.Size() != 3 {
		t.Errorf("expected size 3, got %d", c.Size())
	}

	_, ok := c.Get("a")
	if ok {
		t.Error("expected 'a' to be evicted")
	}

	_, ok = c.Get("d")
	if !ok {
		t.Error("expected 'd' to exist")
	}
}

func TestTTLExpiration(t *testing.T) {
	c := New(100, 50*time.Millisecond)

	c.Set("key1", "value1")
	_, ok := c.Get("key1")
	if !ok {
		t.Error("expected key to exist before expiry")
	}

	time.Sleep(100 * time.Millisecond)

	_, ok = c.Get("key1")
	if ok {
		t.Error("expected key to be expired")
	}
}

func TestClear(t *testing.T) {
	c := New(100, time.Minute)

	c.Set("a", 1)
	c.Set("b", 2)

	c.Clear()

	if c.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", c.Size())
	}
}

func TestKeys(t *testing.T) {
	c := New(100, time.Minute)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	keys := c.Keys()
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}

func TestStats(t *testing.T) {
	c := New(100, time.Minute)

	c.Set("a", 1)
	c.Get("a")
	c.Get("a")
	c.Get("miss")

	stats := c.Stats()
	if stats.Hits != 2 {
		t.Errorf("expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
}

func TestCleanup(t *testing.T) {
	c := New(100, 10*time.Millisecond)

	c.Set("a", 1)
	c.Set("b", 2)

	time.Sleep(20 * time.Millisecond)

	removed := c.Cleanup()
	if removed != 2 {
		t.Errorf("expected 2 removed, got %d", removed)
	}
}

func TestSetWithTTL(t *testing.T) {
	c := New(100, time.Hour)

	c.SetWithTTL("short", "value", 10*time.Millisecond)
	c.SetWithTTL("long", "value", time.Hour)

	time.Sleep(20 * time.Millisecond)

	_, ok := c.Get("short")
	if ok {
		t.Error("expected short-lived key to expire")
	}

	_, ok = c.Get("long")
	if !ok {
		t.Error("expected long-lived key to exist")
	}
}

func TestManager(t *testing.T) {
	m := NewManager()

	m.Create("cache1", 100, time.Minute)
	m.Create("cache2", 200, time.Hour)

	list := m.List()
	if len(list) != 2 {
		t.Errorf("expected 2 caches, got %d", len(list))
	}

	_, ok := m.Get("cache1")
	if !ok {
		t.Error("expected cache1 to exist")
	}

	m.Delete("cache1")
	_, ok = m.Get("cache1")
	if ok {
		t.Error("expected cache1 to be deleted")
	}
}

func TestManagerDuplicate(t *testing.T) {
	m := NewManager()

	c1 := m.Create("cache1", 100, time.Minute)
	c2 := m.Create("cache1", 200, time.Hour)

	if c1 != c2 {
		t.Error("expected same cache instance for duplicate name")
	}
}

func TestManagerClearAll(t *testing.T) {
	m := NewManager()

	c1 := m.Create("cache1", 100, time.Minute)
	c2 := m.Create("cache2", 100, time.Minute)

	c1.Set("a", 1)
	c2.Set("b", 2)

	m.ClearAll()

	if c1.Size() != 0 || c2.Size() != 0 {
		t.Error("expected all caches to be cleared")
	}
}

func TestHTTPCache(t *testing.T) {
	hc := NewHTTPCache(time.Minute)

	hc.Set("key1", "value1")
	val, ok := hc.Get("key1")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}

	hc.Invalidate("key1")
	_, ok = hc.Get("key1")
	if ok {
		t.Error("expected key to be invalidated")
	}
}

func TestHTTPCacheInvalidatePattern(t *testing.T) {
	hc := NewHTTPCache(time.Minute)

	hc.Set("user:1", "data1")
	hc.Set("user:2", "data2")
	hc.Set("post:1", "data3")

	hc.Invalidate("user:*")

	_, ok := hc.Get("user:1")
	if ok {
		t.Error("expected user:1 to be invalidated")
	}

	_, ok = hc.Get("post:1")
	if !ok {
		t.Error("expected post:1 to still exist")
	}
}

func TestHTTPCacheInvalidateAll(t *testing.T) {
	hc := NewHTTPCache(time.Minute)

	hc.Set("a", 1)
	hc.Set("b", 2)

	hc.Invalidate("*")

	if hc.cache.Size() != 0 {
		t.Error("expected all keys to be invalidated")
	}
}

func TestNewWithPolicy(t *testing.T) {
	c := NewWithPolicy(100, time.Minute, PolicyLFU)
	if c == nil {
		t.Fatal("expected cache to be created")
	}
	if c.policy != PolicyLFU {
		t.Errorf("expected LFU policy, got %v", c.policy)
	}
}

func TestLFUEviction(t *testing.T) {
	c := NewWithPolicy(3, time.Minute, PolicyLFU)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	c.Get("a")
	c.Get("a")
	c.Get("b")

	c.Set("d", 4)

	if _, ok := c.Get("c"); ok {
		t.Error("expected 'c' to be evicted (least frequently used)")
	}
	if _, ok := c.Get("a"); !ok {
		t.Error("expected 'a' to be retained (most frequently used)")
	}
}

func TestLFUAllAccessed(t *testing.T) {
	c := NewWithPolicy(3, time.Minute, PolicyLFU)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	c.Get("a")
	c.Get("a")
	c.Get("b")

	c.Set("d", 4)

	if _, ok := c.Get("c"); ok {
		t.Error("expected 'c' to be evicted (least frequently used, freq=0)")
	}
	if _, ok := c.Get("d"); !ok {
		t.Error("expected 'd' to be present")
	}
}

func TestLRUDefaultPolicy(t *testing.T) {
	c := New(3, time.Minute)
	if c.policy != PolicyLRU {
		t.Errorf("expected default LRU policy, got %v", c.policy)
	}
}

func BenchmarkCacheSet(b *testing.B) {
	c := New(10000, time.Minute)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("key-%d", i), i)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	c := New(10000, time.Minute)
	for i := 0; i < 10000; i++ {
		c.Set(fmt.Sprintf("key-%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(fmt.Sprintf("key-%d", i%10000))
	}
}

func BenchmarkCacheDelete(b *testing.B) {
	c := New(10000, time.Minute)
	for i := 0; i < 10000; i++ {
		c.Set(fmt.Sprintf("key-%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Delete(fmt.Sprintf("key-%d", i%10000))
	}
}

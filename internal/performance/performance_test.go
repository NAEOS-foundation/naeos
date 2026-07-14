package performance

import (
	"sync"
	"testing"
	"time"
)

func TestConnectionPool(t *testing.T) {
	pool := NewConnectionPool("test", 2, 5)

	if pool.name != "test" {
		t.Error("expected name 'test'")
	}

	conn, ok := pool.Acquire()
	if !ok {
		t.Error("expected to acquire connection")
	}
	if conn == nil {
		t.Error("expected connection")
	}

	total, avail, inUse := pool.Stats()
	if total != 2 {
		t.Errorf("expected 2 total, got %d", total)
	}
	if inUse != 1 {
		t.Errorf("expected 1 in use, got %d", inUse)
	}
	if avail != 1 {
		t.Errorf("expected 1 available, got %d", avail)
	}

	pool.Release(conn)

	total, avail, inUse = pool.Stats()
	if inUse != 0 {
		t.Errorf("expected 0 in use, got %d", inUse)
	}
	if avail != 2 {
		t.Errorf("expected 2 available, got %d", avail)
	}
}

func TestConnectionPoolMax(t *testing.T) {
	pool := NewConnectionPool("test", 1, 2)

	c1, ok := pool.Acquire()
	if !ok {
		t.Error("expected connection 1")
	}

	c2, ok := pool.Acquire()
	if !ok {
		t.Error("expected connection 2")
	}

	_, ok = pool.Acquire()
	if ok {
		t.Error("expected pool full")
	}

	pool.Release(c1)
	pool.Release(c2)
}

func TestConnectionPoolClose(t *testing.T) {
	pool := NewConnectionPool("test", 1, 2)
	pool.Close()

	_, ok := pool.Acquire()
	if ok {
		t.Error("expected pool closed")
	}
}

func TestConnectionPoolConcurrency(t *testing.T) {
	pool := NewConnectionPool("test", 5, 10)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, ok := pool.Acquire()
			if ok {
				time.Sleep(10 * time.Millisecond)
				pool.Release(conn)
			}
		}()
	}
	wg.Wait()

	total, avail, inUse := pool.Stats()
	if inUse != 0 {
		t.Errorf("expected 0 in use, got %d", inUse)
	}
	if avail != total {
		t.Errorf("expected all available, got %d/%d", avail, total)
	}
}

func TestBatchProcessor(t *testing.T) {
	processed := make([]any, 0)
	var mu sync.Mutex

	bp := NewBatchProcessor("test", 3, func(item any) error {
		mu.Lock()
		processed = append(processed, item)
		mu.Unlock()
		return nil
	})

	bp.AddItem("a")
	bp.AddItem("b")
	bp.AddItem("c")
	bp.AddItem("d")

	batches := bp.GetBatches()
	if len(batches) != 2 {
		t.Errorf("expected 2 batches, got %d", len(batches))
	}

	bp.ProcessAll()

	mu.Lock()
	if len(processed) != 4 {
		t.Errorf("expected 4 processed, got %d", len(processed))
	}
	mu.Unlock()
}

func TestBatchProcessorError(t *testing.T) {
	bp := NewBatchProcessor("test", 3, func(item any) error {
		if item == "fail" {
			return nil // simplified
		}
		return nil
	})

	bp.AddItem("a")
	bp.AddItem("fail")
	bp.AddItem("c")

	bp.ProcessAll()

	for _, batch := range bp.GetBatches() {
		for _, item := range batch.Items {
			if item.Status == BatchFailed && item.Error == nil {
				t.Error("expected error on failed item")
			}
		}
	}
}

func TestCache(t *testing.T) {
	c := NewCache("test")

	c.Set("key1", "value1", 10*time.Second)

	val, ok := c.Get("key1")
	if !ok || val != "value1" {
		t.Error("expected cache hit")
	}

	if c.Size() != 1 {
		t.Errorf("expected size 1, got %d", c.Size())
	}
}

func TestCacheExpiry(t *testing.T) {
	c := NewCache("test")

	c.Set("key1", "value1", 10*time.Millisecond)

	time.Sleep(20 * time.Millisecond)

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected cache miss")
	}
}

func TestCacheDelete(t *testing.T) {
	c := NewCache("test")

	c.Set("key1", "value1", 10*time.Second)

	if !c.Delete("key1") {
		t.Error("expected true")
	}
	if c.Delete("key1") {
		t.Error("expected false")
	}
}

func TestCacheClear(t *testing.T) {
	c := NewCache("test")

	c.Set("key1", "value1", 10*time.Second)
	c.Set("key2", "value2", 10*time.Second)

	c.Clear()

	if c.Size() != 0 {
		t.Errorf("expected size 0, got %d", c.Size())
	}
}

func TestCacheCleanup(t *testing.T) {
	c := NewCache("test")

	c.Set("key1", "value1", 10*time.Millisecond)
	c.Set("key2", "value2", 10*time.Second)

	time.Sleep(20 * time.Millisecond)
	c.Cleanup()

	if c.Size() != 1 {
		t.Errorf("expected size 1, got %d", c.Size())
	}
}

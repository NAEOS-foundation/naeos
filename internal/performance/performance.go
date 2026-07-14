package performance

import (
	"sync"
	"sync/atomic"
	"time"
)

// Connection Pool

type Connection struct {
	ID        string
	CreatedAt time.Time
	LastUsed  time.Time
	InUse     bool
}

type ConnectionPool struct {
	name      string
	minSize   int
	maxSize   int
	conns     []*Connection
	available []*Connection
	inUse     int32
	mu        sync.Mutex
	closed    bool
}

func NewConnectionPool(name string, min, max int) *ConnectionPool {
	pool := &ConnectionPool{
		name:      name,
		minSize:   min,
		maxSize:   max,
		conns:     make([]*Connection, 0, max),
		available: make([]*Connection, 0, max),
	}

	for i := 0; i < min; i++ {
		conn := &Connection{
			ID:        generateConnID(),
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
		}
		pool.conns = append(pool.conns, conn)
		pool.available = append(pool.available, conn)
	}

	return pool
}

func (p *ConnectionPool) Acquire() (*Connection, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, false
	}

	if len(p.available) > 0 {
		conn := p.available[len(p.available)-1]
		p.available = p.available[:len(p.available)-1]
		conn.InUse = true
		conn.LastUsed = time.Now()
		atomic.AddInt32(&p.inUse, 1)
		return conn, true
	}

	if len(p.conns) < p.maxSize {
		conn := &Connection{
			ID:        generateConnID(),
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
			InUse:     true,
		}
		p.conns = append(p.conns, conn)
		atomic.AddInt32(&p.inUse, 1)
		return conn, true
	}

	return nil, false
}

func (p *ConnectionPool) Release(conn *Connection) {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn.InUse = false
	conn.LastUsed = time.Now()
	atomic.AddInt32(&p.inUse, -1)
	p.available = append(p.available, conn)
}

func (p *ConnectionPool) Stats() (total, available, inUse int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.conns), len(p.available), int(atomic.LoadInt32(&p.inUse))
}

func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	p.conns = nil
	p.available = nil
}

// Batch Processor

type BatchItem struct {
	ID      string
	Data    any
	Status  BatchStatus
	Error   error
}

type BatchStatus int

const (
	BatchPending BatchStatus = iota
	BatchProcessing
	BatchCompleted
	BatchFailed
)

type Batch struct {
	ID      string
	Items   []*BatchItem
	Created time.Time
}

type BatchProcessor struct {
	name      string
	batches   []*Batch
	batchSize int
	handler   func(any) error
	mu        sync.Mutex
}

func NewBatchProcessor(name string, batchSize int, handler func(any) error) *BatchProcessor {
	return &BatchProcessor{
		name:      name,
		batches:   make([]*Batch, 0),
		batchSize: batchSize,
		handler:   handler,
	}
}

func (bp *BatchProcessor) AddItem(item any) *BatchItem {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	batchItem := &BatchItem{
		ID:     generateConnID(),
		Data:   item,
		Status: BatchPending,
	}

	if len(bp.batches) == 0 || len(bp.batches[len(bp.batches)-1].Items) >= bp.batchSize {
		batch := &Batch{
			ID:      generateConnID(),
			Items:   []*BatchItem{batchItem},
			Created: time.Now(),
		}
		bp.batches = append(bp.batches, batch)
	} else {
		bp.batches[len(bp.batches)-1].Items = append(bp.batches[len(bp.batches)-1].Items, batchItem)
	}

	return batchItem
}

func (bp *BatchProcessor) ProcessBatch(batchID string) error {
	bp.mu.Lock()
	var batch *Batch
	for _, b := range bp.batches {
		if b.ID == batchID {
			batch = b
			break
		}
	}
	bp.mu.Unlock()

	if batch == nil {
		return nil
	}

	for _, item := range batch.Items {
		item.Status = BatchProcessing
		if err := bp.handler(item.Data); err != nil {
			item.Status = BatchFailed
			item.Error = err
		} else {
			item.Status = BatchCompleted
		}
	}
	return nil
}

func (bp *BatchProcessor) ProcessAll() {
	bp.mu.Lock()
	batches := make([]*Batch, len(bp.batches))
	copy(batches, bp.batches)
	bp.mu.Unlock()

	for _, batch := range batches {
		bp.ProcessBatch(batch.ID)
	}
}

func (bp *BatchProcessor) GetBatches() []*Batch {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return bp.batches
}

func (bp *BatchProcessor) GetBatchByID(id string) *Batch {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	for _, b := range bp.batches {
		if b.ID == id {
			return b
		}
	}
	return nil
}

// Cache

type CacheEntry struct {
	Key       string
	Value     any
	ExpiresAt time.Time
}

type Cache struct {
	name    string
	entries map[string]*CacheEntry
	mu      sync.RWMutex
}

func NewCache(name string) *Cache {
	return &Cache{
		name:    name,
		entries: make(map[string]*CacheEntry),
	}
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &CacheEntry{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(c.entries, key)
		return nil, false
	}

	return entry.Value, true
}

func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.entries[key]; ok {
		delete(c.entries, key)
		return true
	}
	return false
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}

// Helpers

func generateConnID() string {
	return time.Now().Format("20060102150405.000000000")
}

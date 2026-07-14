package cache

import (
	"container/list"
	"sync"
	"time"
)

// Cache Entry

type Entry struct {
	Key       string
	Value     any
	ExpiresAt time.Time
	Size      int
}

func (e *Entry) IsExpired() bool {
	return !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt)
}

// Cache

type Cache struct {
	items    map[string]*list.Element
	lru      *list.List
	capacity int
	ttl      time.Duration
	mu       sync.RWMutex
	stats    *Stats
}

type Stats struct {
	Hits   int64
	Misses int64
	Size   int
}

func New(capacity int, ttl time.Duration) *Cache {
	return &Cache{
		items:    make(map[string]*list.Element),
		lru:      list.New(),
		capacity: capacity,
		ttl:      ttl,
		stats:    &Stats{},
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		entry := elem.Value.(*Entry)
		if entry.IsExpired() {
			c.remove(elem)
			c.stats.Misses++
			return nil, false
		}
		c.lru.MoveToFront(elem)
		c.stats.Hits++
		return entry.Value, true
	}

	c.stats.Misses++
	return nil, false
}

func (c *Cache) Set(key string, value any) {
	c.SetWithTTL(key, value, c.ttl)
}

func (c *Cache) SetWithTTL(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.lru.Remove(elem)
		delete(c.items, key)
	}

	entry := &Entry{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}

	elem := c.lru.PushFront(entry)
	c.items[key] = elem
	c.stats.Size++

	for c.lru.Len() > c.capacity {
		oldest := c.lru.Back()
		if oldest != nil {
			c.remove(oldest)
		}
	}
}

func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.remove(elem)
		return true
	}
	return false
}

func (c *Cache) Exists(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elem, ok := c.items[key]
	if !ok {
		return false
	}

	entry := elem.Value.(*Entry)
	return !entry.IsExpired()
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.lru.Init()
	c.stats.Size = 0
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lru.Len()
}

func (c *Cache) Stats() *Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return &Stats{
		Hits:   c.stats.Hits,
		Misses: c.stats.Misses,
		Size:   c.lru.Len(),
	}
}

func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

func (c *Cache) Cleanup() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	removed := 0
	for elem := c.lru.Back(); elem != nil; {
		entry := elem.Value.(*Entry)
		if entry.IsExpired() {
			prev := elem.Prev()
			c.remove(elem)
			removed++
			elem = prev
		} else {
			break
		}
	}
	return removed
}

func (c *Cache) remove(elem *list.Element) {
	entry := elem.Value.(*Entry)
	delete(c.items, entry.Key)
	c.lru.Remove(elem)
	c.stats.Size--
}

// Cache Manager (multiple caches)

type Manager struct {
	caches map[string]*Cache
	mu     sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		caches: make(map[string]*Cache),
	}
}

func (m *Manager) Create(name string, capacity int, ttl time.Duration) *Cache {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.caches[name]; exists {
		return m.caches[name]
	}

	cache := New(capacity, ttl)
	m.caches[name] = cache
	return cache
}

func (m *Manager) Get(name string) (*Cache, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cache, ok := m.caches[name]
	return cache, ok
}

func (m *Manager) Delete(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.caches, name)
}

func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.caches))
	for name := range m.caches {
		names = append(names, name)
	}
	return names
}

func (m *Manager) ClearAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cache := range m.caches {
		cache.Clear()
	}
}

// HTTP Cache Middleware

type HTTPCache struct {
	cache *Cache
}

func NewHTTPCache(ttl time.Duration) *HTTPCache {
	return &HTTPCache{
		cache: New(1000, ttl),
	}
}

func (h *HTTPCache) Get(key string) (any, bool) {
	return h.cache.Get(key)
}

func (h *HTTPCache) Set(key string, value any) {
	h.cache.Set(key, value)
}

func (h *HTTPCache) Invalidate(pattern string) {
	for _, key := range h.cache.Keys() {
		if matchPattern(key, pattern) {
			h.cache.Delete(key)
		}
	}
}

func matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		return len(key) >= len(pattern)-1 && key[:len(pattern)-1] == pattern[:len(pattern)-1]
	}
	return key == pattern
}

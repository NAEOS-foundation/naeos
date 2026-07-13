package pipelinecache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

type CacheEntry struct {
	Key       string            `json:"key"`
	Result    *pipeline.Result  `json:"-"`
	Timestamp time.Time         `json:"timestamp"`
	HitCount  int               `json:"hit_count"`
}

type Cache struct {
	dir      string
	entries  map[string]*CacheEntry
	maxSize  int
	mu       sync.RWMutex
}

func New(dir string, maxSize int) *Cache {
	if maxSize <= 0 {
		maxSize = 100
	}
	c := &Cache{
		dir:     dir,
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
	}
	c.loadFromDisk()
	return c
}

func (c *Cache) Get(specHash string) (*pipeline.Result, bool) {
	c.mu.RLock()
	entry, ok := c.entries[specHash]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	entry.HitCount++
	return entry.Result, true
}

func (c *Cache) Set(specHash string, result *pipeline.Result) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxSize {
		c.evict()
	}

	c.entries[specHash] = &CacheEntry{
		Key:       specHash,
		Result:    result,
		Timestamp: time.Now(),
	}

	c.saveToDisk(specHash)
}

func (c *Cache) HashSpec(spec string) string {
	h := sha256.Sum256([]byte(spec))
	return fmt.Sprintf("%x", h)
}

func (c *Cache) Invalidate(specHash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, specHash)
	os.Remove(filepath.Join(c.dir, specHash+".json"))
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.entries {
		os.Remove(filepath.Join(c.dir, key+".json"))
	}
	c.entries = make(map[string]*CacheEntry)
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func (c *Cache) evict() {
	var oldest string
	var oldestTime time.Time
	for key, entry := range c.entries {
		if oldest == "" || entry.Timestamp.Before(oldestTime) {
			oldest = key
			oldestTime = entry.Timestamp
		}
	}
	if oldest != "" {
		delete(c.entries, oldest)
		os.Remove(filepath.Join(c.dir, oldest+".json"))
	}
}

func (c *Cache) loadFromDisk() {
	if c.dir == "" {
		return
	}
	os.MkdirAll(c.dir, 0o755)

	matches, err := filepath.Glob(filepath.Join(c.dir, "*.json"))
	if err != nil {
		return
	}

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}
		c.entries[entry.Key] = &entry
	}
}

func (c *Cache) saveToDisk(key string) {
	if c.dir == "" {
		return
	}
	os.MkdirAll(c.dir, 0o755)

	entry := c.entries[key]
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(filepath.Join(c.dir, key+".json"), data, 0o644)
}

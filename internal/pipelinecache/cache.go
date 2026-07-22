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
	Key       string           `json:"key"`
	Result    *pipeline.Result `json:"-"`
	Timestamp time.Time        `json:"timestamp"`
	HitCount  int              `json:"hit_count"`
}

type Cache struct {
	dir     string
	entries map[string]*CacheEntry
	maxSize int
	maxAge  time.Duration
	mu      sync.RWMutex
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

func (c *Cache) SetMaxAge(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxAge = d
}

func (c *Cache) Get(specHash string) (*pipeline.Result, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[specHash]
	if !ok {
		return nil, false
	}

	if c.maxAge > 0 && time.Since(entry.Timestamp) > c.maxAge {
		delete(c.entries, specHash)
		os.Remove(filepath.Join(c.dir, specHash+".json"))
		return nil, false
	}

	entry.HitCount++
	entry.Timestamp = time.Now()
	return entry.Result, true
}

func (c *Cache) Set(specHash string, result *pipeline.Result) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxSize {
		c.evictLRU()
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

func (c *Cache) evictLRU() {
	var lruKey string
	var lruHits int64 = -1
	var lruTime time.Time

	for key, entry := range c.entries {
		score := int64(entry.HitCount)*1000 + entry.Timestamp.UnixNano()/1e9
		lruScore := lruHits*1000 + lruTime.UnixNano()/1e9
		if lruKey == "" || score < lruScore {
			lruKey = key
			lruHits = int64(entry.HitCount)
			lruTime = entry.Timestamp
		}
	}
	if lruKey != "" {
		delete(c.entries, lruKey)
		os.Remove(filepath.Join(c.dir, lruKey+".json"))
	}
}

func (c *Cache) loadFromDisk() {
	if c.dir == "" {
		return
	}
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return
	}

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

type CacheStats struct {
	Size    int            `json:"size"`
	MaxSize int            `json:"max_size"`
	MaxAge  time.Duration  `json:"max_age"`
	Entries []CacheSummary `json:"entries,omitempty"`
}

type CacheSummary struct {
	Key      string `json:"key"`
	HitCount int    `json:"hit_count"`
	Age      string `json:"age"`
}

func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var entries []CacheSummary
	for key, entry := range c.entries {
		entries = append(entries, CacheSummary{
			Key:      key,
			HitCount: entry.HitCount,
			Age:      time.Since(entry.Timestamp).Round(time.Second).String(),
		})
	}
	return CacheStats{
		Size:    len(c.entries),
		MaxSize: c.maxSize,
		MaxAge:  c.maxAge,
		Entries: entries,
	}
}

func (c *Cache) saveToDisk(key string) {
	if c.dir == "" {
		return
	}
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return
	}

	entry := c.entries[key]
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return
	}

	_ = os.WriteFile(filepath.Join(c.dir, key+".json"), data, 0o600)
}

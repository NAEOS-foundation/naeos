package pipeline

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type StageCacheEntry struct {
	Stage     string    `json:"stage"`
	Version   string    `json:"version"`
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

type StageCache struct {
	dir      string
	entries  map[string]*StageCacheEntry
	mu       sync.RWMutex
	versions map[string]string
}

func NewStageCache(dir string) *StageCache {
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "naeos-stage-cache")
	}
	return &StageCache{
		dir:      dir,
		entries:  make(map[string]*StageCacheEntry),
		versions: make(map[string]string),
	}
}

func (s *StageCache) SetVersion(stage, version string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.versions[stage] = version
}

func (s *StageCache) stageKey(stage, version string, data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%s:%s:%x", stage, version, h)
}

func (s *StageCache) Get(stage string, data []byte) ([]byte, bool) {
	s.mu.RLock()
	version := s.versions[stage]
	s.mu.RUnlock()

	key := s.stageKey(stage, version, data)

	s.mu.RLock()
	entry, ok := s.entries[key]
	s.mu.RUnlock()
	if ok {
		return entry.Data, true
	}

	entry, err := s.loadFromDisk(key)
	if err == nil && entry != nil {
		s.mu.Lock()
		s.entries[key] = entry
		s.mu.Unlock()
		return entry.Data, true
	}

	return nil, false
}

func (s *StageCache) Set(stage string, data, output []byte) {
	s.mu.RLock()
	version := s.versions[stage]
	s.mu.RUnlock()

	key := s.stageKey(stage, version, data)

	entry := &StageCacheEntry{
		Stage:     stage,
		Version:   version,
		Data:      output,
		Timestamp: time.Now(),
	}

	s.mu.Lock()
	s.entries[key] = entry
	s.mu.Unlock()

	s.saveToDisk(key, entry)
}

func (s *StageCache) Invalidate(stage string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, entry := range s.entries {
		if entry.Stage == stage {
			delete(s.entries, key)
			os.Remove(filepath.Join(s.dir, key+".json"))
		}
	}
}

func (s *StageCache) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.entries {
		os.Remove(filepath.Join(s.dir, key+".json"))
	}
	s.entries = make(map[string]*StageCacheEntry)
}

func (s *StageCache) Version(stage string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.versions[stage]
}

func (s *StageCache) Stats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stats := make(map[string]int)
	for _, entry := range s.entries {
		stats[entry.Stage]++
	}
	return stats
}

func (s *StageCache) loadFromDisk(key string) (*StageCacheEntry, error) {
	if s.dir == "" {
		return nil, fmt.Errorf("no cache dir")
	}
	data, err := os.ReadFile(filepath.Join(s.dir, key+".json"))
	if err != nil {
		return nil, err
	}
	var entry StageCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *StageCache) saveToDisk(key string, entry *StageCacheEntry) {
	if s.dir == "" {
		return
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(s.dir, key+".json"), data, 0o600)
}

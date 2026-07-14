package configreload

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config Store

type Config struct {
	data      map[string]any
	version   int
	lastMod   time.Time
	filePath  string
	watchers  []ConfigWatcher
	mu        sync.RWMutex
}

type ConfigWatcher func(old, new map[string]any)

func New(filePath string) *Config {
	return &Config{
		data:     make(map[string]any),
		filePath: filePath,
	}
}

func (c *Config) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var newConfig map[string]any

	ext := filepath.Ext(c.filePath)
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &newConfig); err != nil {
			return fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		if err := json.Unmarshal(data, &newConfig); err != nil {
			return fmt.Errorf("failed to parse JSON config: %w", err)
		}
	}

	old := c.data
	c.data = newConfig
	c.version++
	c.lastMod = time.Now()

	// Notify watchers
	for _, watcher := range c.watchers {
		watcher(old, newConfig)
	}

	return nil
}

func (c *Config) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.data[key]
	return val, ok
}

func (c *Config) GetString(key, defaultValue string) string {
	val, ok := c.Get(key)
	if !ok {
		return defaultValue
	}
	if s, ok := val.(string); ok {
		return s
	}
	return defaultValue
}

func (c *Config) GetInt(key string, defaultValue int) int {
	val, ok := c.Get(key)
	if !ok {
		return defaultValue
	}
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return defaultValue
}

func (c *Config) GetBool(key string, defaultValue bool) bool {
	val, ok := c.Get(key)
	if !ok {
		return defaultValue
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultValue
}

func (c *Config) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	old := make(map[string]any)
	for k, v := range c.data {
		old[k] = v
	}

	c.data[key] = value
	c.version++
	c.lastMod = time.Now()

	for _, watcher := range c.watchers {
		watcher(old, c.data)
	}
}

func (c *Config) SetAll(data map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	old := c.data
	c.data = data
	c.version++
	c.lastMod = time.Now()

	for _, watcher := range c.watchers {
		watcher(old, data)
	}
}

func (c *Config) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.data))
	for key := range c.data {
		keys = append(keys, key)
	}
	return keys
}

func (c *Config) Version() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.version
}

func (c *Config) LastModified() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastMod
}

func (c *Config) OnChange(watcher ConfigWatcher) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.watchers = append(c.watchers, watcher)
}

func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ext := filepath.Ext(c.filePath)
	var data []byte
	var err error

	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(c.data)
	default:
		data, err = json.MarshalIndent(c.data, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(c.filePath, data, 0644)
}

func (c *Config) Snapshot() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	snapshot := make(map[string]any)
	for k, v := range c.data {
		snapshot[k] = v
	}
	return snapshot
}

// File Watcher

type FileWatcher struct {
	config   *Config
	interval time.Duration
	stopCh   chan struct{}
	running  bool
	mu       sync.RWMutex
}

func NewFileWatcher(config *Config, interval time.Duration) *FileWatcher {
	return &FileWatcher{
		config:   config,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (fw *FileWatcher) Start() {
	fw.mu.Lock()
	fw.running = true
	fw.stopCh = make(chan struct{})
	fw.mu.Unlock()

	go fw.watch()
}

func (fw *FileWatcher) Stop() {
	fw.mu.Lock()
	fw.running = false
	fw.mu.Unlock()
	close(fw.stopCh)
}

func (fw *FileWatcher) IsRunning() bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.running
}

func (fw *FileWatcher) watch() {
	ticker := time.NewTicker(fw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-fw.stopCh:
			return
		case <-ticker.C:
			fw.config.Load()
		}
	}
}

// Config Manager

type Manager struct {
	configs map[string]*Config
	mu      sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		configs: make(map[string]*Config),
	}
}

func (m *Manager) Add(name string, config *Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configs[name] = config
}

func (m *Manager) Get(name string) (*Config, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, ok := m.configs[name]
	return config, ok
}

func (m *Manager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.configs, name)
}

func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.configs))
	for name := range m.configs {
		names = append(names, name)
	}
	return names
}

func (m *Manager) ReloadAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, config := range m.configs {
		if err := config.Load(); err != nil {
			return fmt.Errorf("failed to reload config '%s': %w", name, err)
		}
	}
	return nil
}

// Diff

type ConfigDiff struct {
	Added    map[string]any
	Removed  map[string]any
	Modified map[string]any
}

func Diff(old, new map[string]any) *ConfigDiff {
	diff := &ConfigDiff{
		Added:    make(map[string]any),
		Removed:  make(map[string]any),
		Modified: make(map[string]any),
	}

	for key, newVal := range new {
		oldVal, exists := old[key]
		if !exists {
			diff.Added[key] = newVal
		} else if fmt.Sprintf("%v", oldVal) != fmt.Sprintf("%v", newVal) {
			diff.Modified[key] = map[string]any{
				"old": oldVal,
				"new": newVal,
			}
		}
	}

	for key, oldVal := range old {
		if _, exists := new[key]; !exists {
			diff.Removed[key] = oldVal
		}
	}

	return diff
}

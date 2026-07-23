package pluginhost

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	goplugin "plugin"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// Manager is the unified plugin manager that handles loading, lifecycle,
// sandboxing, and execution of plugins.
type Manager struct {
	pluginDir string
	plugins   map[string]Plugin
	info      map[string]*PluginInfo
	config    PluginConfig
	sandbox   *Sandbox
	events    *EventBus
	mu        sync.RWMutex
}

// PluginConfig is the persisted plugin configuration.
type PluginConfig struct {
	Plugins []PluginInfo  `json:"plugins"`
	Sandbox SandboxConfig `json:"sandbox,omitempty"`
	Lazy    bool          `json:"lazy,omitempty"`
}

// NewManager creates a new PluginManager for the given directory.
func NewManager(pluginDir string) *Manager {
	return &Manager{
		pluginDir: pluginDir,
		plugins:   make(map[string]Plugin),
		info:      make(map[string]*PluginInfo),
		config:    PluginConfig{Lazy: true},
		sandbox:   NewSandbox(SandboxConfig{}),
		events:    NewEventBus(),
	}
}

func (m *Manager) configPath() string {
	return filepath.Join(m.pluginDir, "plugins.json")
}

// LoadConfig reads the plugin configuration from disk.
func (m *Manager) LoadConfig() error {
	data, err := os.ReadFile(m.configPath())
	if err != nil {
		if os.IsNotExist(err) {
			m.config = PluginConfig{}
			return nil
		}
		return err
	}
	if err := json.Unmarshal(data, &m.config); err != nil {
		return err
	}
	m.sandbox = NewSandbox(m.config.Sandbox)
	return nil
}

// SaveConfig writes the plugin configuration to disk.
func (m *Manager) SaveConfig() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(m.pluginDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(m.configPath(), data, 0o600)
}

// EventBus returns the plugin event bus for subscribing to pipeline events.
func (m *Manager) EventBus() *EventBus {
	return m.events
}

// List returns metadata for all configured plugins.
func (m *Manager) List() []PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.Plugins
}

// Get returns a loaded plugin by name.
func (m *Manager) Get(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[name]
	return p, ok
}

// GetInfo returns plugin info by name.
func (m *Manager) GetInfo(name string) (*PluginInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for i := range m.config.Plugins {
		if m.config.Plugins[i].Name == name {
			return &m.config.Plugins[i], true
		}
	}
	return nil, false
}

// Install registers a Go plugin from a .so file path.
// It reads the exported symbols (PluginName, PluginVersion, PluginDescription)
// to build metadata, then saves to config.
func (m *Manager) Install(path string) (*PluginInfo, error) {
	goPlugin, err := goplugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open plugin %s: %w", path, err)
	}

	symName, err := goPlugin.Lookup("PluginName")
	if err != nil {
		return nil, fmt.Errorf("plugin %s does not export PluginName: %w", path, err)
	}
	namePtr, ok := symName.(*string)
	if !ok {
		return nil, fmt.Errorf("plugin %q: exported PluginName must be a *string; rebuild the plugin with the correct type", path)
	}

	version := "0.0.0"
	if symVersion, err := goPlugin.Lookup("PluginVersion"); err == nil {
		if vPtr, ok := symVersion.(*string); ok {
			version = *vPtr
		}
	}

	description := ""
	if symDesc, err := goPlugin.Lookup("PluginDescription"); err == nil {
		if dPtr, ok := symDesc.(*string); ok {
			description = *dPtr
		}
	}

	author := ""
	if symAuthor, err := goPlugin.Lookup("PluginAuthor"); err == nil {
		if aPtr, ok := symAuthor.(*string); ok {
			author = *aPtr
		}
	}

	pInfo := PluginInfo{
		Name:        *namePtr,
		Version:     version,
		Description: description,
		Author:      author,
		Path:        path,
		Enabled:     true,
		State:       StateCreated,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for i, p := range m.config.Plugins {
		if p.Name == pInfo.Name {
			m.config.Plugins[i] = pInfo
			return &pInfo, m.SaveConfig()
		}
	}

	m.config.Plugins = append(m.config.Plugins, pInfo)
	return &pInfo, m.SaveConfig()
}

// Uninstall removes a plugin from the config.
func (m *Manager) Uninstall(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, p := range m.config.Plugins {
		if p.Name == name {
			m.config.Plugins = append(m.config.Plugins[:i], m.config.Plugins[i+1:]...)
			delete(m.plugins, name)
			delete(m.info, name)
			return m.SaveConfig()
		}
	}
	return 	naeoserr.Wrap(naeoserr.ErrPlugin, fmt.Sprintf("plugin %q not found", name), nil)
}

// Enable enables a plugin by name.
func (m *Manager) Enable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, p := range m.config.Plugins {
		if p.Name == name {
			m.config.Plugins[i].Enabled = true
			return m.SaveConfig()
		}
	}
	return 	naeoserr.Wrap(naeoserr.ErrPlugin, fmt.Sprintf("plugin %q not found", name), nil)
}

// Disable disables a plugin by name.
func (m *Manager) Disable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, p := range m.config.Plugins {
		if p.Name == name {
			m.config.Plugins[i].Enabled = false
			return m.SaveConfig()
		}
	}
	return 	naeoserr.Wrap(naeoserr.ErrPlugin, fmt.Sprintf("plugin %q not found", name), nil)
}

// Register registers a plugin in-memory (for in-process plugins).
func (m *Manager) Register(p Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := p.Name()
	if _, exists := m.plugins[name]; exists {
		return naeoserr.Wrap(naeoserr.ErrConflict, fmt.Sprintf("plugin %q already registered", name), nil)
	}

	m.plugins[name] = p
	m.info[name] = &PluginInfo{
		Name:    name,
		Version: p.Version(),
		State:   StateCreated,
	}
	return nil
}

// Unregister removes a plugin from in-memory registration.
func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; !exists {
		return fmt.Errorf("plugin %q not found in memory; check 'naeos plugin list' and verify it is loaded", name)
		return 	naeoserr.Wrap(naeoserr.ErrPlugin, fmt.Sprintf("plugin %q not found in memory", name), nil)
	}

	delete(m.plugins, name)
	delete(m.info, name)
	return nil
}

// LoadAll loads all enabled plugins from .so files.
// Returns a combined error if any plugins fail to load, but continues loading others.
// When Lazy is true, this still eagerly loads all plugins for backward compatibility.
func (m *Manager) LoadAll(ctx *PluginContext) error {
	m.mu.RLock()
	pluginsCopy := make([]PluginInfo, len(m.config.Plugins))
	copy(pluginsCopy, m.config.Plugins)
	m.mu.RUnlock()

	var errs []string
	for _, pInfo := range pluginsCopy {
		if !pInfo.Enabled || pInfo.Path == "" {
			continue
		}
		if err := m.sandbox.ValidatePath(pInfo.Path); err != nil {
			errs = append(errs, fmt.Sprintf("plugin %s: sandbox validation failed: %v", pInfo.Name, err))
			continue
		}
		p, err := m.loadGoPlugin(pInfo.Path)
		if err != nil {
			errs = append(errs, fmt.Sprintf("plugin %s: load failed: %v", pInfo.Name, err))
			continue
		}
		if err := p.Initialize(ctx); err != nil {
			m.updateState(pInfo.Name, StateError, err)
			errs = append(errs, fmt.Sprintf("plugin %s: init failed: %v", pInfo.Name, err))
			continue
		}
		m.mu.Lock()
		m.plugins[pInfo.Name] = p
		m.updateStateLocked(pInfo.Name, StateInitialized, nil)
		for i := range m.config.Plugins {
			if m.config.Plugins[i].Name == pInfo.Name {
				m.config.Plugins[i].Loaded = true
				break
			}
		}
		m.mu.Unlock()
	}
	if len(errs) > 0 {
		loadErrs := make([]error, len(errs))
		for i, e := range errs {
			loadErrs[i] = naeoserr.Wrap(naeoserr.ErrPlugin, e, nil)
		}
		return naeoserr.Group(loadErrs...)
	}
	return nil
}

func (m *Manager) loadGoPlugin(path string) (Plugin, error) {
	goPlugin, err := goplugin.Open(path)
	if err != nil {
		return nil, err
	}

	sym, err := goPlugin.Lookup("NaeosPlugin")
	if err != nil {
		return nil, fmt.Errorf("plugin does not export NaeosPlugin: %w", err)
	}

	p, ok := sym.(Plugin)
	if !ok {
		return nil, fmt.Errorf("NaeosPlugin does not implement pluginhost.Plugin interface")
	}

	return p, nil
}

// InitializeAll initializes all registered in-process plugins.
func (m *Manager) InitializeAll(ctx *PluginContext) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, p := range m.plugins {
		if err := p.Initialize(ctx); err != nil {
			m.updateStateLocked(name, StateError, err)
			return naeoserr.Wrap(naeoserr.ErrPlugin, fmt.Sprintf("failed to initialize plugin '%s'", name), err)
		}
		m.updateStateLocked(name, StateInitialized, nil)
	}
	return nil
}

// ShutdownAll shuts down all loaded plugins.
func (m *Manager) ShutdownAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, p := range m.plugins {
		if err := p.Shutdown(); err != nil {
			m.updateStateLocked(name, StateError, err)
			lastErr = naeoserr.Wrap(naeoserr.ErrPlugin, fmt.Sprintf("failed to shutdown plugin '%s'", name), err)
		} else {
			m.updateStateLocked(name, StateStopped, nil)
		}
	}
	return lastErr
}

// Execute runs a plugin action with sandbox protections.
// If lazy loading is enabled and the plugin hasn't been loaded yet,
// it loads and initializes the plugin first.
func (m *Manager) Execute(ctx context.Context, name, action string, params map[string]any) (any, error) {
	p, ok := m.Get(name)
	if !ok {
		if m.config.Lazy {
			pluginCtx := &PluginContext{}
			if err := m.lazyLoad(name, pluginCtx); err != nil {
				return nil, naeoserr.Wrapf(err, naeoserr.ErrPlugin, "lazy load plugin %s", name)
			}
			p, ok = m.Get(name)
			if !ok {
				return nil, fmt.Errorf("plugin %q not loaded; check plugin paths and rebuild if needed", name)
			}
		} else {
			return nil, fmt.Errorf("plugin %q not loaded; enable lazy loading or run 'naeos plugin load %s'", name, name)
		return nil, 	naeoserr.Wrap(naeoserr.ErrPlugin, fmt.Sprintf("plugin %s not loaded after lazy load", name), nil)
		}
	} else {
		return nil, 	naeoserr.Wrap(naeoserr.ErrPlugin, fmt.Sprintf("plugin %s not loaded", name), nil)
	}
	}
	if err := m.sandbox.CheckRateLimit(name); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.updateStateLocked(name, StateRunning, nil)
	m.mu.Unlock()

	result, err := m.sandbox.ExecuteWithTimeout(ctx, func() (any, error) {
		return p.Execute(action, params)
	})

	m.mu.Lock()
	if err != nil {
		m.updateStateLocked(name, StateError, err)
	} else {
		m.updateStateLocked(name, StateInitialized, nil)
	}
	m.mu.Unlock()

	return result, err
}

// lazyLoad loads a single plugin by name from its .so file, initializes it,
// and registers it in the manager.
func (m *Manager) lazyLoad(name string, ctx *PluginContext) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.config.Plugins {
		if m.config.Plugins[i].Name == name {
			if m.config.Plugins[i].Loaded {
				return nil
			}
			pInfo := &m.config.Plugins[i]
			if !pInfo.Enabled || pInfo.Path == "" {
				return fmt.Errorf("plugin %q is disabled or has no path; enable it with 'naeos plugin enable %s'", name, name)
				return naeoserr.Wrap(naeoserr.ErrPlugin, fmt.Sprintf("plugin %s is disabled or has no path", name), nil)
			}
			if err := m.sandbox.ValidatePath(pInfo.Path); err != nil {
				return fmt.Errorf("sandbox validation failed for plugin %q: %w", name, err)
			}
			p, err := m.loadGoPlugin(pInfo.Path)
			if err != nil {
				return fmt.Errorf("load plugin %q: %w", name, err)
			}
			if err := p.Initialize(ctx); err != nil {
				m.updateStateLocked(name, StateError, err)
				return fmt.Errorf("init plugin %q: %w", name, err)
				return naeoserr.Wrap(naeoserr.ErrPlugin, fmt.Sprintf("init plugin %s", name), err)
			}
			m.plugins[name] = p
			pInfo.Loaded = true
			m.updateStateLocked(name, StateInitialized, nil)
			return m.SaveConfig()
		}
	}
	return fmt.Errorf("plugin %q not found; run 'naeos plugin list' to see installed plugins", name)
	return 	naeoserr.Wrap(naeoserr.ErrPlugin, fmt.Sprintf("plugin %s not found", name), nil)
}

// Cleanup calls Shutdown on all loaded plugins and releases resources.
func (m *Manager) Cleanup() error {
	var errs []string
	for name, p := range m.plugins {
		if err := p.Shutdown(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
		}
	}
	if len(errs) > 0 {
		cleanupErrs := make([]error, len(errs))
		for i, e := range errs {
			cleanupErrs[i] = naeoserr.Wrap(naeoserr.ErrPlugin, e, nil)
		}
		return naeoserr.Group(cleanupErrs...)
	}
	return nil
}

func (m *Manager) updateState(name string, state PluginState, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateStateLocked(name, state, err)
}

func (m *Manager) updateStateLocked(name string, state PluginState, err error) {
	if info, ok := m.info[name]; ok {
		info.State = state
		if state == StateInitialized || state == StateRunning {
			if info.StartedAt.IsZero() {
				info.StartedAt = time.Now()
			}
		}
		if err != nil {
			info.Error = err
		}
	}
	for i := range m.config.Plugins {
		if m.config.Plugins[i].Name == name {
			m.config.Plugins[i].State = state
			if err != nil {
				m.config.Plugins[i].Error = err
			}
			break
		}
	}
}

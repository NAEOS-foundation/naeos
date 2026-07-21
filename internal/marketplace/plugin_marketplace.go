package marketplace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/securityext"
)

type PluginDependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type VersionEntry struct {
	Version   string    `json:"version"`
	Installed time.Time `json:"installed"`
}

type PluginEntry struct {
	Name         string             `json:"name"`
	Version      string             `json:"version"`
	Description  string             `json:"description"`
	Author       string             `json:"author"`
	Type         string             `json:"type"`
	Tags         []string           `json:"tags"`
	Dependencies []PluginDependency `json:"dependencies,omitempty"`
	Downloads    int                `json:"downloads"`
	Installed    bool               `json:"installed,omitempty"`
	VersionHistory []VersionEntry   `json:"version_history,omitempty"`
	Config       map[string]any     `json:"config,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

type PluginMarketplace struct {
	cacheDir   string
	installDir string
}

func NewPluginMarketplace(cacheDir, installDir string) *PluginMarketplace {
	return &PluginMarketplace{
		cacheDir:   cacheDir,
		installDir: installDir,
	}
}

func (m *PluginMarketplace) Publish(entry PluginEntry) error {
	entries, err := m.loadPlugins()
	if err != nil {
		entries = []PluginEntry{}
	}

	entry.UpdatedAt = time.Now()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	for i, e := range entries {
		if e.Name == entry.Name {
			entries[i] = entry
			return m.savePlugins(entries)
		}
	}

	entries = append(entries, entry)
	return m.savePlugins(entries)
}

func (m *PluginMarketplace) Get(name string) (*PluginEntry, error) {
	entries, err := m.loadPlugins()
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.Name == name {
			return &entry, nil
		}
	}

	return nil, fmt.Errorf("plugin %s not found", name)
}

func (m *PluginMarketplace) Search(query string, tags []string) ([]PluginEntry, error) {
	entries, err := m.loadPlugins()
	if err != nil {
		return nil, err
	}

	var results []PluginEntry
	for _, entry := range entries {
		if query != "" {
			matched := false
			if containsStr(entry.Name, query) || containsStr(entry.Description, query) {
				matched = true
			}
			for _, tag := range entry.Tags {
				if containsStr(tag, query) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		if len(tags) > 0 {
			hasTag := false
			for _, filterTag := range tags {
				for _, entryTag := range entry.Tags {
					if entryTag == filterTag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		results = append(results, entry)
	}

	return results, nil
}

func (m *PluginMarketplace) List() ([]PluginEntry, error) {
	return m.loadPlugins()
}

func (m *PluginMarketplace) Install(name string) error {
	if err := securityext.ValidatePluginName(name); err != nil {
		return fmt.Errorf("invalid plugin name %q: %w", name, err)
	}

	entry, err := m.Get(name)
	if err != nil {
		return err
	}

	if m.IsInstalled(name) {
		return fmt.Errorf("plugin %q is already installed", name)
	}

	if len(entry.Dependencies) > 0 {
		resolved, err := m.ResolveDependencies(entry.Dependencies, nil)
		if err != nil {
			return err
		}
		for _, dep := range resolved {
			if !m.IsInstalled(dep.Name) {
				if err := m.installOne(dep); err != nil {
					return fmt.Errorf("install dependency %q: %w", dep.Name, err)
				}
			}
		}
	}

	return m.installOne(entry)
}

func (m *PluginMarketplace) installOne(entry *PluginEntry) error {
	if err := os.MkdirAll(m.installDir, 0o755); err != nil {
		return err
	}

	pluginDir := filepath.Join(m.installDir, entry.Name)
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		return err
	}

	configFile := filepath.Join(pluginDir, "plugin.json")
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configFile, data, 0o600); err != nil {
		return err
	}

	entries, err := m.loadPlugins()
	if err != nil {
		return err
	}

	now := time.Now()
	for i, e := range entries {
		if e.Name == entry.Name {
			entries[i].Installed = true
			entries[i].Version = entry.Version
			entries[i].Downloads++
			entries[i].UpdatedAt = now
			entries[i].VersionHistory = append(entries[i].VersionHistory, VersionEntry{
				Version:   entry.Version,
				Installed: now,
			})
			return m.savePlugins(entries)
		}
	}

	return nil
}

// ResolveDependencies checks that all dependencies can be satisfied, detecting cycles
// and missing plugins. It returns the list of plugins to install in dependency order.
func (m *PluginMarketplace) ResolveDependencies(deps []PluginDependency, visited []string) ([]*PluginEntry, error) {
	if visited == nil {
		visited = make([]string, 0)
	}

	var resolved []*PluginEntry
	for _, dep := range deps {
		for _, v := range visited {
			if v == dep.Name {
				return nil, fmt.Errorf("circular dependency detected: %s", strings.Join(append(visited, dep.Name), " → "))
			}
		}

		entry, err := m.Get(dep.Name)
		if err != nil {
			return nil, fmt.Errorf("dependency %q not found: %w", dep.Name, err)
		}

		if dep.Version != "" && !versionMatch(dep.Version, entry.Version) {
			return nil, fmt.Errorf("dependency %q requires version %s, available %s", dep.Name, dep.Version, entry.Version)
		}

		if len(entry.Dependencies) > 0 {
			sub, err := m.ResolveDependencies(entry.Dependencies, append(visited, dep.Name))
			if err != nil {
				return nil, fmt.Errorf("resolve %q deps: %w", dep.Name, err)
			}
			resolved = append(resolved, sub...)
		}

		resolved = append(resolved, entry)
	}

	return resolved, nil
}

func versionMatch(required, available string) bool {
	if required == "" || required == "*" {
		return true
	}
	return required == available
}

// VersionHistory returns the install version history for a plugin.
func (m *PluginMarketplace) VersionHistory(name string) ([]VersionEntry, error) {
	if err := securityext.ValidatePluginName(name); err != nil {
		return nil, fmt.Errorf("invalid plugin name %q: %w", name, err)
	}
	entries, err := m.loadPlugins()
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.Name == name {
			if len(e.VersionHistory) == 0 {
				return nil, fmt.Errorf("no version history for %q", name)
			}
			return e.VersionHistory, nil
		}
	}
	return nil, fmt.Errorf("plugin %q not found", name)
}

// Rollback reverts a plugin to a previous version from its version history.
// If version is empty, it rolls back to the second-most-recent version.
func (m *PluginMarketplace) Rollback(name, version string) error {
	if err := securityext.ValidatePluginName(name); err != nil {
		return fmt.Errorf("invalid plugin name %q: %w", name, err)
	}

	history, err := m.VersionHistory(name)
	if err != nil {
		return err
	}

	var target *VersionEntry
	if version == "" {
		if len(history) < 2 {
			return fmt.Errorf("no previous version to roll back to for %q", name)
		}
		target = &history[len(history)-2]
	} else {
		for i := len(history) - 1; i >= 0; i-- {
			if history[i].Version == version {
				target = &history[i]
				break
			}
		}
		if target == nil {
			return fmt.Errorf("version %q not found in history for %q", version, name)
		}
	}

	entry, err := m.Get(name)
	if err != nil {
		return err
	}
	entry.Version = target.Version
	entry.UpdatedAt = time.Now()

	return m.installOne(entry)
}

func (m *PluginMarketplace) Uninstall(name string) error {
	if err := securityext.ValidatePluginName(name); err != nil {
		return fmt.Errorf("invalid plugin name %q: %w", name, err)
	}
	pluginDir := filepath.Join(m.installDir, name)
	if err := os.RemoveAll(pluginDir); err != nil {
		return err
	}

	entries, err := m.loadPlugins()
	if err != nil {
		return err
	}

	for i, entry := range entries {
		if entry.Name == name {
			entries[i].Installed = false
			entries[i].UpdatedAt = time.Now()
			return m.savePlugins(entries)
		}
	}

	return nil
}

func (m *PluginMarketplace) IsInstalled(name string) bool {
	if err := securityext.ValidatePluginName(name); err != nil {
		return false
	}
	pluginDir := filepath.Join(m.installDir, name)
	_, err := os.Stat(pluginDir)
	return err == nil
}

func (m *PluginMarketplace) ListInstalled() ([]PluginEntry, error) {
	entries, err := m.loadPlugins()
	if err != nil {
		return nil, err
	}

	var installed []PluginEntry
	for _, entry := range entries {
		if m.IsInstalled(entry.Name) {
			entry.Installed = true
			installed = append(installed, entry)
		}
	}

	return installed, nil
}

func (m *PluginMarketplace) loadPlugins() ([]PluginEntry, error) {
	path := filepath.Join(m.cacheDir, "plugins.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m.defaultPlugins(), nil
		}
		return nil, err
	}

	var entries []PluginEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func (m *PluginMarketplace) savePlugins(entries []PluginEntry) error {
	if err := os.MkdirAll(m.cacheDir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(m.cacheDir, "plugins.json"), data, 0o600)
}

func (m *PluginMarketplace) defaultPlugins() []PluginEntry {
	return []PluginEntry{
		{
			Name:        "naeos-lint",
			Version:     "1.0.0",
			Description: "Advanced linting rules for NAEOS specs",
			Author:      "naeos",
			Type:        "lint",
			Tags:        []string{"lint", "validation", "quality"},
			Downloads:   500,
		},
		{
			Name:        "naeos-security",
			Version:     "1.0.0",
			Description: "Security audit plugin for specifications",
			Author:      "naeos",
			Type:        "security",
			Tags:        []string{"security", "audit", "compliance"},
			Downloads:   350,
		},
		{
			Name:        "naeos-docs",
			Version:     "1.0.0",
			Description: "Auto-generate documentation from specs",
			Author:      "naeos",
			Type:        "documentation",
			Tags:        []string{"docs", "documentation", "generation"},
			Downloads:   280,
		},
		{
			Name:        "naeos-test",
			Version:     "1.0.0",
			Description: "Test generation and execution plugin",
			Author:      "naeos",
			Type:        "testing",
			Tags:        []string{"test", "testing", "coverage"},
			Downloads:   220,
		},
	}
}

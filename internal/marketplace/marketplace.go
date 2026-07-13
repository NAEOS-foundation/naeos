package marketplace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/version"
)

type RegistryEntry struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Tags        []string  `json:"tags"`
	Downloads   int       `json:"downloads"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SearchFilter struct {
	Query string
	Tags  []string
	Limit int
}

type RegistryClient struct {
	cacheDir string
}

func NewClient(cacheDir string) *RegistryClient {
	return &RegistryClient{cacheDir: cacheDir}
}

func (c *RegistryClient) Search(filter SearchFilter) ([]RegistryEntry, error) {
	entries, err := c.loadCache()
	if err != nil {
		return nil, err
	}

	var results []RegistryEntry
	for _, entry := range entries {
		if filter.Query != "" {
			matched := false
			if contains(entry.Name, filter.Query) || contains(entry.Description, filter.Query) {
				matched = true
			}
			for _, tag := range entry.Tags {
				if contains(tag, filter.Query) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if len(filter.Tags) > 0 {
			hasTag := false
			for _, filterTag := range filter.Tags {
				for _, entryTag := range entry.Tags {
					if entryTag == filterTag {
						hasTag = true
						break
					}
				}
			}
			if !hasTag {
				continue
			}
		}
		results = append(results, entry)
		if filter.Limit > 0 && len(results) >= filter.Limit {
			break
		}
	}
	return results, nil
}

func (c *RegistryClient) Get(name string) (*RegistryEntry, error) {
	entries, err := c.loadCache()
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.Name == name {
			return &entry, nil
		}
	}
	return nil, fmt.Errorf("spec %s not found in registry", name)
}

func (c *RegistryClient) Install(name, targetDir string) error {
	entry, err := c.Get(name)
	if err != nil {
		return err
	}
	specContent := fmt.Sprintf("# %s v%s\n# %s\n# Author: %s\n\nproject: %s\n",
		entry.Name, entry.Version, entry.Description, entry.Author, entry.Name)
	return os.WriteFile(filepath.Join(targetDir, "spec.yaml"), []byte(specContent), 0o600)
}

func (c *RegistryClient) Publish(entry RegistryEntry) error {
	entries, err := c.loadCache()
	if err != nil {
		entries = []RegistryEntry{}
	}
	for i, e := range entries {
		if e.Name == entry.Name {
			entries[i] = entry
			return c.saveCache(entries)
		}
	}
	entries = append(entries, entry)
	return c.saveCache(entries)
}

func (c *RegistryClient) loadCache() ([]RegistryEntry, error) {
	path := filepath.Join(c.cacheDir, "registry.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c.defaultEntries(), nil
		}
		return nil, err
	}
	var entries []RegistryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (c *RegistryClient) saveCache(entries []RegistryEntry) error {
	if err := os.MkdirAll(c.cacheDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.cacheDir, "registry.json"), data, 0o600)
}

func (c *RegistryClient) defaultEntries() []RegistryEntry {
	return []RegistryEntry{
		{Name: "go-http-api", Version: version.DefaultEntryVersion, Description: "Go HTTP API starter", Author: version.ProductName, Tags: []string{"go", "http", "api"}, Downloads: 100},
		{Name: "python-ml-service", Version: version.DefaultEntryVersion, Description: "Python ML service template", Author: version.ProductName, Tags: []string{"python", "ml", "service"}, Downloads: 50},
		{Name: "rust-web-service", Version: version.DefaultModuleVersion, Description: "Rust web service template", Author: version.ProductName, Tags: []string{"rust", "web", "service"}, Downloads: 30},
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

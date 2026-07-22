package schemaregistry

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"golang.org/x/mod/semver"
)

// SchemaEntry represents a versioned schema in the registry.
type SchemaEntry struct {
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Schema    string    `json:"schema"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Summary   string    `json:"summary,omitempty"`
	TenantID  string    `json:"tenant_id,omitempty"`
}

// Registry provides versioned JSON Schema storage and resolution.
type Registry struct {
	mu      sync.RWMutex
	schemas map[string]map[string]*SchemaEntry
}

func New() *Registry {
	return &Registry{
		schemas: make(map[string]map[string]*SchemaEntry),
	}
}

// Register adds or updates a schema version. If the version already exists,
// it is overwritten with the new schema.
func (r *Registry) Register(name, version, schema string) error {
	if name == "" {
		return fmt.Errorf("schema name must not be empty")
	}
	if version == "" {
		return fmt.Errorf("schema version must not be empty")
	}
	if schema == "" {
		return fmt.Errorf("schema body must not be empty")
	}
	if !semver.IsValid(version) {
		return fmt.Errorf("invalid semver version: %q", version)
	}

	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.schemas[name]; !ok {
		r.schemas[name] = make(map[string]*SchemaEntry)
	}

	existing, exists := r.schemas[name][version]
	if exists {
		existing.Schema = schema
		existing.UpdatedAt = now
	} else {
		r.schemas[name][version] = &SchemaEntry{
			Name:      name,
			Version:   version,
			Schema:    schema,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	return nil
}

// Get retrieves a specific schema version.
func (r *Registry) Get(name, version string) (*SchemaEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema %q not found", name)
	}

	if version == "" {
		return r.latest(versions)
	}

	entry, ok := versions[version]
	if !ok {
		return nil, fmt.Errorf("schema %q version %q not found", name, version)
	}

	return entry, nil
}

// List returns all registered schema names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.schemas))
	for n := range r.schemas {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Versions returns all versions for a given schema name, sorted descending.
func (r *Registry) Versions(name string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema %q not found", name)
	}

	vs := make([]string, 0, len(versions))
	for v := range versions {
		vs = append(vs, v)
	}
	sort.Slice(vs, func(i, j int) bool {
		return semver.Compare(vs[i], vs[j]) > 0
	})
	return vs, nil
}

// Delete removes a specific schema version.
func (r *Registry) Delete(name, version string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	versions, ok := r.schemas[name]
	if !ok {
		return fmt.Errorf("schema %q not found", name)
	}

	if _, ok := versions[version]; !ok {
		return fmt.Errorf("schema %q version %q not found", name, version)
	}

	delete(versions, version)

	if len(versions) == 0 {
		delete(r.schemas, name)
	}

	return nil
}

func (r *Registry) latest(versions map[string]*SchemaEntry) (*SchemaEntry, error) {
	var latest string
	for v := range versions {
		if latest == "" || semver.Compare(v, latest) > 0 {
			latest = v
		}
	}
	if latest == "" {
		return nil, fmt.Errorf("no versions found")
	}
	return versions[latest], nil
}

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package schemaregistry

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"golang.org/x/mod/semver"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
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
// Storage is tenant-first: identical schema names and versions may coexist
// independently in different tenants.
type Registry struct {
	mu sync.RWMutex
	// schemas[tenantID][name][version]
	schemas map[string]map[string]map[string]*SchemaEntry
}

func New() *Registry {
	return &Registry{schemas: make(map[string]map[string]map[string]*SchemaEntry)}
}

// Register adds or updates a schema version in the legacy, unscoped namespace.
func (r *Registry) Register(name, version, schema string) error {
	return r.RegisterForTenant(name, version, schema, "")
}

// RegisterForTenant adds or updates a schema version within a tenant scope.
func (r *Registry) RegisterForTenant(name, version, schema, tenantID string) error {
	if name == "" {
		return naeoserr.New(naeoserr.ErrValidation, "schema name must not be empty")
	}
	if version == "" {
		return naeoserr.New(naeoserr.ErrValidation, "schema version must not be empty")
	}
	if schema == "" {
		return naeoserr.New(naeoserr.ErrValidation, "schema body must not be empty")
	}
	if !semver.IsValid(version) {
		return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("invalid semver version: %q", version))
	}

	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.schemas[tenantID] == nil {
		r.schemas[tenantID] = make(map[string]map[string]*SchemaEntry)
	}
	if r.schemas[tenantID][name] == nil {
		r.schemas[tenantID][name] = make(map[string]*SchemaEntry)
	}
	if existing := r.schemas[tenantID][name][version]; existing != nil {
		existing.Schema = schema
		existing.UpdatedAt = now
		return nil
	}
	r.schemas[tenantID][name][version] = &SchemaEntry{Name: name, Version: version, Schema: schema, CreatedAt: now, UpdatedAt: now, TenantID: tenantID}
	return nil
}

// cloneEntry ensures callers cannot mutate registry-owned state.
func cloneEntry(entry *SchemaEntry) *SchemaEntry {
	if entry == nil {
		return nil
	}
	copy := *entry
	return &copy
}

// Get retrieves a specific schema version from the legacy, unscoped namespace.
func (r *Registry) Get(name, version string) (*SchemaEntry, error) {
	return r.getForTenant(name, version, "")
}

func (r *Registry) getForTenant(name, version, tenantID string) (*SchemaEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	versions, ok := r.schemas[tenantID][name]
	if !ok {
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("schema %q not found", name))
	}
	if version == "" {
		var latest *SchemaEntry
		for _, entry := range versions {
			if latest == nil || semver.Compare(entry.Version, latest.Version) > 0 {
				latest = entry
			}
		}
		if latest == nil {
			return nil, naeoserr.New(naeoserr.ErrNotFound, "no versions found")
		}
		return cloneEntry(latest), nil
	}
	entry, ok := versions[version]
	if !ok {
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("schema %q version %q not found", name, version))
	}
	return cloneEntry(entry), nil
}

// List returns schema names in the legacy, unscoped namespace.
func (r *Registry) List() []string { return r.listForTenant("") }

func (r *Registry) listForTenant(tenantID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.schemas[tenantID]))
	for name := range r.schemas[tenantID] {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Versions returns all versions for a schema in the legacy, unscoped namespace.
func (r *Registry) Versions(name string) ([]string, error) { return r.versionsForTenant(name, "") }

func (r *Registry) versionsForTenant(name, tenantID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	versions, ok := r.schemas[tenantID][name]
	if !ok { return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("schema %q not found", name)) }
	out := make([]string, 0, len(versions))
	for v := range versions {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return semver.Compare(out[i], out[j]) > 0 })
	return out, nil
}

// Delete removes a specific schema version from the legacy, unscoped namespace.
func (r *Registry) Delete(name, version string) error { return r.deleteForTenant(name, version, "") }

func (r *Registry) deleteForTenant(name, version, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	versions, ok := r.schemas[tenantID][name]
	if !ok {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("schema %q not found", name))
	}
	if _, ok := versions[version]; !ok {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("schema %q version %q not found", name, version))
	}
	delete(versions, version)
	if len(versions) == 0 {
		delete(r.schemas[tenantID], name)
	}
	if len(r.schemas[tenantID]) == 0 {
		delete(r.schemas, tenantID)
	}
	return nil
}

// GetForTenant retrieves a schema version only from the requested tenant namespace.
func (r *Registry) GetForTenant(name, version, tenantID string) (*SchemaEntry, error) {
	return r.getForTenant(name, version, tenantID)
}

// ListForTenant returns schema names registered for tenantID.
func (r *Registry) ListForTenant(tenantID string) []string { return r.listForTenant(tenantID) }

// VersionsForTenant returns only versions belonging to tenantID.
func (r *Registry) VersionsForTenant(name, tenantID string) ([]string, error) {
	return r.versionsForTenant(name, tenantID)
}

// DeleteForTenant removes a version only from the requested tenant namespace.
func (r *Registry) DeleteForTenant(name, version, tenantID string) error {
	return r.deleteForTenant(name, version, tenantID)
}

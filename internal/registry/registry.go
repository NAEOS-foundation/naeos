package registry

import (
	"fmt"
	"path"
	"sync"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type Entry struct {
	Name      string
	Version   string
	Category  string
	Component any
	Metadata  map[string]string
	Tags      []string
}

type HookFunc func(entry *Entry)

type Snapshot struct {
	Entries []SnapshotEntry `json:"entries"`
}

type SnapshotEntry struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Category string            `json:"category"`
	Metadata map[string]string `json:"metadata"`
	Tags     []string          `json:"tags"`
}

type Registry struct {
	mu           sync.RWMutex
	entries      map[string]*Entry
	tags         map[string]map[string]bool
	onRegister   HookFunc
	onUnregister HookFunc
	onResolve    HookFunc
}

func NewRegistry() *Registry {
	return &Registry{
		entries: make(map[string]*Entry),
		tags:    make(map[string]map[string]bool),
	}
}

func (r *Registry) Register(name string, component any) error {
	return r.RegisterWithMeta(name, "", "", component, nil)
}

func (r *Registry) RegisterWithMeta(name, version, category string, component any, metadata map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return naeoserr.New(naeoserr.ErrValidation, "entry name must not be empty")
	}
	if _, exists := r.entries[name]; exists {
		return naeoserr.New(naeoserr.ErrConflict, fmt.Sprintf("entry %s already registered", name))
	}

	entry := &Entry{
		Name:      name,
		Version:   version,
		Category:  category,
		Component: component,
		Metadata:  metadata,
	}
	r.entries[name] = entry

	if r.onRegister != nil {
		r.onRegister(entry)
	}
	return nil
}

func (r *Registry) Resolve(name string) (any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.entries[name]
	if !exists {
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("entry %s not found", name))
	}
	if r.onResolve != nil {
		r.onResolve(entry)
	}
	return entry.Component, nil
}

func (r *Registry) GetEntry(name string) (*Entry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.entries[name]
	if !exists {
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("entry %s not found", name))
	}
	return entry, nil
}

func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.entries[name]
	if !exists {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("entry %s not found", name))
	}

	if r.onUnregister != nil {
		r.onUnregister(entry)
	}

	delete(r.entries, name)
	delete(r.tags, name)
	return nil
}

func (r *Registry) RegisteredEntries() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, 0, len(r.entries))
	for name := range r.entries {
		result = append(result, name)
	}
	return result
}

func (r *Registry) FindByCategory(category string) []*Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Entry
	for _, entry := range r.entries {
		if entry.Category == category {
			result = append(result, entry)
		}
	}
	return result
}

func (r *Registry) FindByVersion(version string) []*Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Entry
	for _, entry := range r.entries {
		if entry.Version == version {
			result = append(result, entry)
		}
	}
	return result
}

func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}

func (r *Registry) Contains(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.entries[name]
	return exists
}

func (r *Registry) OnRegister(hook HookFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onRegister = hook
}

func (r *Registry) OnUnregister(hook HookFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onUnregister = hook
}

func (r *Registry) OnResolve(hook HookFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onResolve = hook
}

func (e *Entry) AddTags(tags ...string) {
	for _, tag := range tags {
		found := false
		for _, t := range e.Tags {
			if t == tag {
				found = true
				break
			}
		}
		if !found {
			e.Tags = append(e.Tags, tag)
		}
	}
}

func (e *Entry) HasTag(tag string) bool {
	for _, t := range e.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (e *Entry) RemoveTags(tags ...string) {
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}
	filtered := e.Tags[:0]
	for _, t := range e.Tags {
		if !tagSet[t] {
			filtered = append(filtered, t)
		}
	}
	e.Tags = filtered
}

func (r *Registry) AddTags(name string, tags ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.entries[name]
	if !exists {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("entry %s not found", name))
	}

	if r.tags[name] == nil {
		r.tags[name] = make(map[string]bool)
	}

	for _, tag := range tags {
		if !r.tags[name][tag] {
			r.tags[name][tag] = true
			entry.AddTags(tag)
		}
	}
	return nil
}

func (r *Registry) FindByTag(tag string) []*Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Entry
	for name, tagSet := range r.tags {
		if tagSet[tag] {
			if entry, exists := r.entries[name]; exists {
				result = append(result, entry)
			}
		}
	}
	return result
}

func (r *Registry) RemoveTags(name string, tags ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.entries[name]
	if !exists {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("entry %s not found", name))
	}

	tagSet := r.tags[name]
	if tagSet != nil {
		for _, tag := range tags {
			delete(tagSet, tag)
		}
	}

	entry.RemoveTags(tags...)
	return nil
}

func (r *Registry) GetTags(name string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.entries[name]
	if !exists {
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("entry %s not found", name))
	}
	result := make([]string, len(entry.Tags))
	copy(result, entry.Tags)
	return result, nil
}

func (r *Registry) FindByPattern(pattern string) []*Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Entry
	for _, entry := range r.entries {
		matched, _ := path.Match(pattern, entry.Name)
		if matched {
			result = append(result, entry)
		}
	}
	return result
}

func (r *Registry) Snapshot() *Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snap := &Snapshot{
		Entries: make([]SnapshotEntry, 0, len(r.entries)),
	}

	for _, entry := range r.entries {
		se := SnapshotEntry{
			Name:     entry.Name,
			Version:  entry.Version,
			Category: entry.Category,
			Tags:     make([]string, len(entry.Tags)),
		}
		copy(se.Tags, entry.Tags)

		if entry.Metadata != nil {
			se.Metadata = make(map[string]string, len(entry.Metadata))
			for k, v := range entry.Metadata {
				se.Metadata[k] = v
			}
		}

		snap.Entries = append(snap.Entries, se)
	}

	return snap
}

func (r *Registry) Restore(snap *Snapshot) error {
	if snap == nil {
		return naeoserr.New(naeoserr.ErrValidation, "snapshot is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries = make(map[string]*Entry)
	r.tags = make(map[string]map[string]bool)

	for _, se := range snap.Entries {
		if se.Name == "" {
			continue
		}

		entry := &Entry{
			Name:     se.Name,
			Version:  se.Version,
			Category: se.Category,
			Tags:     make([]string, len(se.Tags)),
		}
		copy(entry.Tags, se.Tags)

		if se.Metadata != nil {
			entry.Metadata = make(map[string]string, len(se.Metadata))
			for k, v := range se.Metadata {
				entry.Metadata[k] = v
			}
		}

		r.entries[se.Name] = entry

		if len(se.Tags) > 0 {
			tagSet := make(map[string]bool, len(se.Tags))
			for _, tag := range se.Tags {
				tagSet[tag] = true
			}
			r.tags[se.Name] = tagSet
		}
	}

	return nil
}

func (r *Registry) Replace(name, version, category string, component any, metadata map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.entries[name]
	if !exists {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("entry %s not found", name))
	}

	entry.Version = version
	entry.Category = category
	entry.Component = component

	if metadata != nil {
		entry.Metadata = make(map[string]string, len(metadata))
		for k, v := range metadata {
			entry.Metadata[k] = v
		}
	} else {
		entry.Metadata = nil
	}

	return nil
}

func (r *Registry) FindByMetadata(criteria map[string]string) []*Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Entry
	for _, entry := range r.entries {
		match := true
		for k, v := range criteria {
			if entry.Metadata[k] != v {
				match = false
				break
			}
		}
		if match {
			result = append(result, entry)
		}
	}
	return result
}

func (r *Registry) ForEach(fn func(name string, entry *Entry) bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, entry := range r.entries {
		if !fn(name, entry) {
			break
		}
	}
}

package search

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Search Engine Interface

type Engine interface {
	Name() string
	Index(doc *Document) error
	BulkIndex(docs []*Document) error
	Search(query *Query) (*SearchResult, error)
	Delete(id string) error
	DeleteByQuery(query *Query) (int, error)
	Update(id string, doc *Document) error
	GetByID(id string) (*Document, error)
	Count() int
	Close() error
}

type Document struct {
	ID        string
	Index     string
	Title     string
	Content   string
	Tags      []string
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Query struct {
	Text      string
	Index     string
	Tags      []string
	Limit     int
	Offset    int
	Filters   map[string]any
	SortBy    string
	SortOrder string
}

type SearchResult struct {
	Total   int
	Hits    []*SearchHit
	Took    time.Duration
	Query   string
}

type SearchHit struct {
	Document *Document
	Score    float64
	Highlights map[string][]string
}

// In-Memory Search Engine

type InMemory struct {
	documents map[string]*Document
	mu        sync.RWMutex
}

func NewInMemory() *InMemory {
	return &InMemory{
		documents: make(map[string]*Document),
	}
}

func (e *InMemory) Name() string {
	return "inmemory"
}

func (e *InMemory) Index(doc *Document) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	doc.UpdatedAt = time.Now()
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now()
	}

	e.documents[doc.ID] = doc
	return nil
}

func (e *InMemory) BulkIndex(docs []*Document) error {
	for _, doc := range docs {
		if err := e.Index(doc); err != nil {
			return err
		}
	}
	return nil
}

func (e *InMemory) Search(query *Query) (*SearchResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	start := time.Now()

	var hits []*SearchHit
	for _, doc := range e.documents {
		if e.matchesQuery(doc, query) {
			score := e.calculateScore(doc, query)
			hit := &SearchHit{
				Document: doc,
				Score:    score,
			}
			hits = append(hits, hit)
		}
	}

	// Sort by score
	sort.Slice(hits, func(i, j int) bool {
		return hits[i].Score > hits[j].Score
	})

	// Apply limit and offset
	if query.Limit > 0 {
		if query.Offset >= len(hits) {
			hits = []*SearchHit{}
		} else {
			end := query.Offset + query.Limit
			if end > len(hits) {
				end = len(hits)
			}
			hits = hits[query.Offset:end]
		}
	}

	return &SearchResult{
		Total: len(hits),
		Hits:  hits,
		Took:  time.Since(start),
		Query: query.Text,
	}, nil
}

func (e *InMemory) matchesQuery(doc *Document, query *Query) bool {
	if query.Index != "" && doc.Index != query.Index {
		return false
	}

	if query.Text != "" {
		text := strings.ToLower(doc.Title + " " + doc.Content)
		if !strings.Contains(text, strings.ToLower(query.Text)) {
			return false
		}
	}

	if len(query.Tags) > 0 {
		docTags := make(map[string]bool)
		for _, tag := range doc.Tags {
			docTags[tag] = true
		}
		for _, tag := range query.Tags {
			if !docTags[tag] {
				return false
			}
		}
	}

	return true
}

func (e *InMemory) calculateScore(doc *Document, query *Query) float64 {
	score := 0.0

	if query.Text != "" {
		text := strings.ToLower(doc.Title + " " + doc.Content)
		queryLower := strings.ToLower(query.Text)

		if strings.Contains(strings.ToLower(doc.Title), queryLower) {
			score += 10.0
		}

		if strings.Contains(text, queryLower) {
			score += 5.0
		}

		score += float64(strings.Count(text, queryLower))
	}

	return score
}

func (e *InMemory) Delete(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.documents[id]; ok {
		delete(e.documents, id)
		return nil
	}
	return fmt.Errorf("document not found: %s", id)
}

func (e *InMemory) DeleteByQuery(query *Query) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	deleted := 0
	for id, doc := range e.documents {
		if e.matchesQuery(doc, query) {
			delete(e.documents, id)
			deleted++
		}
	}
	return deleted, nil
}

func (e *InMemory) Update(id string, doc *Document) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.documents[id]; !ok {
		return fmt.Errorf("document not found: %s", id)
	}

	doc.ID = id
	doc.UpdatedAt = time.Now()
	e.documents[id] = doc
	return nil
}

func (e *InMemory) GetByID(id string) (*Document, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	doc, ok := e.documents[id]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", id)
	}
	return doc, nil
}

func (e *InMemory) Count() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.documents)
}

func (e *InMemory) Close() error {
	return nil
}

// Search Manager

type Manager struct {
	engines map[string]Engine
	mu      sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		engines: make(map[string]Engine),
	}
}

// Persistent wraps InMemory with JSON file persistence.
type Persistent struct {
	*InMemory
	filePath string
}

func NewPersistent(dir string) (*Persistent, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create search dir: %w", err)
	}
	p := &Persistent{
		InMemory: NewInMemory(),
		filePath: filepath.Join(dir, "search-index.json"),
	}
	p.load()
	return p, nil
}

func (p *Persistent) save() error {
	p.mu.RLock()
	docs := make([]*Document, 0, len(p.documents))
	for _, doc := range p.documents {
		docs = append(docs, doc)
	}
	p.mu.RUnlock()

	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.filePath, data, 0o644)
}

func (p *Persistent) load() {
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return
	}
	var docs []*Document
	if err := json.Unmarshal(data, &docs); err != nil {
		return
	}
	for _, doc := range docs {
		_ = p.InMemory.Index(doc)
	}
}

func (p *Persistent) Index(doc *Document) error {
	if err := p.InMemory.Index(doc); err != nil {
		return err
	}
	return p.save()
}

func (p *Persistent) Delete(id string) error {
	if err := p.InMemory.Delete(id); err != nil {
		return err
	}
	return p.save()
}

func (p *Persistent) Update(id string, doc *Document) error {
	if err := p.InMemory.Update(id, doc); err != nil {
		return err
	}
	return p.save()
}

func (m *Manager) Register(name string, engine Engine) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.engines[name] = engine
}

func (m *Manager) Get(name string) (Engine, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	engine, ok := m.engines[name]
	return engine, ok
}

func (m *Manager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.engines, name)
}

func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.engines))
	for name := range m.engines {
		names = append(names, name)
	}
	return names
}

func (m *Manager) CloseAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, engine := range m.engines {
		if err := engine.Close(); err != nil {
			return fmt.Errorf("failed to close %s: %w", name, err)
		}
	}
	return nil
}

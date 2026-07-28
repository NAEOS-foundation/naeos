package provenance

import (
	"fmt"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type ProvenanceRecord struct {
	ID            string
	Source        string
	Version       string
	ArtifactID    string
	NEIRReference string
	CreatedBy     string
	Timestamp     time.Time
	PolicyContext string
	ParentID      string
	Metadata      map[string]string
}

type ProvenanceStore struct {
	mu      sync.RWMutex
	records map[string]*ProvenanceRecord
}

func NewStore() *ProvenanceStore {
	return &ProvenanceStore{
		records: make(map[string]*ProvenanceRecord),
	}
}

func (ps *ProvenanceStore) Record(record ProvenanceRecord) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if record.ID == "" {
		return naeoserr.New(naeoserr.ErrValidation, "record ID must not be empty")
	}
	if _, exists := ps.records[record.ID]; exists {
		return naeoserr.New(naeoserr.ErrConflict, fmt.Sprintf("record %s already exists", record.ID))
	}

	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	ps.records[record.ID] = &record
	return nil
}

func (ps *ProvenanceStore) Get(id string) (*ProvenanceRecord, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	record, ok := ps.records[id]
	return record, ok
}

func (ps *ProvenanceStore) FindByArtifact(artifactID string) []*ProvenanceRecord {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var result []*ProvenanceRecord
	for _, r := range ps.records {
		if r.ArtifactID == artifactID {
			result = append(result, r)
		}
	}
	return result
}

func (ps *ProvenanceStore) FindBySource(source string) []*ProvenanceRecord {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var result []*ProvenanceRecord
	for _, r := range ps.records {
		if r.Source == source {
			result = append(result, r)
		}
	}
	return result
}

func (ps *ProvenanceStore) FindByCreator(createdBy string) []*ProvenanceRecord {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var result []*ProvenanceRecord
	for _, r := range ps.records {
		if r.CreatedBy == createdBy {
			result = append(result, r)
		}
	}
	return result
}

func (ps *ProvenanceStore) Lineage(id string) ([]*ProvenanceRecord, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var lineage []*ProvenanceRecord
	currentID := id

	for currentID != "" {
		record, exists := ps.records[currentID]
		if !exists {
			return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("record %s not found in lineage", currentID))
		}
		lineage = append(lineage, record)
		currentID = record.ParentID
	}

	return lineage, nil
}

func (ps *ProvenanceStore) Count() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return len(ps.records)
}

func (ps *ProvenanceStore) Records() []*ProvenanceRecord {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	result := make([]*ProvenanceRecord, 0, len(ps.records))
	for _, r := range ps.records {
		result = append(result, r)
	}
	return result
}

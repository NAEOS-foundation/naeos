// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package provenance

import (
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	s := NewStore()
	if s == nil {
		t.Fatal("expected non-nil store")
	}
	if s.Count() != 0 {
		t.Fatalf("expected 0 records, got %d", s.Count())
	}
}

func TestRecord(t *testing.T) {
	s := NewStore()
	err := s.Record(ProvenanceRecord{
		ID:        "r1",
		Source:    "spec.yaml",
		Version:   "1.0",
		CreatedBy: "pipeline",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Count() != 1 {
		t.Fatalf("expected 1 record, got %d", s.Count())
	}
}

func TestRecordEmptyID(t *testing.T) {
	s := NewStore()
	err := s.Record(ProvenanceRecord{ID: ""})
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestRecordDuplicate(t *testing.T) {
	s := NewStore()
	_ = s.Record(ProvenanceRecord{ID: "r1", Source: "test"})
	err := s.Record(ProvenanceRecord{ID: "r1", Source: "test"})
	if err == nil {
		t.Fatal("expected error for duplicate record")
	}
}

func TestRecordAutoTimestamp(t *testing.T) {
	s := NewStore()
	_ = s.Record(ProvenanceRecord{ID: "r1", Source: "test"})
	record, ok := s.Get("r1")
	if !ok {
		t.Fatal("expected to find record")
	}
	if record.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
}

func TestGet(t *testing.T) {
	s := NewStore()
	_ = s.Record(ProvenanceRecord{ID: "r1", Source: "test", Version: "1.0"})
	record, ok := s.Get("r1")
	if !ok {
		t.Fatal("expected to find record")
	}
	if record.Source != "test" {
		t.Fatalf("expected source 'test', got %s", record.Source)
	}
	if record.Version != "1.0" {
		t.Fatalf("expected version '1.0', got %s", record.Version)
	}
}

func TestFindByArtifact(t *testing.T) {
	s := NewStore()
	_ = s.Record(ProvenanceRecord{ID: "r1", ArtifactID: "a1", Source: "test"})
	_ = s.Record(ProvenanceRecord{ID: "r2", ArtifactID: "a1", Source: "test2"})
	_ = s.Record(ProvenanceRecord{ID: "r3", ArtifactID: "a2", Source: "test"})

	records := s.FindByArtifact("a1")
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
}

func TestFindBySource(t *testing.T) {
	s := NewStore()
	_ = s.Record(ProvenanceRecord{ID: "r1", Source: "spec.yaml"})
	_ = s.Record(ProvenanceRecord{ID: "r2", Source: "spec.yaml"})
	_ = s.Record(ProvenanceRecord{ID: "r3", Source: "other.yaml"})

	records := s.FindBySource("spec.yaml")
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
}

func TestFindByCreator(t *testing.T) {
	s := NewStore()
	_ = s.Record(ProvenanceRecord{ID: "r1", CreatedBy: "pipeline"})
	_ = s.Record(ProvenanceRecord{ID: "r2", CreatedBy: "user"})
	_ = s.Record(ProvenanceRecord{ID: "r3", CreatedBy: "pipeline"})

	records := s.FindByCreator("pipeline")
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
}

func TestLineage(t *testing.T) {
	s := NewStore()
	_ = s.Record(ProvenanceRecord{ID: "r1", Source: "spec.yaml"})
	_ = s.Record(ProvenanceRecord{ID: "r2", Source: "neir", ParentID: "r1"})
	_ = s.Record(ProvenanceRecord{ID: "r3", Source: "artifact", ParentID: "r2"})

	lineage, err := s.Lineage("r3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lineage) != 3 {
		t.Fatalf("expected 3 records in lineage, got %d", len(lineage))
	}
	if lineage[0].ID != "r3" || lineage[2].ID != "r1" {
		t.Fatalf("lineage order incorrect: %v", lineage)
	}
}

func TestLineageNotFound(t *testing.T) {
	s := NewStore()
	_, err := s.Lineage("missing")
	if err == nil {
		t.Fatal("expected error for missing record in lineage")
	}
}

func TestRecords(t *testing.T) {
	s := NewStore()
	_ = s.Record(ProvenanceRecord{ID: "r1", Source: "test1"})
	_ = s.Record(ProvenanceRecord{ID: "r2", Source: "test2"})

	records := s.Records()
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
}

func TestRecordWithMetadata(t *testing.T) {
	s := NewStore()
	err := s.Record(ProvenanceRecord{
		ID:       "r1",
		Source:   "spec.yaml",
		Metadata: map[string]string{"env": "production"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	record, _ := s.Get("r1")
	if record.Metadata["env"] != "production" {
		t.Fatalf("expected metadata env=production, got %v", record.Metadata)
	}
}

func TestRecordWithExplicitTimestamp(t *testing.T) {
	s := NewStore()
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	err := s.Record(ProvenanceRecord{
		ID:        "r1",
		Source:    "test",
		Timestamp: ts,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	record, _ := s.Get("r1")
	if !record.Timestamp.Equal(ts) {
		t.Fatalf("expected explicit timestamp, got %v", record.Timestamp)
	}
}

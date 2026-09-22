// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// RuntimeEvent is an observed lifecycle event that evidence may bind to.
// The event identity, run identity, event name, sequence, and payload digest
// are immutable for the lifetime of the in-memory store.
type RuntimeEvent struct {
	ID           string
	RunID        string
	Name         string
	PayloadDigest string
	Sequence     int
	Timestamp    time.Time
}

// RuntimeEventStore records the runtime observations used by the completion
// boundary. It is deliberately separate from EvidenceStore so an evidence
// record cannot manufacture an event reference during validation.
type RuntimeEventStore struct {
	mu     sync.RWMutex
	events []RuntimeEvent
}

// NewRuntimeEventStore creates an empty runtime event store.
func NewRuntimeEventStore() *RuntimeEventStore {
	return &RuntimeEventStore{}
}

// Append records one observed runtime event and returns its deterministic ID.
func (s *RuntimeEventStore) Append(runID, name, payloadDigest string, sequence int) (RuntimeEvent, error) {
	if s == nil {
		return RuntimeEvent{}, fmt.Errorf("runtime event store is nil")
	}
	if runID == "" || name == "" || payloadDigest == "" || sequence <= 0 {
		return RuntimeEvent{}, fmt.Errorf("runtime event requires run_id, name, payload_digest, and positive sequence")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	h := sha256.Sum256([]byte(fmt.Sprintf("naeos:runtime-event:v1:%s:%s:%s:%d", runID, name, payloadDigest, sequence)))
	event := RuntimeEvent{
		ID:            fmt.Sprintf("evt-%x", h[:]),
		RunID:         runID,
		Name:          name,
		PayloadDigest: payloadDigest,
		Sequence:      sequence,
		Timestamp:     time.Now().UTC(),
	}
	s.events = append(s.events, event)
	return event, nil
}

// Records returns runtime events in append order.
func (s *RuntimeEventStore) Records() []RuntimeEvent {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]RuntimeEvent, len(s.events))
	copy(out, s.events)
	return out
}

// ByID returns a runtime event by its immutable event ID.
func (s *RuntimeEventStore) ByID(id string) *RuntimeEvent {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.events {
		if s.events[i].ID == id {
			event := s.events[i]
			return &event
		}
	}
	return nil
}

// Verify checks event identities against their canonical contents.
func (s *RuntimeEventStore) Verify() error {
	if s == nil {
		return fmt.Errorf("runtime event store is nil")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i, event := range s.events {
		h := sha256.Sum256([]byte(fmt.Sprintf("naeos:runtime-event:v1:%s:%s:%s:%d", event.RunID, event.Name, event.PayloadDigest, event.Sequence)))
		expected := fmt.Sprintf("evt-%x", h[:])
		if event.ID != expected {
			return fmt.Errorf("runtime event identity mismatch at index %d", i)
		}
	}
	return nil
}

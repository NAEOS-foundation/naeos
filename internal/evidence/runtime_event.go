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
	ID            string
	RunID         string
	Name          string
	PayloadDigest string
	Sequence      int
	Timestamp     time.Time
}

// RuntimeEventProducer is the lifecycle-facing capability for publishing
// runtime observations. Evidence builders consume the ledger separately.
type RuntimeEventProducer interface {
	Publish(runID, name, payloadDigest string, sequence int) (RuntimeEvent, error)
}

// RuntimeEventLedger is the append-only runtime observation boundary.
// Producers publish events here; evidence code cannot create event identities.
type RuntimeEventLedger struct {
	store  *RuntimeEventStore
	mu     sync.RWMutex
	sealed bool
}

// NewRuntimeEventLedger creates an independent runtime event ledger.
func NewRuntimeEventLedger() *RuntimeEventLedger {
	return &RuntimeEventLedger{store: NewRuntimeEventStore()}
}

// Publish records an event before any evidence record can reference it.
func (l *RuntimeEventLedger) Publish(runID, name, payloadDigest string, sequence int) (RuntimeEvent, error) {
	if l == nil {
		return RuntimeEvent{}, fmt.Errorf("runtime event ledger is nil")
	}
	l.mu.RLock()
	sealed := l.sealed
	l.mu.RUnlock()
	if sealed {
		return RuntimeEvent{}, fmt.Errorf("runtime event ledger is sealed")
	}
	return l.store.Append(runID, name, payloadDigest, sequence)
}

// Seal closes the observation window. Events published after the boundary are rejected.
func (l *RuntimeEventLedger) Seal() {
	if l == nil {
		return
	}
	l.mu.Lock()
	l.sealed = true
	l.mu.Unlock()
}

// Records returns the ledger snapshot in append order.
func (l *RuntimeEventLedger) Records() []RuntimeEvent {
	if l == nil || l.store == nil {
		return nil
	}
	return l.store.Records()
}

// Verify validates all event identities in the ledger.
func (l *RuntimeEventLedger) Verify() error {
	if l == nil || l.store == nil {
		return fmt.Errorf("runtime event ledger is nil")
	}
	return l.store.Verify()
}

// ByID resolves an event without granting append access.
func (l *RuntimeEventLedger) ByID(id string) *RuntimeEvent {
	if l == nil || l.store == nil {
		return nil
	}
	return l.store.ByID(id)
}

// RuntimeEventStore is the backing immutable event record store.
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

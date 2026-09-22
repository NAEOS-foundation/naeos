// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// DurableRuntimeEventLedger persists canonical runtime events as an append-only
// JSONL journal. Each record is verified before it is accepted.
type DurableRuntimeEventLedger struct {
	mu     sync.Mutex
	path   string
	events []RuntimeEvent
	sealed bool
}

// NewDurableRuntimeEventLedger opens or creates a durable ledger and verifies
// every existing record before returning.
func NewDurableRuntimeEventLedger(path string) (*DurableRuntimeEventLedger, error) {
	if path == "" {
		return nil, fmt.Errorf("durable runtime ledger path is required")
	}
	l := &DurableRuntimeEventLedger{path: filepath.Clean(path)}
	if err := l.load(); err != nil {
		return nil, err
	}
	return l, nil
}

// Publish appends one event and flushes it to durable storage before returning.
func (l *DurableRuntimeEventLedger) Publish(runID, name, payloadDigest string, sequence int) (RuntimeEvent, error) {
	if l == nil {
		return RuntimeEvent{}, fmt.Errorf("durable runtime ledger is nil")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.sealed {
		return RuntimeEvent{}, fmt.Errorf("durable runtime ledger is sealed")
	}
	event, err := canonicalRuntimeEvent(runID, name, payloadDigest, sequence)
	if err != nil {
		return RuntimeEvent{}, err
	}
	if err := l.appendLocked(event); err != nil {
		return RuntimeEvent{}, err
	}
	l.events = append(l.events, event)
	return event, nil
}

// Seal prevents subsequent writes.
func (l *DurableRuntimeEventLedger) Seal() {
	if l != nil {
		l.mu.Lock()
		l.sealed = true
		l.mu.Unlock()
	}
}

// Records returns a defensive snapshot.
func (l *DurableRuntimeEventLedger) Records() []RuntimeEvent {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]RuntimeEvent, len(l.events))
	copy(out, l.events)
	return out
}

// ByID returns a defensive copy of an event.
func (l *DurableRuntimeEventLedger) ByID(id string) *RuntimeEvent {
	for _, event := range l.Records() {
		if event.ID == id {
			copy := event
			return &copy
		}
	}
	return nil
}

// Verify revalidates the complete durable journal against canonical event IDs.
func (l *DurableRuntimeEventLedger) Verify() error {
	if l == nil {
		return fmt.Errorf("durable runtime ledger is nil")
	}
	records := l.Records()
	for i, event := range records {
		expected, err := canonicalRuntimeEvent(event.RunID, event.Name, event.PayloadDigest, event.Sequence)
		if err != nil {
			return fmt.Errorf("ledger record %d invalid: %w", i, err)
		}
		if expected.ID != event.ID {
			return fmt.Errorf("ledger record %d identity mismatch", i)
		}
		if i > 0 && event.Sequence != records[i-1].Sequence+1 {
			return fmt.Errorf("ledger record %d sequence is not contiguous", i)
		}
	}
	return nil
}

func (l *DurableRuntimeEventLedger) load() error {
	data, err := os.ReadFile(l.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read durable runtime ledger: %w", err)
	}
	lines := splitLines(data)
	for i, line := range lines {
		var event RuntimeEvent
		if err := json.Unmarshal(line, &event); err != nil {
			return fmt.Errorf("decode durable runtime ledger record %d: %w", i, err)
		}
		expected, err := canonicalRuntimeEvent(event.RunID, event.Name, event.PayloadDigest, event.Sequence)
		if err != nil || expected.ID != event.ID {
			return fmt.Errorf("durable runtime ledger record %d failed integrity verification", i)
		}
		if i > 0 && event.Sequence != l.events[i-1].Sequence+1 {
			return fmt.Errorf("durable runtime ledger record %d sequence mismatch", i)
		}
		l.events = append(l.events, event)
	}
	return nil
}

func (l *DurableRuntimeEventLedger) appendLocked(event RuntimeEvent) error {
	if err := os.MkdirAll(filepath.Dir(l.path), 0o750); err != nil {
		return fmt.Errorf("create durable runtime ledger directory: %w", err)
	}
	line, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode runtime event: %w", err)
	}
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return fmt.Errorf("open durable runtime ledger: %w", err)
	}
	defer file.Close()
	if _, err := file.Write(append(line, byte(10))); err != nil {
		return fmt.Errorf("append durable runtime ledger: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync durable runtime ledger: %w", err)
	}
	return nil
}

func canonicalRuntimeEvent(runID, name, payloadDigest string, sequence int) (RuntimeEvent, error) {
	if runID == "" || name == "" || payloadDigest == "" || sequence <= 0 {
		return RuntimeEvent{}, fmt.Errorf("runtime event requires run_id, name, payload_digest, and positive sequence")
	}
	h := sha256.Sum256([]byte(fmt.Sprintf("naeos:runtime-event:v1:%s:%s:%s:%d", runID, name, payloadDigest, sequence)))
	return RuntimeEvent{
		ID:            fmt.Sprintf("evt-%x", h[:]),
		RunID:         runID,
		Name:          name,
		PayloadDigest: payloadDigest,
		Sequence:      sequence,
	}, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == byte(10) {
			if i > start {
				lines = append(lines, data[start:i])
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

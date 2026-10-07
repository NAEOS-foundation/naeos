// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type HashedAuditor struct {
	inner  Auditor
	lastID string
	mu     sync.Mutex
}

func NewHashedAuditor(inner Auditor) *HashedAuditor {
	h := &HashedAuditor{inner: inner}
	if file, ok := inner.(*FileAuditor); ok {
		if violations, err := VerifyChainFile(file.path); err == nil && len(violations) == 0 {
			if events, err := readAuditEvents(file.path); err == nil && len(events) > 0 {
				h.lastID = events[len(events)-1].Hash
			}
		}
	}
	return h
}

func readAuditEvents(path string) ([]AuditEvent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	events := make([]AuditEvent, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var event AuditEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (h *HashedAuditor) Log(event AuditEvent) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if event.ID == "" {
		event.ID = generateID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	event.PreviousHash = h.lastID
	event.Hash = computeHash(event)

	h.lastID = event.Hash

	return h.inner.Log(event)
}

func computeHash(event AuditEvent) string {
	event.Hash = ""
	data, _ := json.Marshal(event)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func VerifyChain(events []AuditEvent) []string {
	var violations []string
	var prevHash string

	for i, e := range events {
		if e.PreviousHash != prevHash {
			violations = append(violations,
				fmt.Sprintf("event[%d] (ID: %s): expected previous_hash %q, got %q",
					i, e.ID, prevHash, e.PreviousHash))
		}

		computed := computeHash(e)
		if computed != e.Hash {
			violations = append(violations,
				fmt.Sprintf("event[%d] (ID: %s): hash mismatch — event may be tampered",
					i, e.ID))
		}

		prevHash = e.Hash
	}

	return violations
}

func VerifyChainFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrNotFound, "read audit file")
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var events []AuditEvent
	for _, line := range lines {
		if line == "" {
			continue
		}
		var e AuditEvent
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "parse audit line")
		}
		events = append(events, e)
	}

	return VerifyChain(events), nil
}

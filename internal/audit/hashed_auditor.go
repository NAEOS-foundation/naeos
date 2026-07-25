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
	return &HashedAuditor{inner: inner}
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

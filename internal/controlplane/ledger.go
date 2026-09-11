package controlplane

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LedgerEvent is an append-only record of important control-plane decisions.
type LedgerEvent struct {
	ID           string            `json:"id"`
	Timestamp    time.Time         `json:"timestamp"`
	RequestID    string            `json:"request_id,omitempty"`
	DecisionID   string            `json:"decision_id,omitempty"`
	ExecutionID  string            `json:"execution_id,omitempty"`
	AgentID      string            `json:"agent_id"`
	Capability   Capability        `json:"capability,omitempty"`
	ArtifactHash string            `json:"artifact_hash,omitempty"`
	EventType    string            `json:"event_type"`
	Decision     DecisionStatus    `json:"decision,omitempty"`
	Reason       DecisionReason    `json:"reason,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	PreviousHash string            `json:"previous_hash,omitempty"`
	EventHash    string            `json:"event_hash,omitempty"`
}

// Ledger is an append-only event log for decisions and executions.
type Ledger struct {
	mu                   sync.RWMutex
	events               []LedgerEvent
	nextID               int
	persistencePath      string
	persistenceMu        sync.Mutex
	lastPersistenceError error
}

// NewLedger creates a fresh append-only ledger.
func NewLedger() *Ledger {
	return &Ledger{events: make([]LedgerEvent, 0)}
}

// SetPersistencePath enables snapshot persistence after appends.
func (l *Ledger) SetPersistencePath(path string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	l.persistencePath = path
	l.mu.Unlock()
}

// PersistenceError returns the latest snapshot error, if any.
func (l *Ledger) PersistenceError() error {
	if l == nil {
		return fmt.Errorf("ledger unavailable")
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.lastPersistenceError
}

// Save persists the ledger as a JSON snapshot using an atomic rename.
func (l *Ledger) Save(path string) error {
	if l == nil {
		return fmt.Errorf("ledger unavailable")
	}
	if path == "" {
		return fmt.Errorf("ledger path is required")
	}
	data, err := json.Marshal(l.Events())
	if err != nil {
		return fmt.Errorf("marshal ledger: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create ledger directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".ledger-*.tmp")
	if err != nil {
		return fmt.Errorf("create ledger snapshot: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("protect ledger snapshot: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write ledger snapshot: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync ledger snapshot: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close ledger snapshot: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("commit ledger snapshot: %w", err)
	}
	return nil
}

// LoadLedger restores a ledger snapshot and rejects malformed JSON.
func LoadLedger(path string) (*Ledger, error) {
	if path == "" {
		return nil, fmt.Errorf("ledger path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ledger snapshot: %w", err)
	}
	var events []LedgerEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("decode ledger snapshot: %w", err)
	}
	ledger := NewLedger()
	for _, event := range events {
		if event.ID == "" || event.Timestamp.IsZero() {
			return nil, fmt.Errorf("invalid ledger event %q", event.ID)
		}
		if event.EventHash == "" || event.EventHash != hashLedgerEvent(event) {
			return nil, fmt.Errorf("invalid ledger event hash %q", event.ID)
		}
		if len(ledger.events) == 0 {
			if event.PreviousHash != "" {
				return nil, fmt.Errorf("invalid initial ledger link %q", event.ID)
			}
		} else if event.PreviousHash != ledger.events[len(ledger.events)-1].EventHash {
			return nil, fmt.Errorf("broken ledger hash chain at %q", event.ID)
		}
		ledger.events = append(ledger.events, event)
		ledger.nextID++
	}
	return ledger, nil
}

// Append adds an event to the ledger. It is always append-only and uses a monotonic event ID.
func (l *Ledger) Append(event LedgerEvent) LedgerEvent {
	l.mu.Lock()
	l.nextID++
	if event.ID == "" {
		event.ID = fmt.Sprintf("EVT-%05d", l.nextID)
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if len(l.events) > 0 {
		event.PreviousHash = l.events[len(l.events)-1].EventHash
	}
	event.EventHash = hashLedgerEvent(event)
	l.events = append(l.events, event)
	path := l.persistencePath
	l.mu.Unlock()
	if path != "" {
		if err := l.Save(path); err != nil {
			l.mu.Lock()
			l.lastPersistenceError = err
			l.mu.Unlock()
		} else {
			l.mu.Lock()
			l.lastPersistenceError = nil
			l.mu.Unlock()
		}
	}
	return event
}

func hashLedgerEvent(event LedgerEvent) string {
	event.EventHash = ""
	data, _ := json.Marshal(event)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// Events returns a snapshot of the ledger.
func (l *Ledger) Events() []LedgerEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]LedgerEvent, len(l.events))
	copy(out, l.events)
	return out
}

// EventsForAgent returns all events for a specific agent.
func (l *Ledger) EventsForAgent(agentID string) []LedgerEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]LedgerEvent, 0)
	for _, event := range l.events {
		if event.AgentID == agentID {
			out = append(out, event)
		}
	}
	return out
}

// Query returns evidence matching non-empty correlation filters.
func (l *Ledger) Query(filters map[string]string) []LedgerEvent {
	events := l.Events()
	out := make([]LedgerEvent, 0, len(events))
	for _, event := range events {
		if filters["agent_id"] != "" && event.AgentID != filters["agent_id"] {
			continue
		}
		if filters["request_id"] != "" && event.RequestID != filters["request_id"] {
			continue
		}
		if filters["decision_id"] != "" && event.DecisionID != filters["decision_id"] {
			continue
		}
		if filters["execution_id"] != "" && event.ExecutionID != filters["execution_id"] {
			continue
		}
		if filters["event_type"] != "" && event.EventType != filters["event_type"] {
			continue
		}
		out = append(out, event)
	}
	return out
}

// Decision returns the canonical authorization event for a decision ID.
func (l *Ledger) Decision(decisionID string) (LedgerEvent, bool) {
	if l == nil || decisionID == "" {
		return LedgerEvent{}, false
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	for i := len(l.events) - 1; i >= 0; i-- {
		event := l.events[i]
		if event.EventType == "AUTHORIZATION_DECISION" && event.DecisionID == decisionID {
			return event, true
		}
	}
	return LedgerEvent{}, false
}

// HasExecution reports whether a decision already produced an execution event.
func (l *Ledger) HasExecution(decisionID string) bool {
	if l == nil || decisionID == "" {
		return false
	}
	for _, event := range l.Events() {
		if event.DecisionID == decisionID && event.EventType == "EXECUTION_ALLOWED" {
			return true
		}
	}
	return false
}

// VerificationSummary is the result of a session verification against the evidence ledger.
type VerificationSummary struct {
	AgentID                string   `json:"agent_id"`
	Result                 string   `json:"result"`
	PolicyCompliant        bool     `json:"policy_compliant"`
	Unauthorized           int      `json:"unauthorized"`
	BlockedAttempts        int      `json:"blocked_attempts"`
	UnauthorizedExecutions int      `json:"unauthorized_executions"`
	Issues                 []string `json:"issues,omitempty"`
}

// VerifySession inspects ledger events for unauthorized or blocked actions.
func (l *Ledger) VerifySession(agentID string) VerificationSummary {
	summary := VerificationSummary{
		AgentID:         agentID,
		Result:          "PASS",
		PolicyCompliant: true,
		Issues:          []string{},
	}
	for _, event := range l.EventsForAgent(agentID) {
		if event.Decision == DecisionDeny || event.EventType == "EXECUTION_BLOCKED" {
			summary.BlockedAttempts++
			summary.Issues = append(summary.Issues, fmt.Sprintf("blocked attempt %s for %s: %s", event.ID, event.Capability, event.Reason))
		}
		if event.EventType == "EXECUTION_ALLOWED" && event.Decision != DecisionAllow {
			summary.Unauthorized++
			summary.UnauthorizedExecutions++
			summary.PolicyCompliant = false
			summary.Result = "FAIL"
			summary.Issues = append(summary.Issues, fmt.Sprintf("unauthorized execution evidence %s for %s", event.ID, event.Capability))
		}
	}
	if summary.PolicyCompliant && len(summary.Issues) == 0 {
		summary.Issues = nil
	}
	return summary
}

package controlplane

import (
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
}

// Ledger is an append-only event log for decisions and executions.
type Ledger struct {
	mu     sync.RWMutex
	events []LedgerEvent
	nextID int
}

// NewLedger creates a fresh append-only ledger.
func NewLedger() *Ledger {
	return &Ledger{events: make([]LedgerEvent, 0)}
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
		ledger.events = append(ledger.events, event)
		ledger.nextID++
	}
	return ledger, nil
}

// Append adds an event to the ledger. It is always append-only and uses a monotonic event ID.
func (l *Ledger) Append(event LedgerEvent) LedgerEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nextID++
	if event.ID == "" {
		event.ID = fmt.Sprintf("EVT-%05d", l.nextID)
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	l.events = append(l.events, event)
	return event
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

// Decision returns the canonical authorization event for a decision ID.
func (l *Ledger) Decision(decisionID string) (LedgerEvent, bool) {
	if l == nil || decisionID == "" {
		return LedgerEvent{}, false
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	var latest LedgerEvent
	for i := len(l.events) - 1; i >= 0; i-- {
		event := l.events[i]
		if event.EventType == "AUTHORIZATION_DECISION" && event.DecisionID == decisionID {
			if latest.ID == "" {
				latest = event
			}
			if event.Decision == DecisionPending {
				return event, true
			}
		}
	}
	if latest.ID != "" {
		return latest, true
	}
	return LedgerEvent{}, false
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

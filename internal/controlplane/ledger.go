package controlplane

import (
	"fmt"
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

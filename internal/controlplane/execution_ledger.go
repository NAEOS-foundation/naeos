// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"fmt"
	"time"
)

// ReconcileExecution records an explicit, append-only operator assertion for an
// ambiguous execution. Evidence is a reference to supporting material, not proof
// by itself; callers must authenticate and authorize the reconciler separately.
func (l *Ledger) ReconcileExecution(executionID string, outcome ExecutionOutcome, evidenceRef, reconcilerID string) (LedgerEvent, error) {
	if l == nil {
		return LedgerEvent{}, fmt.Errorf("ledger unavailable")
	}
	if executionID == "" || evidenceRef == "" || reconcilerID == "" {
		return LedgerEvent{}, fmt.Errorf("execution ID, evidence reference, and reconciler identity are required")
	}
	if outcome != ExecutionOutcomeCommitted && outcome != ExecutionOutcomeNotCommitted {
		return LedgerEvent{}, fmt.Errorf("unsupported reconciliation outcome %q", outcome)
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	var start LedgerEvent
	found := false
	unknown := false
	for _, event := range l.events {
		if event.ExecutionID != executionID {
			continue
		}
		switch event.EventType {
		case "EXECUTION_STARTED":
			start, found = event, true
		case "EXECUTION_UNKNOWN":
			unknown = true
		case "EXECUTION_RECONCILED":
			return LedgerEvent{}, fmt.Errorf("execution %s is already reconciled", executionID)
		case "EXECUTION_ALLOWED":
			return LedgerEvent{}, fmt.Errorf("execution %s already has a confirmed successful receipt", executionID)
		}
	}
	if !found || !unknown && !hasUnresolvedStart(l.events, executionID) {
		return LedgerEvent{}, fmt.Errorf("execution %s has no unresolved outcome to reconcile", executionID)
	}
	returned := l.appendLocked(LedgerEvent{
		Timestamp: time.Now().UTC(), RequestID: start.RequestID, DecisionID: start.DecisionID,
		ExecutionID: executionID, AgentID: start.AgentID, Capability: start.Capability,
		ArtifactHash: start.ArtifactHash, EventType: "EXECUTION_RECONCILED",
		Decision: DecisionUnknown, Reason: ReasonExecutionOutcomeUnknown,
		Metadata: map[string]string{
			"reconciliation_outcome": string(outcome),
			"evidence_ref":           evidenceRef,
			"reconciler_id":          reconcilerID,
		},
	})
	if returned.ID == "" {
		return LedgerEvent{}, fmt.Errorf("persist reconciliation event: %w", l.lastPersistenceError)
	}
	return returned, nil
}

func hasUnresolvedStart(events []LedgerEvent, executionID string) bool {
	started, terminal := false, false
	for _, event := range events {
		if event.ExecutionID != executionID {
			continue
		}
		switch event.EventType {
		case "EXECUTION_STARTED":
			started = true
		case "EXECUTION_ALLOWED", "EXECUTION_RECONCILED":
			terminal = true
		}
	}
	return started && !terminal
}

// unresolvedExecutionStarts identifies execution intents that have no recorded terminal outcome.
// A start event is persisted before invoking an external side effect, so it remains evidence of
// uncertainty if the process crashes or the final receipt cannot be durably appended.
func unresolvedExecutionStarts(events []LedgerEvent) []LedgerEvent {
	starts := make(map[string]LedgerEvent)
	resolved := make(map[string]struct{})
	for _, event := range events {
		if event.ExecutionID == "" {
			continue
		}
		switch event.EventType {
		case "EXECUTION_STARTED":
			starts[event.ExecutionID] = event
		case "EXECUTION_ALLOWED":
			resolved[event.ExecutionID] = struct{}{}
		case "EXECUTION_UNKNOWN":
			resolved[event.ExecutionID] = struct{}{}
		case "EXECUTION_RECONCILED":
			if executionHasValidReconciliation(events, event.ExecutionID) {
				resolved[event.ExecutionID] = struct{}{}
			}
		}
	}
	unresolved := make([]LedgerEvent, 0)
	for _, event := range events {
		if event.EventType != "EXECUTION_STARTED" || event.ExecutionID == "" {
			continue
		}
		if _, ok := starts[event.ExecutionID]; !ok {
			continue
		}
		if _, ok := resolved[event.ExecutionID]; !ok {
			unresolved = append(unresolved, event)
			delete(starts, event.ExecutionID)
		}
	}
	return unresolved
}

// executionHasValidReconciliation accepts only a reconciliation event bound to
// the execution ID and carrying all required assertion fields.
func executionHasValidReconciliation(events []LedgerEvent, executionID string) bool {
	if executionID == "" {
		return false
	}
	for _, event := range events {
		if event.ExecutionID != executionID || event.EventType != "EXECUTION_RECONCILED" {
			continue
		}
		if event.Metadata["evidence_ref"] == "" || event.Metadata["reconciler_id"] == "" {
			continue
		}
		outcome := ExecutionOutcome(event.Metadata["reconciliation_outcome"])
		if outcome == ExecutionOutcomeCommitted || outcome == ExecutionOutcomeNotCommitted {
			return true
		}
	}
	return false
}

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newUnknownExecutionLedger() *Ledger {
	ledger := NewLedger()
	ledger.Append(LedgerEvent{
		Timestamp: time.Now().UTC(), RequestID: "REQ-RECONCILE",
		DecisionID: "DEC-RECONCILE", ExecutionID: "EXEC-RECONCILE",
		AgentID: "agent-reconcile", Capability: "repository.write",
		EventType: "EXECUTION_STARTED",
	})
	ledger.Append(LedgerEvent{
		Timestamp: time.Now().UTC(), RequestID: "REQ-RECONCILE",
		DecisionID: "DEC-RECONCILE", ExecutionID: "EXEC-RECONCILE",
		AgentID: "agent-reconcile", Capability: "repository.write",
		EventType: "EXECUTION_UNKNOWN", Decision: DecisionUnknown,
		Reason: ReasonExecutionOutcomeUnknown,
	})
	return ledger
}

func TestLedgerReconcileExecutionResolvesUnknownOutcome(t *testing.T) {
	ledger := newUnknownExecutionLedger()

	event, err := ledger.ReconcileExecution(
		"EXEC-RECONCILE", ExecutionOutcomeCommitted, "evidence://receipt-1", "operator-1",
	)
	if err != nil {
		t.Fatalf("ReconcileExecution() error = %v", err)
	}
	if event.EventType != "EXECUTION_RECONCILED" {
		t.Fatalf("event type = %q, want EXECUTION_RECONCILED", event.EventType)
	}
	if event.Metadata["reconciliation_outcome"] != string(ExecutionOutcomeCommitted) {
		t.Fatalf("reconciliation outcome = %q", event.Metadata["reconciliation_outcome"])
	}
	if event.Metadata["evidence_ref"] != "evidence://receipt-1" || event.Metadata["reconciler_id"] != "operator-1" {
		t.Fatalf("missing reconciliation attribution: %+v", event.Metadata)
	}

	summary := ledger.VerifySession("agent-reconcile")
	if summary.Result != "PASS" || !summary.PolicyCompliant {
		t.Fatalf("VerifySession after valid reconciliation = %+v, want compliant PASS", summary)
	}
}

func TestLedgerReconcileExecutionRejectsInvalidRequestsAndDuplicate(t *testing.T) {
	tests := []struct {
		name         string
		executionID  string
		outcome      ExecutionOutcome
		evidenceRef  string
		reconcilerID string
	}{
		{name: "missing execution ID", outcome: ExecutionOutcomeCommitted, evidenceRef: "evidence://1", reconcilerID: "operator-1"},
		{name: "missing evidence", executionID: "EXEC-RECONCILE", outcome: ExecutionOutcomeCommitted, reconcilerID: "operator-1"},
		{name: "missing reconciler", executionID: "EXEC-RECONCILE", outcome: ExecutionOutcomeCommitted, evidenceRef: "evidence://1"},
		{name: "unsupported outcome", executionID: "EXEC-RECONCILE", outcome: ExecutionOutcome("UNKNOWN"), evidenceRef: "evidence://1", reconcilerID: "operator-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ledger := newUnknownExecutionLedger()
			if _, err := ledger.ReconcileExecution(tt.executionID, tt.outcome, tt.evidenceRef, tt.reconcilerID); err == nil {
				t.Fatal("expected invalid reconciliation to be rejected")
			}
			if got := len(ledger.Events()); got != 2 {
				t.Fatalf("invalid reconciliation changed ledger length to %d, want 2", got)
			}
		})
	}

	ledger := newUnknownExecutionLedger()
	if _, err := ledger.ReconcileExecution("EXEC-RECONCILE", ExecutionOutcomeNotCommitted, "evidence://receipt-2", "operator-2"); err != nil {
		t.Fatalf("first reconciliation failed: %v", err)
	}
	if _, err := ledger.ReconcileExecution("EXEC-RECONCILE", ExecutionOutcomeCommitted, "evidence://receipt-3", "operator-3"); err == nil {
		t.Fatal("expected duplicate reconciliation to be rejected")
	}
	if got := len(ledger.Events()); got != 3 {
		t.Fatalf("duplicate reconciliation changed ledger length to %d, want 3", got)
	}
}

func TestLedgerReconcileExecutionRequiresUnresolvedExecution(t *testing.T) {
	ledger := NewLedger()
	if _, err := ledger.ReconcileExecution("missing-execution", ExecutionOutcomeCommitted, "evidence://receipt", "operator-1"); err == nil {
		t.Fatal("expected unknown execution to be rejected")
	}

	ledger.Append(LedgerEvent{
		Timestamp: time.Now().UTC(), ExecutionID: "EXEC-RESOLVED",
		AgentID: "agent-reconcile", EventType: "EXECUTION_STARTED",
	})
	ledger.Append(LedgerEvent{
		Timestamp: time.Now().UTC(), ExecutionID: "EXEC-RESOLVED",
		AgentID: "agent-reconcile", EventType: "EXECUTION_ALLOWED", Decision: DecisionAllow,
	})
	if _, err := ledger.ReconcileExecution("EXEC-RESOLVED", ExecutionOutcomeCommitted, "evidence://receipt", "operator-1"); err == nil {
		t.Fatal("expected already-confirmed execution to be rejected")
	}
}

func TestLedgerReconcileExecutionDoesNotCommitWhenPersistenceFails(t *testing.T) {
	ledger := newUnknownExecutionLedger()
	root := t.TempDir()
	blocker := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(blocker, []byte("block"), 0o600); err != nil {
		t.Fatal(err)
	}
	ledger.SetPersistencePath(filepath.Join(blocker, "ledger.json"))

	if _, err := ledger.ReconcileExecution("EXEC-RECONCILE", ExecutionOutcomeCommitted, "evidence://receipt", "operator-1"); err == nil {
		t.Fatal("expected persistence failure to be returned")
	}
	if got := len(ledger.Events()); got != 2 {
		t.Fatalf("failed reconciliation changed ledger length to %d, want 2", got)
	}
	if ledger.PersistenceError() == nil {
		t.Fatal("expected persistence error to be recorded")
	}
}

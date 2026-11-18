// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"testing"
	"time"
)

func TestLedgerReconcileExecutionClearsUnknownOnlyWithExplicitEvidence(t *testing.T) {
	ledger := NewLedger()
	start := ledger.Append(LedgerEvent{
		Timestamp: time.Now().UTC(), RequestID: "req-reconcile", DecisionID: "decision-reconcile",
		ExecutionID: "exec-reconcile", AgentID: "agent-reconcile", Capability: "repository.write",
		ArtifactHash: "sha256:reconcile", EventType: "EXECUTION_STARTED", Decision: DecisionAllow,
	})
	if start.ID == "" {
		t.Fatal("expected execution start to be recorded")
	}
	unknown := ledger.Append(LedgerEvent{
		RequestID: start.RequestID, DecisionID: start.DecisionID, ExecutionID: start.ExecutionID,
		AgentID: start.AgentID, Capability: start.Capability, ArtifactHash: start.ArtifactHash,
		EventType: "EXECUTION_UNKNOWN", Decision: DecisionUnknown, Reason: ReasonExecutionOutcomeUnknown,
	})
	if unknown.ID == "" {
		t.Fatal("expected unknown outcome to be recorded")
	}
	if got := ledger.VerifySession(start.AgentID); got.Result != "UNKNOWN" || got.PolicyCompliant {
		t.Fatalf("expected UNKNOWN before reconciliation, got %+v", got)
	}

	if _, err := ledger.ReconcileExecution(start.ExecutionID, ExecutionOutcomeCommitted, "", "operator-1"); err == nil {
		t.Fatal("expected missing evidence reference to be rejected")
	}
	if _, err := ledger.ReconcileExecution(start.ExecutionID, ExecutionOutcome("MAYBE"), "ticket://INC-123", "operator-1"); err == nil {
		t.Fatal("expected unsupported outcome to be rejected")
	}

	reconciled, err := ledger.ReconcileExecution(start.ExecutionID, ExecutionOutcomeCommitted, "ticket://INC-123", "operator-1")
	if err != nil {
		t.Fatalf("reconcile execution: %v", err)
	}
	if reconciled.EventType != "EXECUTION_RECONCILED" || reconciled.Metadata["reconciliation_outcome"] != string(ExecutionOutcomeCommitted) {
		t.Fatalf("unexpected reconciliation event: %+v", reconciled)
	}
	if got := ledger.VerifySession(start.AgentID); got.Result != "PASS" || !got.PolicyCompliant {
		t.Fatalf("expected ledger verification to recover after explicit reconciliation, got %+v", got)
	}
	if got := NewSessionVerifier(ledger, nil).VerifySession(start.AgentID); got.Result != "PASS" || !got.PolicyCompliant {
		t.Fatalf("expected session verifier to recover after explicit reconciliation, got %+v", got)
	}
	if _, err := ledger.ReconcileExecution(start.ExecutionID, ExecutionOutcomeNotCommitted, "ticket://INC-124", "operator-2"); err == nil {
		t.Fatal("expected duplicate reconciliation to be rejected")
	}
	if !ledger.HasExecution(start.DecisionID) {
		t.Fatal("reconciliation must not make the original authorization replayable")
	}
}

func TestInvalidReconciliationDoesNotResolveUnmatchedStart(t *testing.T) {
	ledger := NewLedger()
	start := ledger.Append(LedgerEvent{
		Timestamp: time.Now().UTC(), RequestID: "req-invalid-reconcile", DecisionID: "decision-invalid-reconcile",
		ExecutionID: "exec-invalid-reconcile", AgentID: "agent-invalid-reconcile", Capability: "repository.write",
		ArtifactHash: "sha256:invalid-reconcile", EventType: "EXECUTION_STARTED", Decision: DecisionAllow,
	})
	if start.ID == "" {
		t.Fatal("expected execution start to be recorded")
	}
	if rejected := ledger.Append(LedgerEvent{
		ExecutionID: start.ExecutionID, AgentID: start.AgentID, Capability: start.Capability,
		EventType: "EXECUTION_RECONCILED", Decision: DecisionUnknown,
		Metadata: map[string]string{"reconciliation_outcome": "MAYBE", "evidence_ref": "ticket://INC-999", "reconciler_id": "operator-1"},
	}); rejected.ID != "" {
		t.Fatalf("generic append must reject reconciliation events, got %+v", rejected)
	}
	if got := ledger.VerifySession(start.AgentID); got.Result != "UNKNOWN" || got.PolicyCompliant {
		t.Fatalf("invalid reconciliation must not resolve an unmatched start, got %+v", got)
	}
}

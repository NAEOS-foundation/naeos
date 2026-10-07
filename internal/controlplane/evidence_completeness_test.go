// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import "testing"

func TestVerifyEvidenceRejectsDetachedObservation(t *testing.T) {
	bundle := EvidenceBundle{
		SchemaVersion: "1.2",
		RequestID: "req-1", DecisionID: "dec-1", ExecutionID: "exec-1",
		AgentID: "agent-1", Capability: "repository.write", Decision: DecisionAllow,
		DecisionEvent: LedgerEvent{
			RequestID: "req-1", DecisionID: "dec-1", AgentID: "agent-1",
			Capability: "repository.write", EventType: "AUTHORIZATION_DECISION",
		},
		ExecutionEvent: &LedgerEvent{
			RequestID: "req-1", DecisionID: "dec-1", ExecutionID: "exec-1",
			AgentID: "agent-1", Capability: "repository.write", EventType: "EXECUTION_ALLOWED",
		},
		ObservationEvent: &LedgerEvent{
			RequestID: "req-1", DecisionID: "dec-1", ExecutionID: "exec-attacker",
			AgentID: "agent-1", Capability: "repository.write", EventType: "SIDE_EFFECT_OBSERVED",
		},
	}
	digest, err := evidenceDigest(bundle)
	if err != nil { t.Fatalf("digest: %v", err) }
	bundle.EvidenceDigest = digest
	if err := signEvidenceBundle(&bundle); err != nil { t.Fatalf("sign: %v", err) }
	result := VerifyEvidence(bundle)
	if result.Result != "FAIL" { t.Fatalf("expected detached observation to fail, got %s: %v", result.Result, result.Issues) }
}

func TestVerifyEvidenceRequiresCompleteAllowChain(t *testing.T) {
	bundle := EvidenceBundle{
		SchemaVersion: "1.2",
		RequestID: "req-1", DecisionID: "dec-1", ExecutionID: "exec-1",
		AgentID: "agent-1", Capability: "repository.write", Decision: DecisionAllow,
		DecisionEvent: LedgerEvent{
			RequestID: "req-1", DecisionID: "dec-1", AgentID: "agent-1",
			Capability: "repository.write", EventType: "AUTHORIZATION_DECISION",
		},
		ExecutionEvent: &LedgerEvent{
			RequestID: "req-1", DecisionID: "dec-1", ExecutionID: "exec-1",
			AgentID: "agent-1", Capability: "repository.write", EventType: "EXECUTION_ALLOWED",
		},
	}
	digest, err := evidenceDigest(bundle)
	if err != nil { t.Fatalf("digest: %v", err) }
	bundle.EvidenceDigest = digest
	if err := signEvidenceBundle(&bundle); err != nil { t.Fatalf("sign: %v", err) }
	result := VerifyEvidence(bundle)
	if result.Result != "FAIL" { t.Fatalf("expected incomplete allow chain to fail, got %s: %v", result.Result, result.Issues) }
}

func TestVerifyEvidenceBundleRejectsReorderedLedger(t *testing.T) {
	ledger := NewLedger()
	decision := ledger.Append(LedgerEvent{
		RequestID: "req-1", DecisionID: "dec-1", AgentID: "agent-1",
		Capability: "repository.write", EventType: "AUTHORIZATION_DECISION", Decision: DecisionAllow,
	})
	execution := ledger.Append(LedgerEvent{
		RequestID: "req-1", DecisionID: "dec-1", ExecutionID: "exec-1",
		AgentID: "agent-1", Capability: "repository.write", EventType: "EXECUTION_ALLOWED", Decision: DecisionAllow,
	})
	observation := ledger.Append(LedgerEvent{
		RequestID: "req-1", DecisionID: "dec-1", ExecutionID: "exec-1",
		AgentID: "agent-1", Capability: "repository.write", EventType: "SIDE_EFFECT_OBSERVED", Decision: DecisionAllow,
	})
	bundle := EvidenceBundle{
		RequestID: "req-1", DecisionID: "dec-1", ExecutionID: "exec-1",
		AgentID: "agent-1", Capability: "repository.write", Decision: DecisionAllow,
		DecisionEvent: decision, ExecutionEvent: &execution, ObservationEvent: &observation,
	}
	events := ledger.Events()
	ledger.events = []LedgerEvent{events[0], events[2], events[1]}
	result := ledger.verifyEvidenceBundle(bundle)
	if result.Result != "FAIL" { t.Fatalf("expected reordered ledger to fail, got %s: %v", result.Result, result.Issues) }
}

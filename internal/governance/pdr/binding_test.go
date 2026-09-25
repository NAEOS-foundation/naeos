// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pdr

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
	"github.com/NAEOS-foundation/naeos/internal/evidence"
)

func testEvidenceAndLedger() (*evidence.EvidenceStore, *controlplane.Ledger, evidence.EvidenceRecord) {
	store := evidence.NewStore()
	rec, _ := store.Append(evidence.EvidenceRecord{
		ID: "ev-1", PolicyID: "change-risk-governance", PolicyVersion: "1.0.0",
		Decision: controlplane.DecisionAllow, Actor: "test", Resource: "repo", Action: "change",
	})
	ledger := controlplane.NewLedger()
	ledger.Append(controlplane.LedgerEvent{
		ID: "decision-event-1", DecisionID: "decision-1", EventType: "AUTHORIZATION_DECISION",
		AgentID: "agent-1", Decision: controlplane.DecisionAllow,
		Metadata: map[string]string{"policy_id": "change-risk-governance", "policy_version": "1.0.0"},
	})
	return store, ledger, rec
}

func TestVerifyBindingPasses(t *testing.T) {
	store, ledger, rec := testEvidenceAndLedger()
	err := VerifyBinding(Record{
		RecordVersion: "1.0.0", PolicyID: "change-risk-governance", PolicyVersion: "1.0.0",
		SchemaVersion: "1.0.0", DecisionID: "decision-1",
		Outcome: "allow", Evidence: []EvidenceRef{{ID: rec.ID, Digest: rec.Hash}},
	}, store, ledger)
	if err != nil {
		t.Fatalf("expected valid binding: %v", err)
	}
}

func TestVerifyBindingFailsClosedOnTamperedEvidence(t *testing.T) {
	store, ledger, rec := testEvidenceAndLedger()
	rec.Metadata = map[string]any{"tampered": true}
	store.Append(evidence.EvidenceRecord{ID: "ev-2", PolicyID: "change-risk-governance", PolicyVersion: "1.0.0", Decision: controlplane.DecisionAllow})
	err := VerifyBinding(Record{
		RecordVersion: "1.0.0", PolicyID: "change-risk-governance", PolicyVersion: "1.0.0",
		SchemaVersion: "1.0.0", DecisionID: "decision-1", Outcome: "allow",
		Evidence: []EvidenceRef{{ID: rec.ID, Digest: rec.Hash}},
	}, store, ledger)
	if err == nil {
		t.Fatal("expected tampered evidence binding to fail")
	}
}

func TestVerifyBindingFailsClosedOnMissingEvidence(t *testing.T) {
	store, ledger, _ := testEvidenceAndLedger()
	err := VerifyBinding(Record{
		RecordVersion: "1.0.0", PolicyID: "change-risk-governance", PolicyVersion: "1.0.0",
		SchemaVersion: "1.0.0", DecisionID: "decision-1", Outcome: "allow",
		Evidence: []EvidenceRef{{ID: "missing", Digest: "deadbeef"}},
	}, store, ledger)
	if err == nil {
		t.Fatal("expected missing evidence binding to fail")
	}
}

func TestVerifyBindingFailsClosedOnPolicyMismatch(t *testing.T) {
	store, ledger, rec := testEvidenceAndLedger()
	err := VerifyBinding(Record{
		RecordVersion: "1.0.0", PolicyID: "other-policy", PolicyVersion: "1.0.0",
		SchemaVersion: "1.0.0", DecisionID: "decision-1", Outcome: "allow",
		Evidence: []EvidenceRef{{ID: rec.ID, Digest: rec.Hash}},
	}, store, ledger)
	if err == nil {
		t.Fatal("expected policy mismatch to fail")
	}
}

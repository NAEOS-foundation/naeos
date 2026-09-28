// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pdr

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

func TestVerifyEndToEndPDRBinding(t *testing.T) {
	p, s, l, e := fixture(t)
	r, err := Verify(pdrJSON(e.Hash, "allow", "allow", "passed"), Dependencies{p, s, l})
	if err != nil {
		t.Fatal(err)
	}
	if r.LedgerEventID == "" || len(r.EvidenceIDs) != 1 {
		t.Fatalf("incomplete trace: %+v", r)
	}
}

func TestVerifyRejectsEvidenceTampering(t *testing.T) {
	p, s, l, e := fixture(t)
	s.ByID("ev-001").DecisionReasons = []string{"tampered"}
	if _, err := Verify(pdrJSON(e.Hash, "allow", "allow", "passed"), Dependencies{p, s, l}); err == nil {
		t.Fatal("expected tampering rejection")
	}
}

func TestVerifyRejectsStalePolicy(t *testing.T) {
	p, s, l, e := fixture(t)
	if err := p.Register(&policy.Policy{ID: "change-risk", Version: "2.0.0", Default: policy.DecisionAllow}); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(pdrJSON(e.Hash, "allow", "allow", "passed"), Dependencies{p, s, l}); err == nil {
		t.Fatal("expected stale policy rejection")
	}
}

func TestVerifyRejectsMissingLedgerDecision(t *testing.T) {
	p, s, _, e := fixture(t)
	if _, err := Verify(pdrJSON(e.Hash, "allow", "allow", "passed"), Dependencies{p, s, controlplane.NewLedger()}); err == nil {
		t.Fatal("expected missing ledger decision rejection")
	}
}

func fixture(t *testing.T) (*policy.Registry, *evidence.EvidenceStore, *controlplane.Ledger, evidence.EvidenceRecord) {
	p := policy.NewRegistry()
	if err := p.Register(&policy.Policy{ID: "change-risk", Version: "1.0.0", Default: policy.DecisionAllow}); err != nil {
		t.Fatal(err)
	}
	s := evidence.NewStore()
	e, err := s.Append(evidence.EvidenceRecord{ID: "ev-001", Timestamp: time.Date(2026, 9, 28, 5, 0, 0, 0, time.UTC), Actor: "agent", Resource: "repo", Action: "change", PolicyID: "change-risk", PolicyVersion: "1.0.0", Decision: control.DecisionAllow})
	if err != nil {
		t.Fatal(err)
	}
	l := controlplane.NewLedger()
	l.Append(controlplane.LedgerEvent{DecisionID: "decision-001", AgentID: "agent", EventType: "AUTHORIZATION_DECISION", Decision: controlplane.DecisionAllow})
	return p, s, l, e
}

func pdrJSON(digest, outcome, decision, verification string) []byte {
	r := map[string]any{"record_version": "1.0.0", "record_id": "pdr-001", "policy_id": "change-risk", "policy_version": "1.0.0", "schema_version": "1.0.0", "decision_id": "decision-001", "outcome": outcome, "change_classification": map[string]string{"risk": "medium", "criticality": "medium", "decision": decision}, "evidence": []map[string]string{{"evidence_id": "ev-001", "digest": digest, "kind": "governance"}}, "verification": map[string]string{"status": verification, "evidence_digest": digest}}
	b, _ := json.Marshal(r)
	return b
}

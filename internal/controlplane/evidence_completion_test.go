// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"testing"
	"time"
)

func TestExecuteAtomicWithEvidenceRequiresVerifiedEvidence(t *testing.T) {
	now := time.Now().UTC()
	store := NewPolicyStore()
	policy := &Policy{
		ID:                   "p-evidence",
		Version:              1,
		Status:               "active",
		CreatedAt:            now,
		UpdatedAt:            now,
		AllowedCapabilities:  []Capability{"repository.write"},
		RequiresExplicitAuth: true,
	}
	if err := store.Set(policy); err != nil {
		t.Fatal(err)
	}

	ledger := NewLedger()
	gateway := NewDecisionGateway(NewEvaluator(store), ledger)
	grant := &Grant{
		GrantID:       "g-evidence",
		AgentID:       "agent-evidence",
		PolicyID:      policy.ID,
		PolicyVersion: 1,
		Capabilities:  []Capability{"repository.write"},
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Hour),
		Status:        "active",
	}
	req := AuthorizeRequest{
		RequestID: "REQ-EVIDENCE",
		AgentID:   "agent-evidence",
		Action: Action{
			AgentID:      "agent-evidence",
			Capability:   "repository.write",
			ArtifactHash: "sha256:artifact",
		},
		Grant:     grant,
		Policy:    policy,
		Timestamp: now,
	}
	decision := gateway.Authorize(req)
	if decision.Status != DecisionAllow {
		t.Fatalf("expected allow, got %s", decision.Status)
	}

	out, err := gateway.ExecuteAtomicWithEvidence(req, decision, func() error { return nil }, func(event LedgerEvent) LedgerEvent {\n\t\tevent.EventType = "SIDE_EFFECT_OBSERVED"\n\t\treturn event\n\t})
	if err != nil {
		t.Fatalf("expected evidence-backed completion, got %v", err)
	}
	if out.Result.Status != DecisionAllow {
		t.Fatalf("expected allow result, got %s", out.Result.Status)
	}
	if out.Evidence.ExecutionID == "" {
		t.Fatal("expected execution evidence")
	}
	if out.Evidence.EvidenceDigest == "" {
		t.Fatal("expected evidence digest")
	}
	if VerifyEvidence(out.Evidence).Result != "PASS" {
		t.Fatal("expected independently verifiable evidence")
	}
}

func TestExecuteAtomicWithEvidenceFailsClosedWhenSideEffectFails(t *testing.T) {
	now := time.Now().UTC()
	store := NewPolicyStore()
	policy := &Policy{
		ID:                   "p-fail",
		Version:              1,
		Status:               "active",
		CreatedAt:            now,
		UpdatedAt:            now,
		AllowedCapabilities:  []Capability{"repository.write"},
		RequiresExplicitAuth: true,
	}
	if err := store.Set(policy); err != nil {
		t.Fatal(err)
	}

	ledger := NewLedger()
	gateway := NewDecisionGateway(NewEvaluator(store), ledger)
	grant := &Grant{
		GrantID:       "g-fail",
		AgentID:       "agent-fail",
		PolicyID:      policy.ID,
		PolicyVersion: 1,
		Capabilities:  []Capability{"repository.write"},
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Hour),
		Status:        "active",
	}
	req := AuthorizeRequest{
		RequestID: "REQ-FAIL",
		AgentID:   "agent-fail",
		Action: Action{
			AgentID:      "agent-fail",
			Capability:   "repository.write",
			ArtifactHash: "sha256:artifact",
		},
		Grant:     grant,
		Policy:    policy,
		Timestamp: now,
	}
	decision := gateway.Authorize(req)
	if decision.Status != DecisionAllow {
		t.Fatalf("expected allow, got %s", decision.Status)
	}

	if _, err := gateway.ExecuteAtomicWithEvidence(req, decision, func() error { return assertErr{} }); err == nil {
		t.Fatal("expected failed execution to reject completion")
	}
}

type assertErr struct{}

func (assertErr) Error() string {
	return "side effect failed"
}

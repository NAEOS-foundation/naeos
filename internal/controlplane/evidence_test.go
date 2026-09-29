// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package controlplane

import "testing"

func TestEvidenceBundleReconstructsAllowedExecution(t *testing.T) {
	evaluator := NewEvaluator()
	ledger := NewLedger()
	gateway := NewDecisionGateway(evaluator, ledger)

	policy := &Policy{ID: "POLICY-1", Version: 1, Status: "active", AllowedCapabilities: []Capability{"repository.read"}}
	grant := &Grant{GrantID: "GRANT-1", AgentID: "agent-1", PolicyID: "POLICY-1", PolicyVersion: 1, Capabilities: []Capability{"repository.read"}}
	req := AuthorizeRequest{
		RequestID: "REQ-1",
		AgentID: "agent-1",
		Action: Action{Capability: "repository.read", ArtifactHash: "sha256:artifact"},
		Grant: grant,
		Policy: policy,
	}

	result := gateway.Authorize(req)
	if result.Status != DecisionAllow {
		t.Fatalf("expected ALLOW, got %s", result.Status)
	}
	_, event := gateway.ExecuteDecision(req, result)
	if event.EventType != "EXECUTION_ALLOWED" {
		t.Fatalf("expected execution evidence, got %s", event.EventType)
	}

	bundle, err := ledger.BuildEvidence(result.DecisionID)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.PolicyID != "POLICY-1" || bundle.PolicyVersion != "1" || bundle.GrantID != "GRANT-1" {
		t.Fatalf("missing policy/grant binding: %#v", bundle)
	}
	if bundle.ExecutionEvent == nil || bundle.ExecutionEvent.EventType != "EXECUTION_ALLOWED" {
		t.Fatalf("missing execution evidence: %#v", bundle.ExecutionEvent)
	}
	if got := VerifyEvidence(bundle); got.Result != "PASS" {
		t.Fatalf("expected independent verification PASS, got %#v", got)
	}
}

func TestEvidenceVerificationRejectsTampering(t *testing.T) {
	ledger := NewLedger()
	event := ledger.Append(LedgerEvent{
		RequestID: "REQ-2", DecisionID: "DEC-2", AgentID: "agent-2",
		Capability: "repository.read", ArtifactHash: "sha256:artifact",
		EventType: "AUTHORIZATION_DECISION", Decision: DecisionAllow, Reason: ReasonAllowed,
		Metadata: map[string]string{"policy_id": "POLICY-1", "policy_version": "1", "grant_id": "GRANT-1"},
	})
	ledger.Append(LedgerEvent{
		RequestID: "REQ-2", DecisionID: "DEC-2", ExecutionID: "EXEC-2", AgentID: "agent-2",
		Capability: "repository.read", ArtifactHash: "sha256:artifact",
		EventType: "EXECUTION_ALLOWED", Decision: DecisionAllow, Reason: ReasonAllowed,
	})
	bundle, err := ledger.BuildEvidence(event.DecisionID)
	if err != nil {
		t.Fatal(err)
	}
	bundle.ExecutionEvent.ArtifactHash = "sha256:tampered"
	if got := VerifyEvidence(bundle); got.Result != "FAIL" {
		t.Fatalf("expected tampered evidence FAIL, got %#v", got)
	}
}

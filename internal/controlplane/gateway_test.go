// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"testing"
	"time"
)

func TestDecisionGateway_AllowAndLedger(t *testing.T) {
	store := NewPolicyStore()
	policy := &Policy{
		ID:                  "POLICY-ALLOW",
		Version:             1,
		Status:              "active",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		AllowedCapabilities: []Capability{"repository.read"},
	}
	if err := store.Set(policy); err != nil {
		t.Fatalf("store.Set() error = %v", err)
	}
	grant := &Grant{
		GrantID:       "GRANT-ALLOWED",
		AgentID:       "agent-1",
		PolicyID:      policy.ID,
		PolicyVersion: policy.Version,
		Capabilities:  []Capability{"repository.read"},
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		Status:        "active",
	}
	ledger := NewLedger()
	gateway := NewDecisionGateway(NewEvaluator(store), ledger)
	result := gateway.Authorize(AuthorizeRequest{
		AgentID:   "agent-1",
		Action:    Action{AgentID: "agent-1", Capability: "repository.read"},
		Grant:     grant,
		Policy:    policy,
		Timestamp: time.Now().UTC(),
	})
	if result.Status != DecisionAllow {
		t.Fatalf("expected ALLOW, got %s (%s)", result.Status, result.Message)
	}
	if len(ledger.Events()) != 1 {
		t.Fatalf("expected one ledger event, got %d", len(ledger.Events()))
	}
}

func TestDecisionGateway_DeniesProtectedCapabilityAndRecordsBlock(t *testing.T) {
	store := NewPolicyStore()
	policy := &Policy{
		ID:                    "POLICY-PROTECTED",
		Version:               1,
		Status:                "active",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		AllowedCapabilities:   []Capability{"credential.rotate"},
		ProtectedCapabilities: []Capability{"credential.rotate"},
	}

	if err := store.Set(policy); err != nil {
		t.Fatalf("store.Set() error = %v", err)
	}
	grant := &Grant{
		GrantID:       "GRANT-SECRET",
		AgentID:       "agent-2",
		PolicyID:      policy.ID,
		PolicyVersion: policy.Version,
		Capabilities:  []Capability{"credential.rotate"},
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		Status:        "active",
	}
	ledger := NewLedger()
	gateway := NewDecisionGateway(NewEvaluator(store), ledger)
	result, event := gateway.Execute(AuthorizeRequest{
		AgentID:   "agent-2",
		Action:    Action{AgentID: "agent-2", Capability: "credential.rotate"},
		Grant:     grant,
		Policy:    policy,
		Timestamp: time.Now().UTC(),
	})
	if result.Status != DecisionDeny {
		t.Fatalf("expected DENY, got %s (%s)", result.Status, result.Message)
	}
	if event.EventType != "EXECUTION_BLOCKED" {
		t.Fatalf("expected execution block event, got %s", event.EventType)
	}
	if summary := ledger.VerifySession("agent-2"); summary.Result != "PASS" || summary.BlockedAttempts != 2 {
		t.Fatalf("expected compliant blocked attempt evidence, got result=%s blocked=%d", summary.Result, summary.BlockedAttempts)
	}
}

func TestDecisionGateway_DeniesWhenGatewayEvaluatorIsUnavailable(t *testing.T) {
	var gateway *DecisionGateway
	result := gateway.Authorize(AuthorizeRequest{})
	if result.Status != DecisionDeny {
		t.Fatalf("expected DENY for unavailable gateway, got %s", result.Status)
	}

}

func TestDecisionGateway_RejectsForgedExecutionDecision(t *testing.T) {
	store := NewPolicyStore()
	policy := &Policy{ID: "POLICY-FORGED", Version: 1, Status: "active", AllowedCapabilities: []Capability{"repository.read"}}
	if err := store.Set(policy); err != nil {
		t.Fatal(err)
	}
	ledger := NewLedger()
	gateway := NewDecisionGateway(NewEvaluator(store), ledger)
	result, _ := gateway.ExecuteDecision(AuthorizeRequest{
		RequestID: "REQ-FORGED", AgentID: "agent-1",
		Action: Action{AgentID: "agent-1", Capability: "repository.read"},
	}, DecisionResult{RequestID: "REQ-FORGED", DecisionID: "DEC-FORGED", Status: DecisionAllow})
	if result.Status != DecisionDeny {
		t.Fatalf("expected forged decision to be denied, got %s", result.Status)
	}
}

func TestDecisionGateway_RequiresMatchingApprovalArtifactAndConsumesApproval(t *testing.T) {
	store := NewPolicyStore()
	policy := &Policy{
		ID: "POLICY-APPROVAL", Version: 1, Status: "active",
		ApprovalRequired: []Capability{"production.deploy"}, AllowedCapabilities: []Capability{"production.deploy"},
	}
	if err := store.Set(policy); err != nil {
		t.Fatal(err)
	}
	grant := &Grant{
		GrantID: "GRANT-APPROVAL", AgentID: "agent-1", PolicyID: policy.ID, PolicyVersion: 1,
		Capabilities: []Capability{"production.deploy"}, Status: "active", ExpiresAt: time.Now().Add(time.Hour),
	}
	gateway := NewDecisionGateway(NewEvaluator(store), NewLedger())
	action := Action{AgentID: "agent-1", Capability: "production.deploy", ArtifactHash: "sha256:approved"}
	pending := gateway.Authorize(AuthorizeRequest{RequestID: "REQ-APPROVAL", Action: action, Grant: grant, Policy: policy, Timestamp: time.Now()})
	if pending.Status != DecisionPending {
		t.Fatalf("expected REQUIRE_APPROVAL, got %s", pending.Status)
	}
	approval, err := gateway.RequestApproval(pending, "reviewer-1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gateway.Approve(approval.ID, "reviewed", "sha256:wrong", time.Now()); err == nil {
		t.Fatal("expected mismatched artifact approval to be rejected")
	}
	approval, err = gateway.RequestApproval(pending, "reviewer-1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gateway.Approve(approval.ID, "reviewed", "sha256:approved", time.Now()); err != nil {
		t.Fatal(err)
	}
	approved, _ := gateway.Execute(AuthorizeRequest{
		RequestID: "REQ-APPROVAL", DecisionID: pending.DecisionID, ApprovalID: approval.ID,
		Action: action, Grant: grant, Policy: policy, Timestamp: time.Now(),
	})
	if approved.Status != DecisionAllow {
		t.Fatalf("expected approved execution, got %s (%s)", approved.Status, approved.Message)
	}
	reused, _ := gateway.Execute(AuthorizeRequest{
		RequestID: "REQ-APPROVAL", DecisionID: pending.DecisionID, ApprovalID: approval.ID,
		Action: action, Grant: grant, Policy: policy, Timestamp: time.Now(),
	})
	if reused.Status != DecisionDeny {
		t.Fatalf("expected consumed approval to deny reuse, got %s", reused.Status)
	}
}

func TestDecisionGateway_BlocksStaleAuthorizationAfterPolicyChange(t *testing.T) {
	store := NewPolicyStore()
	now := time.Now().UTC()
	policyV1 := &Policy{
		ID:                  "POLICY-STALE",
		Version:             1,
		Status:              "active",
		CreatedAt:           now,
		UpdatedAt:           now,
		AllowedCapabilities: []Capability{"repository.write"},
	}
	policyV2 := &Policy{
		ID:                  "POLICY-STALE",
		Version:             2,
		Status:              "active",
		CreatedAt:           now.Add(time.Second),
		UpdatedAt:           now.Add(time.Second),
		AllowedCapabilities: []Capability{"repository.write"},
	}
	if err := store.Set(policyV1); err != nil {
		t.Fatal(err)
	}
	grant := &Grant{
		GrantID:       "GRANT-STALE",
		AgentID:       "agent-stale",
		PolicyID:      policyV1.ID,
		PolicyVersion: policyV1.Version,
		Capabilities:  []Capability{"repository.write"},
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Hour),
		Status:        "active",
	}
	ledger := NewLedger()
	gateway := NewDecisionGateway(NewEvaluator(store), ledger)
	req := AuthorizeRequest{
		RequestID: "REQ-STALE",
		AgentID:   "agent-stale",
		Action: Action{
			AgentID:     "agent-stale",
			Capability:  "repository.write",
			ArtifactHash: "sha256:stale",
		},
		Grant:     grant,
		Policy:    policyV1,
		Timestamp: now,
	}
	authorized := gateway.Authorize(req)
	if authorized.Status != DecisionAllow {
		t.Fatalf("expected T0 ALLOW, got %s (%s)", authorized.Status, authorized.Reason)
	}
	if err := store.Set(policyV2); err != nil {
		t.Fatal(err)
	}
	executed, event := gateway.ExecuteDecision(req, authorized)
	if executed.Status != DecisionDeny || executed.Reason != ReasonDeniedStalePolicy {
		t.Fatalf("expected stale authorization DENY, got %s (%s)", executed.Status, executed.Reason)
	}
	if event.EventType != "EXECUTION_BLOCKED" {
		t.Fatalf("expected EXECUTION_BLOCKED, got %s", event.EventType)
	}
	events := ledger.Query(map[string]string{"request_id": req.RequestID})
	if len(events) != 2 {
		t.Fatalf("expected authorization plus blocked execution evidence, got %d events", len(events))
	}
	if events[0].EventType != "AUTHORIZATION_DECISION" || events[1].EventType != "EXECUTION_BLOCKED" {
		t.Fatalf("unexpected evidence sequence: %s -> %s", events[0].EventType, events[1].EventType)
	}
}

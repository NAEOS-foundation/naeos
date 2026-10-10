// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"errors"
	"testing"
	"time"
)

func TestPolicyChangeRevokingCapabilityBlocksStaleAndReauthorizedExecution(t *testing.T) {
	now := time.Now().UTC()
	store := NewPolicyStore()
	policyV1 := &Policy{
		ID: "POLICY-REVOKE", Version: 1, Status: "active",
		CreatedAt: now, UpdatedAt: now,
		AllowedCapabilities: []Capability{"repository.write"},
	}
	policyV2 := &Policy{
		ID: "POLICY-REVOKE", Version: 2, Status: "active",
		CreatedAt: now.Add(time.Second), UpdatedAt: now.Add(time.Second),
		AllowedCapabilities: []Capability{"repository.write"},
		DeniedCapabilities:  []Capability{"repository.write"},
	}
	if err := store.Set(policyV1); err != nil {
		t.Fatal(err)
	}

	ledger := NewLedger()
	gateway := NewDecisionGateway(NewEvaluator(store), ledger)
	grant := &Grant{
		GrantID: "GRANT-REVOKE", AgentID: "agent-revoke", PolicyID: policyV1.ID,
		PolicyVersion: 1, Capabilities: []Capability{"repository.write"},
		CreatedAt: now, ExpiresAt: now.Add(time.Hour), Status: "active",
	}
	req := AuthorizeRequest{
		RequestID: "REQ-REVOKE-STALE", AgentID: "agent-revoke",
		Action: Action{AgentID: "agent-revoke", Capability: "repository.write", ArtifactHash: "sha256:revoke"},
		Grant: grant, Policy: policyV1, Timestamp: now,
	}
	allow := gateway.Authorize(req)
	if allow.Status != DecisionAllow {
		t.Fatalf("expected policy v1 ALLOW, got %s (%s)", allow.Status, allow.Reason)
	}
	if err := store.Set(policyV2); err != nil {
		t.Fatal(err)
	}

	sideEffectCalls := 0
	stale, staleEvent := gateway.ExecuteAtomic(req, allow, func() error {
		sideEffectCalls++
		return errors.New("side effect must not run for stale authorization")
	})
	if stale.Status != DecisionDeny || stale.Reason != ReasonDeniedStalePolicy {
		t.Fatalf("expected stale authorization denial, got %s (%s)", stale.Status, stale.Reason)
	}
	if staleEvent.EventType != "EXECUTION_BLOCKED" {
		t.Fatalf("expected stale execution block evidence, got %s", staleEvent.EventType)
	}
	if sideEffectCalls != 0 {
		t.Fatalf("stale authorization invoked side effect %d times", sideEffectCalls)
	}

	// Rebind the grant to the active version to isolate the policy DENY itself.
	grantV2 := *grant
	grantV2.PolicyVersion = 2
	reqV2 := req
	reqV2.RequestID = "REQ-REVOKE-V2"
	reqV2.Policy = policyV2
	reqV2.Grant = &grantV2
	denied := gateway.Authorize(reqV2)
	if denied.Status != DecisionDeny || denied.Reason != ReasonDeniedByPolicy {
		t.Fatalf("expected policy v2 explicit DENY, got %s (%s)", denied.Status, denied.Reason)
	}

	_, deniedEvent := gateway.ExecuteAtomic(reqV2, denied, func() error {
		sideEffectCalls++
		return errors.New("side effect must not run for denied capability")
	})
	if deniedEvent.EventType != "EXECUTION_BLOCKED" {
		t.Fatalf("expected v2 execution block evidence, got %s", deniedEvent.EventType)
	}
	if sideEffectCalls != 0 {
		t.Fatalf("denied capability invoked side effect %d times", sideEffectCalls)
	}

	staleEvents := ledger.Query(map[string]string{"request_id": req.RequestID})
	v2Events := ledger.Query(map[string]string{"request_id": reqV2.RequestID})
	if !hasTestEvent(staleEvents, "EXECUTION_BLOCKED", ReasonDeniedStalePolicy) {
		t.Fatal("missing stale-authorization block evidence")
	}
	if !hasTestEvent(v2Events, "AUTHORIZATION_DECISION", ReasonDeniedByPolicy) {
		t.Fatal("missing explicit v2 policy-denial evidence")
	}
	if !hasTestEvent(v2Events, "EXECUTION_BLOCKED", ReasonDeniedByPolicy) {
		t.Fatal("missing v2 denied-execution evidence")
	}
	if got := NewSessionVerifier(ledger, NewEvaluator(store)).VerifySession("agent-revoke"); got.Result != "PASS" {
		t.Fatalf("expected independent session verification PASS, got %+v", got)
	}
}

func hasTestEvent(events []LedgerEvent, eventType string, reason DecisionReason) bool {
	for _, event := range events {
		if event.EventType == eventType && event.Reason == reason {
			return true
		}
	}
	return false
}

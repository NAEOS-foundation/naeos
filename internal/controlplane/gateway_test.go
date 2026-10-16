// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"fmt"
	"os"
	"path/filepath"
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
			AgentID:      "agent-stale",
			Capability:   "repository.write",
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

func TestDecisionGateway_ExecuteAtomicRejectsStaleAuthorizationBeforeSideEffect(t *testing.T) {
	store := NewPolicyStore()
	now := time.Now().UTC()
	policyV1 := &Policy{
		ID: "POLICY-ATOMIC-STALE", Version: 1, Status: "active",
		AllowedCapabilities: []Capability{"repository.write"}, CreatedAt: now, UpdatedAt: now,
	}
	policyV2 := &Policy{
		ID: "POLICY-ATOMIC-STALE", Version: 2, Status: "active",
		AllowedCapabilities: []Capability{"repository.write"}, CreatedAt: now.Add(time.Second), UpdatedAt: now.Add(time.Second),
	}
	if err := store.Set(policyV1); err != nil {
		t.Fatal(err)
	}
	grant := &Grant{
		GrantID: "GRANT-ATOMIC-STALE", AgentID: "agent-atomic-stale",
		PolicyID: policyV1.ID, PolicyVersion: policyV1.Version,
		Capabilities: []Capability{"repository.write"}, Status: "active",
		CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}
	ledger := NewLedger()
	gateway := NewDecisionGateway(NewEvaluator(store), ledger)
	req := AuthorizeRequest{
		RequestID: "REQ-ATOMIC-STALE", AgentID: "agent-atomic-stale",
		Action: Action{AgentID: "agent-atomic-stale", Capability: "repository.write", ArtifactHash: "sha256:atomic-stale"},
		Grant:  grant, Policy: policyV1, Timestamp: now,
	}
	authorized := gateway.Authorize(req)
	if authorized.Status != DecisionAllow {
		t.Fatalf("expected T0 ALLOW, got %s (%s)", authorized.Status, authorized.Reason)
	}
	if err := store.Set(policyV2); err != nil {
		t.Fatal(err)
	}

	called := false
	executed, event := gateway.ExecuteAtomic(req, authorized, func() error {
		called = true
		return nil
	})
	if executed.Status != DecisionDeny || executed.Reason != ReasonDeniedStalePolicy {
		t.Fatalf("expected stale atomic authorization DENY, got %s (%s)", executed.Status, executed.Reason)
	}
	if event.EventType != "EXECUTION_BLOCKED" {
		t.Fatalf("expected EXECUTION_BLOCKED, got %s", event.EventType)
	}
	if called {
		t.Fatal("stale authorization must be rejected before the side-effect callback")
	}
	events := ledger.Query(map[string]string{"request_id": req.RequestID})
	if len(events) != 2 || events[1].EventType != "EXECUTION_BLOCKED" {
		t.Fatalf("expected authorization plus blocked execution evidence, got %#v", events)
	}
}

func TestDecisionGateway_ExecuteAtomicSerializesPolicyChangeWithSideEffect(t *testing.T) {
	store := NewPolicyStore()
	now := time.Now().UTC()
	policyV1 := &Policy{
		ID: "POLICY-ATOMIC", Version: 1, Status: "active",
		AllowedCapabilities: []Capability{"repository.write"}, CreatedAt: now, UpdatedAt: now,
	}
	policyV2 := &Policy{
		ID: "POLICY-ATOMIC", Version: 2, Status: "active",
		AllowedCapabilities: []Capability{"repository.write"}, CreatedAt: now.Add(time.Second), UpdatedAt: now.Add(time.Second),
	}
	if err := store.Set(policyV1); err != nil {
		t.Fatal(err)
	}
	grant := &Grant{
		GrantID: "GRANT-ATOMIC", AgentID: "agent-atomic", PolicyID: policyV1.ID, PolicyVersion: 1,
		Capabilities: []Capability{"repository.write"}, Status: "active", CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}
	ledger := NewLedger()
	gateway := NewDecisionGateway(NewEvaluator(store), ledger)
	req := AuthorizeRequest{
		RequestID: "REQ-ATOMIC", AgentID: "agent-atomic",
		Action: Action{AgentID: "agent-atomic", Capability: "repository.write", ArtifactHash: "sha256:atomic"},
		Grant:  grant, Policy: policyV1, Timestamp: now,
	}
	authorized := gateway.Authorize(req)
	if authorized.Status != DecisionAllow {
		t.Fatalf("expected T0 ALLOW, got %s (%s)", authorized.Status, authorized.Reason)
	}

	started := make(chan struct{})
	release := make(chan struct{})
	executionDone := make(chan DecisionResult, 1)
	go func() {
		result, _ := gateway.ExecuteAtomic(req, authorized, func() error {
			close(started)
			<-release
			return nil
		})
		executionDone <- result
	}()
	<-started

	policyUpdateDone := make(chan error, 1)
	go func() { policyUpdateDone <- store.Set(policyV2) }()
	select {
	case err := <-policyUpdateDone:
		t.Fatalf("policy update committed before atomic side effect completed: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	result := <-executionDone
	if result.Status != DecisionAllow {
		t.Fatalf("expected atomic execution ALLOW, got %s (%s)", result.Status, result.Message)
	}
	if err := <-policyUpdateDone; err != nil {
		t.Fatal(err)
	}
	active, err := store.Active(policyV1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if active.Version != 2 {
		t.Fatalf("expected policy v2 after commit, got v%d", active.Version)
	}
	events := ledger.Query(map[string]string{"request_id": req.RequestID})
	if len(events) != 3 || events[1].EventType != "EXECUTION_STARTED" || events[2].EventType != "EXECUTION_ALLOWED" {
		t.Fatalf("expected authorization, durable execution intent, and execution receipt, got %#v", events)
	}
}

func TestDecisionGateway_ExecuteAtomicPreservesUnknownWhenSideEffectReturnsError(t *testing.T) {
	now := time.Now().UTC()
	store := NewPolicyStore()
	policy := &Policy{ID: "POLICY-UNKNOWN-ERROR", Version: 1, Status: "active", AllowedCapabilities: []Capability{"repository.write"}}
	if err := store.Set(policy); err != nil {
		t.Fatal(err)
	}
	grant := &Grant{GrantID: "GRANT-UNKNOWN-ERROR", AgentID: "agent-unknown-error", PolicyID: policy.ID, PolicyVersion: 1, Capabilities: []Capability{"repository.write"}, Status: "active", CreatedAt: now, ExpiresAt: now.Add(time.Hour)}
	ledger := NewLedger()
	gateway := NewDecisionGateway(NewEvaluator(store), ledger)
	req := AuthorizeRequest{RequestID: "REQ-UNKNOWN-ERROR", AgentID: grant.AgentID, Action: Action{AgentID: grant.AgentID, Capability: "repository.write", ArtifactHash: "sha256:unknown-error"}, Grant: grant, Policy: policy, Timestamp: now}
	authorized := gateway.Authorize(req)
	if authorized.Status != DecisionAllow {
		t.Fatalf("expected ALLOW authorization, got %s (%s)", authorized.Status, authorized.Message)
	}

	result, event := gateway.ExecuteAtomic(req, authorized, func() error {
		return fmt.Errorf("upstream timed out after submitting write")
	})
	if result.Status != DecisionUnknown || result.Reason != ReasonExecutionOutcomeUnknown {
		t.Fatalf("expected UNKNOWN execution outcome, got %s (%s)", result.Status, result.Message)
	}
	if event.EventType != "EXECUTION_UNKNOWN" || event.Decision != DecisionUnknown {
		t.Fatalf("expected EXECUTION_UNKNOWN evidence, got %+v", event)
	}
	if got := NewSessionVerifier(ledger, NewEvaluator(store)).VerifySession(grant.AgentID); got.Result != "UNKNOWN" || got.PolicyCompliant {
		t.Fatalf("expected session verification UNKNOWN pending reconciliation, got %+v", got)
	}
}

func TestDecisionGateway_ExecuteAtomicPreservesUnknownWhenReceiptPersistenceFails(t *testing.T) {
	now := time.Now().UTC()
	store := NewPolicyStore()
	policy := &Policy{ID: "POLICY-UNKNOWN-RECEIPT", Version: 1, Status: "active", AllowedCapabilities: []Capability{"repository.write"}}
	if err := store.Set(policy); err != nil {
		t.Fatal(err)
	}
	grant := &Grant{GrantID: "GRANT-UNKNOWN-RECEIPT", AgentID: "agent-unknown-receipt", PolicyID: policy.ID, PolicyVersion: 1, Capabilities: []Capability{"repository.write"}, Status: "active", CreatedAt: now, ExpiresAt: now.Add(time.Hour)}
	root := t.TempDir()
	ledgerPath := filepath.Join(root, "ledger.json")
	ledger := NewLedger()
	ledger.SetPersistencePath(ledgerPath)
	gateway := NewDecisionGateway(NewEvaluator(store), ledger)
	req := AuthorizeRequest{RequestID: "REQ-UNKNOWN-RECEIPT", AgentID: grant.AgentID, Action: Action{AgentID: grant.AgentID, Capability: "repository.write", ArtifactHash: "sha256:unknown-receipt"}, Grant: grant, Policy: policy, Timestamp: now}
	authorized := gateway.Authorize(req)
	if authorized.Status != DecisionAllow {
		t.Fatalf("expected ALLOW authorization, got %s (%s)", authorized.Status, authorized.Message)
	}

	blocker := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(blocker, []byte("block"), 0o600); err != nil {
		t.Fatal(err)
	}
	sideEffectCommitted := false
	result, event := gateway.ExecuteAtomic(req, authorized, func() error {
		// The execution intent has already been durably recorded. Simulate the
		// persistence target becoming unavailable before the final receipt.
		ledger.SetPersistencePath(filepath.Join(blocker, "ledger.json"))
		sideEffectCommitted = true
		return nil
	})
	if !sideEffectCommitted {
		t.Fatal("expected side effect to have run after its intent was persisted")
	}
	if result.Status != DecisionUnknown || result.Reason != ReasonExecutionOutcomeUnknown {
		t.Fatalf("expected UNKNOWN after receipt persistence failure, got %s (%s)", result.Status, result.Message)
	}
	if event.ID != "" {
		t.Fatalf("expected no durable receipt to be returned, got %+v", event)
	}
	if got := ledger.PersistenceError(); got == nil {
		t.Fatal("expected ledger persistence error to be observable")
	}

	// Reload from the last durable snapshot: the unmatched intent must survive
	// restart and prevent a false PASS even though the final receipt is absent.
	reloaded, err := LoadLedger(ledgerPath)
	if err != nil {
		t.Fatalf("load durable ledger snapshot: %v", err)
	}
	verifyUnknown := func(candidate *Ledger) {
		t.Helper()
		if got := NewSessionVerifier(candidate, NewEvaluator(store)).VerifySession(grant.AgentID); got.Result != "UNKNOWN" || got.PolicyCompliant {
			t.Fatalf("expected session verifier to preserve UNKNOWN, got %+v", got)
		}
		if got := candidate.VerifySession(grant.AgentID); got.Result != "UNKNOWN" || got.PolicyCompliant {
			t.Fatalf("expected ledger verifier to preserve UNKNOWN, got %+v", got)
		}
	}
	verifyUnknown(reloaded)

	// A later successful append must carry the unmatched intent forward, not
	// erase uncertainty merely because the ledger can write again.
	ledger.SetPersistencePath(ledgerPath)
	if event := ledger.Append(LedgerEvent{Timestamp: now.Add(time.Second), AgentID: grant.AgentID, EventType: "RECONCILIATION_CHECKPOINT"}); event.ID == "" {
		t.Fatal("expected subsequent ledger append to succeed")
	}
	reloaded, err = LoadLedger(ledgerPath)
	if err != nil {
		t.Fatalf("reload ledger after subsequent append: %v", err)
	}
	verifyUnknown(reloaded)
}

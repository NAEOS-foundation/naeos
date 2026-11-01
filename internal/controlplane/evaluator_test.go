// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"testing"
	"time"
)

func TestEvaluator_AllowsCapabilityInPolicy(t *testing.T) {
	store := NewPolicyStore()
	policy := &Policy{
		ID:                    "POLICY-017",
		Version:               17,
		Status:                "active",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		AllowedCapabilities:   []Capability{"repository.read", "repository.write", "test.execute"},
		DeniedCapabilities:    nil,
		ProtectedCapabilities: []Capability{"credential.rotate", "policy.modify"},
	}
	if err := store.Set(policy); err != nil {
		t.Fatalf("store.Set() error = %v", err)
	}

	grant := &Grant{
		GrantID:       "GRANT-001",
		AgentID:       "agent-payment-01",
		PolicyID:      policy.ID,
		PolicyVersion: policy.Version,
		Capabilities:  []Capability{"repository.read", "repository.write", "test.execute"},
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		Status:        "active",
	}

	evaluator := NewEvaluator(store)
	result := evaluator.EvaluateAction(Action{AgentID: "agent-payment-01", Capability: "repository.read"}, grant, policy, time.Now())
	if result.Status != DecisionAllow {
		t.Fatalf("expected ALLOW verdict, got %s (%s)", result.Status, result.Message)
	}
}

func TestEvaluator_DeniesProtectedCapabilityEvenWhenGranted(t *testing.T) {
	store := NewPolicyStore()
	policy := &Policy{
		ID:                    "POLICY-017",
		Version:               17,
		Status:                "active",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		AllowedCapabilities:   []Capability{"repository.read"},
		ProtectedCapabilities: []Capability{"credential.rotate"},
	}
	if err := store.Set(policy); err != nil {
		t.Fatalf("store.Set() error = %v", err)
	}

	grant := &Grant{
		GrantID:       "GRANT-SECURE",
		AgentID:       "agent-malicious",
		PolicyID:      policy.ID,
		PolicyVersion: policy.Version,
		Capabilities:  []Capability{"credential.rotate", "repository.read"},
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		Status:        "active",
	}

	evaluator := NewEvaluator(store)
	result := evaluator.EvaluateAction(Action{AgentID: "agent-malicious", Capability: "credential.rotate"}, grant, policy, time.Now())
	if result.Status != DecisionDeny {
		t.Fatalf("expected DENY verdict, got %s (%s)", result.Status, result.Message)
	}
	if result.Reason != ReasonDeniedProtected {
		t.Fatalf("expected protected capability reason, got %s", result.Reason)
	}
}

func TestEvaluator_DeniesExpiredGrant(t *testing.T) {
	store := NewPolicyStore()
	policy := &Policy{
		ID:                  "POLICY-017",
		Version:             17,
		Status:              "active",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		AllowedCapabilities: []Capability{"repository.read"},
	}
	if err := store.Set(policy); err != nil {
		t.Fatalf("store.Set() error = %v", err)
	}

	grant := &Grant{
		GrantID:       "GRANT-EXPIRED",
		AgentID:       "agent-expired",
		PolicyID:      policy.ID,
		PolicyVersion: policy.Version,
		Capabilities:  []Capability{"repository.read"},
		CreatedAt:     time.Now().Add(-2 * time.Hour),
		ExpiresAt:     time.Now().Add(-1 * time.Hour),
		Status:        "active",
	}

	evaluator := NewEvaluator(store)
	result := evaluator.EvaluateAction(Action{AgentID: "agent-expired", Capability: "repository.read"}, grant, policy, time.Now())
	if result.Status != DecisionDeny {
		t.Fatalf("expected DENY verdict for expired grant, got %s (%s)", result.Status, result.Message)
	}
}

func TestEvaluator_IsDeterministic(t *testing.T) {
	store := NewPolicyStore()
	policy := &Policy{
		ID:                  "POLICY-017",
		Version:             17,
		Status:              "active",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		AllowedCapabilities: []Capability{"repository.read"},
	}

	if err := store.Set(policy); err != nil {
		t.Fatalf("store.Set() error = %v", err)
	}

	grant := &Grant{
		GrantID:       "GRANT-001",
		AgentID:       "agent-payment-01",
		PolicyID:      policy.ID,
		PolicyVersion: policy.Version,
		Capabilities:  []Capability{"repository.read"},
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		Status:        "active",
	}

	evaluator := NewEvaluator(store)
	first := evaluator.EvaluateAction(Action{AgentID: "agent-payment-01", Capability: "repository.read", Context: map[string]string{"env": "prod"}}, grant, policy, time.Now())
	second := evaluator.EvaluateAction(Action{AgentID: "agent-payment-01", Capability: "repository.read", Context: map[string]string{"env": "prod"}}, grant, policy, time.Now())
	if first.Status != second.Status || first.Reason != second.Reason {
		t.Fatalf("expected deterministic evaluation, got %v and %v", first, second)
	}
}

func TestEvaluator_DeniesGrantPolicyMismatch(t *testing.T) {
	store := NewPolicyStore()
	policy := &Policy{ID: "POLICY-A", Version: 1, Status: "active", AllowedCapabilities: []Capability{"repository.read"}}
	if err := store.Set(policy); err != nil {
		t.Fatal(err)
	}
	grant := &Grant{
		GrantID: "GRANT-A", AgentID: "agent-1", PolicyID: "POLICY-B", PolicyVersion: 1,
		Capabilities: []Capability{"repository.read"}, Status: "active", ExpiresAt: time.Now().Add(time.Hour),
	}
	result := NewEvaluator(store).EvaluateAction(Action{AgentID: "agent-1", Capability: "repository.read"}, grant, policy, time.Now())
	if result.Status != DecisionDeny || result.Reason != ReasonDeniedPolicyMismatch {
		t.Fatalf("expected policy mismatch deny, got %s (%s)", result.Status, result.Reason)
	}
}

func TestEvaluator_DeniesGrantAgentMismatch(t *testing.T) {
	store := NewPolicyStore()
	policy := &Policy{ID: "POLICY-A", Version: 1, Status: "active", AllowedCapabilities: []Capability{"repository.read"}}
	if err := store.Set(policy); err != nil {
		t.Fatal(err)
	}
	grant := &Grant{
		GrantID: "GRANT-A", AgentID: "agent-1", PolicyID: "POLICY-A", PolicyVersion: 1,
		Capabilities: []Capability{"repository.read"}, Status: "active", ExpiresAt: time.Now().Add(time.Hour),
	}
	result := NewEvaluator(store).EvaluateAction(Action{AgentID: "agent-2", Capability: "repository.read"}, grant, policy, time.Now())
	if result.Status != DecisionDeny || result.Reason != ReasonDeniedAgentMismatch {
		t.Fatalf("expected agent mismatch deny, got %s (%s)", result.Status, result.Reason)
	}
}

func TestPolicyStoreIsolatesPolicySnapshots(t *testing.T) {
	store := NewPolicyStore()
	input := &Policy{
		ID: "POLICY-IMMUTABLE", Version: 1, Status: "active",
		AllowedCapabilities: []Capability{"repository.read"},
	}
	if err := store.Set(input); err != nil {
		t.Fatal(err)
	}

	// Mutating the caller-owned input must not mutate the stored snapshot.
	input.AllowedCapabilities[0] = "repository.write"
	input.AllowedCapabilities = append(input.AllowedCapabilities, "production.delete")

	active, err := store.Active("POLICY-IMMUTABLE")
	if err != nil {
		t.Fatal(err)
	}
	if len(active.AllowedCapabilities) != 1 || active.AllowedCapabilities[0] != "repository.read" {
		t.Fatalf("stored policy changed through input pointer: %+v", active.AllowedCapabilities)
	}

	// Mutating returned snapshots must not mutate active or historical policy.
	active.AllowedCapabilities[0] = "production.delete"
	active.Status = "inactive"
	historical, err := store.Version("POLICY-IMMUTABLE", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(historical.AllowedCapabilities) != 1 || historical.AllowedCapabilities[0] != "repository.read" || historical.Status != "active" {
		t.Fatalf("stored policy changed through returned snapshot: %+v", historical)
	}
}

func TestPolicyStoreRejectsVersionReuseAndRollback(t *testing.T) {
	store := NewPolicyStore()
	v1 := &Policy{
		ID: "POLICY-VERSIONED", Version: 1, Status: "active",
		AllowedCapabilities: []Capability{"repository.read"},
	}
	if err := store.Set(v1); err != nil {
		t.Fatal(err)
	}

	mutatedV1 := clonePolicy(v1)
	mutatedV1.AllowedCapabilities = []Capability{"production.delete"}
	if err := store.Set(mutatedV1); err == nil {
		t.Fatal("expected same-version policy mutation to be rejected")
	}

	v2 := clonePolicy(v1)
	v2.Version = 2
	v2.AllowedCapabilities = []Capability{"repository.write"}
	if err := store.Set(v2); err != nil {
		t.Fatal(err)
	}
	if err := store.Set(v1); err == nil {
		t.Fatal("expected policy version rollback to be rejected")
	}

	active, err := store.Active("POLICY-VERSIONED")
	if err != nil {
		t.Fatal(err)
	}
	if active.Version != 2 || len(active.AllowedCapabilities) != 1 || active.AllowedCapabilities[0] != "repository.write" {
		t.Fatalf("unexpected active policy after rejected updates: %+v", active)
	}
}

func TestEvaluator_ReportsIndeterminateWhenEvaluationCannotBeCompleted(t *testing.T) {
	action := Action{AgentID: "agent-1", Capability: "repository.read"}
	if got := NewEvaluator().EvaluateAction(action, nil, nil, time.Now()); got.Status != DecisionIndeterminate || got.Reason != ReasonEvaluationFailed {
		t.Fatalf("unavailable evaluator must be indeterminate, got %+v", got)
	}
	store := NewPolicyStore()
	if got := NewEvaluator(store).Evaluate(AuthorizationContext{PolicyID: "missing", PolicyVersion: 1}, action); got.Status != DecisionIndeterminate || got.Reason != ReasonEvaluationFailed {
		t.Fatalf("missing policy version must be indeterminate, got %+v", got)
	}
	if got := NewEvaluator(store).EvaluateAction(action, nil, nil, time.Now()); got.Status != DecisionIndeterminate || got.Reason != ReasonEvaluationFailed {
		t.Fatalf("nil policy must be indeterminate, got %+v", got)
	}
}

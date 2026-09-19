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

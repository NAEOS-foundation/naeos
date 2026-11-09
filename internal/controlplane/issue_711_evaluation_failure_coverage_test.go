// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"testing"
	"time"
)

// TestEvaluator_EvaluationFailuresRemainIndeterminate expands the #711
// contract over unavailable stores, unresolved selectors, and malformed policy
// references. These are operational evaluation failures, not DENY decisions.
func TestEvaluator_EvaluationFailuresRemainIndeterminate(t *testing.T) {
	now := time.Now().UTC()
	activePolicy := &Policy{
		ID: "POLICY-711-COVERAGE", Version: 1, Status: "active",
		AllowedCapabilities: []Capability{"repository.write"},
	}
	store := NewPolicyStore()
	if err := store.Set(activePolicy); err != nil {
		t.Fatalf("store.Set() error = %v", err)
	}

	tests := []struct {
		name      string
		evaluator *Evaluator
		ctx       AuthorizationContext
	}{
		{
			name:      "evaluator unavailable",
			evaluator: nil,
			ctx:       AuthorizationContext{PolicyID: activePolicy.ID, PolicyVersion: activePolicy.Version, Now: now},
		},
		{
			name:      "policy store unavailable",
			evaluator: NewEvaluator(),
			ctx:       AuthorizationContext{PolicyID: activePolicy.ID, PolicyVersion: activePolicy.Version, Now: now},
		},
		{
			name:      "missing policy selector",
			evaluator: NewEvaluator(store),
			ctx:       AuthorizationContext{Now: now},
		},
		{
			name:      "unresolved policy version",
			evaluator: NewEvaluator(store),
			ctx:       AuthorizationContext{PolicyID: activePolicy.ID, PolicyVersion: 99, Now: now},
		},
		{
			name:      "malformed supplied policy reference",
			evaluator: NewEvaluator(store),
			ctx: AuthorizationContext{
				PolicyID: activePolicy.ID, PolicyVersion: activePolicy.Version,
				Policy: &Policy{ID: "POLICY-TAMPERED", Version: 1, Status: "active"},
				Now:    now,
			},
		},
	}

	action := Action{AgentID: "agent-711", Capability: "repository.write"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.evaluator.Evaluate(tt.ctx, action)
			if result.Status != DecisionIndeterminate || result.Reason != ReasonEvaluationFailed {
				t.Fatalf("expected internal indeterminate evaluation failure, got status=%s reason=%s message=%q",
					result.Status, result.Reason, result.Message)
			}
			if result.DecisionID != "" || result.ApprovedGrant != "" || result.NeedsApproval {
				t.Fatalf("evaluation failure must not mint decision/grant/approval state: %+v", result)
			}
		})
	}
}

// TestEvaluator_PolicyOutcomesRemainExplicit ensures the failure-state
// distinction does not change the valid policy outcomes ALLOW, DENY, or
// REQUIRE_APPROVAL.
func TestEvaluator_PolicyOutcomesRemainExplicit(t *testing.T) {
	now := time.Now().UTC()
	policy := &Policy{
		ID: "POLICY-711-OUTCOMES", Version: 1, Status: "active",
		AllowedCapabilities:  []Capability{"repository.read", "repository.write"},
		DeniedCapabilities:   []Capability{"repository.delete"},
		ApprovalRequired:     []Capability{"repository.write"},
	}
	store := NewPolicyStore()
	if err := store.Set(policy); err != nil {
		t.Fatalf("store.Set() error = %v", err)
	}
	grant := &Grant{
		GrantID: "GRANT-711-OUTCOMES", AgentID: "agent-711",
		PolicyID: policy.ID, PolicyVersion: policy.Version,
		Capabilities: []Capability{"repository.read", "repository.write", "repository.delete"},
		Status:       "active", CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}
	evaluator := NewEvaluator(store)

	tests := []struct {
		capability Capability
		want       DecisionStatus
		wantReason DecisionReason
	}{
		{capability: "repository.read", want: DecisionAllow, wantReason: ReasonAllowed},
		{capability: "repository.delete", want: DecisionDeny, wantReason: ReasonDeniedByPolicy},
		{capability: "repository.write", want: DecisionPending, wantReason: ReasonRequiresApproval},
	}
	for _, tt := range tests {
		t.Run(string(tt.capability), func(t *testing.T) {
			result := evaluator.EvaluateAction(
				Action{AgentID: grant.AgentID, Capability: tt.capability},
				grant, policy, now,
			)
			if result.Status != tt.want || result.Reason != tt.wantReason {
				t.Fatalf("expected status=%s reason=%s, got status=%s reason=%s message=%q",
					tt.want, tt.wantReason, result.Status, result.Reason, result.Message)
			}
		})
	}
}

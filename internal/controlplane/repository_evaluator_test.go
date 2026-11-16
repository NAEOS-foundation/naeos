// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"errors"
	"testing"
	"time"
)

type failingPolicyRepository struct {
	PolicyRepository
	activeErr  error
	versionErr error
	activeNil  bool
	versionNil bool
}

func (r failingPolicyRepository) Active(ctx context.Context, policyID string) (*Policy, error) {
	if r.activeErr != nil {
		return nil, r.activeErr
	}
	if r.activeNil {
		return nil, nil
	}
	return r.PolicyRepository.Active(ctx, policyID)
}

func (r failingPolicyRepository) Version(ctx context.Context, policyID string, version int) (*Policy, error) {
	if r.versionErr != nil {
		return nil, r.versionErr
	}
	if r.versionNil {
		return nil, nil
	}
	return r.PolicyRepository.Version(ctx, policyID, version)
}

func repositoryEvaluatorFixture(t *testing.T) (*RepositoryEvaluator, AuthorizationContext, Action, time.Time) {
	t.Helper()
	now := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	policy := &Policy{
		ID: "policy-repository-evaluator", Version: 1, Status: "active",
		AllowedCapabilities: []Capability{"repo.read"},
	}
	store := NewPolicyStore()
	if err := store.Set(policy); err != nil {
		t.Fatalf("set policy: %v", err)
	}
	repo := NewMemoryPolicyRepository(store)
	grant := &Grant{
		GrantID: "grant-repository-evaluator", AgentID: "agent-1",
		PolicyID: policy.ID, PolicyVersion: policy.Version,
		Capabilities: []Capability{"repo.read"}, Status: "active",
		CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour),
	}
	auth := AuthorizationContext{PolicyID: policy.ID, PolicyVersion: policy.Version, Grant: grant, Now: now}
	action := Action{AgentID: "agent-1", Capability: "repo.read"}
	return NewRepositoryEvaluator(repo), auth, action, now
}

func TestRepositoryEvaluatorAllowsUsingCanonicalPolicy(t *testing.T) {
	evaluator, auth, action, _ := repositoryEvaluatorFixture(t)
	// The supplied policy is only a selector. Its attempted widening must not
	// influence the result; the repository's canonical policy remains authoritative.
	auth.Policy = &Policy{ID: auth.PolicyID, Version: auth.PolicyVersion, Status: "active", AllowedCapabilities: []Capability{"admin.write"}}
	got := evaluator.Evaluate(context.Background(), auth, action)
	if got.Status != DecisionAllow || got.Reason != ReasonAllowed {
		t.Fatalf("Evaluate = %+v, want ALLOW", got)
	}
}

func TestRepositoryEvaluatorDeniesStalePolicyVersion(t *testing.T) {
	_, auth, action, _ := repositoryEvaluatorFixture(t)
	store := NewPolicyStore()
	if err := store.Set(&Policy{ID: auth.PolicyID, Version: 1, Status: "active", AllowedCapabilities: []Capability{"repo.read"}}); err != nil {
		t.Fatal(err)
	}
	if err := store.Set(&Policy{ID: auth.PolicyID, Version: 2, Status: "active", AllowedCapabilities: []Capability{"repo.read"}}); err != nil {
		t.Fatal(err)
	}
	evaluator := NewRepositoryEvaluator(NewMemoryPolicyRepository(store))
	got := evaluator.Evaluate(context.Background(), auth, action)
	if got.Status != DecisionDeny || got.Reason != ReasonDeniedStalePolicy {
		t.Fatalf("Evaluate = %+v, want stale-policy DENY", got)
	}
}

func TestRepositoryEvaluatorRepositoryFailuresAreIndeterminate(t *testing.T) {
	_, auth, action, _ := repositoryEvaluatorFixture(t)
	policyStore := NewPolicyStore()
	if err := policyStore.Set(&Policy{ID: auth.PolicyID, Version: auth.PolicyVersion, Status: "active", AllowedCapabilities: []Capability{"repo.read"}}); err != nil {
		t.Fatal(err)
	}
	base := NewMemoryPolicyRepository(policyStore)
	for name, repo := range map[string]PolicyRepository{
		"version":              failingPolicyRepository{PolicyRepository: base, versionErr: ErrRepositoryUnavailable},
		"active":               failingPolicyRepository{PolicyRepository: base, activeErr: ErrRepositoryUnavailable},
		"version nil response": failingPolicyRepository{PolicyRepository: base, versionNil: true},
		"active nil response":  failingPolicyRepository{PolicyRepository: base, activeNil: true},
	} {
		t.Run(name, func(t *testing.T) {
			got := NewRepositoryEvaluator(repo).Evaluate(context.Background(), auth, action)
			if got.Status != DecisionIndeterminate || got.Reason != ReasonEvaluationFailed {
				t.Fatalf("Evaluate = %+v, want INDETERMINATE", got)
			}
		})
	}
}

func TestRepositoryEvaluatorCanceledContextIsIndeterminate(t *testing.T) {
	evaluator, auth, action, _ := repositoryEvaluatorFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := evaluator.Evaluate(ctx, auth, action)
	if got.Status != DecisionIndeterminate || !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("Evaluate = %+v, context error = %v; want INDETERMINATE and cancellation", got, ctx.Err())
	}
}

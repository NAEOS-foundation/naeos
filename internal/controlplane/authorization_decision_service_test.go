// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"testing"
	"time"
)

func decisionServiceFixture(t *testing.T, approvalRequired bool) (*AuthorizationDecisionService, AuthorizeRequest, *repositoryGatewayTestLedger) {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UTC()
	policy := &Policy{
		ID: "decision-service-policy", Version: 1, Status: "active",
		AllowedCapabilities: []Capability{"artifact.read"},
	}
	if approvalRequired {
		policy.ApprovalRequired = []Capability{"artifact.read"}
	}
	store := NewPolicyStore()
	if err := store.Set(policy); err != nil {
		t.Fatal(err)
	}
	policies := NewMemoryPolicyRepository(store)
	grants := NewMemoryGrantRepository()
	grant := Grant{
		GrantID: "decision-service-grant", AgentID: "decision-service-agent",
		PolicyID: policy.ID, PolicyVersion: policy.Version,
		Capabilities: []Capability{"artifact.read"},
		CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour), Status: "active",
	}
	if err := grants.Create(ctx, grant); err != nil {
		t.Fatal(err)
	}
	ledger := &repositoryGatewayTestLedger{}
	service := NewAuthorizationDecisionService(policies, grants, nil, ledger)
	req := AuthorizeRequest{
		RequestID: "decision-service-request", AgentID: grant.AgentID,
		Action: Action{AgentID: grant.AgentID, Capability: "artifact.read", ArtifactHash: "sha256:decision-service"},
		Timestamp: now,
	}
	return service, req, ledger
}

func TestAuthorizationDecisionServiceDenialHasNoExecutionClaim(t *testing.T) {
	service, req, ledger := decisionServiceFixture(t, false)
	req.Action.Capability = "admin"
	result, err := service.Authorize(context.Background(), req)
	if err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	if result.Status != DecisionDeny {
		t.Fatalf("status = %s; want DENY", result.Status)
	}
	if got := ledger.countType("EXECUTION_STARTED"); got != 0 {
		t.Fatalf("execution claims = %d; want 0", got)
	}
	if got := ledger.countType("AUTHORIZATION_DECISION"); got != 1 {
		t.Fatalf("authorization evidence = %d; want 1", got)
	}
}

func TestAuthorizationDecisionServiceApprovalRequiredHasNoExecutionClaim(t *testing.T) {
	service, req, ledger := decisionServiceFixture(t, true)
	result, err := service.Authorize(context.Background(), req)
	if err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	if result.Status != DecisionPending || !result.NeedsApproval {
		t.Fatalf("result = %+v; want REQUIRE_APPROVAL", result)
	}
	if got := ledger.countType("EXECUTION_STARTED"); got != 0 {
		t.Fatalf("execution claims = %d; want 0", got)
	}
}

func TestAuthorizationDecisionServiceGrantResolutionFailureHasNoDecisionOrExecutionClaim(t *testing.T) {
	service, req, ledger := decisionServiceFixture(t, false)
	service.Grants = unavailableGrantRepository{GrantRepository: service.Grants}
	result, err := service.Authorize(context.Background(), req)
	if err == nil || result.Status != DecisionIndeterminate {
		t.Fatalf("result = %+v, err=%v; want INDETERMINATE with error", result, err)
	}
	if result.DecisionID != "" {
		t.Fatalf("indeterminate result minted decision ID %q", result.DecisionID)
	}
	if got := ledger.countType("AUTHORIZATION_DECISION"); got != 0 {
		t.Fatalf("authorization evidence = %d; want 0", got)
	}
	if got := ledger.countType("EXECUTION_STARTED"); got != 0 {
		t.Fatalf("execution claims = %d; want 0", got)
	}
}

type unavailablePolicyVersionRepository struct {
	PolicyRepository
}

func (r unavailablePolicyVersionRepository) Version(context.Context, string, int) (*Policy, error) {
	return nil, ErrRepositoryUnavailable
}

func TestAuthorizationDecisionServiceEvaluatorFailureHasNoDecisionOrExecutionClaim(t *testing.T) {
	service, req, ledger := decisionServiceFixture(t, false)
	policies := unavailablePolicyVersionRepository{PolicyRepository: service.Policies}
	service.Policies = policies
	service.Evaluator = NewRepositoryEvaluator(policies)

	result, err := service.Authorize(context.Background(), req)
	if err == nil || result.Status != DecisionIndeterminate {
		t.Fatalf("result = %+v, err=%v; want INDETERMINATE with error", result, err)
	}
	if result.DecisionID != "" {
		t.Fatalf("indeterminate result minted decision ID %q", result.DecisionID)
	}
	if got := ledger.countType("AUTHORIZATION_DECISION"); got != 0 {
		t.Fatalf("authorization evidence = %d; want 0", got)
	}
	if got := ledger.countType("EXECUTION_STARTED"); got != 0 {
		t.Fatalf("execution claims = %d; want 0", got)
	}
}

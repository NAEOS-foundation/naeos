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
		CreatedAt:    now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour), Status: "active",
	}
	if err := grants.Create(ctx, grant); err != nil {
		t.Fatal(err)
	}
	ledger := &repositoryGatewayTestLedger{}
	service := NewAuthorizationDecisionService(policies, grants, nil, ledger)
	req := AuthorizeRequest{
		RequestID: "decision-service-request", AgentID: grant.AgentID,
		Action:    Action{AgentID: grant.AgentID, Capability: "artifact.read", ArtifactHash: "sha256:decision-service"},
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


type unavailableApprovalRepository struct {
	ApprovalRepository
}

func (unavailableApprovalRepository) Get(context.Context, string) (Approval, error) {
	return Approval{}, ErrRepositoryUnavailable
}

func decisionServiceApprovedFixture(t *testing.T, decisionID, approvalID string, expiresAt time.Time) (*AuthorizationDecisionService, AuthorizeRequest, *repositoryGatewayTestLedger) {
	t.Helper()
	service, req, ledger := decisionServiceFixture(t, true)
	store := NewApprovalStore()
	repository := NewMemoryApprovalRepository(store)
	approval := Approval{
		ID: approvalID, DecisionID: decisionID, AgentID: req.AgentID,
		Capability: req.Action.Capability, PolicyID: "decision-service-policy",
		PolicyVersion: 1, Approver: "reviewer@example.test", Status: "pending",
		CreatedAt: req.Timestamp.Add(-time.Minute), ExpiresAt: expiresAt,
		ArtifactHash: req.Action.ArtifactHash,
	}
	if err := repository.Create(context.Background(), approval); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Transition(context.Background(), approvalID, "pending", "approved", req.Timestamp); err != nil {
		t.Fatal(err)
	}
	service.Approvals = repository
	req.DecisionID, req.ApprovalID = decisionID, approvalID
	return service, req, ledger
}

func TestAuthorizationDecisionServiceValidApprovalAllowsDecisionOnly(t *testing.T) {
	now := time.Now().UTC()
	service, req, ledger := decisionServiceApprovedFixture(t, "stable-decision-id", "approval-valid", now.Add(time.Hour))
	result, err := service.Authorize(context.Background(), req)
	if err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	if result.Status != DecisionAllow || result.DecisionID != req.DecisionID {
		t.Fatalf("result = %+v; want ALLOW bound to stable decision ID", result)
	}
	if got := ledger.countType("AUTHORIZATION_DECISION"); got != 1 {
		t.Fatalf("authorization evidence = %d; want 1", got)
	}
	if got := ledger.countType("EXECUTION_STARTED"); got != 0 {
		t.Fatalf("execution claims = %d; want 0; ALLOW is decision evidence only", got)
	}
}

func TestAuthorizationDecisionServiceApprovalBindingAndExpiryFailClosed(t *testing.T) {
	tests := []struct {
		name       string
		decisionID string
		artifact   string
		expiresAt  time.Time
	}{
		{name: "decision ID mismatch", decisionID: "different-decision", expiresAt: time.Now().Add(time.Hour)},
		{name: "artifact hash mismatch", decisionID: "stable-decision-id", artifact: "sha256:other", expiresAt: time.Now().Add(time.Hour)},
		{name: "expired approval", decisionID: "stable-decision-id", expiresAt: time.Now().Add(-time.Minute)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().UTC()
			service, req, ledger := decisionServiceApprovedFixture(t, tt.decisionID, "approval-"+tt.name, tt.expiresAt)
			if tt.artifact != "" {
				store := service.Approvals.(*MemoryApprovalRepository).store
				store.mu.Lock()
				approval := store.approvals[req.ApprovalID]
				approval.ArtifactHash = tt.artifact
				store.approvals[req.ApprovalID] = approval
				store.mu.Unlock()
			}
			req.Timestamp = now
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
		})
	}
}

func TestAuthorizationDecisionServiceApprovalRepositoryOutageIsIndeterminateWithoutDecisionEvidence(t *testing.T) {
	service, req, ledger := decisionServiceFixture(t, true)
	service.Approvals = unavailableApprovalRepository{}
	req.DecisionID, req.ApprovalID = "stable-decision-id", "approval-unavailable"
	result, err := service.Authorize(context.Background(), req)
	if err == nil || result.Status != DecisionIndeterminate {
		t.Fatalf("result = %+v, err=%v; want INDETERMINATE with error", result, err)
	}
	if result.DecisionID != "" {
		t.Fatalf("indeterminate result retained decision ID %q", result.DecisionID)
	}
	if got := ledger.countType("AUTHORIZATION_DECISION"); got != 0 {
		t.Fatalf("authorization decision evidence = %d; want 0", got)
	}
	if got := ledger.countType("EXECUTION_STARTED"); got != 0 {
		t.Fatalf("execution claims = %d; want 0", got)
	}
}

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// AuthorizationDecisionService is the decision-only authorization boundary.
// It has no adapter, executor, sandbox, or execution callback dependency.
// Its only side effect is committing the authorization decision evidence.
type AuthorizationDecisionService struct {
	Policies  PolicyRepository
	Grants    GrantRepository
	Approvals ApprovalRepository
	Ledger    LedgerRepository
	Evaluator *RepositoryEvaluator
}

// NewAuthorizationDecisionService wires only repositories and the policy evaluator.
// Nil dependencies are retained so calls fail closed.
func NewAuthorizationDecisionService(policies PolicyRepository, grants GrantRepository, approvals ApprovalRepository, ledger LedgerRepository) *AuthorizationDecisionService {
	return &AuthorizationDecisionService{
		Policies: policies, Grants: grants, Approvals: approvals, Ledger: ledger,
		Evaluator: NewRepositoryEvaluator(policies),
	}
}

// Authorize resolves canonical identity-scoped grant and policy state, evaluates
// the requested action, and returns a decision only after evidence commit.
// An ALLOW result is not an execution instruction; execution must pass through
// a separately composed execution gateway with fresh boundary validation.
func (s *AuthorizationDecisionService) Authorize(ctx context.Context, req AuthorizeRequest) (DecisionResult, error) {
	if ctx == nil || ctx.Err() != nil {
		return DecisionResult{Status: DecisionIndeterminate, Reason: ReasonEvaluationFailed, Message: "authorization context unavailable"}, contextError(ctx)
	}
	if s == nil || s.Policies == nil || s.Grants == nil || s.Ledger == nil || s.Evaluator == nil {
		return DecisionResult{Status: DecisionIndeterminate, Reason: ReasonEvaluationFailed, Message: "repository gateway dependencies unavailable"}, ErrRepositoryUnavailable
	}
	now := req.Timestamp.UTC()
	if req.Timestamp.IsZero() {
		now = time.Now().UTC()
	}
	requestID := req.RequestID
	if requestID == "" {
		requestID = controlPlaneID("REQ")
	}
	action := normalizeRequest(req.Action)
	if req.AgentID == "" || action.AgentID != req.AgentID || action.Capability == "" {
		return DecisionResult{Status: DecisionIndeterminate, Reason: ReasonEvaluationFailed, RequestID: requestID, AgentID: req.AgentID, Requested: action.Capability, Message: "request and action agent/capability bindings are required"}, ErrRepositoryIntegrity
	}

	grant, err := s.Grants.GetByAgent(ctx, req.AgentID)
	if err != nil {
		status, reason := DecisionIndeterminate, ReasonEvaluationFailed
		if errors.Is(err, ErrRepositoryNotFound) || errors.Is(err, ErrRepositoryConflict) {
			status, reason = DecisionDeny, ReasonDeniedByGrant
		}
		return DecisionResult{Status: status, Reason: reason, RequestID: requestID, AgentID: req.AgentID, Requested: action.Capability, Message: fmt.Sprintf("resolve canonical grant: %v", err)}, err
	}
	if grant.AgentID != req.AgentID || grant.GrantID == "" || grant.PolicyID == "" || grant.PolicyVersion < 1 {
		return DecisionResult{Status: DecisionIndeterminate, Reason: ReasonEvaluationFailed, RequestID: requestID, AgentID: req.AgentID, Requested: action.Capability, Message: "canonical grant bindings failed integrity validation"}, ErrRepositoryIntegrity
	}

	active, err := s.Policies.Active(ctx, grant.PolicyID)
	if err != nil {
		return DecisionResult{Status: DecisionIndeterminate, Reason: ReasonEvaluationFailed, RequestID: requestID, AgentID: req.AgentID, Requested: action.Capability, Message: fmt.Sprintf("resolve active policy: %v", err)}, err
	}
	if active == nil || active.ID != grant.PolicyID || active.Status != "active" {
		return DecisionResult{Status: DecisionIndeterminate, Reason: ReasonEvaluationFailed, RequestID: requestID, AgentID: req.AgentID, Requested: action.Capability, Message: "active policy repository response failed integrity validation"}, ErrRepositoryIntegrity
	}

	auth := AuthorizationContext{PolicyID: grant.PolicyID, PolicyVersion: grant.PolicyVersion, Grant: &grant, Now: now}
	result := s.Evaluator.Evaluate(ctx, auth, action)
	result.RequestID, result.AgentID, result.ArtifactHash = requestID, req.AgentID, action.ArtifactHash
	if result.Status == DecisionIndeterminate {
		return result, fmt.Errorf("policy evaluation indeterminate: %s", result.Message)
	}
	if req.DecisionID != "" {
		result.DecisionID = req.DecisionID
	} else {
		result.DecisionID = controlPlaneID("DEC")
	}

	if result.Status == DecisionPending && req.ApprovalID != "" {
		if s.Approvals == nil {
			result.Status, result.Reason, result.Message = DecisionIndeterminate, ReasonEvaluationFailed, "approval repository unavailable"
		} else {
			approval, approvalErr := s.Approvals.Get(ctx, req.ApprovalID)
			switch {
			case approvalErr != nil:
				result.Status, result.Reason = DecisionIndeterminate, ReasonEvaluationFailed
				result.Message = "approval repository lookup failed"
				if errors.Is(approvalErr, ErrRepositoryNotFound) || errors.Is(approvalErr, ErrRepositoryConflict) {
					result.Status, result.Reason = DecisionDeny, ReasonDeniedByPolicy
					result.Message = "approval is missing or no longer consumable"
				}
			case approval.ID != req.ApprovalID || approval.DecisionID != result.DecisionID ||
				approval.AgentID != req.AgentID || approval.Capability != action.Capability ||
				approval.PolicyID != grant.PolicyID || approval.PolicyVersion != grant.PolicyVersion ||
				approval.ArtifactHash != action.ArtifactHash:
				result.Status, result.Reason, result.Message = DecisionDeny, ReasonDeniedByPolicy, "approval binding does not match the requested decision and artifact"
			case !approval.IsValid(now):
				result.Status, result.Reason, result.Message = DecisionDeny, ReasonDeniedByPolicy, "approval is not approved or has expired"
			default:
				result.Status, result.Reason, result.NeedsApproval = DecisionAllow, ReasonAllowed, false
				result.Message = "approval validated against durable repository"
			}
		}
		}
		if result.Status == DecisionIndeterminate {
			result.DecisionID = ""
			return result, fmt.Errorf("approval validation indeterminate: %s", result.Message)
		}
	}

	event := LedgerEvent{
		Timestamp: now, RequestID: requestID, DecisionID: result.DecisionID,
		AgentID: req.AgentID, Capability: action.Capability, ArtifactHash: action.ArtifactHash,
		EventType: "AUTHORIZATION_DECISION", Decision: result.Status, Reason: result.Reason,
		Metadata: map[string]string{
			"policy_id": grant.PolicyID, "policy_version": fmt.Sprintf("%d", grant.PolicyVersion),
			"grant_id": grant.GrantID,
		},
	}
	committed, _, err := s.Ledger.Append(ctx, event)
	if err != nil || committed.ID == "" {
		result.Status, result.DecisionID, result.Reason = DecisionIndeterminate, "", ReasonEvaluationFailed
		result.Message = "authorization evidence append failed; decision is not usable"
		if err != nil {
			return result, err
		}
		return result, ErrRepositoryUnavailable
	}
	return result, nil
}

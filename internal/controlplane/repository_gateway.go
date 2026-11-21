// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// RepositoryDecisionGateway is the context-aware authorization boundary intended
// for production composition. It resolves grants and policy from repositories;
// caller-provided Grant/Policy payloads are never authoritative.
type RepositoryDecisionGateway struct {
	Policies  PolicyRepository
	Grants    GrantRepository
	Approvals ApprovalRepository
	Ledger    LedgerRepository
	Evaluator *RepositoryEvaluator
}

// NewRepositoryDecisionGateway wires repository-backed authorization components.
// Nil dependencies are retained so calls fail closed rather than silently
// substituting in-memory state.
func NewRepositoryDecisionGateway(policies PolicyRepository, grants GrantRepository, approvals ApprovalRepository, ledger LedgerRepository) *RepositoryDecisionGateway {
	return &RepositoryDecisionGateway{
		Policies: policies, Grants: grants, Approvals: approvals, Ledger: ledger,
		Evaluator: NewRepositoryEvaluator(policies),
	}
}

// Authorize resolves the canonical grant and policy, evaluates the request, and
// returns a usable decision only after its evidence append has committed.
func (g *RepositoryDecisionGateway) Authorize(ctx context.Context, req AuthorizeRequest) (DecisionResult, error) {
	if g == nil {
		return DecisionResult{Status: DecisionIndeterminate, Reason: ReasonEvaluationFailed, Message: "repository gateway unavailable"}, ErrRepositoryUnavailable
	}
	service := NewAuthorizationDecisionService(g.Policies, g.Grants, g.Approvals, g.Ledger)
	return service.Authorize(ctx, req)
}

// Execute persists an execution intent before calling the external side effect.
// It revalidates the durable decision, current policy, and grant at the boundary.
// A callback error is recorded as UNKNOWN because the side effect may have
// completed despite a lost/ambiguous response; callers must reconcile, not retry.
func (g *RepositoryDecisionGateway) Execute(ctx context.Context, req AuthorizeRequest, result DecisionResult, sideEffect func(context.Context) error) (DecisionResult, LedgerEvent, error) {
	if ctx == nil || ctx.Err() != nil {
		return result, LedgerEvent{}, contextError(ctx)
	}
	if g == nil || g.Policies == nil || g.Grants == nil || g.Ledger == nil {
		result.Status, result.DecisionID, result.Reason = DecisionIndeterminate, "", ReasonEvaluationFailed
		result.Message = "repository execution boundary unavailable"
		return result, LedgerEvent{}, ErrRepositoryUnavailable
	}
	if result.Status != DecisionAllow || result.DecisionID == "" || sideEffect == nil {
		result.Status, result.Reason = DecisionDeny, ReasonDeniedByPolicy
		result.Message = "only a canonical ALLOW decision and a side-effect callback are executable"
		return result, LedgerEvent{}, ErrRepositoryConflict
	}

	canonical, err := g.Ledger.Decision(ctx, result.DecisionID)
	if err != nil {
		result.Status, result.DecisionID, result.Reason = DecisionIndeterminate, "", ReasonEvaluationFailed
		result.Message = "canonical decision lookup could not be verified"
		if errors.Is(err, ErrRepositoryNotFound) || errors.Is(err, ErrRepositoryConflict) {
			result.Status, result.Reason = DecisionDeny, ReasonDeniedByPolicy
			result.Message = "canonical authorization decision does not exist or is no longer valid"
		}
		return result, LedgerEvent{}, err
	}
	if canonical.Decision != DecisionAllow || canonical.RequestID != result.RequestID ||
		canonical.AgentID != req.AgentID || canonical.Capability != req.Action.Capability ||
		canonical.ArtifactHash != req.Action.ArtifactHash {
		result.Status, result.Reason = DecisionDeny, ReasonDeniedByPolicy
		result.Message = "decision is not canonically bound to this execution request"
		return result, LedgerEvent{}, ErrRepositoryConflict
	}
	already, err := g.Ledger.HasExecution(ctx, result.DecisionID)
	if err != nil {
		result.Status, result.Reason = DecisionIndeterminate, ReasonEvaluationFailed
		result.Message = "execution state could not be verified; side effect was not invoked"
		return result, LedgerEvent{}, err
	}
	if already {
		result.Status, result.Reason = DecisionDeny, ReasonDeniedByPolicy
		result.Message = "execution already exists for this decision"
		return result, LedgerEvent{}, ErrRepositoryConflict
	}

	policyID := canonical.Metadata["policy_id"]
	var policyVersion int
	if _, err := fmt.Sscanf(canonical.Metadata["policy_version"], "%d", &policyVersion); err != nil || policyID == "" || policyVersion < 1 {
		result.Status, result.Reason = DecisionIndeterminate, ReasonEvaluationFailed
		result.Message = "canonical decision policy selector is invalid"
		return result, LedgerEvent{}, ErrRepositoryIntegrity
	}
	if err := g.Policies.CheckActive(ctx, policyID, policyVersion); err != nil {
		result.Status, result.Reason = DecisionIndeterminate, ReasonEvaluationFailed
		result.Message = "active policy state could not be verified"
		if errors.Is(err, ErrRepositoryConflict) {
			result.Status, result.Reason = DecisionDeny, ReasonDeniedStalePolicy
			result.Message = "authorization policy is no longer active"
		}
		return result, LedgerEvent{}, err
	}
	grant, err := g.Grants.GetByAgent(ctx, req.AgentID)
	if err != nil {
		result.Status, result.Reason = DecisionIndeterminate, ReasonEvaluationFailed
		result.Message = "canonical grant state could not be verified"
		if errors.Is(err, ErrRepositoryNotFound) || errors.Is(err, ErrRepositoryConflict) {
			result.Status, result.Reason = DecisionDeny, ReasonDeniedByGrant
			result.Message = "canonical grant is missing or revoked"
		}
		return result, LedgerEvent{}, err
	}
	if !grant.IsValid(time.Now().UTC()) || grant.PolicyID != policyID ||
		grant.PolicyVersion != policyVersion || !containsCapability(grant.Capabilities, req.Action.Capability) {
		result.Status, result.Reason = DecisionDeny, ReasonDeniedByGrant
		result.Message = "canonical grant is invalid or no longer bound to this action"
		return result, LedgerEvent{}, ErrRepositoryConflict
	}

	executionID := controlPlaneID("EXEC")
	intent := LedgerEvent{
		Timestamp: time.Now().UTC(), RequestID: result.RequestID, DecisionID: result.DecisionID,
		ExecutionID: executionID, AgentID: req.AgentID, Capability: req.Action.Capability,
		ArtifactHash: req.Action.ArtifactHash, EventType: "EXECUTION_STARTED",
		Decision: DecisionAllow, Reason: result.Reason,
		Metadata: map[string]string{"policy_id": policyID, "policy_version": fmt.Sprintf("%d", policyVersion), "grant_id": grant.GrantID},
	}
	started, _, err := g.Ledger.Append(ctx, intent)
	if err != nil {
		if errors.Is(err, ErrRepositoryConflict) {
			result.Status, result.Reason = DecisionDeny, ReasonDeniedByPolicy
			result.Message = "another execution already claimed this decision; side effect was not invoked"
			return result, LedgerEvent{}, err
		}
		result.Status, result.DecisionID, result.Reason = DecisionIndeterminate, "", ReasonEvaluationFailed
		result.Message = "durable execution intent could not be committed; side effect was not invoked"
		return result, LedgerEvent{}, err
	}
	if started.ID == "" {
		result.Status, result.DecisionID, result.Reason = DecisionIndeterminate, "", ReasonEvaluationFailed
		result.Message = "durable execution intent returned no committed event; side effect was not invoked"
		return result, LedgerEvent{}, ErrRepositoryUnavailable
	}

	if req.ApprovalID != "" {
		if g.Approvals == nil {
			result.Status, result.Reason = DecisionIndeterminate, ReasonEvaluationFailed
			result.Message = "approval repository unavailable after intent commit; execution requires reconciliation"
			return result, started, ErrRepositoryUnavailable
		}
		if _, err := g.Approvals.Transition(ctx, req.ApprovalID, "approved", "consumed", time.Now().UTC()); err != nil {
			if !errors.Is(err, ErrRepositoryNotFound) && !errors.Is(err, ErrRepositoryConflict) {
				result.Status, result.Reason = DecisionIndeterminate, ReasonEvaluationFailed
				result.Message = "approval consumption could not be verified; execution intent requires reconciliation"
				return result, started, err
			}
			blocked, _, appendErr := g.Ledger.Append(ctx, LedgerEvent{
				Timestamp: time.Now().UTC(), RequestID: result.RequestID, DecisionID: result.DecisionID,
				ExecutionID: executionID, AgentID: req.AgentID, Capability: req.Action.Capability,
				ArtifactHash: req.Action.ArtifactHash, EventType: "EXECUTION_BLOCKED", Decision: DecisionDeny,
				Reason: ReasonDeniedByPolicy, Metadata: map[string]string{"approval_error": err.Error()},
			})
			result.Status, result.Reason = DecisionDeny, ReasonDeniedByPolicy
			result.Message = "approval was missing or already consumed; side effect was not invoked"
			if appendErr != nil || blocked.ID == "" {
				result.Status, result.Reason = DecisionIndeterminate, ReasonEvaluationFailed
				result.Message = "approval rejection evidence append failed; reconcile execution intent"
				if appendErr != nil {
					return result, started, appendErr
				}
				return result, started, ErrRepositoryUnavailable
			}
			return result, blocked, err
		}
	}

	if err := sideEffect(ctx); err != nil {
		unknown, _, appendErr := g.Ledger.Append(ctx, LedgerEvent{
			Timestamp: time.Now().UTC(), RequestID: result.RequestID, DecisionID: result.DecisionID,
			ExecutionID: executionID, AgentID: req.AgentID, Capability: req.Action.Capability,
			ArtifactHash: req.Action.ArtifactHash, EventType: "EXECUTION_UNKNOWN", Decision: DecisionUnknown,
			Reason: ReasonExecutionOutcomeUnknown, Metadata: map[string]string{"side_effect_error": err.Error()},
		})
		result.Status, result.Reason = DecisionUnknown, ReasonExecutionOutcomeUnknown
		result.Message = "side-effect outcome is ambiguous; reconcile before retrying"
		if appendErr != nil || unknown.ID == "" {
			result.Status, result.Reason = DecisionIndeterminate, ReasonEvaluationFailed
			result.Message = "side-effect outcome is ambiguous and UNKNOWN evidence append failed"
			if appendErr != nil {
				return result, started, appendErr
			}
			return result, started, ErrRepositoryUnavailable
		}
		return result, unknown, err
	}

	receipt, _, err := g.Ledger.Append(ctx, LedgerEvent{
		Timestamp: time.Now().UTC(), RequestID: result.RequestID, DecisionID: result.DecisionID,
		ExecutionID: executionID, AgentID: req.AgentID, Capability: req.Action.Capability,
		ArtifactHash: req.Action.ArtifactHash, EventType: "EXECUTION_ALLOWED", Decision: DecisionAllow,
		Reason: result.Reason, Metadata: map[string]string{"policy_id": policyID, "grant_id": grant.GrantID},
	})
	if err != nil || receipt.ID == "" {
		result.Status, result.Reason = DecisionUnknown, ReasonExecutionOutcomeUnknown
		result.Message = "side effect returned success but receipt append failed; reconcile before retrying"
		if err != nil {
			return result, started, err
		}
		return result, started, ErrRepositoryUnavailable
	}
	return result, receipt, nil
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return ErrRepositoryUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return ErrRepositoryUnavailable
}

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
	if g.Evaluator != nil {
		service.Evaluator = g.Evaluator
	}
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

	// Revalidate the exact approval at the execution boundary. ApprovalID is
	// caller-controlled input and may differ from the one checked at Authorize.
	// The approved row is then consumed atomically with the execution intent.
	var approvedApproval *Approval
	if req.ApprovalID != "" {
		if g.Approvals == nil {
			result.Status, result.Reason = DecisionIndeterminate, ReasonEvaluationFailed
			result.Message = "approval repository unavailable; side effect was not invoked"
			return result, LedgerEvent{}, ErrRepositoryUnavailable
		}
		approval, approvalErr := g.Approvals.Get(ctx, req.ApprovalID)
		if approvalErr != nil {
			result.Status, result.Reason = DecisionIndeterminate, ReasonEvaluationFailed
			result.Message = "approval could not be revalidated at execution boundary"
			if errors.Is(approvalErr, ErrRepositoryNotFound) || errors.Is(approvalErr, ErrRepositoryConflict) {
				result.Status, result.Reason = DecisionDeny, ReasonDeniedByPolicy
				result.Message = "approval is missing or no longer consumable"
			}
			return result, LedgerEvent{}, approvalErr
		}
		if approval.ID != req.ApprovalID || approval.DecisionID != canonical.DecisionID ||
			approval.AgentID != req.AgentID || approval.Capability != req.Action.Capability ||
			approval.PolicyID != policyID || approval.PolicyVersion != policyVersion ||
			approval.ArtifactHash != req.Action.ArtifactHash || !approval.IsValid(time.Now().UTC()) {
			result.Status, result.Reason = DecisionDeny, ReasonDeniedByPolicy
			result.Message = "approval is not valid for this canonical execution request"
			return result, LedgerEvent{}, ErrRepositoryConflict
		}
		approvedApproval = &approval
	}

	executionID := controlPlaneID("EXEC")
	intentMetadata := map[string]string{"policy_id": policyID, "policy_version": fmt.Sprintf("%d", policyVersion), "grant_id": grant.GrantID}
	if req.ApprovalID != "" {
		intentMetadata["approval_id"] = req.ApprovalID
	}
	intent := LedgerEvent{
		Timestamp: time.Now().UTC(), RequestID: result.RequestID, DecisionID: result.DecisionID,
		ExecutionID: executionID, AgentID: req.AgentID, Capability: req.Action.Capability,
		ArtifactHash: req.Action.ArtifactHash, EventType: "EXECUTION_STARTED",
		Decision: DecisionAllow, Reason: result.Reason, Metadata: intentMetadata,
	}
	var started LedgerEvent
	if approvedApproval != nil {
		atomicClaims, ok := g.Ledger.(AtomicExecutionClaimRepository)
		if !ok {
			result.Status, result.Reason = DecisionIndeterminate, ReasonEvaluationFailed
			result.Message = "atomic approval consumption is unsupported; side effect was not invoked"
			return result, LedgerEvent{}, ErrRepositoryUnavailable
		}
		started, _, err = atomicClaims.ClaimExecution(ctx, intent, *approvedApproval, time.Now().UTC())
	} else {
		started, _, err = g.Ledger.Append(ctx, intent)
	}
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

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"fmt"
)

// RepositoryEvaluator evaluates actions using authoritative context-aware policy
// persistence. It deliberately constructs an ephemeral in-memory evaluator for
// the pure policy rules after reading canonical policy state; repository state
// remains authoritative and caller-supplied policy contents are never trusted.
type RepositoryEvaluator struct {
	repository PolicyRepository
}

// NewRepositoryEvaluator creates an evaluator over a durable or test policy repository.
func NewRepositoryEvaluator(repository PolicyRepository) *RepositoryEvaluator {
	return &RepositoryEvaluator{repository: repository}
}

// Evaluate resolves the selected immutable policy version and active pointer
// from the repository, then runs the existing deterministic policy rules.
// Any repository failure is INDETERMINATE, not a policy DENY.
func (e *RepositoryEvaluator) Evaluate(ctx context.Context, auth AuthorizationContext, action Action) DecisionResult {
	if ctx == nil {
		return indeterminatePolicyResult(action, ErrRepositoryUnavailable)
	}
	if err := ctx.Err(); err != nil {
		return indeterminatePolicyResult(action, err)
	}
	if e == nil || e.repository == nil {
		return indeterminatePolicyResult(action, ErrRepositoryUnavailable)
	}

	policyID, policyVersion := auth.PolicyID, auth.PolicyVersion
	if auth.Policy != nil {
		if policyID == "" {
			policyID = auth.Policy.ID
		}
		if policyVersion == 0 {
			policyVersion = auth.Policy.Version
		}
		if auth.Policy.ID != policyID || auth.Policy.Version != policyVersion {
			return DecisionResult{
				Status: DecisionIndeterminate, Reason: ReasonEvaluationFailed,
				PolicyID: policyID, PolicyVersion: policyVersion, Requested: action.Capability,
				Message: "supplied policy reference does not match requested policy selector",
			}
		}
	}
	if auth.Grant != nil {
		if policyID == "" {
			policyID = auth.Grant.PolicyID
		}
		if policyVersion == 0 {
			policyVersion = auth.Grant.PolicyVersion
		}
	}

	versioned, err := e.repository.Version(ctx, policyID, policyVersion)
	if err != nil {
		return indeterminatePolicyResult(action, fmt.Errorf("resolve policy version: %w", err))
	}
	if versioned == nil {
		return indeterminatePolicyResult(action, fmt.Errorf("resolve policy version: %w", ErrRepositoryIntegrity))
	}
	active, err := e.repository.Active(ctx, policyID)
	if err != nil {
		return indeterminatePolicyResult(action, fmt.Errorf("resolve active policy: %w", err))
	}
	if active == nil {
		return indeterminatePolicyResult(action, fmt.Errorf("resolve active policy: %w", ErrRepositoryIntegrity))
	}
	if active.ID != policyID || active.Status != "active" {
		return indeterminatePolicyResult(action, ErrRepositoryIntegrity)
	}
	if versioned.ID != policyID || versioned.Version != policyVersion {
		return indeterminatePolicyResult(action, ErrRepositoryIntegrity)
	}
	if active.Version != policyVersion {
		return DecisionResult{
			Status: DecisionDeny, Reason: ReasonDeniedStalePolicy,
			PolicyID: active.ID, PolicyVersion: active.Version, Requested: action.Capability,
			Message: fmt.Sprintf("selected policy %s v%d is not active (active v%d)", policyID, policyVersion, active.Version),
		}
	}
	if versioned.Status != "active" {
		return DecisionResult{
			Status: DecisionDeny, Reason: ReasonDeniedStalePolicy,
			PolicyID: versioned.ID, PolicyVersion: versioned.Version, Requested: action.Capability,
			Message: fmt.Sprintf("policy %s v%d is not active", versioned.ID, versioned.Version),
		}
	}

	store := NewPolicyStore()
	if err := store.Set(active); err != nil {
		return indeterminatePolicyResult(action, fmt.Errorf("load canonical active policy: %w", err))
	}
	canonicalAuth := auth
	canonicalAuth.PolicyID = active.ID
	canonicalAuth.PolicyVersion = active.Version
	canonicalAuth.Policy = nil // policy contents are loaded from the repository above
	return NewEvaluator(store).Evaluate(canonicalAuth, normalizeRequest(action))
}

func indeterminatePolicyResult(action Action, err error) DecisionResult {
	return DecisionResult{
		Status:    DecisionIndeterminate,
		Reason:    ReasonEvaluationFailed,
		Requested: action.Capability,
		Message:   err.Error(),
	}
}

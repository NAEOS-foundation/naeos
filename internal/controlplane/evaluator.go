// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"
)

// PolicyStore stores policy definitions by ID and version.
type PolicyStore struct {
	mu        sync.RWMutex
	current   map[string]*Policy
	byVersion map[string]map[int]*Policy
}

// NewPolicyStore creates a policy store for deterministic evaluation.
func NewPolicyStore() *PolicyStore {
	return &PolicyStore{
		current:   make(map[string]*Policy),
		byVersion: make(map[string]map[int]*Policy),
	}
}

// Set stores an immutable snapshot of a policy as the active version.
// Callers cannot mutate stored policy state through the input pointer.
func (ps *PolicyStore) Set(policy *Policy) error {
	if policy == nil {
		return fmt.Errorf("policy must not be nil")
	}
	candidate := clonePolicy(policy)
	if candidate.ID == "" {
		return fmt.Errorf("policy ID is required")
	}
	if candidate.Version <= 0 {
		return fmt.Errorf("policy version must be greater than zero")
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if current, ok := ps.current[candidate.ID]; ok {
		switch {
		case candidate.Version < current.Version:
			return fmt.Errorf("policy version rollback is not allowed: %s v%d is active", candidate.ID, current.Version)
		case candidate.Version == current.Version:
			if !reflect.DeepEqual(current, candidate) {
				return fmt.Errorf("policy contents changed without incrementing policy version: %s v%d", candidate.ID, candidate.Version)
			}
			return nil
		}
	}
	if _, ok := ps.byVersion[candidate.ID]; !ok {
		ps.byVersion[candidate.ID] = make(map[int]*Policy)
	}
	ps.current[candidate.ID] = candidate
	ps.byVersion[candidate.ID][candidate.Version] = candidate
	return nil
}

// clonePolicy copies all mutable fields so policy snapshots cannot be changed
// through slices or pointers retained by callers.
func clonePolicy(policy *Policy) *Policy {
	if policy == nil {
		return nil
	}
	cloned := *policy
	cloned.AllowedCapabilities = append([]Capability(nil), policy.AllowedCapabilities...)
	cloned.DeniedCapabilities = append([]Capability(nil), policy.DeniedCapabilities...)
	cloned.ProtectedCapabilities = append([]Capability(nil), policy.ProtectedCapabilities...)
	cloned.ApprovalRequired = append([]Capability(nil), policy.ApprovalRequired...)
	return &cloned
}

// Active returns the active policy for the given policy ID.
func (ps *PolicyStore) Active(policyID string) (*Policy, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	policy, ok := ps.current[policyID]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPolicyNotFound, policyID)
	}
	if policy.Status != "active" {
		return nil, fmt.Errorf("%w: %s (status=%s)", ErrPolicyInactive, policyID, policy.Status)
	}
	return clonePolicy(policy), nil
}

// Version returns a specific policy version.
func (ps *PolicyStore) Version(policyID string, version int) (*Policy, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	byVersion, ok := ps.byVersion[policyID]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPolicyNotFound, policyID)
	}
	policy, ok := byVersion[version]
	if !ok {
		return nil, fmt.Errorf("%w: %s v%d", ErrPolicyVersionNotFound, policyID, version)
	}
	return clonePolicy(policy), nil
}

// WithExecutionLock validates the active policy and executes fn while policy updates are excluded.
// This provides the in-process atomic commit boundary used by ExecuteAtomic.
func (ps *PolicyStore) WithExecutionLock(policyID string, version int, fn func(*Policy) error) error {
	if ps == nil {
		return fmt.Errorf("policy store unavailable")
	}
	if fn == nil {
		return fmt.Errorf("execution callback is required")
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()
	policy, ok := ps.current[policyID]
	if !ok {
		return fmt.Errorf("%w: %s", ErrPolicyNotFound, policyID)
	}
	if policy.Status != "active" {
		return fmt.Errorf("%w: %s (status=%s)", ErrPolicyInactive, policyID, policy.Status)
	}
	if policy.Version != version {
		return fmt.Errorf("%w: %s v%d is not active", ErrPolicyVersionNotFound, policyID, version)
	}
	return fn(clonePolicy(policy))
}

// Evaluator evaluates a request against an active policy and grant.
type Evaluator struct {
	store *PolicyStore
}

// NewEvaluator creates a deterministic policy evaluator.
func NewEvaluator(store ...*PolicyStore) *Evaluator {
	var s *PolicyStore
	if len(store) > 0 {
		s = store[0]
	}
	return &Evaluator{store: s}
}

// ActivePolicy returns the currently active policy for a policy ID.
// Execution-boundary freshness checks use this view rather than agent-held state.
func (e *Evaluator) ActivePolicy(policyID string) (*Policy, error) {
	if e == nil || e.store == nil {
		return nil, fmt.Errorf("policy evaluator unavailable")
	}
	return e.store.Active(policyID)
}

// ErrPolicyNotFound indicates the policy does not exist.
var ErrPolicyNotFound = fmt.Errorf("policy not found")

// ErrPolicyInactive indicates the policy exists but is not active.
var ErrPolicyInactive = fmt.Errorf("policy inactive")

// ErrPolicyVersionNotFound indicates the version is unavailable.
var ErrPolicyVersionNotFound = fmt.Errorf("policy version not found")

// Evaluate decides whether an action is allowed, denied, or requires approval.
func (e *Evaluator) Evaluate(ctx AuthorizationContext, action Action) DecisionResult {
	if e == nil || e.store == nil {
		return DecisionResult{
			Status:    DecisionIndeterminate,
			Reason:    ReasonEvaluationFailed,
			Requested: action.Capability,
			Message:   "policy evaluator unavailable",
		}
	}

	policyID, policyVersion := ctx.PolicyID, ctx.PolicyVersion
	if ctx.Policy != nil {
		if policyID == "" {
			policyID = ctx.Policy.ID
		}
		if policyVersion == 0 {
			policyVersion = ctx.Policy.Version
		}
		if ctx.Policy.ID != policyID || ctx.Policy.Version != policyVersion {
			return DecisionResult{
				Status: DecisionIndeterminate, Reason: ReasonEvaluationFailed,
				PolicyID: policyID, Requested: action.Capability,
				Message: "supplied policy identity does not match the requested policy reference",
			}
		}
	}

	canonical, err := e.store.Version(policyID, policyVersion)
	if err != nil {
		return DecisionResult{
			Status: DecisionIndeterminate, Reason: ReasonEvaluationFailed,
			PolicyID: policyID, Requested: action.Capability, Message: err.Error(),
		}
	}
	if canonical.Status != "active" {
		return DecisionResult{
			Status: DecisionDeny, Reason: ReasonDeniedStalePolicy,
			PolicyID: canonical.ID, PolicyVersion: canonical.Version,
			Requested: action.Capability,
			Message: fmt.Sprintf("policy %s v%d is not active", canonical.ID, canonical.Version),
		}
	}
	active, err := e.store.Active(policyID)
	if err != nil || active.Version != policyVersion {
		return DecisionResult{
			Status: DecisionDeny, Reason: ReasonDeniedStalePolicy,
			PolicyID: canonical.ID, PolicyVersion: canonical.Version,
			Requested: action.Capability,
			Message: fmt.Sprintf("policy %s v%d is not the active policy version", canonical.ID, canonical.Version),
		}
	}
	if ctx.Policy != nil && !reflect.DeepEqual(ctx.Policy, canonical) {
		return DecisionResult{
			Status: DecisionIndeterminate, Reason: ReasonEvaluationFailed,
			PolicyID: canonical.ID, PolicyVersion: canonical.Version,
			Requested: action.Capability,
			Message: "supplied policy contents do not match the canonical policy-store snapshot",
		}
	}
	ctx.Policy = canonical
	policy := ctx.Policy

	if ctx.Grant != nil && !ctx.Grant.IsValid(ctx.Now) {
		return DecisionResult{
			Status:        DecisionDeny,
			Reason:        ReasonDeniedByGrant,
			PolicyID:      policy.ID,
			PolicyVersion: policy.Version,
			Requested:     action.Capability,
			ApprovedGrant: ctx.Grant.GrantID,
			Message:       fmt.Sprintf("grant %s is invalid or expired", ctx.Grant.GrantID),
		}
	}

	if ctx.Grant != nil && ctx.Grant.PolicyID != policy.ID {
		return DecisionResult{
			Status:        DecisionDeny,
			Reason:        ReasonDeniedPolicyMismatch,
			PolicyID:      policy.ID,
			PolicyVersion: policy.Version,
			Requested:     action.Capability,
			ApprovedGrant: ctx.Grant.GrantID,
			Message:       fmt.Sprintf("grant %s belongs to policy %s, not %s", ctx.Grant.GrantID, ctx.Grant.PolicyID, policy.ID),
		}
	}

	if ctx.Grant != nil && ctx.Grant.PolicyVersion != policy.Version {
		return DecisionResult{
			Status:        DecisionDeny,
			Reason:        ReasonDeniedVersionMismatch,
			PolicyID:      policy.ID,
			PolicyVersion: policy.Version,
			Requested:     action.Capability,
			ApprovedGrant: ctx.Grant.GrantID,
			Message:       fmt.Sprintf("grant %s is bound to policy version %d, current policy is version %d", ctx.Grant.GrantID, ctx.Grant.PolicyVersion, policy.Version),
		}
	}

	if ctx.Grant != nil && ctx.Grant.AgentID != action.AgentID {
		return DecisionResult{
			Status:        DecisionDeny,
			Reason:        ReasonDeniedAgentMismatch,
			PolicyID:      policy.ID,
			PolicyVersion: policy.Version,
			Requested:     action.Capability,
			ApprovedGrant: ctx.Grant.GrantID,
			Message:       fmt.Sprintf("grant %s is scoped to agent %s, not %s", ctx.Grant.GrantID, ctx.Grant.AgentID, action.AgentID),
		}
	}

	if ctx.Grant != nil && !containsCapability(ctx.Grant.Capabilities, action.Capability) {
		return DecisionResult{
			Status:        DecisionDeny,
			Reason:        ReasonDeniedByScope,
			PolicyID:      policy.ID,
			PolicyVersion: policy.Version,
			Requested:     action.Capability,
			ApprovedGrant: ctx.Grant.GrantID,
			Message:       fmt.Sprintf("capability %s is not granted to agent %s", action.Capability, action.AgentID),
		}
	}

	for _, protected := range policy.ProtectedCapabilities {
		if protected == action.Capability {
			return DecisionResult{
				Status:        DecisionDeny,
				Reason:        ReasonDeniedProtected,
				PolicyID:      policy.ID,
				PolicyVersion: policy.Version,
				Requested:     action.Capability,
				Message:       fmt.Sprintf("capability %s is protected by policy %s v%d", action.Capability, policy.ID, policy.Version),
			}
		}
	}

	for _, denied := range policy.DeniedCapabilities {
		if denied == action.Capability {
			return DecisionResult{
				Status:        DecisionDeny,
				Reason:        ReasonDeniedByPolicy,
				PolicyID:      policy.ID,
				PolicyVersion: policy.Version,
				Requested:     action.Capability,
				Message:       fmt.Sprintf("capability %s is explicitly denied by policy %s v%d", action.Capability, policy.ID, policy.Version),
			}
		}
	}

	if containsCapability(policy.AllowedCapabilities, action.Capability) {
		if containsCapability(policy.ApprovalRequired, action.Capability) {
			return DecisionResult{
				Status:        DecisionPending,
				Reason:        ReasonRequiresApproval,
				PolicyID:      policy.ID,
				PolicyVersion: policy.Version,
				Requested:     action.Capability,
				NeedsApproval: true,
				Message:       fmt.Sprintf("capability %s requires approval under policy %s v%d", action.Capability, policy.ID, policy.Version),
			}
		}
		return DecisionResult{
			Status:        DecisionAllow,
			Reason:        ReasonAllowed,
			PolicyID:      policy.ID,
			PolicyVersion: policy.Version,
			Requested:     action.Capability,
			Message:       fmt.Sprintf("capability %s is allowed by policy %s v%d", action.Capability, policy.ID, policy.Version),
		}
	}

	return DecisionResult{
		Status:        DecisionDeny,
		Reason:        ReasonDeniedByPolicy,
		PolicyID:      policy.ID,
		PolicyVersion: policy.Version,
		Requested:     action.Capability,
		Message:       fmt.Sprintf("capability %s is not in the allow list of policy %s v%d", action.Capability, policy.ID, policy.Version),
	}
}

// normalizeRequest canonicalizes decision inputs to keep evaluation deterministic.
func normalizeRequest(action Action) Action {
	if action.Context == nil {
		action.Context = map[string]string{}
	}
	keys := make([]string, 0, len(action.Context))
	for k := range action.Context {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	clean := make(map[string]string, len(keys))
	for _, k := range keys {
		clean[k] = action.Context[k]
	}
	action.Context = clean
	return action
}

func containsCapability(values []Capability, target Capability) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

// EvaluateAction is a convenience wrapper that normalizes the request before evaluation.
func (e *Evaluator) EvaluateAction(action Action, grant *Grant, policy *Policy, now time.Time) DecisionResult {
	action = normalizeRequest(action)
	ctx := AuthorizationContext{
		Grant:  grant,
		Policy: policy,
		Now:    now,
	}
	if policy != nil {
		ctx.PolicyID = policy.ID
		ctx.PolicyVersion = policy.Version
	}
	if ctx.PolicyID == "" && policy != nil {
		ctx.PolicyID = policy.ID
	}
	if ctx.PolicyVersion == 0 && policy != nil {
		ctx.PolicyVersion = policy.Version
	}
	return e.Evaluate(ctx, action)
}

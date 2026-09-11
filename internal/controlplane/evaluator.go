package controlplane

import (
	"fmt"
	"sort"
	"time"
)

// PolicyStore stores policy definitions by ID and version.
type PolicyStore struct {
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

// Set stores a policy as the active policy and tracks versions.
func (ps *PolicyStore) Set(policy *Policy) error {
	if policy == nil {
		return fmt.Errorf("policy must not be nil")
	}
	if policy.ID == "" {
		return fmt.Errorf("policy ID is required")
	}
	if policy.Version <= 0 {
		return fmt.Errorf("policy version must be greater than zero")
	}
	ps.current[policy.ID] = policy
	if _, ok := ps.byVersion[policy.ID]; !ok {
		ps.byVersion[policy.ID] = make(map[int]*Policy)
	}
	ps.byVersion[policy.ID][policy.Version] = policy
	return nil
}

// Active returns the active policy for the given policy ID.
func (ps *PolicyStore) Active(policyID string) (*Policy, error) {
	policy, ok := ps.current[policyID]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPolicyNotFound, policyID)
	}
	if policy.Status != "active" {
		return nil, fmt.Errorf("%w: %s (status=%s)", ErrPolicyInactive, policyID, policy.Status)
	}
	return policy, nil
}

// Version returns a specific policy version.
func (ps *PolicyStore) Version(policyID string, version int) (*Policy, error) {
	byVersion, ok := ps.byVersion[policyID]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPolicyNotFound, policyID)
	}
	policy, ok := byVersion[version]
	if !ok {
		return nil, fmt.Errorf("%w: %s v%d", ErrPolicyVersionNotFound, policyID, version)
	}
	return policy, nil
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
			Status:    DecisionDeny,
			Reason:    ReasonNoPolicyFound,
			Requested: action.Capability,
			Message:   "policy evaluator unavailable",
		}
	}

	if ctx.Policy == nil {
		policy, err := e.store.Version(ctx.PolicyID, ctx.PolicyVersion)
		if err != nil {
			return DecisionResult{
				Status:    DecisionDeny,
				Reason:    ReasonNoPolicyFound,
				PolicyID:  ctx.PolicyID,
				Requested: action.Capability,
				Message:   err.Error(),
			}
		}
		ctx.Policy = policy
	}

	policy := ctx.Policy
	if policy == nil {
		return DecisionResult{
			Status:    DecisionDeny,
			Reason:    ReasonNoPolicyFound,
			PolicyID:  ctx.PolicyID,
			Requested: action.Capability,
			Message:   "policy is nil",
		}
	}

	if policy.Status != "active" {
		return DecisionResult{
			Status:        DecisionDeny,
			Reason:        ReasonDeniedStalePolicy,
			PolicyID:      policy.ID,
			PolicyVersion: policy.Version,
			Requested:     action.Capability,
			Message:       fmt.Sprintf("policy %s v%d is not active", policy.ID, policy.Version),
		}
	}

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
		PolicyID:      policy.ID,
		PolicyVersion: policy.Version,
		Grant:         grant,
		Policy:        policy,
		Now:           now,
	}
	if ctx.PolicyID == "" && policy != nil {
		ctx.PolicyID = policy.ID
	}
	if ctx.PolicyVersion == 0 && policy != nil {
		ctx.PolicyVersion = policy.Version
	}
	return e.Evaluate(ctx, action)
}

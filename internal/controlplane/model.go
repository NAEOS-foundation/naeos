package controlplane

import (
	"fmt"
	"time"
)

// Capability is a single permission or action boundary that can be granted or denied.
type Capability string

// DecisionStatus is the deterministic outcome of policy evaluation.
type DecisionStatus string

const (
	DecisionAllow   DecisionStatus = "ALLOW"
	DecisionDeny    DecisionStatus = "DENY"
	DecisionPending DecisionStatus = "REQUIRE_APPROVAL"
)

// DecisionReason describes why a decision was made.
type DecisionReason string

const (
	ReasonAllowed               DecisionReason = "allowed_by_policy"
	ReasonDeniedByPolicy        DecisionReason = "denied_by_policy"
	ReasonDeniedProtected       DecisionReason = "protected_capability"
	ReasonDeniedByGrant         DecisionReason = "grant_not_valid"
	ReasonDeniedByScope         DecisionReason = "capability_not_in_grant"
	ReasonDeniedMalformedGrant  DecisionReason = "malformed_grant"
	ReasonDeniedPolicyMismatch  DecisionReason = "grant_policy_mismatch"
	ReasonDeniedVersionMismatch DecisionReason = "grant_version_mismatch"
	ReasonDeniedAgentMismatch   DecisionReason = "agent_scope_mismatch"
	ReasonDeniedStalePolicy     DecisionReason = "stale_policy"
	ReasonRequiresApproval      DecisionReason = "approval_required"
	ReasonNoPolicyFound         DecisionReason = "policy_not_found"
)

// Policy defines the active rules for a named policy version.
type Policy struct {
	ID                    string       `json:"id"`
	Version               int          `json:"version"`
	Status                string       `json:"status"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
	AllowedCapabilities   []Capability `json:"allowed_capabilities"`
	DeniedCapabilities    []Capability `json:"denied_capabilities"`
	ProtectedCapabilities []Capability `json:"protected_capabilities"`
	ApprovalRequired      []Capability `json:"approval_required"`
	RequiresExplicitAuth  bool         `json:"requires_explicit_auth"`
}

// Grant models the authority assigned to an agent for a policy version.
type Grant struct {
	GrantID       string       `json:"grant_id"`
	AgentID       string       `json:"agent_id"`
	PolicyID      string       `json:"policy_id"`
	PolicyVersion int          `json:"policy_version"`
	Capabilities  []Capability `json:"capabilities"`
	CreatedAt     time.Time    `json:"created_at"`
	ExpiresAt     time.Time    `json:"expires_at"`
	Status        string       `json:"status"`
	Revoked       bool         `json:"revoked"`
	RevokedAt     *time.Time   `json:"revoked_at,omitempty"`
	RevokeReason  string       `json:"revoke_reason,omitempty"`
}

// IsValid returns whether the grant is still valid and current.
func (g Grant) IsValid(now time.Time) bool {
	if g.Revoked {
		return false
	}
	if g.Status != "active" {
		return false
	}
	if now.After(g.ExpiresAt) {
		return false
	}
	if g.PolicyID == "" || g.GrantID == "" {
		return false
	}
	return true
}

// Action is the request being evaluated by the control plane.
type Action struct {
	AgentID      string            `json:"agent_id"`
	Capability   Capability        `json:"capability"`
	Payload      interface{}       `json:"payload,omitempty"`
	ArtifactHash string            `json:"artifact_hash,omitempty"`
	Context      map[string]string `json:"context,omitempty"`
}

// AuthorizationContext carries the active policy and grant state used to decide a request.
type AuthorizationContext struct {
	PolicyID      string    `json:"policy_id"`
	PolicyVersion int       `json:"policy_version"`
	Grant         *Grant    `json:"grant,omitempty"`
	Policy        *Policy   `json:"policy,omitempty"`
	Now           time.Time `json:"now"`
}

// DecisionResult is the deterministic result of evaluating a request.
type DecisionResult struct {
	DecisionID    string         `json:"decision_id,omitempty"`
	RequestID     string         `json:"request_id,omitempty"`
	AgentID       string         `json:"agent_id,omitempty"`
	Status        DecisionStatus `json:"status"`
	Reason        DecisionReason `json:"reason"`
	PolicyID      string         `json:"policy_id,omitempty"`
	PolicyVersion int            `json:"policy_version,omitempty"`
	Requested     Capability     `json:"requested,omitempty"`
	ArtifactHash  string         `json:"artifact_hash,omitempty"`
	ApprovedGrant string         `json:"approved_grant,omitempty"`
	NeedsApproval bool           `json:"needs_approval,omitempty"`
	Message       string         `json:"message,omitempty"`
}

func (d DecisionResult) String() string {
	if d.Message != "" {
		return fmt.Sprintf("%s: %s", d.Status, d.Message)
	}
	return string(d.Status)
}

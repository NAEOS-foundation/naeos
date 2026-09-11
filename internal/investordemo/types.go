// Package investordemo contains the NAEOS investor demo implementation.
// This demonstrates NAEOS as a control plane for AI agent authorization, policy enforcement,
// and independent verification.
package investordemo

import (
	"time"
)

// ============================================================================
// Core Types
// ============================================================================

// Capability represents a single permission that an agent may request or be granted.
type Capability string

// ReplayProtection prevents the same handoff from being executed multiple times.
type ReplayProtection struct {
	Nonce     string    `json:"nonce"`
	Timestamp time.Time `json:"timestamp"`
}

// ============================================================================
// Policy
// ============================================================================

// Policy defines what capabilities are allowed or denied for agents.
type Policy struct {
	PolicyID              string       `json:"policy_id"`
	Version               int          `json:"version"`
	Status                string       `json:"status"` // active, inactive, superseded
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
	AllowedCapabilities   []Capability `json:"allowed_capabilities"`
	DeniedCapabilities    []Capability `json:"denied_capabilities"`
	ProtectedCapabilities []Capability `json:"protected_capabilities"` // e.g., iam.modify, policy.modify
	ApprovalRequired      []Capability `json:"approval_required"`
	RequiresExplicitAuth  bool         `json:"requires_explicit_authorization"`
}

// ============================================================================
// Capability Grant
// ============================================================================

// CapabilityGrant gives an agent explicit permission to use specific capabilities.
type CapabilityGrant struct {
	GrantID       string       `json:"grant_id"`
	AgentID       string       `json:"agent_id"`
	PolicyID      string       `json:"policy_id"`
	PolicyVersion int          `json:"policy_version"`
	Capabilities  []Capability `json:"capabilities"`
	CreatedAt     time.Time    `json:"created_at"`
	ExpiresAt     time.Time    `json:"expires_at"`
	Status        string       `json:"status"` // active, expired, revoked
	Revoked       bool         `json:"revoked"`
	RevokedAt     *time.Time   `json:"revoked_at,omitempty"`
	RevokeReason  string       `json:"revoke_reason,omitempty"`
}

// IsValid checks if the grant is still valid.
func (g *CapabilityGrant) IsValid() bool {
	if g.Revoked {
		return false
	}
	if g.Status != "active" {
		return false
	}
	if time.Now().After(g.ExpiresAt) {
		return false
	}
	return true
}

// ============================================================================
// Handoff Contract
// ============================================================================

// HandoffContract is the protocol for passing authorization from one component to another.
// It includes cryptographic protections against tampering and replay attacks.
type HandoffContract struct {
	ContractVersion        string                 `json:"contract_version"`
	CanonicalVersion       string                 `json:"canonicalization_version"`
	Initiator              string                 `json:"initiator"` // agent-id or component-id
	RequestedCapability    Capability             `json:"requested_capability"`
	AuthorizedCapabilities []Capability           `json:"authorized_capabilities"`
	PayloadDigest          string                 `json:"payload_digest"`
	Payload                map[string]interface{} `json:"payload,omitempty"`
	PolicyID               string                 `json:"policy_id"`
	PolicyVersion          int                    `json:"policy_version"`
	Provenance             map[string]interface{} `json:"provenance"`
	CreatedAt              time.Time              `json:"created_at"`
	ExpiresAt              time.Time              `json:"expires_at"`
	ReplayProtection       ReplayProtection       `json:"replay_protection"`
	Signature              string                 `json:"signature,omitempty"`
	DownstreamHandoff      *HandoffContract       `json:"downstream_handoff,omitempty"`
}

// ============================================================================
// Authorization Decision
// ============================================================================

// AuthorizationDecision represents the result of authorization checking.
type AuthorizationDecision struct {
	DecisionID          string     `json:"decision_id"`
	Timestamp           time.Time  `json:"timestamp"`
	AgentID             string     `json:"agent_id"`
	RequestedCapability Capability `json:"requested_capability"`
	PolicyID            string     `json:"policy_id"`
	PolicyVersion       int        `json:"policy_version"`
	Authorized          bool       `json:"authorized"`
	Reason              string     `json:"reason"`
	GrantID             string     `json:"grant_id,omitempty"`
	BlockReason         string     `json:"block_reason,omitempty"`
}

// ============================================================================
// Handoff Validation Result
// ============================================================================

// HandoffValidationResult represents the result of validating a handoff contract.
type HandoffValidationResult struct {
	ValidationID               string    `json:"validation_id"`
	Timestamp                  time.Time `json:"timestamp"`
	ContractVersion            string    `json:"contract_version"`
	Valid                      bool      `json:"valid"`
	Errors                     []string  `json:"errors"`
	Warnings                   []string  `json:"warnings"`
	CapabilityWideningDetected bool      `json:"capability_widening_detected"`
	ReplayDetected             bool      `json:"replay_detected"`
	ExpiredDetected            bool      `json:"expired_detected"`
	PayloadTampered            bool      `json:"payload_tampered"`
	ProvenanceMismatch         bool      `json:"provenance_mismatch"`
}

// ============================================================================
// Verification Result
// ============================================================================

// VerificationResult represents the independent verification of an action.
type VerificationResult struct {
	VerificationID      string    `json:"verification_id"`
	Timestamp           time.Time `json:"timestamp"`
	AgentID             string    `json:"agent_id"`
	Result              string    `json:"result"` // PASS, FAIL, INCONCLUSIVE
	PolicyCompliance    bool      `json:"policy_compliance"`
	TestsPassed         bool      `json:"tests_passed"`
	UnauthorizedActions int       `json:"unauthorized_actions"`
	Issues              []string  `json:"issues"`
	Warnings            []string  `json:"warnings"`
	Reason              string    `json:"reason"`
	Summary             string    `json:"summary"`
}

// ============================================================================
// Audit Event
// ============================================================================

// AuditEvent records all important security events in the system.
type AuditEvent struct {
	EventID             string                 `json:"event_id"`
	Timestamp           time.Time              `json:"timestamp"`
	EventType           string                 `json:"event_type"` // AUTHORIZATION_REQUESTED, AUTHORIZATION_GRANTED, AUTHORIZATION_DENIED, etc.
	AgentID             string                 `json:"agent_id"`
	RequestedCapability Capability             `json:"requested_capability,omitempty"`
	PolicyID            string                 `json:"policy_id,omitempty"`
	PolicyVersion       int                    `json:"policy_version,omitempty"`
	Decision            string                 `json:"decision,omitempty"` // ALLOW, BLOCK
	Reason              string                 `json:"reason,omitempty"`
	Details             map[string]interface{} `json:"details,omitempty"`
}

// ============================================================================
// Execution Request
// ============================================================================

// ExecutionRequest represents a request to execute an action.
type ExecutionRequest struct {
	RequestID       string                 `json:"request_id"`
	Timestamp       time.Time              `json:"timestamp"`
	AgentID         string                 `json:"agent_id"`
	Capability      Capability             `json:"capability"`
	Payload         map[string]interface{} `json:"payload"`
	HandoffContract *HandoffContract       `json:"handoff_contract,omitempty"`
}

// ============================================================================
// Execution Result
// ============================================================================

// ExecutionResult represents the result of executing a capability.
type ExecutionResult struct {
	ExecutionID        string                 `json:"execution_id"`
	RequestID          string                 `json:"request_id"`
	DecisionID         string                 `json:"decision_id,omitempty"`
	EvidenceID         string                 `json:"evidence_id,omitempty"`
	Timestamp          time.Time              `json:"timestamp"`
	AgentID            string                 `json:"agent_id"`
	Capability         Capability             `json:"capability"`
	Authorized         bool                   `json:"authorized"`
	Executed           bool                   `json:"executed"`
	Result             map[string]interface{} `json:"result,omitempty"`
	Error              string                 `json:"error,omitempty"`
	VerificationStatus string                 `json:"verification_status,omitempty"`
}

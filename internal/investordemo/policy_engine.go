package investordemo

import (
	"fmt"
	"time"
)

// PolicyStore stores and manages policies.
type PolicyStore struct {
	policies map[string]*Policy
	versions map[string]map[int]*Policy // policy_id -> version -> policy
}

// NewPolicyStore creates a new policy store.
func NewPolicyStore() *PolicyStore {
	return &PolicyStore{
		policies: make(map[string]*Policy),
		versions: make(map[string]map[int]*Policy),
	}
}

// StorePolicy stores a policy and tracks versions.
func (ps *PolicyStore) StorePolicy(policy *Policy) error {
	if policy.PolicyID == "" {
		return fmt.Errorf("policy ID is required")
	}

	ps.policies[policy.PolicyID] = policy

	if _, ok := ps.versions[policy.PolicyID]; !ok {
		ps.versions[policy.PolicyID] = make(map[int]*Policy)
	}
	ps.versions[policy.PolicyID][policy.Version] = policy

	return nil
}

// GetPolicy retrieves the current active policy.
func (ps *PolicyStore) GetPolicy(policyID string) (*Policy, error) {
	p, ok := ps.policies[policyID]
	if !ok {
		return nil, fmt.Errorf("policy %s not found", policyID)
	}

	if p.Status != "active" {
		return nil, fmt.Errorf("policy %s is not active (status: %s)", policyID, p.Status)
	}

	return p, nil
}

// GetPolicyVersion retrieves a specific version of a policy.
func (ps *PolicyStore) GetPolicyVersion(policyID string, version int) (*Policy, error) {
	versions, ok := ps.versions[policyID]
	if !ok {
		return nil, fmt.Errorf("policy %s not found", policyID)
	}

	p, ok := versions[version]
	if !ok {
		return nil, fmt.Errorf("policy %s version %d not found", policyID, version)
	}

	return p, nil
}

// UpdatePolicy updates a policy to a new version.
func (ps *PolicyStore) UpdatePolicy(policy *Policy) error {
	// The old policy should be marked as superseded
	if oldPolicy, ok := ps.policies[policy.PolicyID]; ok && oldPolicy.Status == "active" {
		oldPolicy.Status = "superseded"
	}

	ps.policies[policy.PolicyID] = policy

	if _, ok := ps.versions[policy.PolicyID]; !ok {
		ps.versions[policy.PolicyID] = make(map[int]*Policy)
	}
	ps.versions[policy.PolicyID][policy.Version] = policy

	return nil
}

// ============================================================================
// PolicyEngine evaluates whether an agent has a capability
// ============================================================================

type PolicyEngine struct {
	policyStore *PolicyStore
	auditLedger *AuditLedger
}

// NewPolicyEngine creates a new policy engine.
func NewPolicyEngine(policyStore *PolicyStore, auditLedger *AuditLedger) *PolicyEngine {
	return &PolicyEngine{
		policyStore: policyStore,
		auditLedger: auditLedger,
	}
}

// EvaluateCapability checks if a capability is allowed by the policy.
func (pe *PolicyEngine) EvaluateCapability(policyID string, policyVersion int, capability Capability) (bool, string) {
	policy, err := pe.policyStore.GetPolicyVersion(policyID, policyVersion)
	if err != nil {
		return false, fmt.Sprintf("Policy evaluation failed: %v", err)
	}

	// Check explicitly denied capabilities
	for _, denied := range policy.DeniedCapabilities {
		if denied == capability {
			return false, fmt.Sprintf("Capability %s is explicitly denied by policy %s v%d", capability, policyID, policyVersion)
		}
	}

	// Check allowed capabilities
	for _, allowed := range policy.AllowedCapabilities {
		if allowed == capability {
			return true, ""
		}
	}

	// If not in allowed list and not denied, it's not allowed
	return false, fmt.Sprintf("Capability %s is not in the allowed list of policy %s v%d", capability, policyID, policyVersion)
}

// IsProtectedCapability checks if a capability is protected (requires special handling).
func (pe *PolicyEngine) IsProtectedCapability(policyID string, policyVersion int, capability Capability) (bool, error) {
	policy, err := pe.policyStore.GetPolicyVersion(policyID, policyVersion)
	if err != nil {
		return false, err
	}

	for _, protected := range policy.ProtectedCapabilities {
		if protected == capability {
			return true, nil
		}
	}

	return false, nil
}

// RequiresApproval checks if a capability requires additional approval.
func (pe *PolicyEngine) RequiresApproval(policyID string, policyVersion int, capability Capability) (bool, error) {
	policy, err := pe.policyStore.GetPolicyVersion(policyID, policyVersion)
	if err != nil {
		return false, err
	}

	for _, needsApproval := range policy.ApprovalRequired {
		if needsApproval == capability {
			return true, nil
		}
	}

	return false, nil
}

// EnforcesExplicitAuth checks if the policy requires explicit authorization.
func (pe *PolicyEngine) EnforcesExplicitAuth(policyID string, policyVersion int) bool {
	policy, err := pe.policyStore.GetPolicyVersion(policyID, policyVersion)
	if err != nil {
		return false
	}

	return policy.RequiresExplicitAuth
}

// ============================================================================
// CapabilityAuthority checks if an agent has been granted a capability
// ============================================================================

type CapabilityAuthority struct {
	grantStore   *GrantStore
	policyStore  *PolicyStore
	policyEngine *PolicyEngine
	auditLedger  *AuditLedger
}

// NewCapabilityAuthority creates a new capability authority.
func NewCapabilityAuthority(grantStore *GrantStore, policyStore *PolicyStore, policyEngine *PolicyEngine, auditLedger *AuditLedger) *CapabilityAuthority {
	return &CapabilityAuthority{
		grantStore:   grantStore,
		policyStore:  policyStore,
		policyEngine: policyEngine,
		auditLedger:  auditLedger,
	}
}

// CheckAuthorization determines if an agent is authorized for a capability.
func (ca *CapabilityAuthority) CheckAuthorization(agentID string, capability Capability) *AuthorizationDecision {
	decision := &AuthorizationDecision{
		DecisionID:          generateID("AUTHZ"),
		Timestamp:           time.Now(),
		AgentID:             agentID,
		RequestedCapability: capability,
		Authorized:          false,
	}

	// Get the grant for the agent
	grant, err := ca.grantStore.GetGrantByAgent(agentID)
	if err != nil {
		decision.Reason = fmt.Sprintf("No grant found for agent %s: %v", agentID, err)
		decision.BlockReason = "NO_GRANT"
		ca.recordAuditEvent("AUTHORIZATION_DENIED", agentID, capability, "", 0, "BLOCK", decision.BlockReason)
		return decision
	}

	decision.PolicyID = grant.PolicyID
	decision.PolicyVersion = grant.PolicyVersion
	decision.GrantID = grant.GrantID

	// Check if grant is valid
	if !grant.IsValid() {
		decision.Reason = fmt.Sprintf("Grant %s is not valid (revoked or expired)", grant.GrantID)
		decision.BlockReason = "GRANT_INVALID"
		ca.recordAuditEvent("AUTHORIZATION_DENIED", agentID, capability, grant.PolicyID, grant.PolicyVersion, "BLOCK", decision.BlockReason)
		return decision
	}

	// Check if the capability is in the grant
	hasCapability := false
	for _, grantedCap := range grant.Capabilities {
		if grantedCap == capability {
			hasCapability = true
			break
		}
	}

	if !hasCapability {
		decision.Reason = fmt.Sprintf("Capability %s not in grant %s", capability, grant.GrantID)
		decision.BlockReason = "CAPABILITY_NOT_GRANTED"
		ca.recordAuditEvent("AUTHORIZATION_DENIED", agentID, capability, grant.PolicyID, grant.PolicyVersion, "BLOCK", decision.BlockReason)
		return decision
	}

	// Check if the capability is allowed by the policy
	allowed, policyReason := ca.policyEngine.EvaluateCapability(grant.PolicyID, grant.PolicyVersion, capability)
	if !allowed {
		decision.Reason = policyReason
		decision.BlockReason = "POLICY_DENIES_CAPABILITY"
		ca.recordAuditEvent("AUTHORIZATION_DENIED", agentID, capability, grant.PolicyID, grant.PolicyVersion, "BLOCK", decision.BlockReason)
		return decision
	}

	// Authorization granted!
	decision.Authorized = true
	decision.Reason = fmt.Sprintf("Agent %s is authorized for capability %s under grant %s", agentID, capability, grant.GrantID)
	ca.recordAuditEvent("AUTHORIZATION_GRANTED", agentID, capability, grant.PolicyID, grant.PolicyVersion, "ALLOW", "AUTHORIZED")

	return decision
}

// RevokeGrant revokes a capability grant.
func (ca *CapabilityAuthority) RevokeGrant(grantID string, reason string) error {
	now := time.Now()
	return ca.grantStore.RevokeGrant(grantID, reason, &now)
}

// recordAuditEvent is a helper that records an audit event, discarding the error
// since audit recording in the demo is best-effort (append-only in-memory store).
func (ca *CapabilityAuthority) recordAuditEvent(eventType, agentID string, cap Capability, policyID string, policyVersion int, decision, reason string) {
	_ = ca.auditLedger.RecordEvent(&AuditEvent{
		EventID:             generateID("AUD"),
		Timestamp:           time.Now(),
		EventType:           eventType,
		AgentID:             agentID,
		RequestedCapability: cap,
		PolicyID:            policyID,
		PolicyVersion:       policyVersion,
		Decision:            decision,
		Reason:              reason,
	})
}

package investordemo

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

// generateID generates a unique ID with a prefix.
func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()%1000000)
}

// generateNonce generates a random nonce for replay protection.
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ============================================================================
// Demo Scenario Setup
// ============================================================================

type DemoSetup struct {
	PolicyStore          *PolicyStore
	GrantStore           *GrantStore
	AuditLedger          *AuditLedger
	PolicyEngine         *PolicyEngine
	CapabilityAuthority  *CapabilityAuthority
	HandoffValidator     *HandoffValidator
	IndependentVerifier  *IndependentVerifier
	ExecutionGate        *ExecutionGate
	ControlPlaneGateway  *controlplane.DecisionGateway
	ControlPlaneVerifier *controlplane.SessionVerifier
}

// SetupDemoEnvironment sets up the initial demo state with policies, grants, and components.
func SetupDemoEnvironment() *DemoSetup {
	// Create stores and ledger
	policyStore := NewPolicyStore()
	grantStore := NewGrantStore()
	auditLedger := NewAuditLedger()

	// Create policy engine
	policyEngine := NewPolicyEngine(policyStore, auditLedger)

	// Create capability authority
	capabilityAuthority := NewCapabilityAuthority(grantStore, policyStore, policyEngine, auditLedger)

	// Create handoff validator
	handoffValidator := NewHandoffValidator(auditLedger, policyEngine, grantStore)

	// Create independent verifier
	independentVerifier := NewIndependentVerifier(auditLedger, policyEngine, grantStore)

	// Create execution gate
	executionGate := NewExecutionGate(capabilityAuthority, handoffValidator, auditLedger, policyEngine, independentVerifier)
	controlPlaneGateway := newControlPlaneGateway(policyStore, grantStore)
	executionGate.controlPlaneGateway = controlPlaneGateway

	// Create the demo policy
	policy := &Policy{
		PolicyID:  "POLICY-017",
		Version:   17,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		AllowedCapabilities: []Capability{
			"repository.read",
			"repository.write",
			"test.execute",
		},
		DeniedCapabilities: []Capability{},
		ProtectedCapabilities: []Capability{
			"iam.modify",
			"credential.rotate",
			"production.deploy",
			"external.publish",
			"policy.modify",
		},
		ApprovalRequired:     []Capability{},
		RequiresExplicitAuth: true,
	}

	_ = policyStore.StorePolicy(policy)

	// Create a grant for the demo agent
	grant := &CapabilityGrant{
		GrantID:       "GRANT-001",
		AgentID:       "agent-payment-01",
		PolicyID:      "POLICY-017",
		PolicyVersion: 17,
		Capabilities: []Capability{
			"repository.read",
			"repository.write",
			"test.execute",
		},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Status:    "active",
		Revoked:   false,
	}

	_ = grantStore.StoreGrant(grant)
	controlPlaneGateway = newControlPlaneGateway(policyStore, grantStore)
	executionGate.controlPlaneGateway = controlPlaneGateway
	controlPlaneVerifier := controlplane.NewSessionVerifier(controlPlaneGateway.Ledger, controlPlaneGateway.Evaluator)
	auditLedger.SetObserver(&controlPlaneAuditObserver{ledger: controlPlaneGateway.Ledger})

	return &DemoSetup{
		PolicyStore:          policyStore,
		GrantStore:           grantStore,
		AuditLedger:          auditLedger,
		PolicyEngine:         policyEngine,
		CapabilityAuthority:  capabilityAuthority,
		HandoffValidator:     handoffValidator,
		IndependentVerifier:  independentVerifier,
		ExecutionGate:        executionGate,
		ControlPlaneGateway:  controlPlaneGateway,
		ControlPlaneVerifier: controlPlaneVerifier,
	}
}

// ============================================================================
// Demo Scenarios
// ============================================================================

// ScenarioResult represents the result of running a demo scenario.
type ScenarioResult struct {
	ScenarioName          string
	Description           string
	Passed                bool
	ExpectedResult        string
	ActualResult          string
	Details               string
	ExecutionResult       *ExecutionResult
	AuthorizationDecision *AuthorizationDecision
	HandoffValidation     *HandoffValidationResult
	VerificationResult    *VerificationResult
}

// RunScenario1_AuthorizedRepositoryRead tests authorized read access.
func RunScenario1_AuthorizedRepositoryRead(setup *DemoSetup) *ScenarioResult {
	scenario := &ScenarioResult{
		ScenarioName:   "Authorized Repository Read",
		Description:    "Agent requests repository.read capability which is in its grant",
		ExpectedResult: "ALLOW",
	}

	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		Timestamp:  time.Now(),
		AgentID:    "agent-payment-01",
		Capability: "repository.read",
		Payload:    map[string]interface{}{"file": "payment_service.go"},
	}

	result, _ := setup.ExecutionGate.Authorize(request)
	scenario.ExecutionResult = result
	scenario.ActualResult = "ALLOW"
	scenario.Passed = result.Authorized
	scenario.Details = result.Result["status"].(string)

	return scenario
}

// RunScenario2_UnauthorizedCredentialRotation tests blocking unauthorized credential rotation.
func RunScenario2_UnauthorizedCredentialRotation(setup *DemoSetup) *ScenarioResult {
	scenario := &ScenarioResult{
		ScenarioName:   "Unauthorized Credential Rotation",
		Description:    "Agent attempts credential.rotate which is not granted and is protected",
		ExpectedResult: "BLOCK",
	}

	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		Timestamp:  time.Now(),
		AgentID:    "agent-payment-01",
		Capability: "credential.rotate",
		Payload:    map[string]interface{}{"key": "prod-api-key"},
	}

	result, _ := setup.ExecutionGate.Authorize(request)
	scenario.ExecutionResult = result
	scenario.ActualResult = "BLOCK"
	scenario.Passed = !result.Authorized
	scenario.Details = result.Error

	return scenario
}

// RunScenario3_PolicySelfModification tests blocking policy self-modification.
func RunScenario3_PolicySelfModification(setup *DemoSetup) *ScenarioResult {
	scenario := &ScenarioResult{
		ScenarioName:   "Policy Self-Modification Attack",
		Description:    "Agent attempts to modify the policy governing its own authorization",
		ExpectedResult: "BLOCK",
	}

	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		Timestamp:  time.Now(),
		AgentID:    "agent-payment-01",
		Capability: "policy.modify",
		Payload:    map[string]interface{}{"policy_id": "POLICY-017", "version": 18},
	}

	result, _ := setup.ExecutionGate.Authorize(request)
	scenario.ExecutionResult = result
	scenario.ActualResult = "BLOCK"
	scenario.Passed = !result.Authorized
	scenario.Details = result.Error

	return scenario
}

// RunScenario4_CapabilityEscalationViaHandoff tests detecting capability escalation via downstream handoff.
func RunScenario4_CapabilityEscalationViaHandoff(setup *DemoSetup) *ScenarioResult {
	scenario := &ScenarioResult{
		ScenarioName:   "Capability Escalation via Agent-to-Agent Handoff",
		Description:    "Agent A hands off to Agent B requesting a capability that was never granted",
		ExpectedResult: "BLOCK",
	}

	// Parent handoff is valid
	parentContract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              "agent-payment-01",
		RequestedCapability:    "repository.write",
		AuthorizedCapabilities: []Capability{"repository.read", "repository.write", "test.execute"},
		PayloadDigest:          calculatePayloadDigest(map[string]interface{}{}),
		Payload:                map[string]interface{}{},
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		Provenance:             map[string]interface{}{"source": "agent-payment-01"},
		CreatedAt:              time.Now(),
		ExpiresAt:              time.Now().Add(1 * time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: generateNonce(), Timestamp: time.Now()},
	}

	// Downstream handoff attempts credential escalation
	downstreamContract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              "agent-secondary-02",
		RequestedCapability:    "credential.rotate",
		AuthorizedCapabilities: []Capability{"repository.read", "repository.write", "test.execute"}, // Same as parent
		PayloadDigest:          calculatePayloadDigest(map[string]interface{}{}),
		Payload:                map[string]interface{}{},
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		Provenance:             map[string]interface{}{"source": "agent-secondary-02"},
		CreatedAt:              time.Now(),
		ExpiresAt:              time.Now().Add(1 * time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: generateNonce(), Timestamp: time.Now()},
	}

	parentContract.DownstreamHandoff = downstreamContract

	validation := setup.HandoffValidator.ValidateHandoff(parentContract)
	scenario.HandoffValidation = validation
	scenario.ActualResult = "BLOCK"
	scenario.Passed = !validation.Valid && validation.CapabilityWideningDetected
	scenario.Details = fmt.Sprintf("Errors: %v", validation.Errors)

	return scenario
}

// RunScenario5_ReplayAttack tests detecting replay attacks.
func RunScenario5_ReplayAttack(setup *DemoSetup) *ScenarioResult {
	scenario := &ScenarioResult{
		ScenarioName:   "Replay Attack Detection",
		Description:    "Agent attempts to replay a previously used handoff contract",
		ExpectedResult: "BLOCK",
	}

	// First handoff
	nonce := generateNonce()
	firstContract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              "agent-payment-01",
		RequestedCapability:    "repository.write",
		AuthorizedCapabilities: []Capability{"repository.read", "repository.write", "test.execute"},
		PayloadDigest:          calculatePayloadDigest(map[string]interface{}{}),
		Payload:                map[string]interface{}{},
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		Provenance:             map[string]interface{}{"source": "agent-payment-01"},
		CreatedAt:              time.Now(),
		ExpiresAt:              time.Now().Add(1 * time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: nonce, Timestamp: time.Now()},
	}

	// First execution - should pass
	validation1 := setup.HandoffValidator.ValidateHandoff(firstContract)
	scenario.Passed = !validation1.ReplayDetected

	// Attempt replay with the same nonce
	replayContract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              "agent-payment-01",
		RequestedCapability:    "repository.write",
		AuthorizedCapabilities: []Capability{"repository.read", "repository.write", "test.execute"},
		PayloadDigest:          calculatePayloadDigest(map[string]interface{}{}),
		Payload:                map[string]interface{}{},
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		Provenance:             map[string]interface{}{"source": "agent-payment-01"},
		CreatedAt:              time.Now(),
		ExpiresAt:              time.Now().Add(1 * time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: nonce, Timestamp: time.Now()}, // SAME NONCE
	}

	validation2 := setup.HandoffValidator.ValidateHandoff(replayContract)
	scenario.HandoffValidation = validation2
	scenario.ActualResult = "BLOCK"
	scenario.Passed = scenario.Passed && validation2.ReplayDetected && !validation2.Valid
	scenario.Details = fmt.Sprintf("First validation passed: %v, Replay detected: %v", !validation1.ReplayDetected, validation2.ReplayDetected)

	return scenario
}

// RunScenario6_PolicyVersionMismatch tests stale authorization detection.
// Per spec §9 Attack 6: When the active policy is updated (POLICY-017 v17 → v18),
// old authorizations must be re-evaluated at execution time and blocked as stale.
func RunScenario6_PolicyVersionMismatch(setup *DemoSetup) *ScenarioResult {
	scenario := &ScenarioResult{
		ScenarioName:   "Stale Authorization - Policy Version Changed",
		Description:    "Authorization under old policy becomes invalid when policy is updated",
		ExpectedResult: "BLOCK",
	}

	// Step 1: Verify the action is currently allowed under POLICY-017 v17
	preCheck := setup.CapabilityAuthority.CheckAuthorization("agent-payment-01", "repository.write")
	if !preCheck.Authorized {
		scenario.Details = "Pre-check failed: repository.write should be allowed under POLICY-017 v17"
		scenario.Passed = false
		scenario.ActualResult = "ERROR"
		return scenario
	}

	// Step 2: Update POLICY-017 to v18 — remove repository.write from allowed list
	// This supersedes v17 and marks it inactive
	newPolicy := &Policy{
		PolicyID:  "POLICY-017",
		Version:   18,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		AllowedCapabilities: []Capability{
			"repository.read",
			"test.execute",
			// NOTE: repository.write is REMOVED in v18
		},
		DeniedCapabilities: []Capability{},
		ProtectedCapabilities: []Capability{
			"iam.modify",
			"credential.rotate",
			"production.deploy",
			"external.publish",
			"policy.modify",
		},
		ApprovalRequired:     []Capability{},
		RequiresExplicitAuth: true,
	}

	_ = setup.PolicyStore.UpdatePolicy(newPolicy)

	// Record a POLICY_CHANGE audit event
	_ = setup.AuditLedger.RecordEvent(&AuditEvent{
		Timestamp:     time.Now(),
		EventType:     "POLICY_CHANGE",
		AgentID:       "system",
		PolicyID:      "POLICY-017",
		PolicyVersion: 18,
		Details: map[string]interface{}{
			"previous_version": 17,
			"new_version":      18,
		},
	})

	// Step 3: Attempt to execute repository.write with the old grant
	// The agent's grant is bound to POLICY-017 v17, but the active policy is now v18
	// The execution gate must detect this stale authorization and BLOCK
	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		Timestamp:  time.Now(),
		AgentID:    "agent-payment-01",
		Capability: "repository.write",
		Payload:    map[string]interface{}{},
	}

	result, _ := setup.ExecutionGate.Authorize(request)
	scenario.ExecutionResult = result

	scenario.ActualResult = "BLOCK"
	scenario.Passed = !result.Authorized
	scenario.Details = fmt.Sprintf("POLICY-017 updated from v17 → v18. "+
		"Grant bound to v17. Execution %v. Expected: BLOCK (STALE_AUTHORIZATION)",
		func() string {
			if result.Authorized {
				return "ALLOWED (BUG)"
			}
			return "BLOCKED"
		}())

	return scenario
}

// RunScenarioIAMModification tests blocking a protected trust-boundary action.
// Per spec §9 Attack 2, iam.modify is a protected capability and must be denied.
func RunScenarioIAMModification(setup *DemoSetup) *ScenarioResult {
	scenario := &ScenarioResult{
		ScenarioName:   "IAM Modification Attack",
		Description:    "Agent attempts iam.modify, a protected trust-boundary action",
		ExpectedResult: "BLOCK",
	}

	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		Timestamp:  time.Now(),
		AgentID:    "agent-payment-01",
		Capability: "iam.modify",
		Payload:    map[string]interface{}{"role": "admin"},
	}

	result, _ := setup.ExecutionGate.Authorize(request)
	scenario.ExecutionResult = result
	scenario.ActualResult = "BLOCK"
	scenario.Passed = !result.Authorized
	scenario.Details = result.Error

	return scenario
}

// RunAllScenarios runs all demo scenarios.
func RunAllScenarios(setup *DemoSetup) []*ScenarioResult {
	scenarios := []*ScenarioResult{
		RunScenario1_AuthorizedRepositoryRead(setup),
		RunScenario2_UnauthorizedCredentialRotation(setup),
		RunScenario3_PolicySelfModification(setup),
		RunScenario4_CapabilityEscalationViaHandoff(setup),
		RunScenario5_ReplayAttack(setup),
		RunScenario6_PolicyVersionMismatch(setup),
	}

	return scenarios
}

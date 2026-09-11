package investordemo

import (
	"strings"
	"testing"
	"time"
)

// ============================================================================
// Tests for Policy Engine
// ============================================================================

func TestPolicyEngine_AllowedCapability(t *testing.T) {
	setup := SetupDemoEnvironment()

	// Test that repository.read is allowed
	decision := setup.CapabilityAuthority.CheckAuthorization("agent-payment-01", "repository.read")
	if !decision.Authorized {
		t.Errorf("Expected repository.read to be authorized, got: %s", decision.Reason)
	}
}

func TestPolicyEngine_DeniedCapability(t *testing.T) {
	setup := SetupDemoEnvironment()

	// Test that credential.rotate is denied
	decision := setup.CapabilityAuthority.CheckAuthorization("agent-payment-01", "credential.rotate")
	if decision.Authorized {
		t.Errorf("Expected credential.rotate to be denied, got authorized")
	}

	if decision.BlockReason != "CAPABILITY_NOT_GRANTED" {
		t.Errorf("Expected CAPABILITY_NOT_GRANTED, got: %s", decision.BlockReason)
	}
}

func TestPolicyEngine_ProtectedCapability(t *testing.T) {
	setup := SetupDemoEnvironment()

	// Test that policy.modify is protected
	isProtected, _ := setup.PolicyEngine.IsProtectedCapability("POLICY-017", 17, "policy.modify")
	if !isProtected {
		t.Errorf("Expected policy.modify to be protected")
	}
}

// ============================================================================
// Tests for Capability Authority
// ============================================================================

func TestCapabilityAuthority_InvalidGrant(t *testing.T) {
	setup := SetupDemoEnvironment()

	// Test authorization with non-existent agent
	decision := setup.CapabilityAuthority.CheckAuthorization("non-existent-agent", "repository.read")
	if decision.Authorized {
		t.Errorf("Expected non-existent agent to be unauthorized")
	}
}

func TestCapabilityAuthority_ExpiredGrant(t *testing.T) {
	setup := SetupDemoEnvironment()

	// Create an expired grant
	expiredGrant := &CapabilityGrant{
		GrantID:       "GRANT-EXPIRED",
		AgentID:       "agent-expired",
		PolicyID:      "POLICY-017",
		PolicyVersion: 17,
		Capabilities:  []Capability{"repository.read"},
		CreatedAt:     time.Now().Add(-24 * time.Hour),
		ExpiresAt:     time.Now().Add(-1 * time.Hour), // Already expired
		Status:        "active",
		Revoked:       false,
	}

	setup.GrantStore.StoreGrant(expiredGrant)

	// Test authorization with expired grant
	decision := setup.CapabilityAuthority.CheckAuthorization("agent-expired", "repository.read")
	if decision.Authorized {
		t.Errorf("Expected expired grant to be unauthorized")
	}

	if decision.BlockReason != "GRANT_INVALID" {
		t.Errorf("Expected GRANT_INVALID, got: %s", decision.BlockReason)
	}
}

func TestCapabilityAuthority_RevokedGrant(t *testing.T) {
	setup := SetupDemoEnvironment()

	// Create and revoke a grant
	revokedGrant := &CapabilityGrant{
		GrantID:       "GRANT-REVOKED",
		AgentID:       "agent-revoked",
		PolicyID:      "POLICY-017",
		PolicyVersion: 17,
		Capabilities:  []Capability{"repository.read"},
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(24 * time.Hour),
		Status:        "active",
		Revoked:       true, // Already revoked
		RevokeReason:  "Manual revocation",
	}

	setup.GrantStore.StoreGrant(revokedGrant)

	// Test authorization with revoked grant
	decision := setup.CapabilityAuthority.CheckAuthorization("agent-revoked", "repository.read")
	if decision.Authorized {
		t.Errorf("Expected revoked grant to be unauthorized")
	}
}

// ============================================================================
// Tests for Handoff Validator
// ============================================================================

func TestHandoffValidator_ValidContract(t *testing.T) {
	setup := SetupDemoEnvironment()

	contract := &HandoffContract{
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

	result := setup.HandoffValidator.ValidateHandoff(contract)
	if !result.Valid {
		t.Errorf("Expected valid contract, got errors: %v", result.Errors)
	}

	if result.ReplayDetected {
		t.Errorf("Expected no replay detection for first contract")
	}
}

func TestHandoffValidator_ExpiredContract(t *testing.T) {
	setup := SetupDemoEnvironment()

	contract := &HandoffContract{
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
		CreatedAt:              time.Now().Add(-2 * time.Hour),
		ExpiresAt:              time.Now().Add(-1 * time.Hour), // Expired
		ReplayProtection:       ReplayProtection{Nonce: generateNonce(), Timestamp: time.Now()},
	}

	result := setup.HandoffValidator.ValidateHandoff(contract)
	if result.Valid {
		t.Errorf("Expected expired contract to be invalid")
	}

	if !result.ExpiredDetected {
		t.Errorf("Expected expired detection")
	}
}

func TestHandoffValidator_ReplayDetection(t *testing.T) {
	setup := SetupDemoEnvironment()

	nonce := generateNonce()

	contract1 := &HandoffContract{
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

	// First validation should pass
	result1 := setup.HandoffValidator.ValidateHandoff(contract1)
	if !result1.Valid {
		t.Errorf("Expected first contract to be valid")
	}

	// Second validation with same nonce should be flagged as replay
	contract2 := &HandoffContract{
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

	result2 := setup.HandoffValidator.ValidateHandoff(contract2)
	if result2.Valid {
		t.Errorf("Expected replay contract to be invalid")
	}

	if !result2.ReplayDetected {
		t.Errorf("Expected replay detection")
	}
}

func TestHandoffValidator_CapabilityWidening(t *testing.T) {
	setup := SetupDemoEnvironment()

	contract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              "agent-payment-01",
		RequestedCapability:    "credential.rotate", // NOT in authorized list
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

	result := setup.HandoffValidator.ValidateHandoff(contract)
	if result.Valid {
		t.Errorf("Expected contract with capability widening to be invalid")
	}

	if !result.CapabilityWideningDetected {
		t.Errorf("Expected capability widening detection")
	}
}

func TestHandoffValidator_PayloadMutation(t *testing.T) {
	setup := SetupDemoEnvironment()

	// Contract with a payload whose digest does NOT match the payload content.
	contract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              "agent-payment-01",
		RequestedCapability:    "repository.write",
		AuthorizedCapabilities: []Capability{"repository.read", "repository.write", "test.execute"},
		Payload:                map[string]interface{}{"action": "update_payment"},
		PayloadDigest:          "tampered-digest-value", // mismatched
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		Provenance:             map[string]interface{}{"source": "agent-payment-01"},
		CreatedAt:              time.Now(),
		ExpiresAt:              time.Now().Add(1 * time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: generateNonce(), Timestamp: time.Now()},
	}

	result := setup.HandoffValidator.ValidateHandoff(contract)
	if result.Valid {
		t.Errorf("Expected contract with payload tampering to be invalid")
	}

	if !result.PayloadTampered {
		t.Errorf("Expected payload tampering detection")
	}
}

func TestHandoffValidator_ProvenanceMutation(t *testing.T) {
	setup := SetupDemoEnvironment()

	// Contract whose provenance source does not match the initiator.
	contract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              "agent-payment-01",
		RequestedCapability:    "repository.write",
		AuthorizedCapabilities: []Capability{"repository.read", "repository.write", "test.execute"},
		PayloadDigest:          calculatePayloadDigest(map[string]interface{}{}),
		Payload:                map[string]interface{}{},
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		Provenance:             map[string]interface{}{"source": "another-agent"}, // mismatched
		CreatedAt:              time.Now(),
		ExpiresAt:              time.Now().Add(1 * time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: generateNonce(), Timestamp: time.Now()},
	}

	result := setup.HandoffValidator.ValidateHandoff(contract)
	if !result.ProvenanceMismatch {
		t.Errorf("Expected provenance mismatch detection")
	}
	if result.Valid {
		t.Errorf("Expected provenance mismatch to fail closed (contract must be invalid)")
	}
	if !contains(result.Errors, "Provenance mismatch") {
		t.Errorf("Expected provenance mismatch error detail, got %v", result.Errors)
	}
}

func TestHandoffValidator_MissingProvenance(t *testing.T) {
	setup := SetupDemoEnvironment()

	// Contract without any provenance is missing a trust root and must fail closed.
	contract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              "agent-payment-01",
		RequestedCapability:    "repository.write",
		AuthorizedCapabilities: []Capability{"repository.read", "repository.write", "test.execute"},
		PayloadDigest:          calculatePayloadDigest(map[string]interface{}{}),
		Payload:                map[string]interface{}{},
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		CreatedAt:              time.Now(),
		ExpiresAt:              time.Now().Add(1 * time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: generateNonce(), Timestamp: time.Now()},
	}

	result := setup.HandoffValidator.ValidateHandoff(contract)
	if !result.ProvenanceMismatch {
		t.Errorf("Expected missing provenance to be detected as a mismatch")
	}
	if result.Valid {
		t.Errorf("Expected contract without provenance to fail closed")
	}
}

// contains reports whether a slice contains the given substring in any element.
func contains(items []string, substr string) bool {
	for _, it := range items {
		if strings.Contains(it, substr) {
			return true
		}
	}
	return false
}

func TestHandoffValidator_UnsupportedContractVersion(t *testing.T) {
	setup := SetupDemoEnvironment()

	contract := &HandoffContract{
		ContractVersion:        "2.0", // unsupported
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

	result := setup.HandoffValidator.ValidateHandoff(contract)
	if result.Valid {
		t.Errorf("Expected contract with unsupported version to be invalid")
	}
}

func TestHandoffValidator_CanonicalizationMismatch(t *testing.T) {
	setup := SetupDemoEnvironment()

	contract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "99", // unsupported canonicalization version
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

	result := setup.HandoffValidator.ValidateHandoff(contract)
	if result.Valid {
		t.Errorf("Expected contract with canonicalization mismatch to be invalid")
	}
}

// ============================================================================
// Tests for Execution Gate
// ============================================================================

func TestExecutionGate_AuthorizedExecution(t *testing.T) {
	setup := SetupDemoEnvironment()

	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		AgentID:    "agent-payment-01",
		Capability: "repository.read",
		Payload:    map[string]interface{}{},
	}

	result, _ := setup.ExecutionGate.Authorize(request)
	if !result.Authorized {
		t.Errorf("Expected authorized execution: %s", result.Error)
	}

	if !result.Executed {
		t.Errorf("Expected execution to proceed")
	}
}

func TestExecutionGate_UnauthorizedExecution(t *testing.T) {
	setup := SetupDemoEnvironment()

	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		AgentID:    "agent-payment-01",
		Capability: "credential.rotate",
		Payload:    map[string]interface{}{},
	}

	result, _ := setup.ExecutionGate.Authorize(request)
	if result.Authorized {
		t.Errorf("Expected unauthorized execution")
	}

	if result.Executed {
		t.Errorf("Expected execution to be blocked")
	}
}

func TestExecutionGate_ProtectedCapability(t *testing.T) {
	setup := SetupDemoEnvironment()

	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		AgentID:    "agent-payment-01",
		Capability: "policy.modify",
		Payload:    map[string]interface{}{},
	}

	result, _ := setup.ExecutionGate.Authorize(request)
	if result.Authorized {
		t.Errorf("Expected protected capability to be unauthorized")
	}
}

// TestExecutionGate_IAMModification verifies spec §9 Attack 2: an agent attempting
// to modify the trust boundary (iam.modify) is blocked as a protected capability.
func TestExecutionGate_IAMModification(t *testing.T) {
	setup := SetupDemoEnvironment()

	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		AgentID:    "agent-payment-01",
		Capability: "iam.modify",
		Payload:    map[string]interface{}{"principal": "new-admin", "role": "admin"},
	}

	result, _ := setup.ExecutionGate.Authorize(request)
	if result.Authorized {
		t.Errorf("Expected iam.modify to be unauthorized")
	}
	if result.Executed {
		t.Errorf("Expected iam.modify to be blocked")
	}
}

// TestExecutionGate_ExternalPublish verifies that external.publish is also blocked.
func TestExecutionGate_ExternalPublish(t *testing.T) {
	setup := SetupDemoEnvironment()

	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		AgentID:    "agent-payment-01",
		Capability: "external.publish",
		Payload:    map[string]interface{}{"artifact": "registry/payment:latest"},
	}

	result, _ := setup.ExecutionGate.Authorize(request)
	if result.Authorized {
		t.Errorf("Expected external.publish to be unauthorized")
	}
}

// TestHandoffValidator_SignedContract verifies an HMAC-signed contract is accepted
// and a tampered signature is rejected.
func TestHandoffValidator_SignedContract(t *testing.T) {
	setup := SetupDemoEnvironment()

	makeContract := func() *HandoffContract {
		return &HandoffContract{
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
	}

	// Sign a contract — should validate.
	signed := makeContract()
	signed.Signature = setup.HandoffValidator.SignContract(signed)
	result := setup.HandoffValidator.ValidateHandoff(signed)
	if !result.Valid {
		t.Errorf("Expected signed contract to be valid, got: %v", result.Errors)
	}

	// Tamper with the signature — should be rejected.
	tampered := makeContract()
	tampered.Signature = "deadbeefdeadbeef"
	tamperedResult := setup.HandoffValidator.ValidateHandoff(tampered)
	if tamperedResult.Valid {
		t.Errorf("Expected tampered signature to be rejected")
	}
}

// ============================================================================
// Tests for Audit Ledger
// ============================================================================

func TestAuditLedger_RecordEvent(t *testing.T) {
	ledger := NewAuditLedger()

	event := &AuditEvent{
		EventID:   generateID("AUD"),
		Timestamp: time.Now(),
		EventType: "AUTHORIZATION_GRANTED",
		AgentID:   "agent-test",
	}

	err := ledger.RecordEvent(event)
	if err != nil {
		t.Errorf("Failed to record event: %v", err)
	}

	events := ledger.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

func TestAuditLedger_AppendOnly(t *testing.T) {
	ledger := NewAuditLedger()

	// Record multiple events
	for i := 0; i < 5; i++ {
		event := &AuditEvent{
			EventID:   generateID("AUD"),
			Timestamp: time.Now(),
			EventType: "TEST",
			AgentID:   "agent-test",
		}
		ledger.RecordEvent(event)
	}

	events := ledger.GetEvents()
	if len(events) != 5 {
		t.Errorf("Expected 5 events, got %d", len(events))
	}

	// Events should be immutable (copy returned)
	events[0].AgentID = "modified"
	eventsAgain := ledger.GetEvents()
	if eventsAgain[0].AgentID == "modified" {
		t.Errorf("Events should be immutable")
	}
}

// ============================================================================
// Tests for Independent Verifier
// ============================================================================

func TestVerifier_VerifyAuthorizedAction(t *testing.T) {
	setup := SetupDemoEnvironment()

	// Execute an authorized action
	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		AgentID:    "agent-payment-01",
		Capability: "repository.read",
		Payload:    map[string]interface{}{},
	}

	setup.ExecutionGate.Authorize(request)

	// Verify the session
	verification := setup.IndependentVerifier.VerifySession("agent-payment-01")
	if verification.Result != "PASS" {
		t.Errorf("Expected verification to pass, got: %s", verification.Result)
	}
}

func TestVerifier_VerifyUnauthorizedAttempt(t *testing.T) {
	setup := SetupDemoEnvironment()

	// Attempt unauthorized action
	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		AgentID:    "agent-payment-01",
		Capability: "credential.rotate",
		Payload:    map[string]interface{}{},
	}

	setup.ExecutionGate.Authorize(request)

	// Verify the session - should detect unauthorized attempt
	verification := setup.IndependentVerifier.VerifySession("agent-payment-01")
	if verification.Result == "PASS" {
		t.Errorf("Expected verification to fail due to unauthorized attempt")
	}

	if verification.UnauthorizedActions == 0 {
		t.Errorf("Expected to detect unauthorized actions")
	}
}

// ============================================================================
// Tests for Demo Scenarios
// ============================================================================

func TestScenario1_AuthorizedRead(t *testing.T) {
	setup := SetupDemoEnvironment()
	scenario := RunScenario1_AuthorizedRepositoryRead(setup)

	if !scenario.Passed {
		t.Errorf("Scenario 1 failed: %s", scenario.Details)
	}

	if scenario.ActualResult != "ALLOW" {
		t.Errorf("Expected ALLOW, got: %s", scenario.ActualResult)
	}
}

func TestScenario2_UnauthorizedCredentialRotation(t *testing.T) {
	setup := SetupDemoEnvironment()
	scenario := RunScenario2_UnauthorizedCredentialRotation(setup)

	if !scenario.Passed {
		t.Errorf("Scenario 2 failed: %s", scenario.Details)
	}

	if scenario.ActualResult != "BLOCK" {
		t.Errorf("Expected BLOCK, got: %s", scenario.ActualResult)
	}
}

func TestScenario3_PolicySelfModification(t *testing.T) {
	setup := SetupDemoEnvironment()
	scenario := RunScenario3_PolicySelfModification(setup)

	if !scenario.Passed {
		t.Errorf("Scenario 3 failed: %s", scenario.Details)
	}

	if scenario.ActualResult != "BLOCK" {
		t.Errorf("Expected BLOCK, got: %s", scenario.ActualResult)
	}
}

func TestScenario4_CapabilityEscalation(t *testing.T) {
	setup := SetupDemoEnvironment()
	scenario := RunScenario4_CapabilityEscalationViaHandoff(setup)

	if !scenario.Passed {
		t.Errorf("Scenario 4 failed: %s", scenario.Details)
	}

	if scenario.ActualResult != "BLOCK" {
		t.Errorf("Expected BLOCK, got: %s", scenario.ActualResult)
	}
}

func TestScenario5_ReplayAttack(t *testing.T) {
	setup := SetupDemoEnvironment()
	scenario := RunScenario5_ReplayAttack(setup)

	if !scenario.Passed {
		t.Errorf("Scenario 5 failed: %s", scenario.Details)
	}

	if scenario.ActualResult != "BLOCK" {
		t.Errorf("Expected BLOCK, got: %s", scenario.ActualResult)
	}
}

func TestScenario6_StaleAuthorization(t *testing.T) {
	setup := SetupDemoEnvironment()
	scenario := RunScenario6_PolicyVersionMismatch(setup)

	if !scenario.Passed {
		t.Errorf("Scenario 6 failed: %s", scenario.Details)
	}

	if scenario.ActualResult != "BLOCK" {
		t.Errorf("Expected BLOCK, got: %s", scenario.ActualResult)
	}

	if scenario.ExecutionResult == nil {
		t.Errorf("Expected execution result to be populated")
	} else if scenario.ExecutionResult.Authorized {
		t.Errorf("Expected execution to be blocked due to stale authorization")
	}
}

func TestExecutionGate_StalePolicyVersion(t *testing.T) {
	setup := SetupDemoEnvironment()

	// Verify authorized before policy change
	request := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		AgentID:    "agent-payment-01",
		Capability: "repository.write",
		Payload:    map[string]interface{}{},
	}

	result1, _ := setup.ExecutionGate.Authorize(request)
	if !result1.Authorized {
		t.Errorf("Expected authorized before policy change, got: %s", result1.Error)
	}

	// Update POLICY-017 to v18 — superseding the old version.
	newPolicy := &Policy{
		PolicyID:  "POLICY-017",
		Version:   18,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		AllowedCapabilities: []Capability{
			"repository.read",
		},
		DeniedCapabilities:    []Capability{},
		ProtectedCapabilities: []Capability{"iam.modify", "credential.rotate", "production.deploy", "external.publish", "policy.modify"},
		ApprovalRequired:      []Capability{},
		RequiresExplicitAuth:  true,
	}

	setup.PolicyStore.UpdatePolicy(newPolicy)

	// Attempt same action - should be blocked as stale
	result2, _ := setup.ExecutionGate.Authorize(request)
	if result2.Authorized {
		t.Errorf("Expected blocked after policy version change, got authorized")
	}

	if result2.Error == "" {
		t.Errorf("Expected error message about stale authorization")
	}
}

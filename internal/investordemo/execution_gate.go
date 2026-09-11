package investordemo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// Handoff Contract Validator
// ============================================================================

type HandoffValidator struct {
	auditLedger  *AuditLedger
	policyEngine *PolicyEngine
	grantStore   *GrantStore
	signingKey   []byte               // HMAC-SHA256 signing key
	seenNonces   map[string]time.Time // nonce -> timestamp (for replay detection)
	mu           sync.RWMutex
}

// NewHandoffValidator creates a new handoff validator.
func NewHandoffValidator(auditLedger *AuditLedger, policyEngine *PolicyEngine, grantStore *GrantStore) *HandoffValidator {
	return &HandoffValidator{
		auditLedger:  auditLedger,
		policyEngine: policyEngine,
		grantStore:   grantStore,
		signingKey:   []byte("naeos-demo-signing-key-change-in-production"),
		seenNonces:   make(map[string]time.Time),
	}
}

// SignContract produces an HMAC-SHA256 signature over the contract's canonical fields.
func (hv *HandoffValidator) SignContract(contract *HandoffContract) string {
	canonical := hv.buildSigningPayload(contract)
	mac := hmac.New(sha256.New, hv.signingKey)
	mac.Write([]byte(canonical)) //nolint:errcheck // write to hash never fails
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyContractSignature verifies the HMAC-SHA256 signature of a handoff contract.
func (hv *HandoffValidator) VerifyContractSignature(contract *HandoffContract) bool {
	if contract.Signature == "" {
		return false
	}
	expected := hv.SignContract(contract)
	return hmac.Equal([]byte(expected), []byte(contract.Signature))
}

// buildSigningPayload creates a canonical string from the contract's key fields.
func (hv *HandoffValidator) buildSigningPayload(contract *HandoffContract) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("version:%s", contract.ContractVersion))
	parts = append(parts, fmt.Sprintf("canonical:%s", contract.CanonicalVersion))
	parts = append(parts, fmt.Sprintf("initiator:%s", contract.Initiator))
	parts = append(parts, fmt.Sprintf("requested:%s", contract.RequestedCapability))
	parts = append(parts, fmt.Sprintf("policy:%s:%d", contract.PolicyID, contract.PolicyVersion))
	parts = append(parts, fmt.Sprintf("digest:%s", contract.PayloadDigest))
	parts = append(parts, fmt.Sprintf("nonce:%s", contract.ReplayProtection.Nonce))
	parts = append(parts, fmt.Sprintf("expires:%s", contract.ExpiresAt.UTC().Format(time.RFC3339)))

	// Sort authorized capabilities for deterministic output
	caps := make([]string, len(contract.AuthorizedCapabilities))
	for i, c := range contract.AuthorizedCapabilities {
		caps[i] = string(c)
	}
	sort.Strings(caps)
	parts = append(parts, fmt.Sprintf("caps:%s", strings.Join(caps, ",")))

	return strings.Join(parts, "|")
}

// ValidateHandoff validates a handoff contract and detects attacks.
func (hv *HandoffValidator) ValidateHandoff(contract *HandoffContract) *HandoffValidationResult {
	result := &HandoffValidationResult{
		ValidationID:    generateID("HVAL"),
		Timestamp:       time.Now(),
		ContractVersion: contract.ContractVersion,
		Valid:           true,
		Errors:          []string{},
		Warnings:        []string{},
	}

	// Check contract version
	if contract.ContractVersion != "1.0" {
		result.Errors = append(result.Errors, fmt.Sprintf("Unsupported contract version: %s", contract.ContractVersion))
		result.Valid = false
	}

	// Check canonicalization version
	if contract.CanonicalVersion != "1" {
		result.Errors = append(result.Errors, fmt.Sprintf("Unsupported canonicalization version: %s", contract.CanonicalVersion))
		result.Valid = false
	}

	// Verify HMAC-SHA256 signature
	if contract.Signature != "" {
		if !hv.VerifyContractSignature(contract) {
			result.Errors = append(result.Errors, "Contract signature verification failed - contract may have been tampered with")
			result.Valid = false
			hv.recordAuditEvent("CONTRACT_TAMPER_DETECTED", contract.Initiator, map[string]interface{}{
				"initiator": contract.Initiator,
			})
		}
	}

	// Check if contract is expired
	if time.Now().After(contract.ExpiresAt) {
		result.Errors = append(result.Errors, "Contract is expired")
		result.ExpiredDetected = true
		result.Valid = false
	}

	// Check replay protection - detect if this nonce has been seen before
	if hv.isReplayedNonce(contract.ReplayProtection.Nonce) {
		result.Errors = append(result.Errors, fmt.Sprintf("Replay attack detected: nonce %s has been seen before", contract.ReplayProtection.Nonce))
		result.ReplayDetected = true
		result.Valid = false
		hv.recordAuditEvent("REPLAY_ATTACK_DETECTED", contract.Initiator, map[string]interface{}{
			"nonce": contract.ReplayProtection.Nonce,
		})
	} else {
		hv.recordNonce(contract.ReplayProtection.Nonce)
	}

	// Validate payload digest
	if contract.Payload != nil {
		calculatedDigest := calculatePayloadDigest(contract.Payload)
		if calculatedDigest != contract.PayloadDigest {
			result.Errors = append(result.Errors, "Payload digest mismatch - payload may have been tampered with")
			result.PayloadTampered = true
			result.Valid = false
		}
	}

	// Check for capability widening
	// The requested capability must be in the authorized capabilities
	capabilityFound := false
	for _, authedCap := range contract.AuthorizedCapabilities {
		if authedCap == contract.RequestedCapability {
			capabilityFound = true
			break
		}
	}

	if !capabilityFound {
		result.Errors = append(result.Errors, fmt.Sprintf("Requested capability %s not in authorized capabilities", contract.RequestedCapability))
		result.CapabilityWideningDetected = true
		result.Valid = false
	}

	// Check for downstream capability escalation
	if contract.DownstreamHandoff != nil {
		// The downstream handoff cannot request a capability that was not in the parent handoff
		hasCapability := false
		for _, cap := range contract.AuthorizedCapabilities {
			if cap == contract.DownstreamHandoff.RequestedCapability {
				hasCapability = true
				break
			}
		}

		if !hasCapability {
			result.Errors = append(result.Errors, fmt.Sprintf("Downstream handoff requests capability %s which was not authorized in parent handoff", contract.DownstreamHandoff.RequestedCapability))
			result.CapabilityWideningDetected = true
			result.Valid = false
			hv.recordAuditEvent("CAPABILITY_ESCALATION_DETECTED", contract.Initiator, map[string]interface{}{
				"parent_authorized":    contract.AuthorizedCapabilities,
				"downstream_requested": contract.DownstreamHandoff.RequestedCapability,
			})
		}
	}

	// Check provenance.
	// Provenance mismatch is a hard failure: a contract that claims a source
	// different from its initiator must be rejected (fail closed).
	if contract.Provenance == nil {
		result.Errors = append(result.Errors, "Contract provenance is missing")
		result.ProvenanceMismatch = true
		result.Valid = false
	} else if source, ok := contract.Provenance["source"]; !ok || source != contract.Initiator {
		result.Errors = append(result.Errors, "Provenance mismatch: source does not match initiator")
		result.ProvenanceMismatch = true
		result.Valid = false
		hv.recordAuditEvent("PROVENANCE_MISMATCH_DETECTED", contract.Initiator, map[string]interface{}{
			"initiator": contract.Initiator,
			"source":    contract.Provenance["source"],
		})
	}

	// Record validation event
	hv.recordAuditEvent("HANDOFF_VALIDATION", contract.Initiator, map[string]interface{}{
		"valid":            result.Valid,
		"errors":           result.Errors,
		"contract_version": contract.ContractVersion,
	})

	return result
}

// recordAuditEvent is a helper that records an audit event, discarding the error
// since audit recording in the demo is best-effort (append-only in-memory store).
func (hv *HandoffValidator) recordAuditEvent(eventType, agentID string, details map[string]interface{}) {
	_ = hv.auditLedger.RecordEvent(&AuditEvent{
		Timestamp: time.Now(),
		EventType: eventType,
		AgentID:   agentID,
		Details:   details,
	})
}

// isReplayedNonce checks if a nonce has been seen before.
func (hv *HandoffValidator) isReplayedNonce(nonce string) bool {
	hv.mu.RLock()
	defer hv.mu.RUnlock()

	_, exists := hv.seenNonces[nonce]
	return exists
}

// recordNonce records a nonce as seen.
func (hv *HandoffValidator) recordNonce(nonce string) {
	hv.mu.Lock()
	defer hv.mu.Unlock()

	hv.seenNonces[nonce] = time.Now()
}

// Reset clears the nonce ledger. Used when the demo is reset.
func (hv *HandoffValidator) Reset() {
	hv.mu.Lock()
	defer hv.mu.Unlock()
	hv.seenNonces = make(map[string]time.Time)
}

// calculatePayloadDigest calculates a canonical SHA256 digest of a payload.
// Uses deterministic JSON serialization with sorted keys for consistency.
func calculatePayloadDigest(payload map[string]interface{}) string {
	canonical := canonicalizeMap(payload)
	data, err := json.Marshal(canonical)
	if err != nil {
		// Fallback: hash the error itself (deterministic but invalid)
		hash := sha256.Sum256([]byte(fmt.Sprintf("digest-error:%v", err)))
		return hex.EncodeToString(hash[:])
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// canonicalizeMap produces a deterministically-ordered representation of a map.
func canonicalizeMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = canonicalizeValue(v)
	}
	return result
}

// canonicalizeValue recursively sorts map keys in nested structures.
func canonicalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		return canonicalizeMap(val)
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = canonicalizeValue(item)
		}
		return result
	case map[string]string:
		sorted := make(map[string]interface{}, len(val))
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sorted[k] = val[k]
		}
		return sorted
	default:
		return val
	}
}

// ============================================================================
// Execution Gate
// ============================================================================

type ExecutionGate struct {
	capabilityAuthority *CapabilityAuthority
	handoffValidator    *HandoffValidator
	auditLedger         *AuditLedger
	policyEngine        *PolicyEngine
	verifier            *IndependentVerifier
}

// NewExecutionGate creates a new execution gate.
func NewExecutionGate(
	capabilityAuthority *CapabilityAuthority,
	handoffValidator *HandoffValidator,
	auditLedger *AuditLedger,
	policyEngine *PolicyEngine,
	verifier *IndependentVerifier,
) *ExecutionGate {
	return &ExecutionGate{
		capabilityAuthority: capabilityAuthority,
		handoffValidator:    handoffValidator,
		auditLedger:         auditLedger,
		policyEngine:        policyEngine,
		verifier:            verifier,
	}
}

// Authorize determines if an execution request should be allowed or blocked.
// This is the main gate that all consequential actions pass through.
func (eg *ExecutionGate) Authorize(request *ExecutionRequest) (*ExecutionResult, error) {
	result := &ExecutionResult{
		ExecutionID: generateID("EXEC"),
		RequestID:   request.RequestID,
		Timestamp:   time.Now(),
		AgentID:     request.AgentID,
		Capability:  request.Capability,
		Authorized:  false,
		Executed:    false,
	}

	// Step 1: Get the agent's grant and check authorization
	authDecision := eg.capabilityAuthority.CheckAuthorization(request.AgentID, request.Capability)
	if !authDecision.Authorized {
		result.Error = fmt.Sprintf("Authorization denied: %s", authDecision.Reason)
		eg.recordAuditEvent("EXECUTION_BLOCKED", request.AgentID, request.Capability, "", 0, "BLOCK", authDecision.BlockReason)
		return result, nil
	}

	// Step 2: If a handoff contract was provided, validate it
	if request.HandoffContract != nil {
		handoffValidation := eg.handoffValidator.ValidateHandoff(request.HandoffContract)
		if !handoffValidation.Valid {
			result.Error = fmt.Sprintf("Handoff validation failed: %v", handoffValidation.Errors)
			eg.recordAuditEvent("EXECUTION_BLOCKED", request.AgentID, request.Capability, "", 0, "BLOCK", "HANDOFF_VALIDATION_FAILED")
			return result, nil
		}
	}

	// Step 3: Re-check authorization at execution time using current policy
	// This ensures that policy changes are respected even if the authorization was valid before
	currentGrant, err := eg.capabilityAuthority.grantStore.GetGrantByAgent(request.AgentID)
	if err != nil || !currentGrant.IsValid() {
		result.Error = "Authorization is stale or invalid at execution time"
		eg.recordAuditEvent("EXECUTION_BLOCKED", request.AgentID, request.Capability, "", 0, "BLOCK", "STALE_AUTHORIZATION")
		return result, nil
	}

	// Step 3b: Verify the grant's policy version still matches the active policy.
	// If the policy has been superseded (e.g., POLICY-017 v17 → v18), the old grant
	// is stale and must be blocked. This enforces Demo Principle #6:
	// "Bind authorization to policy version."
	activePolicy, policyErr := eg.policyEngine.policyStore.GetPolicy(currentGrant.PolicyID)
	if policyErr != nil || activePolicy.Version != currentGrant.PolicyVersion {
		result.Error = fmt.Sprintf("Authorization is stale: granted under %s v%d but current active policy has changed",
			currentGrant.PolicyID, currentGrant.PolicyVersion)
		activeVersion := 0
		if policyErr == nil {
			activeVersion = activePolicy.Version
		}
		_ = eg.auditLedger.RecordEvent(&AuditEvent{
			Timestamp:           time.Now(),
			EventType:           "EXECUTION_BLOCKED",
			AgentID:             request.AgentID,
			RequestedCapability: request.Capability,
			PolicyID:            currentGrant.PolicyID,
			PolicyVersion:       currentGrant.PolicyVersion,
			Decision:            "BLOCK",
			Reason:              "STALE_AUTHORIZATION",
			Details: map[string]interface{}{
				"grant_policy_version":  currentGrant.PolicyVersion,
				"active_policy_version": activeVersion,
			},
		})
		return result, nil
	}

	// Check if the capability is protected (like iam.modify, policy.modify, credential.rotate)
	// Protected capabilities should never be executed
	isProtected, _ := eg.policyEngine.IsProtectedCapability(currentGrant.PolicyID, currentGrant.PolicyVersion, request.Capability)
	if isProtected {
		result.Error = fmt.Sprintf("Capability %s is protected and cannot be executed", request.Capability)
		eg.recordAuditEvent("EXECUTION_BLOCKED", request.AgentID, request.Capability, currentGrant.PolicyID, currentGrant.PolicyVersion, "BLOCK", "PROTECTED_CAPABILITY")
		return result, nil
	}

	// Step 4: Authorization granted - execute the capability
	result.Authorized = true
	result.Executed = true
	result.Result = map[string]interface{}{
		"status":     "executed",
		"capability": request.Capability,
	}

	// Step 5: Run independent verification
	verificationResult := eg.verifier.Verify(request.AgentID, request.Capability, currentGrant.PolicyID)
	result.VerificationStatus = verificationResult.Result

	eg.recordAuditEvent("EXECUTION_ALLOWED", request.AgentID, request.Capability, currentGrant.PolicyID, currentGrant.PolicyVersion, "ALLOW", "AUTHORIZED")

	return result, nil
}

// recordAuditEvent is a helper that records an execution audit event.
func (eg *ExecutionGate) recordAuditEvent(eventType, agentID string, cap Capability, policyID string, policyVersion int, decision, reason string) {
	_ = eg.auditLedger.RecordEvent(&AuditEvent{
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

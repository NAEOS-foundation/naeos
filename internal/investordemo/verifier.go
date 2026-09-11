package investordemo

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Audit Ledger - Append-only event log
// ============================================================================

type AuditLedger struct {
	mu     sync.RWMutex
	events []*AuditEvent
}

// NewAuditLedger creates a new audit ledger.
func NewAuditLedger() *AuditLedger {
	return &AuditLedger{
		events: make([]*AuditEvent, 0),
	}
}

// RecordEvent records an audit event (append-only).
func (al *AuditLedger) RecordEvent(event *AuditEvent) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	if event.EventID == "" {
		return fmt.Errorf("event ID is required")
	}

	// Events are append-only - cannot be modified
	al.events = append(al.events, event)
	return nil
}

// GetEvents returns all events.
func (al *AuditLedger) GetEvents() []*AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	// Return deep copies to prevent external modification
	eventsCopy := make([]*AuditEvent, len(al.events))
	for i, ev := range al.events {
		evCopy := *ev
		eventsCopy[i] = &evCopy
	}
	return eventsCopy
}

// GetEventsByAgent returns all events for a specific agent.
func (al *AuditLedger) GetEventsByAgent(agentID string) []*AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var result []*AuditEvent
	for _, event := range al.events {
		if event.AgentID == agentID {
			result = append(result, event)
		}
	}
	return result
}

// GetEventsByType returns all events of a specific type.
func (al *AuditLedger) GetEventsByType(eventType string) []*AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var result []*AuditEvent
	for _, event := range al.events {
		if event.EventType == eventType {
			result = append(result, event)
		}
	}
	return result
}

// CountEventsByType counts events of a specific type.
func (al *AuditLedger) CountEventsByType(eventType string) int {
	al.mu.RLock()
	defer al.mu.RUnlock()

	count := 0
	for _, event := range al.events {
		if event.EventType == eventType {
			count++
		}
	}
	return count
}

// ============================================================================
// Independent Verifier
// ============================================================================

type IndependentVerifier struct {
	auditLedger  *AuditLedger
	policyEngine *PolicyEngine
	grantStore   *GrantStore
}

// NewIndependentVerifier creates a new independent verifier.
func NewIndependentVerifier(auditLedger *AuditLedger, policyEngine *PolicyEngine, grantStore *GrantStore) *IndependentVerifier {
	return &IndependentVerifier{
		auditLedger:  auditLedger,
		policyEngine: policyEngine,
		grantStore:   grantStore,
	}
}

// Verify independently verifies that an agent's action was authorized.
// The verifier is separate from the execution decision to ensure independent review.
func (iv *IndependentVerifier) Verify(agentID string, capability Capability, policyID string) *VerificationResult {
	result := &VerificationResult{
		VerificationID:      generateID("VER"),
		Timestamp:           time.Now(),
		AgentID:             agentID,
		Result:              "PASS",
		PolicyCompliance:    true,
		TestsPassed:         true,
		UnauthorizedActions: 0,
		Issues:              []string{},
	}

	// Check 1: Verify that the grant exists and is valid
	grant, err := iv.grantStore.GetGrantByAgent(agentID)
	if err != nil {
		result.Result = "FAIL"
		result.PolicyCompliance = false
		result.Issues = append(result.Issues, fmt.Sprintf("No valid grant found for agent %s", agentID))
		return result
	}

	if !grant.IsValid() {
		result.Result = "FAIL"
		result.PolicyCompliance = false
		result.Issues = append(result.Issues, fmt.Sprintf("Grant %s is not valid", grant.GrantID))
		return result
	}

	// Check 2: Verify the capability is in the grant
	hasCapability := false
	for _, grantedCap := range grant.Capabilities {
		if grantedCap == capability {
			hasCapability = true
			break
		}
	}

	if !hasCapability {
		result.Result = "FAIL"
		result.PolicyCompliance = false
		result.UnauthorizedActions = 1
		result.Issues = append(result.Issues, fmt.Sprintf("Capability %s not in grant", capability))
		return result
	}

	// Check 3: Verify the policy allows this capability
	allowed, policyReason := iv.policyEngine.EvaluateCapability(policyID, grant.PolicyVersion, capability)
	if !allowed {
		result.Result = "FAIL"
		result.PolicyCompliance = false
		result.UnauthorizedActions = 1
		result.Issues = append(result.Issues, policyReason)
		return result
	}

	// Check 4: Verify no unauthorized actions were recorded in the audit log
	agentEvents := iv.auditLedger.GetEventsByAgent(agentID)
	unauthorizedCount := 0
	for _, event := range agentEvents {
		if event.EventType == "AUTHORIZATION_DENIED" || event.EventType == "EXECUTION_BLOCKED" {
			unauthorizedCount++
		}
	}

	if unauthorizedCount > 0 {
		result.Result = "FAIL"
		result.PolicyCompliance = false
		result.UnauthorizedActions = unauthorizedCount
		result.Issues = append(result.Issues, fmt.Sprintf("%d unauthorized actions detected in audit log", unauthorizedCount))
		return result
	}

	// All checks passed
	result.Result = "PASS"
	result.Summary = fmt.Sprintf("Agent %s is authorized for capability %s under policy %s", agentID, capability, policyID)

	return result
}

// VerifySession verifies an entire agent session (multiple capabilities).
func (iv *IndependentVerifier) VerifySession(agentID string) *VerificationResult {
	result := &VerificationResult{
		VerificationID:      generateID("VER"),
		Timestamp:           time.Now(),
		AgentID:             agentID,
		Result:              "PASS",
		PolicyCompliance:    true,
		TestsPassed:         true,
		UnauthorizedActions: 0,
		Issues:              []string{},
	}

	// Get all events for this agent
	agentEvents := iv.auditLedger.GetEventsByAgent(agentID)

	// Check if there are any blocked or denied actions
	for _, event := range agentEvents {
		if event.EventType == "AUTHORIZATION_DENIED" {
			result.UnauthorizedActions++
		}
		if event.EventType == "EXECUTION_BLOCKED" {
			result.UnauthorizedActions++
		}
		if event.EventType == "CAPABILITY_ESCALATION_DETECTED" {
			result.UnauthorizedActions++
			result.Issues = append(result.Issues, "Capability escalation was detected and prevented")
		}
		if event.EventType == "REPLAY_ATTACK_DETECTED" {
			result.UnauthorizedActions++
			result.Issues = append(result.Issues, "Replay attack was detected and prevented")
		}
		if event.EventType == "POLICY_CHANGE" {
			// Policy changed - need to re-evaluate all subsequent actions
			result.Warnings = append(result.Warnings, "Policy was changed during session")
		}
	}

	if result.UnauthorizedActions > 0 {
		result.Result = "FAIL"
		result.PolicyCompliance = false
	}

	if len(result.Issues) > 0 {
		result.Summary = fmt.Sprintf("Agent %s had %d unauthorized actions", agentID, result.UnauthorizedActions)
	} else {
		result.Summary = fmt.Sprintf("Agent %s session passed all verifications", agentID)
	}

	return result
}

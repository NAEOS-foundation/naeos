package investordemo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ============================================================================
// HTTP Server API for the Demo
// ============================================================================

type APIServer struct {
	setup *DemoSetup
	mux   *http.ServeMux
}

// NewAPIServer creates a new API server for the demo.
func NewAPIServer(setup *DemoSetup) *APIServer {
	server := &APIServer{
		setup: setup,
		mux:   http.NewServeMux(),
	}

	// Register endpoints
	server.mux.HandleFunc("/api/health", server.handleHealth)
	server.mux.HandleFunc("/api/authorize", server.handleAuthorize)
	server.mux.HandleFunc("/api/execute", server.handleExecute)
	server.mux.HandleFunc("/api/audit", server.handleAuditEvents)
	server.mux.HandleFunc("/api/scenarios", server.handleRunScenarios)
	server.mux.HandleFunc("/api/investor-demo", server.handleInvestorDemo)
	server.mux.HandleFunc("/api/policy", server.handleGetPolicy)
	server.mux.HandleFunc("/api/grants", server.handleGetGrants)
	server.mux.HandleFunc("/api/verification", server.handleVerifySession)
	server.mux.HandleFunc("/api/reset", server.handleReset)

	return server
}

// ServeHTTP implements the http.Handler interface.
func (as *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	as.mux.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// handleHealth returns server health status.
func (as *APIServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// handleAuthorize checks if an agent is authorized for a capability (without executing).
func (as *APIServer) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AgentID    string `json:"agent_id"`
		Capability string `json:"capability"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	decision := as.setup.CapabilityAuthority.CheckAuthorization(req.AgentID, Capability(req.Capability))
	writeJSON(w, decision)
}

// handleExecute executes an action through the execution gate.
func (as *APIServer) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AgentID    string                 `json:"agent_id"`
		Capability string                 `json:"capability"`
		Payload    map[string]interface{} `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	execRequest := &ExecutionRequest{
		RequestID:  generateID("REQ"),
		AgentID:    req.AgentID,
		Capability: Capability(req.Capability),
		Payload:    req.Payload,
	}

	result, _ := as.setup.ExecutionGate.Authorize(execRequest)
	writeJSON(w, result)
}

// handleAuditEvents returns recent audit events.
func (as *APIServer) handleAuditEvents(w http.ResponseWriter, _ *http.Request) {
	events := as.setup.AuditLedger.GetEvents()
	writeJSON(w, map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}

// handleRunScenarios runs all demo scenarios.
func (as *APIServer) handleRunScenarios(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scenarios := RunAllScenarios(as.setup)
	writeJSON(w, map[string]interface{}{
		"scenarios": scenarios,
		"total":     len(scenarios),
	})
}

// handleGetPolicy returns the current policy.
func (as *APIServer) handleGetPolicy(w http.ResponseWriter, _ *http.Request) {
	policy, _ := as.setup.PolicyStore.GetPolicy("POLICY-017")
	writeJSON(w, policy)
}

// handleGetGrants returns all grants.
func (as *APIServer) handleGetGrants(w http.ResponseWriter, _ *http.Request) {
	grants := as.setup.GrantStore.ListGrants()
	writeJSON(w, map[string]interface{}{
		"grants": grants,
		"total":  len(grants),
	})
}

// handleVerifySession runs independent verification on an agent's session.
func (as *APIServer) handleVerifySession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AgentID string `json:"agent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	verification := as.setup.IndependentVerifier.VerifySession(req.AgentID)
	writeJSON(w, verification)
}

// handleReset resets the demo to initial state.
func (as *APIServer) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	as.setup.HandoffValidator.Reset()
	as.setup = SetupDemoEnvironment()

	writeJSON(w, map[string]string{
		"status":  "reset",
		"message": "Demo environment reset to initial state",
	})
}

// DemoStep represents a single step in the scripted investor demo.
type DemoStep struct {
	Step        int                    `json:"step"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Action      string                 `json:"action"`
	Result      string                 `json:"result"`
	Blocked     bool                   `json:"blocked"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// handleInvestorDemo runs the scripted investor demo sequence per spec §13.
func (as *APIServer) handleInvestorDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Reset to clean state before running demo
	as.setup.HandoffValidator.Reset()
	as.setup = SetupDemoEnvironment()

	steps := []DemoStep{}
	agentID := "agent-payment-01"

	// Step 1: Agent receives task
	steps = append(steps, DemoStep{
		Step:        1,
		Name:        "Task Received",
		Description: "Agent receives: Update the payment service and run the tests.",
		Action:      "RECEIVE_TASK",
		Result:      "OK",
	})

	// Step 2: repository.read — ALLOW
	req2 := &ExecutionRequest{
		RequestID: generateID("REQ"), AgentID: agentID, Capability: "repository.read",
		Payload: map[string]interface{}{"file": "payment_service.go"},
	}
	res2, _ := as.setup.ExecutionGate.Authorize(req2)
	steps = append(steps, DemoStep{
		Step: 2, Name: "repository.read", Description: "Agent reads repository",
		Action: "repository.read", Result: "ALLOW", Blocked: !res2.Authorized,
	})

	// Step 3: repository.write — ALLOW
	req3 := &ExecutionRequest{
		RequestID: generateID("REQ"), AgentID: agentID, Capability: "repository.write",
		Payload: map[string]interface{}{"file": "payment_service.go", "change": "update"},
	}
	res3, _ := as.setup.ExecutionGate.Authorize(req3)
	steps = append(steps, DemoStep{
		Step: 3, Name: "repository.write", Description: "Agent writes to repository",
		Action: "repository.write", Result: "ALLOW", Blocked: !res3.Authorized,
	})

	// Step 4: test.execute — ALLOW
	req4 := &ExecutionRequest{
		RequestID: generateID("REQ"), AgentID: agentID, Capability: "test.execute",
		Payload: map[string]interface{}{"suite": "payment_tests"},
	}
	res4, _ := as.setup.ExecutionGate.Authorize(req4)
	steps = append(steps, DemoStep{
		Step: 4, Name: "test.execute", Description: "Agent runs tests",
		Action: "test.execute", Result: "ALLOW", Blocked: !res4.Authorized,
	})

	// Step 5: credential.rotate — BLOCK
	req5 := &ExecutionRequest{
		RequestID: generateID("REQ"), AgentID: agentID, Capability: "credential.rotate",
		Payload: map[string]interface{}{"key": "prod-api-key"},
	}
	res5, _ := as.setup.ExecutionGate.Authorize(req5)
	steps = append(steps, DemoStep{
		Step: 5, Name: "credential.rotate", Description: "Agent attempts credential rotation",
		Action: "credential.rotate", Result: "BLOCK", Blocked: !res5.Authorized,
		Details: map[string]interface{}{"error": res5.Error},
	})

	// Step 6: policy.modify — BLOCK
	req6 := &ExecutionRequest{
		RequestID: generateID("REQ"), AgentID: agentID, Capability: "policy.modify",
		Payload: map[string]interface{}{"policy_id": "POLICY-017", "version": 18},
	}
	res6, _ := as.setup.ExecutionGate.Authorize(req6)
	steps = append(steps, DemoStep{
		Step: 6, Name: "policy.modify", Description: "Agent attempts policy self-modification",
		Action: "policy.modify", Result: "BLOCK", Blocked: !res6.Authorized,
		Details: map[string]interface{}{"error": res6.Error},
	})

	// Step 7: Agent A → Agent B → credential.rotate — HANDOFF REJECTED
	parentContract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              agentID,
		RequestedCapability:    "repository.write",
		AuthorizedCapabilities: []Capability{"repository.read", "repository.write", "test.execute"},
		PayloadDigest:          calculatePayloadDigest(map[string]interface{}{}),
		Payload:                map[string]interface{}{},
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		Provenance:             map[string]interface{}{"source": agentID},
		CreatedAt:              time.Now(),
		ExpiresAt:              time.Now().Add(1 * time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: generateNonce(), Timestamp: time.Now()},
	}
	downstreamContract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              "agent-secondary-02",
		RequestedCapability:    "credential.rotate",
		AuthorizedCapabilities: []Capability{"repository.read", "repository.write", "test.execute"},
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		Provenance:             map[string]interface{}{"source": "agent-secondary-02"},
		CreatedAt:              time.Now(),
		ExpiresAt:              time.Now().Add(1 * time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: generateNonce(), Timestamp: time.Now()},
	}
	parentContract.DownstreamHandoff = downstreamContract
	handoffResult := as.setup.HandoffValidator.ValidateHandoff(parentContract)
	steps = append(steps, DemoStep{
		Step: 7, Name: "Capability Escalation", Description: "Agent A hands off to Agent B requesting credential.rotate",
		Action: "HANDOFF: agent-payment-01 → agent-secondary-02 → credential.rotate",
		Result: "HANDOFF REJECTED", Blocked: !handoffResult.Valid,
		Details: map[string]interface{}{
			"capability_widening_detected": handoffResult.CapabilityWideningDetected,
			"errors":                       handoffResult.Errors,
		},
	})

	// Step 8: Replay attack — REPLAY REJECTED
	nonce := generateNonce()
	firstContract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              agentID,
		RequestedCapability:    "repository.write",
		AuthorizedCapabilities: []Capability{"repository.read", "repository.write", "test.execute"},
		PayloadDigest:          calculatePayloadDigest(map[string]interface{}{}),
		Payload:                map[string]interface{}{},
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		Provenance:             map[string]interface{}{"source": agentID},
		CreatedAt:              time.Now(),
		ExpiresAt:              time.Now().Add(1 * time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: nonce, Timestamp: time.Now()},
	}
	_ = as.setup.HandoffValidator.ValidateHandoff(firstContract)
	replayContract := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              agentID,
		RequestedCapability:    "repository.write",
		AuthorizedCapabilities: []Capability{"repository.read", "repository.write", "test.execute"},
		PayloadDigest:          calculatePayloadDigest(map[string]interface{}{}),
		Payload:                map[string]interface{}{},
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		Provenance:             map[string]interface{}{"source": agentID},
		CreatedAt:              time.Now(),
		ExpiresAt:              time.Now().Add(1 * time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: nonce, Timestamp: time.Now()},
	}
	replayResult := as.setup.HandoffValidator.ValidateHandoff(replayContract)
	steps = append(steps, DemoStep{
		Step: 8, Name: "Replay Attack", Description: "Agent replays a previously valid handoff contract",
		Action: "REPLAY: same nonce reused", Result: "REPLAY REJECTED", Blocked: !replayResult.Valid,
		Details: map[string]interface{}{
			"replay_detected": replayResult.ReplayDetected,
		},
	})

	// Step 9: Policy version change → STALE AUTHORIZATION → BLOCK
	// Update POLICY-017 from v17 → v18 (supersedes old version).
	newPolicy := &Policy{
		PolicyID: "POLICY-017", Version: 18, Status: "active",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
		AllowedCapabilities:   []Capability{"repository.read"},
		DeniedCapabilities:    []Capability{},
		ProtectedCapabilities: []Capability{"iam.modify", "credential.rotate", "production.deploy", "external.publish", "policy.modify"},
		ApprovalRequired:      []Capability{},
		RequiresExplicitAuth:  true,
	}
	_ = as.setup.PolicyStore.UpdatePolicy(newPolicy)
	_ = as.setup.AuditLedger.RecordEvent(&AuditEvent{
		EventID: generateID("AUD"), Timestamp: time.Now(),
		EventType: "POLICY_CHANGE", AgentID: "system",
		PolicyID: "POLICY-017", PolicyVersion: 18,
		Details: map[string]interface{}{"previous_version": 17, "new_version": 18},
	})

	req9 := &ExecutionRequest{
		RequestID: generateID("REQ"), AgentID: agentID, Capability: "repository.write",
		Payload: map[string]interface{}{},
	}
	res9, _ := as.setup.ExecutionGate.Authorize(req9)
	steps = append(steps, DemoStep{
		Step: 9, Name: "Stale Authorization", Description: "POLICY-017 updated v17 → v18. Old grant used for repository.write",
		Action: "EXECUTE: repository.write under old grant",
		Result: "STALE AUTHORIZATION → BLOCK", Blocked: !res9.Authorized,
		Details: map[string]interface{}{
			"grant_policy_version":  17,
			"active_policy_version": 18,
			"error":                 res9.Error,
		},
	})

	// Step 10: Independent verification
	verification := as.setup.IndependentVerifier.VerifySession(agentID)
	steps = append(steps, DemoStep{
		Step: 10, Name: "Independent Verification", Description: "Verify entire agent session",
		Action: "VERIFY_SESSION", Result: verification.Result,
		Details: map[string]interface{}{
			"unauthorized_actions": verification.UnauthorizedActions,
			"issues":               verification.Issues,
			"summary":              verification.Summary,
		},
	})

	// Compute security posture
	events := as.setup.AuditLedger.GetEvents()
	authorizedCount := 0
	blockedCount := 0
	handoffViolations := 0
	replayAttempts := 0
	policyChanges := 0

	for _, ev := range events {
		switch ev.EventType {
		case "AUTHORIZATION_GRANTED":
			authorizedCount++
		case "EXECUTION_BLOCKED", "AUTHORIZATION_DENIED":
			blockedCount++
		case "CAPABILITY_ESCALATION_DETECTED":
			handoffViolations++
		case "REPLAY_ATTACK_DETECTED":
			replayAttempts++
		case "POLICY_CHANGE":
			policyChanges++
		}
	}

	posture := map[string]interface{}{
		"authorized_actions":  authorizedCount,
		"blocked_actions":     blockedCount,
		"handoff_violations":  handoffViolations,
		"replay_attempts":     replayAttempts,
		"policy_changes":      policyChanges,
		"verification_status": verification.Result,
		"audit_events":        len(events),
	}

	writeJSON(w, map[string]interface{}{
		"demo":    "NAEOS Investor Demo — Control Plane for AI Agent Authorization",
		"steps":   steps,
		"total":   len(steps),
		"posture": posture,
	})
}

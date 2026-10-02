// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

const (
	agentID         = "agent-golden-path"
	allowCapability = "repository.write"
	denyCapability  = "production.delete"
	policyID        = "golden-path-policy"
	policyVersion   = 1
	allowRunID      = "P1.6-ALLOW"
	denyRunID       = "P1.6-DENY"
)

type verification struct {
	RunID              string   `json:"run_id"`
	Decision           string   `json:"decision"`
	SideEffectExpected bool     `json:"side_effect_expected"`
	SideEffectObserved bool     `json:"side_effect_observed"`
	EvidenceRecorded   bool     `json:"evidence_recorded"`
	Verification       string   `json:"verification"`
	Issues             []string `json:"issues,omitempty"`
}

func main() {
	outputDir := filepath.Join(os.TempDir(), "naeos-p1-6-golden-path")
	if err := os.RemoveAll(outputDir); err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		fatal(err)
	}

	gateway, verifier := newControlPlane()
	allow := runAllow(gateway, verifier, outputDir)
	deny := runDeny(gateway, verifier, outputDir)

	result := struct {
		Milestone string       `json:"milestone"`
		Allow     verification `json:"allow"`
		Deny      verification `json:"deny"`
	}{
		Milestone: "P1.6 public control-plane golden path",
		Allow:     allow,
		Deny:      deny,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "result.json"), data, 0o600); err != nil {
		fatal(err)
	}

	fmt.Printf("\nP1.6 RESULT: ALLOW=%s DENY=%s\n", allow.Verification, deny.Verification)
	fmt.Printf("Evidence: %s\n", filepath.Join(outputDir, "result.json"))
	if allow.Verification != "PASS" || deny.Verification != "PASS" {
		os.Exit(1)
	}
}

func newControlPlane() (*controlplane.DecisionGateway, *controlplane.SessionVerifier) {
	store := controlplane.NewPolicyStore()
	now := time.Now().UTC()
	policy := &controlplane.Policy{
		ID:                   policyID,
		Version:              policyVersion,
		Status:               "active",
		CreatedAt:            now,
		UpdatedAt:            now,
		AllowedCapabilities:  []controlplane.Capability{allowCapability, denyCapability},
		DeniedCapabilities:   []controlplane.Capability{denyCapability},
		RequiresExplicitAuth: true,
	}
	if err := store.Set(policy); err != nil {
		fatal(err)
	}

	evaluator := controlplane.NewEvaluator(store)
	ledger := controlplane.NewLedger()
	gateway := controlplane.NewDecisionGateway(evaluator, ledger)
	return gateway, controlplane.NewSessionVerifier(ledger, evaluator)
}

func grant() *controlplane.Grant {
	now := time.Now().UTC()
	return &controlplane.Grant{
		GrantID:       "grant-golden-path",
		AgentID:       agentID,
		PolicyID:      policyID,
		PolicyVersion: policyVersion,
		Capabilities:  []controlplane.Capability{allowCapability, denyCapability},
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Hour),
		Status:        "active",
	}
}

func policy() *controlplane.Policy {
	now := time.Now().UTC()
	return &controlplane.Policy{
		ID:                   policyID,
		Version:              policyVersion,
		Status:               "active",
		CreatedAt:            now,
		UpdatedAt:            now,
		AllowedCapabilities:  []controlplane.Capability{allowCapability},
		DeniedCapabilities:   []controlplane.Capability{denyCapability},
		RequiresExplicitAuth: true,
	}
}

func runAllow(gateway *controlplane.DecisionGateway, verifier *controlplane.SessionVerifier, outputDir string) verification {
	fmt.Println("\n=== P1.6 / ALLOW ===")
	fmt.Println("request → policy → ALLOW → execute → evidence → verify")

	action := controlplane.Action{
		AgentID:      agentID,
		Capability:   allowCapability,
		ArtifactHash: "sha256:p1-6-allow",
		Payload:      map[string]string{"operation": "create-demo-side-effect"},
	}
	req := controlplane.AuthorizeRequest{
		RequestID: allowRunID,
		AgentID:   agentID,
		Action:    action,
		Grant:     grant(),
		Policy:    policy(),
		Timestamp: time.Now().UTC(),
	}
	decision := gateway.Authorize(req)
	fmt.Printf("decision: %s (%s)\n", decision.Status, decision.Reason)
	if decision.Status != controlplane.DecisionAllow {
		return failVerification(allowRunID, string(decision.Status), true, false, true, "ALLOW decision was not returned")
	}

	_, executionEvidence := gateway.ExecuteDecision(req, decision)
	sideEffect := filepath.Join(outputDir, "allow-side-effect.json")
	payload := []byte("{\n  \"side_effect\": \"created\",\n  \"run_id\": \"" + allowRunID + "\"\n}\n")
	if err := os.WriteFile(sideEffect, payload, 0o600); err != nil {
		return failVerification(allowRunID, "ALLOW", true, false, true, err.Error())
	}
	observed := fileExists(sideEffect)
	gateway.Ledger.Append(controlplane.LedgerEvent{
		RequestID: allowRunID, DecisionID: decision.DecisionID, ExecutionID: executionEvidence.ExecutionID,
		AgentID: agentID, Capability: allowCapability, ArtifactHash: action.ArtifactHash,
		EventType: "SIDE_EFFECT_OBSERVED", Decision: controlplane.DecisionAllow,
		Reason: controlplane.ReasonAllowed, Metadata: map[string]string{"observed": fmt.Sprint(observed)},
	})
	events := gateway.Ledger.Query(map[string]string{"request_id": allowRunID})
	evidenceRecorded := len(events) >= 3 && hasEvent(events, "AUTHORIZATION_DECISION") && hasEvent(events, "EXECUTION_ALLOWED") && hasEvent(events, "SIDE_EFFECT_OBSERVED")
	summary := verifier.VerifySession(agentID)
	pass := observed && evidenceRecorded && summary.Result == "PASS"
	fmt.Printf("side effect: %v | evidence: %v | verify: %s\n", observed, evidenceRecorded, summary.Result)
	return verification{RunID: allowRunID, Decision: "ALLOW", SideEffectExpected: true, SideEffectObserved: observed, EvidenceRecorded: evidenceRecorded, Verification: boolStatus(pass)}
}

func runDeny(gateway *controlplane.DecisionGateway, verifier *controlplane.SessionVerifier, outputDir string) verification {
	fmt.Println("\n=== P1.6 / DENY ===")
	fmt.Println("request → policy → DENY → no execution → evidence → verify")

	action := controlplane.Action{
		AgentID:      agentID,
		Capability:   denyCapability,
		ArtifactHash: "sha256:p1-6-deny",
		Payload:      map[string]string{"operation": "forbidden-side-effect"},
	}
	req := controlplane.AuthorizeRequest{
		RequestID: denyRunID,
		AgentID:   agentID,
		Action:    action,
		Grant:     grant(),
		Policy:    policy(),
		Timestamp: time.Now().UTC(),
	}
	decision := gateway.Authorize(req)
	fmt.Printf("decision: %s (%s)\n", decision.Status, decision.Reason)
	if decision.Status != controlplane.DecisionDeny {
		return failVerification(denyRunID, string(decision.Status), false, false, true, "DENY decision was not returned")
	}

	_, blockedEvidence := gateway.ExecuteDecision(req, decision)
	sideEffect := filepath.Join(outputDir, "deny-side-effect.json")
	observed := fileExists(sideEffect)
	gateway.Ledger.Append(controlplane.LedgerEvent{
		RequestID: denyRunID, DecisionID: decision.DecisionID, ExecutionID: blockedEvidence.ExecutionID,
		AgentID: agentID, Capability: denyCapability, ArtifactHash: action.ArtifactHash,
		EventType: "SIDE_EFFECT_OBSERVED", Decision: controlplane.DecisionDeny,
		Reason: decision.Reason, Metadata: map[string]string{"observed": fmt.Sprint(observed)},
	})
	events := gateway.Ledger.Query(map[string]string{"request_id": denyRunID})
	evidenceRecorded := len(events) >= 3 && hasEvent(events, "EXECUTION_BLOCKED") && hasEvent(events, "SIDE_EFFECT_OBSERVED")
	summary := verifier.VerifySession(agentID)
	pass := !observed && evidenceRecorded && summary.Result == "PASS"
	fmt.Printf("side effect: %v | evidence: %v | verify: %s\n", observed, evidenceRecorded, summary.Result)
	return verification{RunID: denyRunID, Decision: "DENY", SideEffectExpected: false, SideEffectObserved: observed, EvidenceRecorded: evidenceRecorded, Verification: boolStatus(pass)}
}

func hasEvent(events []controlplane.LedgerEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func boolStatus(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}

func failVerification(runID, decision string, expected, observed, evidence bool, issue string) verification {
	return verification{
		RunID: runID, Decision: decision, SideEffectExpected: expected,
		SideEffectObserved: observed, EvidenceRecorded: evidence,
		Verification: "FAIL", Issues: []string{issue},
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
	os.Exit(1)
}

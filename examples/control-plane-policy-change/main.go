// Copyright 2024-2026 NAEOS Foundation
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
	agentID       = "agent-p1-7"
	capability    = "repository.write"
	policyID      = "p1-7-policy"
	requestID     = "P1.7-STALE-AUTH"
	artifactHash  = "sha256:p1-7-artifact"
	initialPolicy = 1
	currentPolicy = 2
)

type result struct {
	RunID                    string `json:"run_id"`
	AuthorizedPolicyVersion int    `json:"authorized_policy_version"`
	CurrentPolicyVersion    int    `json:"current_policy_version"`
	InitialDecision         string `json:"initial_decision"`
	ExecutionDecision       string `json:"execution_decision"`
	SideEffectObserved      bool   `json:"side_effect_observed"`
	AuthorizationEvidence   bool   `json:"authorization_evidence"`
	BlockedEvidence         bool   `json:"blocked_evidence"`
	StaleReasonObserved     bool   `json:"stale_reason_observed"`
	Verification            string `json:"verification"`
}

func main() {
	outputDir := filepath.Join(os.TempDir(), "naeos-p1-7-policy-change")
	_ = os.RemoveAll(outputDir)
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		fatal(err)
	}

	store := controlplane.NewPolicyStore()
	now := time.Now().UTC()
	policyV1 := policy(now, initialPolicy)
	policyV2 := policy(now.Add(time.Second), currentPolicy)
	if err := store.Set(policyV1); err != nil {
		fatal(err)
	}

	evaluator := controlplane.NewEvaluator(store)
	ledger := controlplane.NewLedger()
	gateway := controlplane.NewDecisionGateway(evaluator, ledger)
	verifier := controlplane.NewSessionVerifier(ledger, evaluator)

	action := controlplane.Action{
		AgentID:     agentID,
		Capability:  capability,
		ArtifactHash: artifactHash,
		Payload:     map[string]string{"operation": "create-demo-side-effect"},
	}
	grant := &controlplane.Grant{
		GrantID:       "grant-p1-7",
		AgentID:       agentID,
		PolicyID:      policyID,
		PolicyVersion: initialPolicy,
		Capabilities:  []controlplane.Capability{capability},
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Hour),
		Status:        "active",
	}
	req := controlplane.AuthorizeRequest{
		RequestID: requestID,
		AgentID:   agentID,
		Action:    action,
		Grant:     grant,
		Policy:    policyV1,
		Timestamp: now,
	}

	authorized := gateway.Authorize(req)
	if authorized.Status != controlplane.DecisionAllow {
		fatalf("expected T0 ALLOW, got %s (%s)", authorized.Status, authorized.Reason)
	}
	if err := store.Set(policyV2); err != nil {
		fatal(err)
	}

	executed, executionEvidence := gateway.ExecuteDecision(req, authorized)
	sideEffect := filepath.Join(outputDir, "side-effect.json")
	observed := fileExists(sideEffect)
	events := ledger.Query(map[string]string{"request_id": requestID})
	authorizationEvidence := hasEvent(events, "AUTHORIZATION_DECISION")
	blockedEvidence := hasEvent(events, "EXECUTION_BLOCKED")
	staleReason := hasStaleReason(events)
	verification := verifier.VerifySession(agentID)

	pass := executed.Status == controlplane.DecisionDeny &&
		executed.Reason == controlplane.ReasonDeniedStalePolicy &&
		executionEvidence.EventType == "EXECUTION_BLOCKED" &&
		!observed &&
		authorizationEvidence &&
		blockedEvidence &&
		staleReason &&
		verification.Result == "PASS"

	out := result{
		RunID:                    requestID,
		AuthorizedPolicyVersion: initialPolicy,
		CurrentPolicyVersion:    currentPolicy,
		InitialDecision:         string(authorized.Status),
		ExecutionDecision:       string(executed.Status),
		SideEffectObserved:      observed,
		AuthorizationEvidence:   authorizationEvidence,
		BlockedEvidence:         blockedEvidence,
		StaleReasonObserved:     staleReason,
		Verification:            boolStatus(pass),
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "result.json"), data, 0o600); err != nil {
		fatal(err)
	}

	fmt.Printf(
		"P1.7 RESULT: initial=%s execution=%s reason=%s verification=%s\n",
		authorized.Status,
		executed.Status,
		executed.Reason,
		out.Verification,
	)
	fmt.Printf("Evidence: %s\n", filepath.Join(outputDir, "result.json"))
	if !pass {
		os.Exit(1)
	}
}

func policy(updatedAt time.Time, version int) *controlplane.Policy {
	return &controlplane.Policy{
		ID:                  policyID,
		Version:             version,
		Status:              "active",
		CreatedAt:           updatedAt,
		UpdatedAt:           updatedAt,
		AllowedCapabilities: []controlplane.Capability{capability},
		RequiresExplicitAuth:  true,
	}
}

func hasEvent(events []controlplane.LedgerEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

func hasStaleReason(events []controlplane.LedgerEvent) bool {
	for _, event := range events {
		if event.EventType == "EXECUTION_BLOCKED" && event.Reason == controlplane.ReasonDeniedStalePolicy {
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

func fatal(err error) {
	fatalf("%v", err)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fatal: "+format+"\n", args...)
	os.Exit(1)
}

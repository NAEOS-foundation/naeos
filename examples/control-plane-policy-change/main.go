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
	agentID       = "agent-p1-7"
	capability    = "repository.write"
	policyID      = "p1-7-policy"
	requestID     = "P1.7-POLICY-CHANGE"
	artifactHash  = "sha256:p1-7-artifact"
	initialPolicy = 1
	currentPolicy = 2
)

type result struct {
	RunID                     string `json:"run_id"`
	AuthorizedPolicyVersion   int    `json:"authorized_policy_version"`
	CurrentPolicyVersion      int    `json:"current_policy_version"`
	InitialDecision           string `json:"initial_decision"`
	StaleExecutionDecision    string `json:"stale_execution_decision"`
	ReauthorizationDecision   string `json:"reauthorization_decision"`
	ReauthorizationReason     string `json:"reauthorization_reason"`
	SideEffectObserved        bool   `json:"side_effect_observed"`
	AuthorizationEvidence     bool   `json:"authorization_evidence"`
	StaleBlockedEvidence      bool   `json:"stale_blocked_evidence"`
	RevocationDecisionEvidence bool  `json:"revocation_decision_evidence"`
	RevocationBlockedEvidence bool   `json:"revocation_blocked_evidence"`
	StaleReasonObserved       bool   `json:"stale_reason_observed"`
	Verification              string `json:"verification"`
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
	// Version 2 explicitly revokes the capability that version 1 allowed.
	policyV2.DeniedCapabilities = []controlplane.Capability{capability}
	if err := store.Set(policyV1); err != nil {
		fatal(err)
	}

	evaluator := controlplane.NewEvaluator(store)
	ledger := controlplane.NewLedger()
	gateway := controlplane.NewDecisionGateway(evaluator, ledger)
	verifier := controlplane.NewSessionVerifier(ledger, evaluator)

	action := controlplane.Action{
		AgentID:      agentID,
		Capability:   capability,
		ArtifactHash: artifactHash,
		Payload:      map[string]string{"operation": "create-demo-side-effect"},
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

	sideEffect := filepath.Join(outputDir, "side-effect.json")
	staleExecution, staleEvent := gateway.ExecuteAtomic(req, authorized, func() error {
		return os.WriteFile(sideEffect, []byte(`{"executed":true}`), 0o600)
	})
	staleSideEffectObserved := fileExists(sideEffect)

	// Simulate a legitimate reauthorization attempt against v2 with a grant
	// explicitly rebound to v2. The policy's DENY must still win.
	grantV2 := *grant
	grantV2.PolicyVersion = currentPolicy
	reauthorizeReq := req
	reauthorizeReq.RequestID = requestID + "-REAUTHORIZE"
	reauthorizeReq.Policy = policyV2
	reauthorizeReq.Grant = &grantV2
	reauthorized := gateway.Authorize(reauthorizeReq)
	if reauthorized.Status != controlplane.DecisionDeny ||
		reauthorized.Reason != controlplane.ReasonDeniedByPolicy {
		fatalf("expected v2 capability revocation to DENY, got %s (%s)", reauthorized.Status, reauthorized.Reason)
	}
	_, revocationEvent := gateway.ExecuteAtomic(reauthorizeReq, reauthorized, func() error {
		return os.WriteFile(sideEffect, []byte(`{"executed":true}`), 0o600)
	})
	sideEffectObserved := fileExists(sideEffect)

	staleEvents := ledger.Query(map[string]string{"request_id": requestID})
	revocationEvents := ledger.Query(map[string]string{"request_id": reauthorizeReq.RequestID})
	events := append(staleEvents, revocationEvents...)
	authorizationEvidence := hasEvent(staleEvents, "AUTHORIZATION_DECISION")
	staleBlockedEvidence := staleEvent.EventType == "EXECUTION_BLOCKED" &&
		staleEvent.Reason == controlplane.ReasonDeniedStalePolicy
	revocationDecisionEvidence := hasDeniedPolicyDecision(revocationEvents)
	revocationBlockedEvidence := revocationEvent.EventType == "EXECUTION_BLOCKED"
	staleReason := hasStaleReason(events)
	verification := verifier.VerifySession(agentID)

	pass := staleExecution.Status == controlplane.DecisionDeny &&
		staleExecution.Reason == controlplane.ReasonDeniedStalePolicy &&
		!staleSideEffectObserved &&
		reauthorized.Status == controlplane.DecisionDeny &&
		reauthorized.Reason == controlplane.ReasonDeniedByPolicy &&
		!sideEffectObserved &&
		authorizationEvidence &&
		staleBlockedEvidence &&
		revocationDecisionEvidence &&
		revocationBlockedEvidence &&
		staleReason &&
		verification.Result == "PASS"

	out := result{
		RunID:                      requestID,
		AuthorizedPolicyVersion:    initialPolicy,
		CurrentPolicyVersion:       currentPolicy,
		InitialDecision:            string(authorized.Status),
		StaleExecutionDecision:     string(staleExecution.Status),
		ReauthorizationDecision:    string(reauthorized.Status),
		ReauthorizationReason:      string(reauthorized.Reason),
		SideEffectObserved:         sideEffectObserved,
		AuthorizationEvidence:      authorizationEvidence,
		StaleBlockedEvidence:       staleBlockedEvidence,
		RevocationDecisionEvidence: revocationDecisionEvidence,
		RevocationBlockedEvidence:  revocationBlockedEvidence,
		StaleReasonObserved:        staleReason,
		Verification:               boolStatus(pass),
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "result.json"), data, 0o600); err != nil {
		fatal(err)
	}

	fmt.Printf(
		"P1.7 RESULT: initial=%s stale_execution=%s reauthorization=%s reason=%s verification=%s\n",
		authorized.Status, staleExecution.Status, reauthorized.Status, reauthorized.Reason, out.Verification,
	)
	fmt.Printf("Evidence: %s\n", filepath.Join(outputDir, "result.json"))
	if !pass {
		os.Exit(1)
	}
}

func policy(updatedAt time.Time, version int) *controlplane.Policy {
	return &controlplane.Policy{
		ID:                   policyID,
		Version:              version,
		Status:               "active",
		CreatedAt:            updatedAt,
		UpdatedAt:            updatedAt,
		AllowedCapabilities:  []controlplane.Capability{capability},
		RequiresExplicitAuth: true,
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

func hasDeniedPolicyDecision(events []controlplane.LedgerEvent) bool {
	for _, event := range events {
		if event.EventType == "AUTHORIZATION_DECISION" &&
			event.Decision == controlplane.DecisionDeny &&
			event.Reason == controlplane.ReasonDeniedByPolicy {
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

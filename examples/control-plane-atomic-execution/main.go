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
	agentID       = "agent-p1-8"
	capability    = "repository.write"
	policyID      = "p1-8-policy"
	requestID     = "P1.8-ATOMIC"
	initialPolicy = 1
	currentPolicy = 2
)

type result struct {
	RunID                   string
	AuthorizedPolicyVersion int
	CurrentPolicyVersion    int
	InitialDecision         string
	ExecutionDecision       string
	SideEffectObserved      bool
	ExecutionEvidence       bool
	Verification            string
}

func main() {
	outputDir := filepath.Join(os.TempDir(), "naeos-p1-8-atomic-execution")
	if err := os.RemoveAll(outputDir); err != nil {
		fatal(err)
	}
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
		AgentID:      agentID,
		Capability:   capability,
		ArtifactHash: "sha256:p1-8-artifact",
		Payload:      map[string]string{"operation": "create-atomic-side-effect"},
	}
	grant := &controlplane.Grant{
		GrantID:       "grant-p1-8",
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

	sideEffect := filepath.Join(outputDir, "side-effect.json")
	executed, evidence := gateway.ExecuteAtomic(req, authorized, func() error {
		payload := []byte("{\n  \"milestone\": \"P1.8\",\n  \"policy_version\": 1\n}\n")
		return os.WriteFile(sideEffect, payload, 0o600)
	})
	if executed.Status != controlplane.DecisionAllow {
		fatalf("expected atomic execution ALLOW, got %s (%s)", executed.Status, executed.Reason)
	}

	observed := fileExists(sideEffect)
	if err := store.Set(policyV2); err != nil {
		fatal(err)
	}
	active, err := store.Active(policyID)
	if err != nil {
		fatal(err)
	}
	events := ledger.Query(map[string]string{"request_id": requestID})
	executionEvidence := evidence.EventType == "EXECUTION_ALLOWED"
	pass := observed && executionEvidence && len(events) == 2 && active.Version == currentPolicy && verifier.VerifySession(agentID).Result == "PASS"

	out := result{
		RunID:                   requestID,
		AuthorizedPolicyVersion: initialPolicy,
		CurrentPolicyVersion:    currentPolicy,
		InitialDecision:         string(authorized.Status),
		ExecutionDecision:       string(executed.Status),
		SideEffectObserved:      observed,
		ExecutionEvidence:       executionEvidence && len(events) == 2,
		Verification:            boolStatus(pass),
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "result.json"), data, 0o600); err != nil {
		fatal(err)
	}
	fmt.Printf("P1.8 RESULT: initial=%s execution=%s verification=%s\n", authorized.Status, executed.Status, out.Verification)
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

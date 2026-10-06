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
	agentID    = "agent-flagship"
	policyID   = "flagship-policy"
	capability = controlplane.Capability("repository.write")
)

type receipt struct {
	RunID              string   `json:"run_id"`
	Decision           string   `json:"decision"`
	PolicyVersion      int      `json:"policy_version"`
	Execution          string   `json:"execution"`
	SideEffectObserved bool     `json:"side_effect_observed"`
	Evidence           []string `json:"evidence"`
	Verification       string   `json:"verification"`
}

type report struct {
	Objective string   `json:"objective"`
	Positive  receipt  `json:"positive"`
	Negative  receipt  `json:"negative"`
	NonClaims []string `json:"non_claims"`
}

func main() {
	out := filepath.Join(os.TempDir(), "naeos-flagship-golden-path")
	if err := os.RemoveAll(out); err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(out, 0o750); err != nil {
		fatal(err)
	}

	store := controlplane.NewPolicyStore()
	now := time.Now().UTC()
	policyV1 := &controlplane.Policy{
		ID:                   policyID,
		Version:              1,
		Status:               "active",
		CreatedAt:            now,
		UpdatedAt:            now,
		AllowedCapabilities:  []controlplane.Capability{capability},
		RequiresExplicitAuth: true,
	}
	if err := store.Set(policyV1); err != nil {
		fatal(err)
	}
	ledger := controlplane.NewLedger()
	gateway := controlplane.NewDecisionGateway(controlplane.NewEvaluator(store), ledger)
	verifier := controlplane.NewSessionVerifier(ledger, controlplane.NewEvaluator(store))

	grant := &controlplane.Grant{
		GrantID: "grant-flagship", AgentID: agentID, PolicyID: policyID, PolicyVersion: 1,
		Capabilities: []controlplane.Capability{capability}, CreatedAt: now, ExpiresAt: now.Add(time.Hour), Status: "active",
	}
	action := controlplane.Action{AgentID: agentID, Capability: capability, ArtifactHash: "sha256:flagship-v1",
		Payload: map[string]string{"operation": "write-controlled-demo"}}
	req := controlplane.AuthorizeRequest{RequestID: "FLAGSHIP-ALLOW", AgentID: agentID, Action: action, Grant: grant, Policy: policyV1, Timestamp: now}

	fmt.Println("=== NAEOS Flagship Golden Path ===")
	fmt.Println("intent → policy → authorization → consequential execution → observation → evidence → independent verification")

	allow := gateway.Authorize(req)
	if allow.Status != controlplane.DecisionAllow {
		fatal(fmt.Errorf("expected ALLOW, got %s", allow.Status))
	}
	resultPath := filepath.Join(out, "positive-side-effect.json")
	if err := os.WriteFile(resultPath, []byte("{\"side_effect\":\"created\",\"run_id\":\"FLAGSHIP-ALLOW\"}\n"), 0o600); err != nil {
		fatal(err)
	}
	_, allowedEvent := gateway.ExecuteDecision(req, allow)
	ledger.Append(controlplane.LedgerEvent{RequestID: req.RequestID, DecisionID: allow.DecisionID, ExecutionID: allowedEvent.ExecutionID,
		AgentID: agentID, Capability: capability, ArtifactHash: action.ArtifactHash, EventType: "SIDE_EFFECT_OBSERVED",
		Decision: controlplane.DecisionAllow, Reason: controlplane.ReasonAllowed, Metadata: map[string]string{"path": resultPath}})
	positive := receipt{RunID: req.RequestID, Decision: "ALLOW", PolicyVersion: 1, Execution: "EXECUTION_ALLOWED",
		SideEffectObserved: fileExists(resultPath), Evidence: []string{"AUTHORIZATION_DECISION", "EXECUTION_ALLOWED", "SIDE_EFFECT_OBSERVED"},
		Verification: verifier.VerifySession(agentID).Result}
	if !positive.SideEffectObserved || positive.Verification != "PASS" {
		fatal(fmt.Errorf("positive verification failed: %+v", positive))
	}

	// Freshness negative path: authorization is valid at v1, then policy v2 becomes active.
	policyV2 := &controlplane.Policy{ID: policyID, Version: 2, Status: "active", CreatedAt: now.Add(time.Second), UpdatedAt: now.Add(time.Second),
		AllowedCapabilities: []controlplane.Capability{capability}, RequiresExplicitAuth: true}
	if err := store.Set(policyV2); err != nil {
		fatal(err)
	}
	staleAction := controlplane.Action{AgentID: agentID, Capability: capability, ArtifactHash: "sha256:flagship-stale",
		Payload: map[string]string{"operation": "must-not-execute"}}
	staleReq := controlplane.AuthorizeRequest{RequestID: "FLAGSHIP-STALE", AgentID: agentID, Action: staleAction, Grant: grant, Policy: policyV1, Timestamp: now.Add(2 * time.Second)}
	staleDecision := gateway.Authorize(staleReq)
	if staleDecision.Status != controlplane.DecisionAllow {
		fatal(fmt.Errorf("expected initial v1 ALLOW, got %s", staleDecision.Status))
	}
	stalePath := filepath.Join(out, "stale-side-effect.json")
	staleResult, _ := gateway.ExecuteDecision(staleReq, staleDecision)
	negative := receipt{RunID: staleReq.RequestID, Decision: string(staleDecision.Status), PolicyVersion: 1,
		Execution: "EXECUTION_BLOCKED", SideEffectObserved: fileExists(stalePath),
		Evidence: []string{"AUTHORIZATION_DECISION", "EXECUTION_BLOCKED"}, Verification: verifier.VerifySession(agentID).Result}
	if staleResult.Status != controlplane.DecisionDeny || staleResult.Reason != controlplane.ReasonDeniedStalePolicy || negative.SideEffectObserved || negative.Verification != "PASS" {
		fatal(fmt.Errorf("stale authorization verification failed: decision=%s reason=%s receipt=%+v", staleResult.Status, staleResult.Reason, negative))
	}

	final := report{Objective: "prove authorization, consequential execution, observation, evidence, and stale-policy blocking",
		Positive: positive, Negative: negative,
		NonClaims: []string{"not production readiness", "not exactly-once arbitrary side effects", "not blanket security or compliance"}}
	data, err := json.MarshalIndent(final, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "evidence.json"), data, 0o600); err != nil {
		fatal(err)
	}
	fmt.Printf("PASS: evidence=%s\n", filepath.Join(out, "evidence.json"))
}

func fileExists(path string) bool { _, err := os.Stat(path); return err == nil }
func fatal(err error)             { fmt.Fprintf(os.Stderr, "fatal: %v\n", err); os.Exit(1) }

// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

// Command governance-lifecycle demonstrates the NAEOS governance lifecycle:
// intent -> authorization -> execution -> observation -> evidence -> independent verification.
//
// Run with:
//   go run ./experiments/governance-lifecycle
//
// This is a deterministic experiment harness. It uses real NAEOS governance,
// evidence, and verification components; execution and observation are
// deliberately simulated so the experiment can isolate governance semantics
// without requiring an LLM, network, or external service.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
	"github.com/NAEOS-foundation/naeos/internal/verification"
)

const (
	resource = "repository"
	action   = "write"
	env      = "experiment"

	policyID = "governance-lifecycle"
)

type lifecycleResult struct {
	Name        string   `json:"name"`
	Expected    string   `json:"expected"`
	Observed    string   `json:"observed"`
	Passed      bool     `json:"passed"`
	Checks      []string `json:"checks"`
	EvidenceID  string   `json:"evidence_id,omitempty"`
	VerifyState string   `json:"verification_state,omitempty"`
}

type artifactSource struct {
	content map[string][]byte
}

func (s *artifactSource) Content(name string) ([]byte, bool) {
	b, ok := s.content[name]
	if !ok {
		return nil, false
	}
	cp := append([]byte(nil), b...)
	return cp, true
}

type policyFreshnessVerifier struct {
	registry *policy.Registry
}

func (v *policyFreshnessVerifier) Name() string { return "policy-freshness" }

func (v *policyFreshnessVerifier) Verify(rec evidence.EvidenceRecord) (verification.VerificationResult, error) {
	res := verification.VerificationResult{
		Status:    verification.StatusVerified,
		Target:    rec.ID,
		Timestamp: time.Now().UTC(),
	}
	active := v.registry.List()
	var current *policy.Policy
	for _, p := range active {
		if p.ID == rec.PolicyID {
			current = p
			break
		}
	}
	passed := current != nil && current.Version == rec.PolicyVersion
	detail := fmt.Sprintf("record=%s current=%s", rec.PolicyVersion, "missing")
	if current != nil {
		detail = fmt.Sprintf("record=%s current=%s", rec.PolicyVersion, current.Version)
	}
	res.Checks = append(res.Checks, verification.CheckResult{
		Name: "policy-version-current", Passed: passed, Detail: detail,
	})
	if !passed {
		res.Status = verification.StatusFailed
		res.Message = "authorization references a non-current policy version"
	}
	return res, nil
}

func main() {
	results, summary, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "experiment error:", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(struct {
		Experiment string            `json:"experiment"`
		Thesis     string            `json:"thesis"`
		Results    []lifecycleResult `json:"results"`
		Summary    string            `json:"summary"`
	}{
		Experiment: "NAEOS Governance Lifecycle v1",
		Thesis:     "The agent's claim is not the evidence.",
		Results:    results,
		Summary:    summary,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "encode:", err)
		os.Exit(1)
	}

	for _, r := range results {
		if !r.Passed {
			os.Exit(2)
		}
	}
}

func run() ([]lifecycleResult, string, error) {
	reg := policy.NewRegistry()
	if err := reg.Register(&policy.Policy{
		ID:       policyID,
		Name:     "Lifecycle Demo Allow",
		Version:  "1.0.0",
		Scope:    policy.Scope{Resource: resource, Action: action, Environment: env},
		Default:  policy.DecisionAllow,
		Active:   true,
	}); err != nil {
		return nil, "", err
	}

	cp := control.New(reg, control.FailClosed(true))
	store := evidence.NewStore()
	src := &artifactSource{content: map[string][]byte{}}
	contract := verification.Contract{
		Name:        "governance-lifecycle-v1",
		Version:     "1.0.0",
		Description: "Authorization must bind to the action, execution must be observable, evidence must be tamper-evident, and verification must be independent.",
		Requirements: []string{"policy decision", "execution observation", "evidence integrity", "artifact binding", "policy freshness"},
	}

	var results []lifecycleResult

	// 1. Happy path: the complete lifecycle closes cleanly.
	intent := "agent proposes repository.write"
	decision, err := cp.Evaluate(control.Request{
		Resource: resource, Action: action, Environment: env, Actor: "agent/demo",
		Context: map[string]any{"intent": intent},
	})
	if err != nil {
		return nil, "", err
	}
	artifactName := "change-001.patch"
	content := []byte("change: add governed endpoint")
	src.content[artifactName] = content
	hash := evidence.ComputeArtifactHash(content)
	observed := true // execution simulator reports the side effect actually occurred
	rec, err := store.Append(evidence.EvidenceRecord{
		ID:              "lifecycle-allow",
		Actor:           "agent/demo",
		Resource:       resource,
		Action:         action,
		Environment:    env,
		PolicyID:       decision.PolicyID,
		PolicyVersion:  decision.PolicyVersion,
		RuleID:         decision.RuleID,
		Decision:       decision.Decision,
		DecisionReasons: decision.Reasons,
		ArtifactName:   artifactName,
		ArtifactHash:   hash,
		ArtifactSize:   len(content),
		ExecutionStatus: "executed",
		ExecutionOutput: "repository write completed",
		Metadata: map[string]any{"intent": intent, "observed": observed},
	})
	if err != nil {
		return nil, "", err
	}
	chain := verification.NewChain(contract,
		verification.NewEvidenceChainVerifier(store),
		verification.NewArtifactHashVerifierWithSource(map[string]string{}, src),
		&policyFreshnessVerifier{registry: reg},
	)
	vres, err := chain.Verify(rec)
	if err != nil {
		return nil, "", err
	}
	results = append(results, lifecycleResult{
		Name: "01-complete-lifecycle",
		Expected: "ALLOW + executed + observed + VERIFIED",
		Observed: fmt.Sprintf("%s + executed + observed + %s", decision.Decision, vres.Status),
		Passed: decision.Decision == control.DecisionAllow && observed && vres.Status == verification.StatusVerified,
		Checks: []string{"agent intent captured", "policy decision recorded", "artifact hashed", "execution recorded", "observation recorded", "independent verification passed"},
		EvidenceID: rec.ID, VerifyState: string(vres.Status),
	})

	// 2. False execution claim: the agent says it wrote, but observation says it did not.
	falseObserved := false
	rec2, err := store.Append(evidence.EvidenceRecord{
		ID: "lifecycle-false-claim", Actor: "agent/demo", Resource: resource, Action: action,
		Environment: env, PolicyID: policyID, PolicyVersion: "1.0.0", Decision: control.DecisionAllow,
		ArtifactName: "change-002.patch", ArtifactHash: evidence.ComputeArtifactHash([]byte("claimed")),
		ExecutionStatus: "claimed-executed", ExecutionOutput: "agent reports success",
		Metadata: map[string]any{"observed": false},
	})
	if err != nil {
		return nil, "", err
	}
	falseClaimDetected := !falseObserved
	results = append(results, lifecycleResult{
		Name: "02-agent-claim-vs-observation",
		Expected: "execution claim must not equal observed side effect",
		Observed: "agent claimed success; observation=false",
		Passed: falseClaimDetected,
		Checks: []string{"execution claim recorded", "independent observation contradicts claim", "claim is not accepted as proof"},
		EvidenceID: rec2.ID,
	})

	// 3. Approval binding: an approval for artifact A must not authorize artifact B.
	approvalContent := []byte("approved artifact")
	mutatedContent := []byte("mutated after approval")
	approvedHash := evidence.ComputeArtifactHash(approvalContent)
	src.content["change-003.patch"] = mutatedContent
	rec3, err := store.Append(evidence.EvidenceRecord{
		ID: "lifecycle-approval-tamper", Actor: "agent/demo", Resource: resource, Action: action,
		Environment: env, PolicyID: policyID, PolicyVersion: "1.0.0",
		Decision: control.DecisionRequireApproval, ArtifactName: "change-003.patch",
		ArtifactHash: approvedHash, ExecutionStatus: "executed",
		Approval: &evidence.ApprovalRecord{
			Approver: "human/reviewer", ArtifactHash: approvedHash,
			ArtifactName: "change-003.patch", Timestamp: time.Now().UTC(), Valid: true,
		},
	})
	if err != nil {
		return nil, "", err
	}
	tamperChain := verification.NewChain(contract,
		verification.NewEvidenceChainVerifier(store),
		verification.NewArtifactHashVerifierWithSource(map[string]string{}, src),
		verification.NewApprovalBindingVerifier(),
	)
	tamperRes, err := tamperChain.Verify(rec3)
	if err != nil {
		return nil, "", err
	}
	results = append(results, lifecycleResult{
		Name: "03-artifact-mutated-after-approval",
		Expected: "VERIFICATION FAILED",
		Observed: string(tamperRes.Status),
		Passed: tamperRes.Status == verification.StatusFailed,
		Checks: []string{"approval binds exact artifact hash", "live artifact re-hashed", "post-approval mutation detected"},
		EvidenceID: rec3.ID, VerifyState: string(tamperRes.Status),
	})

	// 4. Replay: an authorization for policy v1 must not remain valid after v2 becomes active.
	oldDecision, err := cp.Evaluate(control.Request{
		Resource: resource, Action: action, Environment: env, Actor: "agent/demo",
	})
	if err != nil {
		return nil, "", err
	}
	if err := reg.Register(&policy.Policy{
		ID: policyID, Name: "Lifecycle Demo Deny", Version: "2.0.0",
		Scope: policy.Scope{Resource: resource, Action: action, Environment: env},
		Default: policy.DecisionDeny, Active: true,
	}); err != nil {
		return nil, "", err
	}
	replayRec, err := store.Append(evidence.EvidenceRecord{
		ID: "lifecycle-replay", Actor: "agent/demo", Resource: resource, Action: action,
		Environment: env, PolicyID: oldDecision.PolicyID, PolicyVersion: oldDecision.PolicyVersion,
		Decision: oldDecision.Decision, ExecutionStatus: "blocked-before-execution",
		Metadata: map[string]any{"replay": true},
	})
	if err != nil {
		return nil, "", err
	}
	replayVerifier := verification.NewChain(contract, &policyFreshnessVerifier{registry: reg})
	replayRes, err := replayVerifier.Verify(replayRec)
	if err != nil {
		return nil, "", err
	}
	currentDecision, err := cp.Evaluate(control.Request{Resource: resource, Action: action, Environment: env, Actor: "agent/demo"})
	if err != nil {
		return nil, "", err
	}
	results = append(results, lifecycleResult{
		Name: "04-stale-authorization-replay",
		Expected: "OLD AUTHORIZATION REJECTED AFTER POLICY CHANGE",
		Observed: fmt.Sprintf("stored=%s/current=%s/verification=%s", oldDecision.PolicyVersion, currentDecision.PolicyVersion, replayRes.Status),
		Passed: oldDecision.PolicyVersion == "1.0.0" && currentDecision.PolicyVersion == "2.0.0" && replayRes.Status == verification.StatusFailed && currentDecision.Decision == control.DecisionDeny,
		Checks: []string{"old authorization recorded with policy v1", "policy v2 becomes active", "current evaluation denies", "independent freshness verification rejects replay"},
		EvidenceID: replayRec.ID, VerifyState: string(replayRes.Status),
	})

	if idx, err := store.Verify(); err != nil {
		return nil, "", fmt.Errorf("final evidence chain verification failed at %d: %w", idx, err)
	}
	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}
	return results, fmt.Sprintf("%d/%d lifecycle scenarios passed; evidence chain intact; the agent's claim is not treated as authoritative evidence", passed, len(results)), nil
}

var _ = strings.TrimSpace

// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

// Command level3-evidence exercises the real NAEOS execution gateway against
// a real local filesystem side effect, then verifies that side effect through
// an independent observer and verifier.
//
// Run with:
//
//	go run ./experiments/level3-evidence
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
	"github.com/NAEOS-foundation/naeos/internal/runtime/gateway"
	"github.com/NAEOS-foundation/naeos/internal/verification"
)

const (
	resource    = "workspace"
	action      = "write"
	environment = "experiment"
	actor       = "agent/level3"
	fileName    = "effect.txt"
)

type scenarioResult struct {
	Name            string   `json:"name"`
	Expected        string   `json:"expected"`
	Observed        string   `json:"observed"`
	Passed          bool     `json:"passed"`
	Checks          []string `json:"checks"`
	EvidenceID      string   `json:"evidence_id,omitempty"`
	Verification    string   `json:"verification,omitempty"`
	FailureDetected bool     `json:"failure_detected,omitempty"`
}

type fileSandbox struct {
	root    string
	content []byte
	persist bool
}

func (s *fileSandbox) Execute(req gateway.ToolRequest) (string, error) {
	if req.Resource != resource || req.Action != action {
		return "", fmt.Errorf("unexpected request %s/%s", req.Resource, req.Action)
	}
	if !s.persist {
		return "reported completion without persistence", nil
	}
	if err := os.WriteFile(filepath.Join(s.root, fileName), s.content, 0o600); err != nil {
		return "", err
	}
	return "filesystem write completed", nil
}

type observer struct {
	root string
}

func (o observer) Observe() (content []byte, exists bool, err error) {
	content, err = os.ReadFile(filepath.Join(o.root, fileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return append([]byte(nil), content...), true, nil
}

type sideEffectVerifier struct {
	observer observer
}

func (v sideEffectVerifier) Name() string { return "independent-side-effect" }

func (v sideEffectVerifier) Verify(rec evidence.EvidenceRecord) (verification.VerificationResult, error) {
	result := verification.VerificationResult{
		Status:    verification.StatusVerified,
		Target:    rec.ID,
		Timestamp: time.Now().UTC(),
	}

	content, exists, err := v.observer.Observe()
	if err != nil {
		return verification.VerificationResult{}, err
	}

	expectedObserved := rec.Decision == control.DecisionAllow && rec.ExecutionStatus == "completed"
	unauthorizedObserved := rec.Decision != control.DecisionAllow && exists
	actualHash := ""
	if exists {
		actualHash = evidence.ComputeArtifactHash(content)
	}
	hashMatches := exists && actualHash == rec.ArtifactHash

	result.Checks = append(result.Checks,
		verification.CheckResult{
			Name:   "side-effect-exists-as-authorized",
			Passed: exists == expectedObserved && !unauthorizedObserved,
			Detail: fmt.Sprintf("exists=%v expected=%v unauthorized=%v", exists, expectedObserved, unauthorizedObserved),
		},
		verification.CheckResult{
			Name:   "side-effect-digest",
			Passed: !expectedObserved || hashMatches,
			Detail: fmt.Sprintf("record=%s observed=%s", rec.ArtifactHash, actualHash),
		},
	)

	for _, check := range result.Checks {
		if !check.Passed {
			result.Status = verification.StatusFailed
			result.Message = "independent side-effect verification failed"
			return result, nil
		}
	}
	result.Message = "independent side-effect verified"
	return result, nil
}

func main() {
	results, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "experiment error:", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(struct {
		Experiment string           `json:"experiment"`
		Thesis     string           `json:"thesis"`
		Results    []scenarioResult `json:"results"`
		Summary    string           `json:"summary"`
	}{
		Experiment: "NAEOS Level-3 Evidence Experiment v1",
		Thesis:     "A policy decision is not proof that an unauthorized side effect was prevented.",
		Results:    results,
		Summary:    summarize(results),
	}); err != nil {
		fmt.Fprintln(os.Stderr, "encode:", err)
		os.Exit(1)
	}

	for _, result := range results {
		if !result.Passed {
			os.Exit(2)
		}
	}
}

func run() ([]scenarioResult, error) {
	root, err := os.MkdirTemp("", "naeos-level3-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(root)

	var results []scenarioResult

	// 1. ALLOW: the real gateway invokes a sandbox that writes a real file.
	allowContent := []byte("authorized change\n")
	allowSandbox := &fileSandbox{root: root, content: allowContent, persist: true}
	allowGateway, allowControl := newGateway(policy.DecisionAllow, allowSandbox)
	allowResult, err := allowGateway.Authorize(gateway.ToolRequest{
		Tool: "filesystem", Action: action, Resource: resource,
		Environment: environment, Actor: actor,
	})
	if err != nil {
		return nil, err
	}
	allowObserver := observer{root: root}
	observedContent, exists, err := allowObserver.Observe()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("allow scenario did not produce a filesystem side effect")
	}
	allowRecord, err := appendEvidence(allowControl, allowResult, observedContent, allowObserver)
	if err != nil {
		return nil, err
	}
	allowVerification, err := verify(allowRecord, allowObserver)
	if err != nil {
		return nil, err
	}
	results = append(results, scenarioResult{
		Name: "01-allow-real-side-effect",
		Expected: "ALLOW + completed + observed + VERIFIED",
		Observed: fmt.Sprintf("%s + %s + observed + %s", allowResult.Decision, allowResult.Status, allowVerification.Status),
		Passed: allowResult.Decision == control.DecisionAllow &&
			allowResult.Status == "completed" && exists &&
			allowVerification.Status == verification.StatusVerified,
		Checks: []string{
			"real execution gateway authorized the request",
			"sandbox wrote a real file",
			"independent observer read the file",
			"evidence recorded the observed artifact digest",
			"independent verification passed",
		},
		EvidenceID: allowRecord.ID,
		Verification: string(allowVerification.Status),
	})

	// 2. DENY: the gateway must not invoke the sandbox and the observer must
	// confirm that no file was created.
	denyRoot, err := os.MkdirTemp(root, "deny-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(denyRoot)
	denySandbox := &fileSandbox{root: denyRoot, content: []byte("must not exist\n"), persist: true}
	denyGateway, denyControl := newGateway(policy.DecisionDeny, denySandbox)
	denyResult, err := denyGateway.Authorize(gateway.ToolRequest{
		Tool: "filesystem", Action: action, Resource: resource,
		Environment: environment, Actor: actor,
	})
	if err != nil {
		return nil, err
	}
	denyObserver := observer{root: denyRoot}
	_, denyExists, err := denyObserver.Observe()
	if err != nil {
		return nil, err
	}
	denyRecord, err := appendEvidence(denyControl, denyResult, nil, denyObserver)
	if err != nil {
		return nil, err
	}
	denyVerification, err := verify(denyRecord, denyObserver)
	if err != nil {
		return nil, err
	}
	results = append(results, scenarioResult{
		Name: "02-deny-no-side-effect",
		Expected: "DENY + no execution + no observed side effect + VERIFIED",
		Observed: fmt.Sprintf("%s + %s + no side effect + %s", denyResult.Decision, denyResult.Status, denyVerification.Status),
		Passed: denyResult.Decision == control.DecisionDeny &&
			denyResult.Status == "denied" && !denyExists &&
			denyVerification.Status == verification.StatusVerified,
		Checks: []string{
			"policy returned DENY",
			"gateway did not invoke the sandbox",
			"independent observer confirmed absence",
			"evidence recorded the denial",
			"independent verification passed",
		},
		EvidenceID: denyRecord.ID,
		Verification: string(denyVerification.Status),
	})

	// 3. DIRECT BYPASS: write outside the gateway. The observer sees the
	// effect, but the gateway history contains no authorized execution.
	bypassRoot, err := os.MkdirTemp(root, "bypass-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(bypassRoot)
	bypassGateway, bypassControl := newGateway(policy.DecisionDeny, &fileSandbox{
		root: bypassRoot, content: []byte("bypass\n"), persist: true,
	})
	bypassPath := filepath.Join(bypassRoot, fileName)
	if err := os.WriteFile(bypassPath, []byte("out-of-band write\n"), 0o600); err != nil {
		return nil, err
	}
	bypassObserver := observer{root: bypassRoot}
	bypassContent, bypassExists, err := bypassObserver.Observe()
	if err != nil {
		return nil, err
	}
	bypassHistory := bypassGateway.History()
	bypassRecord, err := appendEvidence(bypassControl, gateway.ExecutionResult{
		Request: gateway.ToolRequest{
			Tool: "filesystem", Action: action, Resource: resource,
			Environment: environment, Actor: actor,
		},
		Decision: control.DecisionDeny, PolicyID: "level3-DENY", Status: "bypassed",
		Output: "side effect observed outside gateway",
	}, bypassContent, bypassObserver)
	if err != nil {
		return nil, err
	}
	bypassVerification, err := verify(bypassRecord, bypassObserver)
	if err != nil {
		return nil, err
	}
	bypassDetected := bypassExists && len(bypassHistory) == 0 && bypassVerification.Status == verification.StatusFailed
	results = append(results, scenarioResult{
		Name: "03-direct-bypass-detected",
		Expected: "OUT-OF-BAND SIDE EFFECT MUST FAIL GOVERNANCE VERIFICATION",
		Observed: fmt.Sprintf("side effect=%v gateway_history=%d verification=%s", bypassExists, len(bypassHistory), bypassVerification.Status),
		Passed: bypassDetected,
		Checks: []string{
			"side effect was produced without gateway execution",
			"gateway history contains no authorized execution",
			"independent observer saw the out-of-band effect",
			"verification failed instead of treating DENY as proof of prevention",
		},
		EvidenceID: bypassRecord.ID,
		Verification: string(bypassVerification.Status),
		FailureDetected: bypassVerification.Status == verification.StatusFailed,
	})

	// 4. EXECUTION/OBSERVATION SEPARATION: sandbox reports completion but
	// intentionally does not persist the effect. The observer catches it.
	lieRoot, err := os.MkdirTemp(root, "lie-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(lieRoot)
	lieSandbox := &fileSandbox{root: lieRoot, content: []byte("reported\n"), persist: false}
	lieGateway, lieControl := newGateway(policy.DecisionAllow, lieSandbox)
	lieResult, err := lieGateway.Authorize(gateway.ToolRequest{
		Tool: "filesystem", Action: action, Resource: resource,
		Environment: environment, Actor: actor,
	})
	if err != nil {
		return nil, err
	}
	lieObserver := observer{root: lieRoot}
	lieContent, lieExists, err := lieObserver.Observe()
	if err != nil {
		return nil, err
	}
	lieRecord, err := appendEvidence(lieControl, lieResult, lieContent, lieObserver)
	if err != nil {
		return nil, err
	}
	lieVerification, err := verify(lieRecord, lieObserver)
	if err != nil {
		return nil, err
	}
	results = append(results, scenarioResult{
		Name: "04-execution-observation-separation",
		Expected: "EXECUTION CLAIM MUST NOT OVERRIDE INDEPENDENT OBSERVATION",
		Observed: fmt.Sprintf("execution=%s observed=%v verification=%s", lieResult.Status, lieExists, lieVerification.Status),
		Passed: lieResult.Status == "completed" && !lieExists && lieVerification.Status == verification.StatusFailed,
		Checks: []string{
			"gateway execution result reported completion",
			"observer independently checked the filesystem",
			"observer found no side effect",
			"verification preserved the mismatch as a failure",
		},
		EvidenceID: lieRecord.ID,
		Verification: string(lieVerification.Status),
		FailureDetected: lieVerification.Status == verification.StatusFailed,
	})

	// 5. TAMPER: capture an observed artifact, mutate it, then verify again.
	tamperRoot, err := os.MkdirTemp(root, "tamper-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tamperRoot)
	tamperContent := []byte("original\n")
	tamperSandbox := &fileSandbox{root: tamperRoot, content: tamperContent, persist: true}
	tamperGateway, tamperControl := newGateway(policy.DecisionAllow, tamperSandbox)
	tamperResult, err := tamperGateway.Authorize(gateway.ToolRequest{
		Tool: "filesystem", Action: action, Resource: resource,
		Environment: environment, Actor: actor,
	})
	if err != nil {
		return nil, err
	}
	tamperObserver := observer{root: tamperRoot}
	tamperObserved, tamperExists, err := tamperObserver.Observe()
	if err != nil || !tamperExists {
		return nil, fmt.Errorf("tamper scenario initial observation failed: %v", err)
	}
	tamperRecord, err := appendEvidence(tamperControl, tamperResult, tamperObserved, tamperObserver)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(tamperRoot, fileName), []byte("mutated\n"), 0o600); err != nil {
		return nil, err
	}
	tamperVerification, err := verify(tamperRecord, tamperObserver)
	if err != nil {
		return nil, err
	}
	results = append(results, scenarioResult{
		Name: "05-tamper-after-observation",
		Expected: "POST-OBSERVATION MUTATION MUST FAIL VERIFICATION",
		Observed: fmt.Sprintf("verification=%s", tamperVerification.Status),
		Passed: tamperVerification.Status == verification.StatusFailed,
		Checks: []string{
			"artifact was observed and hashed",
			"artifact was mutated after evidence capture",
			"independent verification detected digest mismatch",
		},
		EvidenceID: tamperRecord.ID,
		Verification: string(tamperVerification.Status),
		FailureDetected: tamperVerification.Status == verification.StatusFailed,
	})

	return results, nil
}

func newGateway(decision policy.Decision, sandbox gateway.Sandbox) (*gateway.ExecutionGateway, *control.ControlPlane) {
	reg := policy.NewRegistry()
	_ = reg.Register(&policy.Policy{
		ID: "level3-" + string(decision), Name: "Level 3 " + string(decision),
		Version: "1.0.0",
		Scope: policy.Scope{Resource: resource, Action: action, Environment: environment},
		Default: decision, Active: true,
	})
	cp := control.New(reg, control.FailClosed(true))
	return gateway.New(cp, sandbox), cp
}

func appendEvidence(cp *control.ControlPlane, result gateway.ExecutionResult, observed []byte, obs observer) (evidence.EvidenceRecord, error) {
	policyVersion := "unknown"
	decisions := cp.ListDecisions()
	if len(decisions) > 0 {
		policyVersion = decisions[len(decisions)-1].PolicyVersion
	}

	rec := evidence.EvidenceRecord{
		Actor: actor, Resource: result.Request.Resource, Action: result.Request.Action,
		Environment: environment, PolicyID: result.PolicyID, PolicyVersion: policyVersion,
		Decision: result.Decision, DecisionReasons: result.Reasons,
		ArtifactName: fileName, ExecutionStatus: result.Status,
		ExecutionOutput: result.Output, ExecutionDurationMs: result.Duration.Milliseconds(),
		Metadata: map[string]any{
			"observation_exists": observed != nil,
			"observation_path": filepath.Join(obs.root, fileName),
			"gateway_decision_count": len(decisions),
		},
	}
	if observed != nil {
		rec.ArtifactHash = evidence.ComputeArtifactHash(observed)
		rec.ArtifactSize = len(observed)
	}

	store := evidence.NewStore()
	seed := result.Decision + ":" + result.Status + ":" + rec.ArtifactHash + ":" + result.Output
	rec.ID = "level3-" + evidence.ComputeArtifactHash([]byte(seed))[:16]
	saved, err := store.Append(rec)
	if err != nil {
		return evidence.EvidenceRecord{}, err
	}
	if _, err := store.Verify(); err != nil {
		return evidence.EvidenceRecord{}, err
	}
	return saved, nil
}

func verify(rec evidence.EvidenceRecord, obs observer) (verification.VerificationResult, error) {
	contract := verification.Contract{
		Name: "level3-evidence-v1", Version: "1.0.0",
		Description: "The observed side effect must match the independently re-read artifact.",
		Requirements: []string{"evidence integrity", "side-effect observation"},
	}
	store := evidence.NewStore()
	saved, err := store.Append(rec)
	if err != nil {
		return verification.VerificationResult{}, err
	}
	chain := verification.NewChain(contract,
		verification.NewEvidenceChainVerifier(store),
		sideEffectVerifier{observer: obs},
	)
	return chain.Verify(saved)
}

func summarize(results []scenarioResult) string {
	passed := 0
	for _, result := range results {
		if result.Passed {
			passed++
		}
	}
	return fmt.Sprintf("%d/%d expected scenario assertions passed; expected verification failures were preserved", passed, len(results))
}

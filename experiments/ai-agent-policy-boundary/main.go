// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

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
 resource = "filesystem"
 action = "write"
 environment = "flagship"
 actor = "ai-agent/demo"
 fileName = "agent-action.txt"
)

type scenarioResult struct {
 Name string
 Expected string
 Observed string
 Passed bool
 Decision string
 GatewayState string
 ObservedSideEffect bool
 Verification string
 Checks []string
}

type fileSandbox struct {
 root string
 content []byte
 persist bool
}

func (s *fileSandbox) Execute(req gateway.ToolRequest) (string, error) {
 if !s.persist { return "sandbox reported completion without persistence", nil }
 if err := os.WriteFile(filepath.Join(s.root, fileName), s.content, 0600); err != nil { return "", err }
 return fmt.Sprintf("wrote %s", fileName), nil
}

type fileObserver struct{ root string }

func (o fileObserver) Observe() ([]byte, bool, error) {
 b, err := os.ReadFile(filepath.Join(o.root, fileName))
 if os.IsNotExist(err) { return nil, false, nil }
 if err != nil { return nil, false, err }
 return b, true, nil
}

type observationVerifier struct{ observer fileObserver }

func (v observationVerifier) Name() string { return "independent-observation" }

func (v observationVerifier) Verify(rec evidence.EvidenceRecord) (verification.VerificationResult, error) {
 b, exists, err := v.observer.Observe()
 if err != nil { return verification.VerificationResult{}, err }
 res := verification.VerificationResult{Status: verification.StatusVerified, Target: rec.ID, Timestamp: time.Now().UTC()}
 expectedExists := rec.ArtifactHash != ""
 actualHash := ""
 if exists { actualHash = evidence.ComputeArtifactHash(b) }
 match := exists == expectedExists && (!exists || actualHash == rec.ArtifactHash)
 res.Checks = append(res.Checks, verification.CheckResult{Name: "observed-side-effect", Passed: match, Detail: fmt.Sprintf("expected_exists=%v actual_exists=%v expected_hash=%s actual_hash=%s", expectedExists, exists, rec.ArtifactHash, actualHash)})
 if !match { res.Status = verification.StatusFailed; res.Message = "independent observation does not match evidence" }
 return res, nil
}

type gatewayVerifier struct{ gateway *gateway.ExecutionGateway }

func (v gatewayVerifier) Name() string { return "gateway-authorization" }

func (v gatewayVerifier) Verify(rec evidence.EvidenceRecord) (verification.VerificationResult, error) {
 res := verification.VerificationResult{Status: verification.StatusVerified, Target: rec.ID, Timestamp: time.Now().UTC()}
 authorizedExecution := false
 for _, item := range v.gateway.History() {
  if item.Decision == control.DecisionAllow && item.Status == "completed" { authorizedExecution = true; break }
 }
 expectedAuthorized := rec.Decision == control.DecisionAllow && rec.ExecutionStatus == "completed"
 match := authorizedExecution == expectedAuthorized
 res.Checks = append(res.Checks, verification.CheckResult{Name: "authorized-execution", Passed: match, Detail: fmt.Sprintf("expected_authorized=%v gateway_authorized_execution=%v", expectedAuthorized, authorizedExecution)})
 if !match { res.Status = verification.StatusFailed; res.Message = "observed execution is inconsistent with gateway authorization history" }
 return res, nil
}

type observerSource struct{ observer fileObserver }

func (s observerSource) Content(name string) ([]byte, bool) {
 b, ok, err := s.observer.Observe()
 if err != nil { return nil, false }
 return b, ok
}

func newControl(decision policy.Decision) *control.ControlPlane {
 reg := policy.NewRegistry()
 _ = reg.Register(&policy.Policy{
  ID: "ai-agent-boundary", Name: "AI Agent Boundary", Version: "1.0.0",
  Scope: policy.Scope{Resource: resource, Action: action, Environment: environment},
  Default: decision, Active: true,
 })
 return control.New(reg, control.FailClosed(true))
}

func appendEvidence(result gateway.ExecutionResult, observed []byte) (evidence.EvidenceRecord, *evidence.EvidenceStore, error) {
 store := evidence.NewStore()
 rec := evidence.EvidenceRecord{
  ID: "ai-agent-boundary-" + fmt.Sprintf("%d", time.Now().UnixNano()),
  Actor: actor, Resource: result.Request.Resource, Action: result.Request.Action, Environment: environment,
  PolicyID: result.PolicyID, PolicyVersion: "1.0.0", RuleID: result.RuleID, Decision: result.Decision,
  DecisionReasons: result.Reasons, ArtifactName: fileName, ExecutionStatus: result.Status,
  ExecutionOutput: result.Output, ExecutionDurationMs: result.Duration.Milliseconds(),
 }
 if observed != nil { rec.ArtifactHash = evidence.ComputeArtifactHash(observed); rec.ArtifactSize = len(observed) }
 saved, err := store.Append(rec)
 return saved, store, err
}

func verify(rec evidence.EvidenceRecord, store *evidence.EvidenceStore, gw *gateway.ExecutionGateway, observer fileObserver) (verification.VerificationResult, error) {
 contract := verification.Contract{
  Name: "ai-agent-policy-boundary-v1", Version: "1.0.0",
  Description: "Authorization, execution, observation, and evidence must remain distinct.",
  Requirements: []string{"evidence integrity", "independent observation", "gateway authorization"},
 }
 chain := verification.NewChain(contract,
  verification.NewEvidenceChainVerifier(store),
  verification.NewArtifactHashVerifierWithSource(nil, observerSource{observer: observer}),
  observationVerifier{observer: observer},
  gatewayVerifier{gateway: gw},
 )
 return chain.Verify(rec)
}

func runScenario(name string, decision policy.Decision, persist bool, bypass bool) (scenarioResult, error) {
 root, err := os.MkdirTemp("", "naeos-ai-agent-boundary-*")
 if err != nil { return scenarioResult{}, err }
 defer os.RemoveAll(root)

 cp := newControl(decision)
 sandbox := &fileSandbox{root: root, content: []byte("authorized agent change\n"), persist: persist}
 gw := gateway.New(cp, sandbox)
 observer := fileObserver{root: root}

 var result gateway.ExecutionResult
 if bypass {
  if err := os.WriteFile(filepath.Join(root, fileName), []byte("out-of-band agent change\n"), 0600); err != nil { return scenarioResult{}, err }
  result = gateway.ExecutionResult{
   Request: gateway.ToolRequest{Tool: "filesystem", Action: action, Resource: resource, Environment: environment, Actor: actor},
   Decision: control.DecisionDeny, PolicyID: "ai-agent-boundary", Status: "bypassed",
   Output: "side effect occurred outside gateway", Timestamp: time.Now().UTC(),
  }
 } else {
  result, err = gw.Authorize(gateway.ToolRequest{Tool: "filesystem", Action: action, Resource: resource, Environment: environment, Actor: actor})
  if err != nil { return scenarioResult{}, err }
 }

 observed, exists, err := observer.Observe()
 if err != nil { return scenarioResult{}, err }
 rec, store, err := appendEvidence(result, observed)
 if err != nil { return scenarioResult{}, err }
 ver, err := verify(rec, store, gw, observer)
 if err != nil { return scenarioResult{}, err }

 expectedVerification := verification.StatusVerified
 expected := ""
 passed := false
 switch name {
 case "ALLOW":
  expected = "gateway ALLOW -> real side effect -> independent observation -> VERIFIED"
  passed = result.Decision == control.DecisionAllow && result.Status == "completed" && exists && ver.Status == expectedVerification
 case "DENY":
  expected = "gateway DENY -> no sandbox execution -> no side effect -> VERIFIED"
  passed = result.Decision == control.DecisionDeny && result.Status == "denied" && !exists && ver.Status == expectedVerification
 case "REQUIRE_APPROVAL":
  expected = "approval required -> no sandbox execution -> no side effect"
  passed = result.Decision == control.DecisionRequireApproval && result.Status == "denied" && !exists && ver.Status == expectedVerification
 case "DIRECT_BYPASS":
  expected = "out-of-band side effect -> no authorized gateway execution -> verification FAILED"
  passed = exists && len(gw.History()) == 0 && ver.Status == verification.StatusFailed
 }

 return scenarioResult{
  Name: name, Expected: expected,
  Observed: fmt.Sprintf("decision=%s gateway=%s side_effect=%v verification=%s", result.Decision, result.Status, exists, ver.Status),
  Passed: passed, Decision: string(result.Decision), GatewayState: result.Status,
  ObservedSideEffect: exists, Verification: string(ver.Status),
  Checks: []string{
   "agent intent normalized as a ToolRequest",
   "policy decision produced by the real control plane",
   "execution crossed the real gateway or intentionally bypassed it",
   "side effect checked by an independent observer",
   "evidence recorded separately from execution result",
   "verification evaluated independently",
  },
 }, nil
}

func main() {
 scenarios := []struct{name string; decision policy.Decision; persist bool; bypass bool}{
  {"ALLOW", policy.DecisionAllow, true, false},
  {"DENY", policy.DecisionDeny, true, false},
  {"REQUIRE_APPROVAL", policy.DecisionRequireApproval, true, false},
  {"DIRECT_BYPASS", policy.DecisionDeny, false, true},
 }
 results := make([]scenarioResult, 0, len(scenarios))
 for _, s := range scenarios {
  result, err := runScenario(s.name, s.decision, s.persist, s.bypass)
  if err != nil { fmt.Fprintln(os.Stderr, "experiment error:", err); os.Exit(1) }
  results = append(results, result)
 }
 passed := 0
 for _, r := range results { if r.Passed { passed++ } }
 enc := json.NewEncoder(os.Stdout)
 enc.SetIndent("", "  ")
 _ = enc.Encode(struct{Experiment string; Thesis string; Results []scenarioResult; Summary string}{
  Experiment: "NAEOS AI Agent Policy Boundary v1",
  Thesis: "An AI agent may propose an action, but authorization and proof of side effect remain outside the agent.",
  Results: results, Summary: fmt.Sprintf("%d/%d expected scenario assertions passed", passed, len(results)),
 })
 for _, r := range results { if !r.Passed { os.Exit(2) } }
}

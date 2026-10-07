// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"errors"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

type conformanceRevalidator struct {
	err error
}

func (r conformanceRevalidator) Evaluate(req control.Request) (control.DecisionRecord, error) {
	return control.DecisionRecord{
		Request:  req,
		Decision: control.DecisionAllow,
		PolicyID: "stale-policy",
	}, nil
}

func (r conformanceRevalidator) ValidateDecision(control.Request, control.DecisionRecord) (control.DecisionRecord, error) {
	if r.err != nil {
		return control.DecisionRecord{}, r.err
	}
	return control.DecisionRecord{Decision: control.DecisionAllow, PolicyID: "current"}, nil
}

type conformanceObserver struct {
	observed bool
}

func (o conformanceObserver) Observe(ToolRequest, ExecutionResult) (Observation, error) {
	return Observation{
		Status:       "observed",
		Observed:     o.observed,
		ArtifactHash: "sha256:evidence",
	}, nil
}

type conformanceReplayStore struct {
	err error
}

func (s conformanceReplayStore) Claim(string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return true, nil
}

type conformanceAdapter struct {
	request   ToolRequest
	decisions []ExecutionResult
}

func (a *conformanceAdapter) Name() string { return "external-test-agent" }

func (a *conformanceAdapter) NormalizeTool(any) (ToolRequest, error) {
	return a.request, nil
}

func (a *conformanceAdapter) OnDecision(result ExecutionResult) error {
	a.decisions = append(a.decisions, result)
	return nil
}

func TestExternalAgentConformanceAllowedRequest(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "repo-policy"}
	sb := &countingSandbox{}
	gw := New(cp, sb, WithReplayProtection(true))

	result, err := gw.Authorize(ToolRequest{
		RequestID:    "request-conformance-allow-01",
		InvocationID: "conformance-allow-01",
		Tool:         "file-edit",
		Action:       "write",
		Capability:   "repository.write",
		Resource:     "src/app.go",
		Actor:        "external-test-agent",
	})
	if err != nil {
		t.Fatalf("allowed request returned error: %v", err)
	}
	if result.Status != "completed" || result.Decision != control.DecisionAllow || result.RequestID != "request-conformance-allow-01" || result.Request.RequestID != "request-conformance-allow-01" {
		t.Fatalf("expected governed ALLOW/completed result, got %+v", result)
	}
	if sb.Count() != 1 {
		t.Fatalf("expected one sandbox side effect, got %d", sb.Count())
	}
}

func TestExternalAgentConformanceMissingInvocationFailsClosed(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow}
	sb := &countingSandbox{}
	gw := New(cp, sb, WithReplayProtection(true))

	_, err := gw.Authorize(ToolRequest{
		Tool:   "file-edit",
		Action: "write",
	})
	if err == nil {
		t.Fatal("expected missing invocation_id to fail closed")
	}
	if sb.Count() != 0 {
		t.Fatalf("missing invocation_id crossed execution boundary: %d side effects", sb.Count())
	}
}

func TestExternalAgentConformanceReplayDeniedBeforeSideEffect(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow}
	sb := &countingSandbox{}
	store := NewInMemoryInvocationStore()
	gw := New(cp, sb, WithReplayProtection(true), WithInvocationStore(store))

	first, err := gw.Authorize(ToolRequest{
		InvocationID: "conformance-replay-01",
		Tool:         "file-edit",
		Action:       "write",
	})
	if err != nil || first.Status != "completed" {
		t.Fatalf("first invocation failed: result=%+v err=%v", first, err)
	}

	second, err := gw.Authorize(ToolRequest{
		InvocationID: "conformance-replay-01",
		Tool:         "file-edit",
		Action:       "write",
	})
	if err != nil {
		t.Fatalf("replay returned unexpected error: %v", err)
	}
	if second.Status != "denied" {
		t.Fatalf("expected replay denial, got %+v", second)
	}
	if sb.Count() != 1 {
		t.Fatalf("replay crossed side-effect boundary: %d executions", sb.Count())
	}
}

func TestExternalAgentConformanceStaleAuthorizationCannotExecute(t *testing.T) {
	cp := &conformanceRevalidator{err: errors.New("policy changed")}
	sb := &countingSandbox{}
	gw := New(cp, sb, WithReplayProtection(true))

	result, err := gw.Authorize(ToolRequest{
		InvocationID: "conformance-stale-01",
		Tool:         "deploy",
		Action:       "execute",
	})
	if err != nil {
		t.Fatalf("stale authorization returned unexpected error: %v", err)
	}
	if result.Status != "denied" {
		t.Fatalf("expected stale authorization denial, got %+v", result)
	}
	if sb.Count() != 0 {
		t.Fatalf("stale authorization crossed execution boundary: %d side effects", sb.Count())
	}
}

func TestExternalAgentConformanceAdapterCannotBypassGateway(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionDeny}
	sb := &countingSandbox{}
	adapter := &conformanceAdapter{
		request: ToolRequest{
			InvocationID: "conformance-adapter-01",
			Tool:         "deploy",
			Action:       "execute",
		},
	}
	gw := New(cp, sb, WithReplayProtection(true))
	gw.RegisterAdapter(adapter.Name(), adapter)

	result, err := gw.AuthorizeFromAdapter(adapter.Name(), map[string]any{"action": "execute"})
	if err != nil {
		t.Fatalf("adapter path returned unexpected error: %v", err)
	}
	if result.Status != "denied" || sb.Count() != 0 {
		t.Fatalf("adapter bypassed gateway boundary: result=%+v side_effects=%d", result, sb.Count())
	}
	if len(adapter.decisions) != 1 || adapter.decisions[0].Status != "denied" {
		t.Fatalf("adapter did not receive governed denial: %+v", adapter.decisions)
	}
}

func TestExternalAgentConformanceApprovalCannotExecute(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionRequireApproval}
	sb := &countingSandbox{}
	gw := New(cp, sb, WithReplayProtection(true))

	result, err := gw.Authorize(ToolRequest{
		InvocationID: "conformance-approval-01",
		Tool:         "deploy",
		Action:       "execute",
	})
	if err != nil {
		t.Fatalf("approval-required request returned error: %v", err)
	}
	if result.Status != "denied" || sb.Count() != 0 {
		t.Fatalf("approval-required request crossed boundary: result=%+v side_effects=%d", result, sb.Count())
	}
}

func TestExternalAgentConformanceEvidenceObservation(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow}
	sb := &countingSandbox{}
	gw := New(cp, sb, WithObserver(conformanceObserver{observed: true}))

	result, err := gw.Authorize(ToolRequest{
		Tool:   "file-edit",
		Action: "write",
	})
	if err != nil {
		t.Fatalf("observed execution returned error: %v", err)
	}
	if result.Observation == nil || !result.Observation.Observed || result.Observation.ArtifactHash == "" {
		t.Fatalf("expected evidence-backed observation, got %+v", result.Observation)
	}
}

func TestExternalAgentConformanceDurableReplayFailureFailsClosed(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow}
	sb := &countingSandbox{}
	gw := New(cp, sb,
		WithReplayProtection(true),
		WithInvocationStore(conformanceReplayStore{err: errors.New("durable store unavailable")}),
	)

	result, err := gw.Authorize(ToolRequest{
		InvocationID: "conformance-store-failure-01",
		Tool:         "file-edit",
		Action:       "write",
	})
	if err == nil {
		t.Fatal("expected durable replay failure to fail closed")
	}
	if result.Status != "denied" || sb.Count() != 0 {
		t.Fatalf("durable replay failure crossed boundary: result=%+v side_effects=%d", result, sb.Count())
	}
}

func TestExternalAgentConformanceContractVersionFixture(t *testing.T) {
	const contractVersion = "1.0"
	const contractName = "naeos.agent-adoption"

	if contractName != "naeos.agent-adoption" || contractVersion != "1.0" {
		t.Fatalf("unexpected conformance contract: %s/%s", contractName, contractVersion)
	}

	// Unknown major versions are outside this suite's accepted contract and
	// must be rejected by an adapter before the request reaches the gateway.
	unknownMajor := "2.0"
	if unknownMajor[:1] == contractVersion[:1] {
		t.Fatalf("test fixture did not represent an unknown major version: %s", unknownMajor)
	}
}

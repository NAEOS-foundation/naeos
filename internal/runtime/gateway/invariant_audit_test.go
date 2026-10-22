// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"errors"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

type mutatingPolicyControlPlane struct {
	*control.ControlPlane
	registry *policy.Registry
}

func (c mutatingPolicyControlPlane) ValidateDecision(req control.Request, issued control.DecisionRecord) (control.DecisionRecord, error) {
	if err := c.registry.Register(&policy.Policy{
		ID:      "gateway-mutation-policy",
		Version: "2.0.0",
		Scope:   policy.Scope{Resource: "filesystem", Action: "write"},
		Default: policy.DecisionDeny,
	}); err != nil {
		return control.DecisionRecord{}, err
	}
	return c.ControlPlane.ValidateDecision(req, issued)
}

func TestGatewayInvariantRealPolicyMutationCannotExecute(t *testing.T) {
	registry := policy.NewRegistry()
	if err := registry.Register(&policy.Policy{
		ID:      "gateway-mutation-policy",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "filesystem", Action: "write"},
		Default: policy.DecisionAllow,
	}); err != nil {
		t.Fatalf("register initial policy: %v", err)
	}

	cp := mutatingPolicyControlPlane{
		ControlPlane: control.New(registry),
		registry:     registry,
	}
	sb := &countingSandbox{}
	gw := New(cp, sb, WithRequiredRevalidation(true))

	result, err := gw.Authorize(ToolRequest{
		RequestID: "audit-policy-mutation",
		Tool:      "filesystem",
		Action:    "write",
	})
	if err != nil {
		t.Fatalf("policy mutation should be handled as a governed denial: %v", err)
	}
	if result.Status != "denied" {
		t.Fatalf("policy mutation was not denied: result=%+v", result)
	}
	if sb.Count() != 0 {
		t.Fatalf("policy mutation crossed execution boundary: side_effects=%d", sb.Count())
	}
	active, ok := registry.GetActive("gateway-mutation-policy")
	if !ok || active.Version != "2.0.0" || active.Default != policy.DecisionDeny {
		t.Fatalf("test did not apply the intended real policy mutation: active=%+v found=%t", active, ok)
	}
}

type auditRevalidator struct {
	decision control.Decision
	err      error
}

func (r auditRevalidator) Evaluate(req control.Request) (control.DecisionRecord, error) {
	return control.DecisionRecord{
		Request:  req,
		Decision: control.DecisionAllow,
		PolicyID: "audit-policy",
	}, nil
}

func (r auditRevalidator) ValidateDecision(req control.Request, issued control.DecisionRecord) (control.DecisionRecord, error) {
	if r.err != nil {
		return control.DecisionRecord{}, r.err
	}
	issued.Decision = r.decision
	return issued, nil
}

func TestGatewayInvariantRequiredRevalidationFailsClosedWhenUnsupported(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow}
	sb := &countingSandbox{}
	gw := New(cp, sb, WithRequiredRevalidation(true))

	result, err := gw.Authorize(ToolRequest{RequestID: "audit-no-revalidator", Tool: "filesystem", Action: "write"})
	if err == nil {
		t.Fatal("expected missing decision revalidation to fail closed")
	}
	if result.Status != "denied" {
		t.Fatalf("expected denied result, got %q", result.Status)
	}
	if sb.Count() != 0 {
		t.Fatalf("missing revalidation reached sandbox: %d calls", sb.Count())
	}
}

func TestGatewayInvariantUnknownDecisionFailsClosed(t *testing.T) {
	cp := &stubControlPlane{decision: control.Decision("UNKNOWN")}
	sb := &countingSandbox{}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{RequestID: "audit-unknown", Tool: "filesystem", Action: "write"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "denied" || sb.Count() != 0 {
		t.Fatalf("unknown decision crossed execution boundary: result=%+v side_effects=%d", result, sb.Count())
	}
}

func TestGatewayInvariantInitialDecisionsCannotReachSandbox(t *testing.T) {
	for _, decision := range []control.Decision{
		control.DecisionDeny,
		control.DecisionRequireApproval,
	} {
		t.Run(string(decision), func(t *testing.T) {
			cp := &stubControlPlane{decision: decision}
			sb := &countingSandbox{}
			gw := New(cp, sb)

			result, err := gw.Authorize(ToolRequest{Tool: "filesystem", Action: "write"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Status != "denied" || sb.Count() != 0 {
				t.Fatalf("initial %s crossed execution boundary: result=%+v side_effects=%d", decision, result, sb.Count())
			}
		})
	}
}

func TestGatewayInvariantRevalidationRequireApprovalCannotExecute(t *testing.T) {
	cp := auditRevalidator{decision: control.DecisionRequireApproval}
	sb := &countingSandbox{}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{Tool: "filesystem", Action: "write"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "denied" || result.Decision != control.DecisionRequireApproval || sb.Count() != 0 {
		t.Fatalf("revalidated approval crossed execution boundary: result=%+v side_effects=%d", result, sb.Count())
	}
}

func TestGatewayInvariantRevalidationDenyCannotExecute(t *testing.T) {
	cp := auditRevalidator{decision: control.DecisionDeny}
	sb := &countingSandbox{}
	gw := New(cp, sb, WithRequiredRevalidation(true))

	result, err := gw.Authorize(ToolRequest{Tool: "filesystem", Action: "write"})
	if err != nil {
		t.Fatalf("revalidation DENY should be represented as governed denial: %v", err)
	}
	if result.Status != "denied" || result.Decision != control.DecisionDeny || sb.Count() != 0 {
		t.Fatalf("revalidation DENY crossed execution boundary: result=%+v side_effects=%d", result, sb.Count())
	}
}

func TestGatewayInvariantRevalidationErrorFailsClosed(t *testing.T) {
	cp := auditRevalidator{err: errors.New("policy changed during revalidation")}
	sb := &countingSandbox{}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{Tool: "filesystem", Action: "write"})
	if err != nil {
		t.Fatalf("revalidation error should be represented as governed denial: %v", err)
	}
	if result.Status != "denied" || sb.Count() != 0 {
		t.Fatalf("revalidation error crossed execution boundary: result=%+v side_effects=%d", result, sb.Count())
	}
}

type auditObserver struct {
	observation Observation
}

func (o auditObserver) Observe(req ToolRequest, result ExecutionResult) (Observation, error) {
	return o.observation, nil
}

func TestGatewayInvariantObservationMismatchFailsClosed(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow}
	sb := &countingSandbox{}
	gw := New(cp, sb, WithObserver(auditObserver{
		observation: Observation{Status: "mismatch", Observed: true},
	}))

	result, err := gw.Authorize(ToolRequest{Tool: "filesystem", Action: "write"})
	if err == nil {
		t.Fatal("expected observation mismatch to fail closed")
	}
	if result.Status != "failed" || sb.Count() != 1 {
		t.Fatalf("unexpected mismatch handling: result=%+v side_effects=%d", result, sb.Count())
	}
}

func TestGatewayInvariantObservationMustBindIdentity(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow}
	sb := &countingSandbox{}
	gw := New(cp, sb, WithObserver(auditObserver{
		observation: Observation{
			RequestID:    "other-request",
			InvocationID: "other-invocation",
			Status:       "observed",
			Observed:     true,
		},
	}))

	result, err := gw.Authorize(ToolRequest{
		RequestID:    "audit-request-01",
		InvocationID: "audit-invocation-01",
		Tool:         "filesystem",
		Action:       "write",
	})
	if err == nil {
		t.Fatal("expected mismatched observation identity to fail closed")
	}
	if result.Status != "failed" || sb.Count() != 1 {
		t.Fatalf("unexpected observation identity handling: result=%+v side_effects=%d", result, sb.Count())
	}
}

func TestGatewayInvariantEvidenceCannotBeClaimedFromSandboxOutputAlone(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow}
	sb := &stubSandbox{output: "claimed side effect"}
	gw := New(cp, sb, WithObserver(auditObserver{
		observation: Observation{Status: "absent", Observed: false},
	}))

	result, err := gw.Authorize(ToolRequest{Tool: "filesystem", Action: "write"})
	if err == nil {
		t.Fatal("expected sandbox output without observed evidence to fail closed")
	}
	if result.Status != "failed" || result.Observation == nil || result.Observation.Observed {
		t.Fatalf("sandbox output was treated as evidence: result=%+v", result)
	}
}

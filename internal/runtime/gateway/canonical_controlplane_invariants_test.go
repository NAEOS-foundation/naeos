// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway_test

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
	"github.com/NAEOS-foundation/naeos/internal/runtime/gateway"
)

type countingSandbox struct {
	calls int32
}

func (s *countingSandbox) Execute(_ gateway.ToolRequest) (string, error) {
	atomic.AddInt32(&s.calls, 1)
	return "executed", nil
}

func realControlPlane(t *testing.T, policies ...*policy.Policy) (*control.ControlPlane, *policy.Registry) {
	t.Helper()
	reg := policy.NewRegistry()
	for _, p := range policies {
		if err := reg.Register(p); err != nil {
			t.Fatalf("register policy: %v", err)
		}
	}
	return control.New(reg), reg
}

func TestCanonicalControlPlaneDenialBlocksSideEffect(t *testing.T) {
	cp, _ := realControlPlane(t, &policy.Policy{
		ID:      "deny-deploy",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run"},
		Default: policy.DecisionDeny,
	})
	sb := &countingSandbox{}
	gw := gateway.New(cp, sb)

	result, err := gw.Authorize(gateway.ToolRequest{Tool: "deploy", Action: "run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != control.DecisionDeny || result.Status != "denied" {
		t.Fatalf("expected canonical ControlPlane DENY, got decision=%s status=%s", result.Decision, result.Status)
	}
	if got := atomic.LoadInt32(&sb.calls); got != 0 {
		t.Fatalf("side effect executed after DENY: sandbox calls=%d", got)
	}
}

func TestCanonicalControlPlaneAllowPermitsSideEffect(t *testing.T) {
	cp, _ := realControlPlane(t, &policy.Policy{
		ID:      "allow-deploy",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run"},
		Default: policy.DecisionAllow,
	})
	sb := &countingSandbox{}
	gw := gateway.New(cp, sb)

	result, err := gw.Authorize(gateway.ToolRequest{Tool: "deploy", Action: "run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != control.DecisionAllow || result.Status != "completed" {
		t.Fatalf("expected canonical ControlPlane ALLOW and completed execution, got decision=%s status=%s", result.Decision, result.Status)
	}
	if got := atomic.LoadInt32(&sb.calls); got != 1 {
		t.Fatalf("expected exactly one side effect after ALLOW, got %d", got)
	}
}

func TestCanonicalControlPlaneEvaluatorFailureBlocksSideEffect(t *testing.T) {
	cp, _ := realControlPlane(t, &policy.Policy{
		ID:      "invalid-policy",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "ship", Action: "run"},
		Default: policy.DecisionAllow,
		Rules: []policy.PolicyRule{{
			RuleID:    "invalid",
			Condition: "",
			Decision:  policy.DecisionAllow,
			Priority:  1,
		}},
	})
	sb := &countingSandbox{}
	gw := gateway.New(cp, sb)

	result, err := gw.Authorize(gateway.ToolRequest{
		Tool:    "ship",
		Action:  "run",
		Context: map[string]any{"project": "naeos"},
	})
	if err == nil {
		t.Fatal("expected evaluator failure to be surfaced as an internal error")
	}
	var evaluationFailure *control.EvaluationFailure
	if !errors.As(err, &evaluationFailure) {
		t.Fatalf("expected typed EvaluationFailure, got %T: %v", err, err)
	}
	if evaluationFailure.PolicyID != "invalid-policy" || evaluationFailure.RuleID != "invalid" {
		t.Fatalf("expected policy and rule attribution, got %+v", evaluationFailure)
	}
	if result.Decision != "" || result.Status != "failed" {
		t.Fatalf("evaluation failure must remain a non-decision failed state, got decision=%q status=%q", result.Decision, result.Status)
	}
	if result.EvaluationFailure == nil || result.EvaluationFailure.Code != evaluationFailure.Code {
		t.Fatalf("expected structured failure metadata in execution result, got %+v", result.EvaluationFailure)
	}
	if got := atomic.LoadInt32(&sb.calls); got != 0 {
		t.Fatalf("side effect executed after evaluator failure: sandbox calls=%d", got)
	}
	history := gw.History()
	if len(history) != 1 || history[0].Decision != "" || history[0].Status != "failed" {
		t.Fatalf("expected one failed non-decision history record, got %+v", history)
	}
}

type mutatingControlPlane struct {
	inner *control.ControlPlane
	reg   *policy.Registry
}

func (m *mutatingControlPlane) Evaluate(req control.Request) (control.DecisionRecord, error) {
	rec, err := m.inner.Evaluate(req)
	if err != nil {
		return rec, err
	}
	// Simulate a policy change after the first authorization decision.
	_ = m.reg.Register(&policy.Policy{
		ID:      "deploy-policy",
		Version: "2.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run"},
		Default: policy.DecisionDeny,
	})
	return rec, nil
}

func (m *mutatingControlPlane) ValidateDecision(req control.Request, issued control.DecisionRecord) (control.DecisionRecord, error) {
	return m.inner.ValidateDecision(req, issued)
}

func TestCanonicalControlPlaneRevalidationBlocksPolicyChange(t *testing.T) {
	cp, reg := realControlPlane(t, &policy.Policy{
		ID:      "deploy-policy",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run"},
		Default: policy.DecisionAllow,
	})
	sb := &countingSandbox{}
	gw := gateway.New(&mutatingControlPlane{inner: cp, reg: reg}, sb)

	result, err := gw.Authorize(gateway.ToolRequest{Tool: "deploy", Action: "run"})
	if err != nil {
		t.Fatalf("unexpected gateway error: %v", err)
	}
	if result.Status != "denied" {
		t.Fatalf("expected revalidation to block execution after policy change, got status=%s decision=%s", result.Status, result.Decision)
	}
	if got := atomic.LoadInt32(&sb.calls); got != 0 {
		t.Fatalf("side effect executed after policy change: sandbox calls=%d", got)
	}
}

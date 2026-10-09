// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"fmt"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

type stubControlPlane struct {
	decision control.Decision
	policyID string
	ruleID   string
}

func (s *stubControlPlane) Evaluate(req control.Request) (control.DecisionRecord, error) {
	return control.DecisionRecord{
		Request:  req,
		Decision: s.decision,
		PolicyID: s.policyID,
		RuleID:   s.ruleID,
		Reasons:  []string{fmt.Sprintf("stub: %s", s.decision)},
	}, nil
}

type stubSandbox struct {
	output string
	err    error
}

func (s *stubSandbox) Execute(req ToolRequest) (string, error) {
	return s.output, s.err
}

type stubAdapter struct {
	name       string
	normalized ToolRequest
	normErr    error
	decisions  []ExecutionResult
}

func (a *stubAdapter) Name() string { return a.name }

func (a *stubAdapter) NormalizeTool(raw any) (ToolRequest, error) {
	return a.normalized, a.normErr
}

func (a *stubAdapter) OnDecision(result ExecutionResult) error {
	a.decisions = append(a.decisions, result)
	return nil
}

func TestAdapterPrivilegeBoundaryDeniesWithoutGrant(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow}
	sb := &countingSandbox{}
	gw := New(cp, sb)
	adapter := JSONToolAdapter{}
	gw.RegisterAdapter(adapter.Name(), adapter)

	result, err := gw.AuthorizeFromAdapter(adapter.Name(), map[string]any{
		"request_id": "req-p25-deny",
		"tool":       "filesystem",
		"action":     "write",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "denied" {
		t.Fatalf("expected denial, got %q", result.Status)
	}
	if sb.Count() != 0 {
		t.Fatalf("adapter without grant reached sandbox: %d calls", sb.Count())
	}
	if result.RequestID != "req-p25-deny" {
		t.Fatalf("request_id not preserved: %q", result.RequestID)
	}
	history := gw.History()
	if len(history) != 1 || history[0].RequestID != "req-p25-deny" || history[0].Status != "denied" {
		t.Fatalf("adapter denial was not recorded in gateway history: %+v", history)
	}
}

func TestAdapterPrivilegeBoundaryAllowsExplicitGrant(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow}
	sb := &countingSandbox{}
	gw := New(cp, sb)
	adapter := JSONToolAdapter{}
	gw.RegisterAdapter(adapter.Name(), adapter)
	if err := gw.GrantAdapterPolicy(adapter.Name(), AdapterPolicy{AllowedTools: []string{"filesystem"}, AllowedActions: []string{"write"}}); err != nil {
		t.Fatalf("grant adapter policy: %v", err)
	}

	result, err := gw.AuthorizeFromAdapter(adapter.Name(), map[string]any{"request_id": "req-p25-allow", "tool": "filesystem", "action": "write"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Fatalf("expected completion, got %q", result.Status)
	}
	if sb.Count() != 1 {
		t.Fatalf("expected one sandbox call, got %d", sb.Count())
	}
}

func TestGatewayAllowExecution(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{
		Tool:   "shell",
		Action: "run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Fatalf("expected completed, got %s", result.Status)
	}
	if result.Decision != control.DecisionAllow {
		t.Fatalf("expected ALLOW, got %s", result.Decision)
	}
	if result.Hash == "" {
		t.Fatal("expected hash to be set")
	}
}

func TestGatewayDenyExecution(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionDeny, policyID: "p1"}
	sb := &stubSandbox{output: "should not run"}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{
		Tool:   "deploy",
		Action: "run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "denied" {
		t.Fatalf("expected denied, got %s", result.Status)
	}
}

func TestGatewayApprovalRequired(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionRequireApproval, policyID: "p1"}
	sb := &stubSandbox{output: "should not run"}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{
		Tool:   "deploy",
		Action: "run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "denied" {
		t.Fatalf("expected denied (approval required), got %s", result.Status)
	}
}

func TestGatewayRestrictionDeniesAllowedPolicy(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	gw.AddRestriction(Restriction{
		Tool:   "rm",
		Reason: "destructive operations blocked",
	})

	result, err := gw.Authorize(ToolRequest{
		Tool:   "rm",
		Action: "run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "denied" {
		t.Fatalf("expected denied due to restriction, got %s", result.Status)
	}
}

func TestGatewayRestrictionWildcardMatch(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	gw.AddRestriction(Restriction{
		Tool:        "shell*",
		Environment: "production",
		Reason:      "shell blocked in prod",
	})

	// Should match: shell in production
	result, err := gw.Authorize(ToolRequest{
		Tool:        "shell",
		Action:      "run",
		Environment: "production",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "denied" {
		t.Fatalf("expected denied, got %s", result.Status)
	}

	// Should not match: shell in staging
	result, err = gw.Authorize(ToolRequest{
		Tool:        "shell",
		Action:      "run",
		Environment: "staging",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Fatalf("expected completed (different env), got %s", result.Status)
	}
}

func TestGatewayAdapterIntegration(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "adapter output"}
	gw := New(cp, sb)

	adapter := &stubAdapter{
		name: "claude",
		normalized: ToolRequest{
			Tool:   "file-edit",
			Action: "write",
		},
	}
	gw.RegisterAdapter("claude", adapter)
	if err := gw.GrantAdapterPolicy("claude", AdapterPolicy{AllowedTools: []string{"*"}, AllowedActions: []string{"*"}}); err != nil {
		t.Fatal(err)
	}

	result, err := gw.AuthorizeFromAdapter("claude", map[string]any{"path": "foo.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Fatalf("expected completed, got %s", result.Status)
	}
	if len(adapter.decisions) != 1 {
		t.Fatalf("expected adapter to receive 1 decision, got %d", len(adapter.decisions))
	}
}

func TestGatewayAdapterNotFound(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	_, err := gw.AuthorizeFromAdapter("unknown", nil)
	if err == nil {
		t.Fatal("expected error for unknown adapter")
	}
}

func TestGatewayHistory(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	gw.Authorize(ToolRequest{Tool: "t1", Action: "a1"})
	gw.Authorize(ToolRequest{Tool: "t2", Action: "a2"})

	history := gw.History()
	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}
}

func TestGatewayDenials(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionDeny, policyID: "p1"}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	gw.Authorize(ToolRequest{Tool: "t1", Action: "a1"})
	gw.Authorize(ToolRequest{Tool: "t2", Action: "a2"})

	denials := gw.Denials()
	if len(denials) != 2 {
		t.Fatalf("expected 2 denials, got %d", len(denials))
	}
}

func TestGatewayMissingTool(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	_, err := gw.Authorize(ToolRequest{Action: "run"})
	if err == nil {
		t.Fatal("expected error for missing tool")
	}
}

func TestGatewaySandboxFailure(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{err: fmt.Errorf("sandbox crash")}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{Tool: "shell", Action: "run"})
	if err == nil {
		t.Fatal("expected error from sandbox failure")
	}
	if result.Status != "failed" {
		t.Fatalf("expected failed, got %s", result.Status)
	}
}

func TestGatewayFailOpen(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{err: fmt.Errorf("sandbox crash")}
	gw := New(cp, sb, FailClosed(false))

	result, err := gw.Authorize(ToolRequest{Tool: "shell", Action: "run"})
	if err != nil {
		t.Fatalf("expected no error in fail-open mode, got %v", err)
	}
	if result.Status != "failed" {
		t.Fatalf("expected failed, got %s", result.Status)
	}
}

func TestGatewayDeterministicDecision(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	// Same input must produce the same decision.
	for i := 0; i < 10; i++ {
		result, err := gw.Authorize(ToolRequest{
			Tool:   "deploy",
			Action: "run",
			Context: map[string]any{
				"version": "1.0.0",
			},
		})
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if result.Decision != control.DecisionAllow {
			t.Fatalf("iteration %d: expected ALLOW, got %s", i, result.Decision)
		}
		if result.PolicyID != "p1" {
			t.Fatalf("iteration %d: expected policy p1, got %s", i, result.PolicyID)
		}
	}
}

func TestGatewayTimestampOrdering(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	for i := 0; i < 5; i++ {
		gw.Authorize(ToolRequest{Tool: fmt.Sprintf("t%d", i), Action: "run"})
	}

	history := gw.History()
	for i := 1; i < len(history); i++ {
		if history[i].Timestamp.Before(history[i-1].Timestamp) {
			t.Fatalf("history not in timestamp order at index %d", i)
		}
	}
}

func TestGatewayPolicyMetadata(t *testing.T) {
	cp := &stubControlPlane{
		decision: control.DecisionDeny,
		policyID: "production-deploy",
		ruleID:   "require-approval",
	}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{Tool: "deploy", Action: "run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PolicyID != "production-deploy" {
		t.Fatalf("expected policy production-deploy, got %s", result.PolicyID)
	}
	if result.RuleID != "require-approval" {
		t.Fatalf("expected rule require-approval, got %s", result.RuleID)
	}
	if len(result.Reasons) == 0 {
		t.Fatal("expected reasons to be populated")
	}
}

func TestGatewayContextInjection(t *testing.T) {
	// Verify that the gateway passes context to the control plane.
	var captured control.Request
	cp := &captureCP{capture: &captured, decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	gw.Authorize(ToolRequest{
		Capability:  "deploy.execute",
		Tool:        "deploy",
		Action:      "run",
		Resource:    "app",
		Environment: "production",
		Actor:       "ci-bot",
		Context:     map[string]any{"version": "2.0.0"},
	})

	if captured.Capability != "deploy.execute" {
		t.Fatalf("expected capability deploy.execute, got %s", captured.Capability)
	}
	if captured.Resource != "app" {
		t.Fatalf("expected resource app, got %s", captured.Resource)
	}
	if captured.Environment != "production" {
		t.Fatalf("expected environment production, got %s", captured.Environment)
	}
	if captured.Actor != "ci-bot" {
		t.Fatalf("expected actor ci-bot, got %s", captured.Actor)
	}
	if v, ok := captured.Context["version"]; !ok || v != "2.0.0" {
		t.Fatalf("expected context version 2.0.0, got %v", v)
	}
}

type captureCP struct {
	capture  *control.Request
	decision control.Decision
	policyID string
}

func (c *captureCP) Evaluate(req control.Request) (control.DecisionRecord, error) {
	*c.capture = req
	return control.DecisionRecord{
		Request:  req,
		Decision: c.decision,
		PolicyID: c.policyID,
	}, nil
}

func TestGatewayRestriction(t *testing.T) {
	r := Restriction{Tool: "rm", Reason: "blocked"}
	if !r.Matches(ToolRequest{Tool: "rm"}) {
		t.Fatal("expected match")
	}
	if r.Matches(ToolRequest{Tool: "ls"}) {
		t.Fatal("expected no match")
	}
}

func TestGatewayRestrictionWildcard(t *testing.T) {
	r := Restriction{Tool: "shell*", Environment: "production"}
	if !r.Matches(ToolRequest{Tool: "shell-run", Environment: "production"}) {
		t.Fatal("expected wildcard match")
	}
	if r.Matches(ToolRequest{Tool: "shell-run", Environment: "staging"}) {
		t.Fatal("expected no match on env")
	}
}

func TestGatewayDurationTracked(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{Tool: "t", Action: "a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Duration < 0 {
		t.Fatal("expected non-negative duration")
	}
}

func TestGatewayHistoryConcurrent(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "ok"}
	gw := New(cp, sb)

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(n int) {
			gw.Authorize(ToolRequest{Tool: fmt.Sprintf("t%d", n), Action: "a"})
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 10; i++ {
		<-done
	}

	history := gw.History()
	if len(history) != 10 {
		t.Fatalf("expected 10 history entries, got %d", len(history))
	}
}

type mutatingControlPlane struct {
	inner    *control.ControlPlane
	registry *policy.Registry
	mutated  bool
}

func (m *mutatingControlPlane) Evaluate(req control.Request) (control.DecisionRecord, error) {
	rec, err := m.inner.Evaluate(req)
	if err != nil {
		return rec, err
	}
	if !m.mutated {
		m.mutated = true
		if err := m.registry.Register(&policy.Policy{
			ID:      "deploy",
			Name:    "Deploy Policy",
			Version: "2.0.0",
			Scope:   policy.Scope{},
			Default: policy.DecisionDeny,
			Active:  true,
		}); err != nil {
			return control.DecisionRecord{}, err
		}
	}
	return rec, nil
}

func (m *mutatingControlPlane) ValidateDecision(req control.Request, issued control.DecisionRecord) (control.DecisionRecord, error) {
	return m.inner.ValidateDecision(req, issued)
}

func TestGatewayPolicyMutationInvalidatesAuthorization(t *testing.T) {
	reg := policy.NewRegistry()
	if err := reg.Register(&policy.Policy{
		ID:      "deploy",
		Name:    "Deploy Policy",
		Version: "1.0.0",
		Scope:   policy.Scope{},
		Default: policy.DecisionAllow,
		Active:  true,
	}); err != nil {
		t.Fatalf("register initial policy: %v", err)
	}

	cp := &mutatingControlPlane{
		inner:    control.New(reg),
		registry: reg,
	}
	sb := &stubSandbox{output: "MUST NOT EXECUTE"}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{
		Tool:   "deploy",
		Action: "run",
	})
	if err != nil {
		t.Fatalf("unexpected authorization error: %v", err)
	}
	if result.Status != "denied" {
		t.Fatalf("expected policy mutation to invalidate authorization, got %s", result.Status)
	}
	if result.Decision != control.DecisionAllow {
		t.Fatalf("expected original decision to remain observable in audit result, got %s", result.Decision)
	}
	if len(gw.History()) != 1 {
		t.Fatalf("expected one recorded invalidated execution attempt, got %d", len(gw.History()))
	}
}

func TestAuthorizationBindsCapability(t *testing.T) {
	reg := policy.NewRegistry()
	if err := reg.Register(&policy.Policy{
		ID:      "filesystem",
		Name:    "Filesystem Policy",
		Version: "1.0.0",
		Scope:   policy.Scope{},
		Default: policy.DecisionAllow,
		Active:  true,
	}); err != nil {
		t.Fatalf("register policy: %v", err)
	}

	cp := control.New(reg)
	issued, err := cp.Evaluate(control.Request{
		Capability: "filesystem.read",
		Resource:   "filesystem",
		Action:     "read",
	})
	if err != nil {
		t.Fatalf("initial authorization failed: %v", err)
	}
	if issued.Decision != control.DecisionAllow {
		t.Fatalf("expected initial read capability to be allowed, got %s", issued.Decision)
	}

	// Reusing the read authorization for a broader write capability must fail
	// even when the same policy would independently allow a fresh request.
	_, err = cp.ValidateDecision(control.Request{
		Capability: "filesystem.write",
		Resource:   "filesystem",
		Action:     "write",
	}, issued)
	if err == nil {
		t.Fatal("expected capability escalation to invalidate the original authorization")
	}
}

// Ensure the full integration path with real policy/control works.
func TestGatewayFullIntegration(t *testing.T) {
	reg := policy.NewRegistry()
	reg.Register(&policy.Policy{
		ID:      "deploy",
		Name:    "Deploy Policy",
		Version: "1.0.0",
		Scope:   policy.Scope{}, // wildcard: matches any request
		Default: policy.DecisionAllow,
		Rules: []policy.PolicyRule{
			{
				RuleID:    "require-env",
				Condition: "not_empty:environment",
				Decision:  policy.DecisionRequireApproval,
				Priority:  10,
			},
		},
		Active: true,
	})

	cp := control.New(reg)
	sb := NewDefaultSandbox(SandboxConfig{})
	gw := New(cp, sb)

	// When environment is not set, the rule condition fails (empty string
	// in context), so the control plane issues DENY.
	result, err := gw.Authorize(ToolRequest{
		Tool:   "deploy",
		Action: "run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != control.DecisionDeny {
		t.Fatalf("expected DENY when rule fails, got %s", result.Decision)
	}
	if result.Status != "denied" {
		t.Fatalf("expected denied, got %s", result.Status)
	}

	// When environment is provided, the rule passes (not_empty), so the
	// policy decision is REQUIRE_APPROVAL (from the passing rule).
	result, err = gw.Authorize(ToolRequest{
		Tool:        "deploy",
		Action:      "run",
		Environment: "staging",
		Context:     map[string]any{"environment": "staging"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != control.DecisionRequireApproval {
		t.Fatalf("expected REQUIRE_APPROVAL, got %s", result.Decision)
	}
	if result.Status != "denied" {
		t.Fatalf("expected denied (approval required), got %s", result.Status)
	}

	// When environment is set and rule is disabled, the policy default
	// (ALLOW) is used. We register a new policy with no rules.
	reg2 := policy.NewRegistry()
	reg2.Register(&policy.Policy{
		ID:      "deploy2",
		Name:    "Deploy Policy 2",
		Version: "1.0.0",
		Scope:   policy.Scope{},
		Default: policy.DecisionAllow,
		Active:  true,
	})

	cp2 := control.New(reg2)
	gw2 := New(cp2, sb)

	result, err = gw2.Authorize(ToolRequest{
		Tool:        "deploy",
		Action:      "run",
		Environment: "staging",
		Context:     map[string]any{"environment": "staging"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != control.DecisionAllow {
		t.Fatalf("expected ALLOW (default), got %s", result.Decision)
	}
	if result.Status != "completed" {
		t.Fatalf("expected completed, got %s", result.Status)
	}
}

type rotatingControlPlane struct {
	decisions []control.Decision
	calls     int
}

func (c *rotatingControlPlane) Evaluate(req control.Request) (control.DecisionRecord, error) {
	decision := c.decisions[c.calls]
	c.calls++
	return control.DecisionRecord{
		Request:       req,
		Decision:      decision,
		PolicyID:      "rotating-policy",
		PolicyVersion: fmt.Sprintf("1.0.%d", c.calls),
		RuleID:        "fresh-decision",
		Reasons:       []string{"decision evaluated at authorization time"},
	}, nil
}

func TestGatewayAuthorizationReplayRequiresFreshDecision(t *testing.T) {
	cp := &rotatingControlPlane{decisions: []control.Decision{
		control.DecisionAllow,
		control.DecisionDeny,
	}}
	sb := &stubSandbox{output: "side effect"}
	gw := New(cp, sb)

	first, err := gw.Authorize(ToolRequest{
		Tool:   "filesystem",
		Action: "write",
	})
	if err != nil {
		t.Fatalf("first authorization failed: %v", err)
	}
	if first.Decision != control.DecisionAllow || first.Status != "completed" {
		t.Fatalf("expected first request to complete under ALLOW, got decision=%s status=%s", first.Decision, first.Status)
	}

	// A previously returned ALLOW must not be reusable as authority for a
	// subsequent request. The gateway has no API that accepts a stale
	// DecisionRecord; it must consult the control plane again.
	second, err := gw.Authorize(ToolRequest{
		Tool:   "filesystem",
		Action: "write",
	})
	if err != nil {
		t.Fatalf("second authorization failed unexpectedly: %v", err)
	}
	if second.Decision != control.DecisionDeny {
		t.Fatalf("expected fresh authorization to return DENY, got %s", second.Decision)
	}
	if second.Status != "denied" {
		t.Fatalf("expected replay attempt to be denied before execution, got %s", second.Status)
	}
	if cp.calls != 2 {
		t.Fatalf("expected control plane to be evaluated twice, got %d calls", cp.calls)
	}
}

type replayCountingSandbox struct {
	count int
}

func (s *replayCountingSandbox) Execute(req ToolRequest) (string, error) {
	s.count++
	return "executed", nil
}

func TestGatewayReplayProtectionRejectsDuplicateInvocation(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &replayCountingSandbox{}
	gw := New(cp, sb, WithReplayProtection(true))

	first, err := gw.Authorize(ToolRequest{
		InvocationID: "inv-001",
		Tool:         "filesystem",
		Action:       "write",
	})
	if err != nil {
		t.Fatalf("first authorization failed: %v", err)
	}
	if first.Status != "completed" {
		t.Fatalf("expected first invocation to complete, got %s", first.Status)
	}

	second, err := gw.Authorize(ToolRequest{
		InvocationID: "inv-001",
		Tool:         "filesystem",
		Action:       "write",
	})
	if err != nil {
		t.Fatalf("replay authorization failed unexpectedly: %v", err)
	}
	if second.Status != "denied" {
		t.Fatalf("expected replay to be denied, got %s", second.Status)
	}
	if second.Output != "invocation already consumed" {
		t.Fatalf("expected replay reason, got %q", second.Output)
	}
	if sb.count != 1 {
		t.Fatalf("expected exactly one sandbox execution, got %d", sb.count)
	}
}

func TestGatewayReplayProtectionRequiresInvocationID(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &replayCountingSandbox{}
	gw := New(cp, sb, WithReplayProtection(true))

	_, err := gw.Authorize(ToolRequest{
		Tool:   "filesystem",
		Action: "write",
	})
	if err == nil {
		t.Fatal("expected missing invocation identity to fail closed")
	}
	if sb.count != 0 {
		t.Fatalf("expected no sandbox execution, got %d", sb.count)
	}
}

func TestGatewayReplayProtectionAllowsDistinctInvocations(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &replayCountingSandbox{}
	gw := New(cp, sb, WithReplayProtection(true))

	for _, id := range []string{"inv-001", "inv-002"} {
		result, err := gw.Authorize(ToolRequest{
			InvocationID: id,
			Tool:         "filesystem",
			Action:       "write",
		})
		if err != nil {
			t.Fatalf("invocation %s failed: %v", id, err)
		}
		if result.Status != "completed" {
			t.Fatalf("invocation %s expected completed, got %s", id, result.Status)
		}
	}
	if sb.count != 2 {
		t.Fatalf("expected two sandbox executions, got %d", sb.count)
	}
}

func TestInMemoryInvocationStoreConcurrentClaim(t *testing.T) {
	store := NewInMemoryInvocationStore()
	const attempts = 32
	results := make(chan bool, attempts)
	errs := make(chan error, attempts)

	for i := 0; i < attempts; i++ {
		go func() {
			claimed, err := store.Claim("concurrent-invocation")
			results <- claimed
			errs <- err
		}()
	}

	claimedCount := 0
	for i := 0; i < attempts; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("claim returned error: %v", err)
		}
		if <-results {
			claimedCount++
		}
	}
	if claimedCount != 1 {
		t.Fatalf("expected exactly one successful claim, got %d", claimedCount)
	}
}

func TestGatewayReplayStoreFailureFailsClosed(t *testing.T) {
	store := &failingInvocationStore{err: fmt.Errorf("store unavailable")}
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &replayCountingSandbox{}
	gw := New(cp, sb, WithReplayProtection(true), WithInvocationStore(store))

	result, err := gw.Authorize(ToolRequest{
		InvocationID: "inv-storage-failure",
		Tool:         "filesystem",
		Action:       "write",
	})
	if err == nil {
		t.Fatal("expected replay store failure to fail closed")
	}
	if result.Status != "denied" {
		t.Fatalf("expected denied result, got %s", result.Status)
	}
	if sb.count != 0 {
		t.Fatalf("expected no sandbox execution, got %d", sb.count)
	}
}

type failingInvocationStore struct {
	err error
}

func (s *failingInvocationStore) Claim(string) (bool, error) {
	return false, s.err
}

func TestGatewayRevalidationDecisionMustBlockExecution(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1", ruleID: "r1"}
	sb := &countingSandbox{}
	gw := New(cp, sb)
	gw.controlPlane = revalidationControlPlane{
		stubControlPlane: cp,
		decision:         control.DecisionDeny,
	}

	result, err := gw.Authorize(ToolRequest{
		RequestID: "req-revalidation-deny",
		Tool:      "filesystem",
		Action:    "write",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "denied" {
		t.Fatalf("expected revalidation denial, got %q", result.Status)
	}
	if result.Decision != control.DecisionDeny {
		t.Fatalf("expected revalidated DENY, got %q", result.Decision)
	}
	if sb.Count() != 0 {
		t.Fatalf("revalidated non-ALLOW decision reached sandbox: %d calls", sb.Count())
	}
}

type revalidationControlPlane struct {
	*stubControlPlane
	decision control.Decision
}

func (c revalidationControlPlane) ValidateDecision(req control.Request, issued control.DecisionRecord) (control.DecisionRecord, error) {
	issued.Decision = c.decision
	issued.Reasons = []string{"revalidation changed decision"}
	return issued, nil
}


// invariantAuditRevalidator lets tests exercise outcomes at the final authorization boundary.
type invariantAuditRevalidator struct {
	*stubControlPlane
	decision control.Decision
	err      error
}

func (c invariantAuditRevalidator) ValidateDecision(_ control.Request, issued control.DecisionRecord) (control.DecisionRecord, error) {
	if c.err != nil {
		return control.DecisionRecord{}, c.err
	}
	issued.Decision = c.decision
	issued.Reasons = []string{"invariant audit revalidation"}
	return issued, nil
}

func TestGatewayRevalidationNonAllowAndErrorFailClosed(t *testing.T) {
	tests := []struct {
		name     string
		decision control.Decision
		err      error
	}{
		{name: "require approval", decision: control.DecisionRequireApproval},
		{name: "unknown decision", decision: control.Decision("UNKNOWN")},
		{name: "revalidation error", decision: control.DecisionAllow, err: fmt.Errorf("policy store unavailable")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
			sb := &countingSandbox{}
			gw := New(cp, sb)
			gw.controlPlane = invariantAuditRevalidator{stubControlPlane: cp, decision: tt.decision, err: tt.err}

			result, err := gw.Authorize(ToolRequest{RequestID: "req-audit-" + tt.name, Tool: "filesystem", Action: "write"})
			if tt.err != nil && err != nil {
				// Fail-closed may surface the revalidation failure as an error.
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Status != "denied" {
				t.Fatalf("expected fail-closed denial, got status=%q err=%v", result.Status, err)
			}
			if sb.Count() != 0 {
				t.Fatalf("non-ALLOW/error revalidation reached sandbox: %d calls", sb.Count())
			}
		})
	}
}

type invariantAuditObserver struct {
	observation Observation
	err         error
}

func (o invariantAuditObserver) Observe(ToolRequest, ExecutionResult) (Observation, error) {
	return o.observation, o.err
}

func TestGatewayObserverFailuresCannotProduceSuccess(t *testing.T) {
	tests := []struct {
		name        string
		observation Observation
		err         error
	}{
		{name: "not observed", observation: Observation{Status: "absent", Observed: false}},
		{name: "mismatched request identity", observation: Observation{RequestID: "other-request", Status: "observed", Observed: true}},
		{name: "mismatched invocation identity", observation: Observation{RequestID: "req-evidence", InvocationID: "other-invocation", Status: "observed", Observed: true}},
		{name: "observer error", observation: Observation{Status: "observed", Observed: true}, err: fmt.Errorf("observer unavailable")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
			sb := &countingSandbox{}
			obs := tt.observation
			req := ToolRequest{RequestID: "req-evidence", InvocationID: "inv-evidence", Tool: "filesystem", Action: "write"}
			if tt.name == "mismatched request identity" {
				obs.InvocationID = req.InvocationID
			} else if tt.name == "mismatched invocation identity" {
				obs.RequestID = req.RequestID
			} else if tt.name == "not observed" {
				obs.RequestID = req.RequestID
				obs.InvocationID = req.InvocationID
			} else if tt.name == "observer error" {
				obs.RequestID = req.RequestID
				obs.InvocationID = req.InvocationID
			}
			obs.InvocationDigest = InvocationDigest(req)
			gw := New(cp, sb, WithObserver(invariantAuditObserver{observation: obs, err: tt.err}))

			result, err := gw.Authorize(req)
			if err == nil {
				t.Fatalf("expected observer failure to fail closed")
			}
			if result.Status != "failed" {
				t.Fatalf("expected failed result, got %q", result.Status)
			}
			if result.Observation == nil {
				t.Fatal("expected observer result to be retained for audit")
			}
		})
	}
}

func TestGatewaySandboxOutputAloneIsNotObservationEvidence(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "claimed side effect succeeded"}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{RequestID: "req-no-observer", InvocationID: "inv-no-observer", Tool: "filesystem", Action: "write"})
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if result.Status != "completed" {
		t.Fatalf("expected execution status to reflect sandbox completion, got %q", result.Status)
	}
	if result.Observation != nil {
		t.Fatal("sandbox output must not manufacture an independent observation")
	}
}

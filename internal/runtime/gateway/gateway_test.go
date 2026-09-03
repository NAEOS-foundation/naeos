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
		Tool:        "deploy",
		Action:      "run",
		Resource:    "app",
		Environment: "production",
		Actor:       "ci-bot",
		Context:     map[string]any{"version": "2.0.0"},
	})

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

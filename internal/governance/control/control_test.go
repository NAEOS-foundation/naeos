package control

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

func newTestPlane(t *testing.T, failClosed bool, policies ...*policy.Policy) *ControlPlane {
	t.Helper()
	reg := policy.NewRegistry()
	for _, p := range policies {
		if err := reg.Register(p); err != nil {
			t.Fatalf("register policy: %v", err)
		}
	}
	return New(reg, FailClosed(failClosed))
}

func TestEvaluateNoPolicyFailClosed(t *testing.T) {
	c := newTestPlane(t, true)
	rec, err := c.Evaluate(Request{Resource: "deploy", Action: "run", Context: map[string]any{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Decision != DecisionDeny {
		t.Fatalf("expected DENY (fail closed), got %s", rec.Decision)
	}
	if !rec.Deterministic {
		t.Fatal("expected deterministic flag")
	}
}

func TestEvaluateNoPolicyFailOpen(t *testing.T) {
	c := newTestPlane(t, false)
	rec, _ := c.Evaluate(Request{Resource: "deploy", Action: "run"})
	if rec.Decision != DecisionAllow {
		t.Fatalf("expected ALLOW (fail open), got %s", rec.Decision)
	}
}

func TestEvaluateMissingArguments(t *testing.T) {
	c := newTestPlane(t, true)
	if _, err := c.Evaluate(Request{Action: "run"}); err == nil {
		t.Fatal("expected error when resource missing")
	}
	if _, err := c.Evaluate(Request{Resource: "deploy"}); err == nil {
		t.Fatal("expected error when action missing")
	}
}

func TestEvaluateApprovalRequired(t *testing.T) {
	p := &policy.Policy{
		ID:      "prod-deploy",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run", Environment: "production"},
		Default: policy.DecisionRequireApproval,
	}
	c := newTestPlane(t, true, p)

	rec, err := c.Evaluate(Request{Resource: "deploy", Action: "run", Environment: "production"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Decision != DecisionRequireApproval {
		t.Fatalf("expected REQUIRE_APPROVAL, got %s", rec.Decision)
	}
	if rec.PolicyID != "prod-deploy" {
		t.Fatalf("expected policy prod-deploy, got %s", rec.PolicyID)
	}
}

func TestEvaluateScopeOutOfEnvironment(t *testing.T) {
	p := &policy.Policy{
		ID:      "prod-deploy",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run", Environment: "production"},
		Default: policy.DecisionRequireApproval,
	}
	c := newTestPlane(t, true, p)

	// Same action but staging environment => no matching policy => deny.
	rec, _ := c.Evaluate(Request{Resource: "deploy", Action: "run", Environment: "staging"})
	if rec.Decision != DecisionDeny {
		t.Fatalf("expected DENY for non-matching scope, got %s", rec.Decision)
	}
}

func TestEvaluateDenyRuleOverridesDefault(t *testing.T) {
	p := &policy.Policy{
		ID:      "tag-tls",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "network_connection"},
		Default: policy.DecisionAllow,
		Rules: []policy.PolicyRule{
			{RuleID: "tls-min", Condition: "gte:tls_version,1.3", Decision: policy.DecisionAllow, Priority: 1},
		},
	}
	c := newTestPlane(t, true, p)

	// tls_version 1.2 -> rule fails -> DENY.
	rec, _ := c.Evaluate(Request{
		Resource: "network_connection", Action: "open",
		Context: map[string]any{"tls_version": "1.2"},
	})
	if rec.Decision != DecisionDeny {
		t.Fatalf("expected DENY when rule fails, got %s", rec.Decision)
	}
	if rec.RuleID != "tls-min" {
		t.Fatalf("expected failing rule tls-min, got %s", rec.RuleID)
	}
}

func TestEvaluateHighestPriorityApprovalWins(t *testing.T) {
	// Two policies: one allows deploy-run generally, one requires approval
	// for the narrow production scope. The strictest (approval) must win.
	base := &policy.Policy{
		ID:      "allow-deploy",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run"},
		Default: policy.DecisionAllow,
	}
	prod := &policy.Policy{
		ID:      "approve-prod",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run", Environment: "production"},
		Default: policy.DecisionRequireApproval,
	}
	c := newTestPlane(t, true, base, prod)

	rec, _ := c.Evaluate(Request{Resource: "deploy", Action: "run", Environment: "production"})
	if rec.Decision != DecisionRequireApproval {
		t.Fatalf("expected REQUIRE_APPROVAL (strictest wins), got %s", rec.Decision)
	}
	if rec.PolicyID != "approve-prod" {
		t.Fatalf("expected decision attributed to approve-prod, got %s", rec.PolicyID)
	}
}

func TestEvaluateDeterministic(t *testing.T) {
	p := &policy.Policy{
		ID:      "x",
		Version: "1.0.0",
		Default: policy.DecisionRequireApproval,
	}
	c := newTestPlane(t, true, p)

	req := Request{Resource: "r", Action: "a", Context: map[string]any{"k": "v"}}
	first, _ := c.Evaluate(req)
	second, _ := c.Evaluate(req)
	if first.Decision != second.Decision {
		t.Fatalf("decisions not deterministic: %s vs %s", first.Decision, second.Decision)
	}
}

func TestListDecisions(t *testing.T) {
	c := newTestPlane(t, true)
	if _, err := c.Evaluate(Request{Resource: "r", Action: "a"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Evaluate(Request{Resource: "r2", Action: "b"}); err != nil {
		t.Fatal(err)
	}
	if got := c.ListDecisions(); len(got) != 2 {
		t.Fatalf("expected 2 recorded decisions, got %d", len(got))
	}
}

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package control

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

func newTestPlane(t *testing.T, policies ...*policy.Policy) *ControlPlane {
	return newTestPlaneWithOptions(t, nil, policies...)
}

func newTestPlaneWithOptions(t *testing.T, opts []Option, policies ...*policy.Policy) *ControlPlane {
	t.Helper()
	reg := policy.NewRegistry()
	for _, p := range policies {
		if err := reg.Register(p); err != nil {
			t.Fatalf("register policy: %v", err)
		}
	}
	return New(reg, opts...)
}

func TestEvaluateNoPolicyFailClosed(t *testing.T) {
	c := newTestPlane(t)
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

func TestEvaluateMissingArguments(t *testing.T) {
	c := newTestPlane(t)
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
	c := newTestPlane(t, p)

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
	c := newTestPlane(t, p)

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
	c := newTestPlane(t, p)

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
	c := newTestPlane(t, base, prod)

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
	c := newTestPlane(t, p)

	req := Request{Resource: "r", Action: "a", Context: map[string]any{"k": "v"}}
	first, _ := c.Evaluate(req)
	second, _ := c.Evaluate(req)
	if first.Decision != second.Decision {
		t.Fatalf("decisions not deterministic: %s vs %s", first.Decision, second.Decision)
	}
}

func TestListDecisions(t *testing.T) {
	c := newTestPlane(t)
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

type errorEvaluator struct{}

func (errorEvaluator) Evaluate(map[string]any) error {
	return fmt.Errorf("synthetic evaluator failure")
}

func (errorEvaluator) EvaluateRules([]policy.Rule, map[string]any) ([]policy.EvaluationResult, error) {
	return nil, fmt.Errorf("synthetic evaluator failure")
}

func TestEvaluateEvaluatorErrorFailsClosed(t *testing.T) {
	p := &policy.Policy{
		ID:      "deploy-policy",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run"},
		Default: policy.DecisionAllow,
		Rules: []policy.PolicyRule{
			{RuleID: "must-evaluate", Condition: "exists:project", Decision: policy.DecisionAllow, Priority: 1},
		},
	}
	c := newTestPlane(t, p)
	c.evaluator = errorEvaluator{}

	rec, err := c.Evaluate(Request{Resource: "deploy", Action: "run", Context: map[string]any{"project": "naeos"}})
	if err == nil {
		t.Fatal("expected evaluator failure to remain an internal error")
	}
	if rec.Decision != "" {
		t.Fatalf("evaluation failure must not be represented as a policy decision, got %s", rec.Decision)
	}
	var failure *EvaluationFailure
	if !errors.As(err, &failure) {
		t.Fatalf("expected typed EvaluationFailure, got %T: %v", err, err)
	}
	if failure.PolicyID != "deploy-policy" || failure.PolicyVersion != "1.0.0" || failure.RuleID != "must-evaluate" {
		t.Fatalf("expected policy/rule attribution, got %+v", failure)
	}
	if failure.Stage != "rule_evaluation" || failure.Code != "policy_evaluator_error" || failure.Timestamp.IsZero() {
		t.Fatalf("expected structured failure metadata, got %+v", failure)
	}
	if got := c.ListDecisions(); len(got) != 0 {
		t.Fatalf("evaluation failure must not be recorded as a decision, got %d records", len(got))
	}
}

func TestEvaluateBindsCapabilityGrantDigest(t *testing.T) {
	p := &policy.Policy{
		ID:      "repository-write",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "repository", Action: "write"},
		Default: policy.DecisionAllow,
	}
	c := newTestPlane(t, p)

	rec, err := c.Evaluate(Request{
		Capability: "repository.write",
		Resource:   "repository",
		Action:     "write",
		Actor:      "codex",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.GrantDigest == "" {
		t.Fatal("expected authorization grant digest")
	}
	if !grantDigestMatches(rec) {
		t.Fatal("expected grant digest to bind the issued authorization identity")
	}
}

func TestValidateDecisionRejectsTamperedCapabilityGrant(t *testing.T) {
	p := &policy.Policy{
		ID:      "repository-write",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "repository", Action: "write"},
		Default: policy.DecisionAllow,
	}
	c := newTestPlane(t, p)

	req := Request{
		Capability: "repository.write",
		Resource:   "repository",
		Action:     "write",
		Actor:      "codex",
	}
	issued, err := c.Evaluate(req)
	if err != nil {
		t.Fatalf("issue authorization: %v", err)
	}

	issued.Request.Capability = "repository.admin"
	if _, err := c.ValidateDecision(req, issued); err == nil {
		t.Fatal("expected tampered grant to be rejected")
	}
}

func TestValidateDecisionRejectsTamperedGrantDigest(t *testing.T) {
	p := &policy.Policy{
		ID:      "repository-write",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "repository", Action: "write"},
		Default: policy.DecisionAllow,
	}
	c := newTestPlane(t, p)

	req := Request{
		Capability: "repository.write",
		Resource:   "repository",
		Action:     "write",
		Actor:      "codex",
	}
	issued, err := c.Evaluate(req)
	if err != nil {
		t.Fatalf("issue authorization: %v", err)
	}

	issued.GrantDigest = "tampered"
	if _, err := c.ValidateDecision(req, issued); err == nil {
		t.Fatal("expected tampered grant digest to be rejected")
	}
}

func TestValidateDecisionRejectsExpiredGrant(t *testing.T) {
	p := &policy.Policy{
		ID:      "repository-write",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "repository", Action: "write"},
		Default: policy.DecisionAllow,
	}
	c := newTestPlaneWithOptions(t, []Option{WithGrantTTL(time.Millisecond)}, p)
	req := Request{
		Capability: "repository.write",
		Resource:   "repository",
		Action:     "write",
		Actor:      "codex",
	}
	issued, err := c.Evaluate(req)
	if err != nil {
		t.Fatalf("issue authorization: %v", err)
	}
	if issued.ExpiresAt.IsZero() {
		t.Fatal("expected grant expiry to be issued")
	}
	time.Sleep(5 * time.Millisecond)
	if _, err := c.ValidateDecision(req, issued); err == nil {
		t.Fatal("expected expired grant to be rejected")
	}
}

func TestValidateDecisionRejectsRevokedGrant(t *testing.T) {
	p := &policy.Policy{
		ID:      "repository-write",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "repository", Action: "write"},
		Default: policy.DecisionAllow,
	}
	c := newTestPlane(t, p)
	req := Request{
		Capability: "repository.write",
		Resource:   "repository",
		Action:     "write",
		Actor:      "codex",
	}
	issued, err := c.Evaluate(req)
	if err != nil {
		t.Fatalf("issue authorization: %v", err)
	}
	if err := c.RevokeGrant(issued.GrantDigest); err != nil {
		t.Fatalf("revoke grant: %v", err)
	}
	if _, err := c.ValidateDecision(req, issued); err == nil {
		t.Fatal("expected revoked grant to be rejected")
	}
}

func TestValidateDecisionRejectsTamperedGrantExpiry(t *testing.T) {
	p := &policy.Policy{
		ID:      "repository-write",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "repository", Action: "write"},
		Default: policy.DecisionAllow,
	}
	c := newTestPlaneWithOptions(t, []Option{WithGrantTTL(time.Minute)}, p)
	req := Request{
		Capability: "repository.write",
		Resource:   "repository",
		Action:     "write",
		Actor:      "codex",
	}
	issued, err := c.Evaluate(req)
	if err != nil {
		t.Fatalf("issue authorization: %v", err)
	}
	issued.ExpiresAt = issued.ExpiresAt.Add(time.Minute)
	if _, err := c.ValidateDecision(req, issued); err == nil {
		t.Fatal("expected tampered grant expiry to be rejected")
	}
}

type recordingRuleEvaluator struct {
	calls []string
}

func (e *recordingRuleEvaluator) Evaluate(ctx map[string]any) error {
	return policy.NewEvaluator().Evaluate(ctx)
}

func (e *recordingRuleEvaluator) EvaluateRules(rules []policy.Rule, ctx map[string]any) ([]policy.EvaluationResult, error) {
	if len(rules) > 0 {
		e.calls = append(e.calls, rules[0].RuleID)
	}
	return policy.NewEvaluator().EvaluateRules(rules, ctx)
}

func TestEvaluateRulePriorityControlsOrderButNotDecisionSeverity(t *testing.T) {
	p := &policy.Policy{
		ID: "priority-order", Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run"},
		Default: policy.DecisionAllow,
		Rules: []policy.PolicyRule{
			{RuleID: "approval-late", Condition: "exists:approval_flag", Decision: policy.DecisionRequireApproval, Priority: 4},
			{RuleID: "allow-first", Condition: "exists:allow_flag", Decision: policy.DecisionAllow, Priority: 1},
		},
	}
	c := newTestPlane(t, p)
	evaluator := &recordingRuleEvaluator{}
	c.evaluator = evaluator

	rec, err := c.Evaluate(Request{
		Resource: "deploy", Action: "run",
		Context: map[string]any{"allow_flag": true, "approval_flag": true},
	})
	if err != nil {
		t.Fatalf("unexpected evaluation error: %v", err)
	}
	if rec.Decision != DecisionRequireApproval || rec.RuleID != "approval-late" {
		t.Fatalf("lower-priority ALLOW must not override REQUIRE_APPROVAL: %+v", rec)
	}
	if fmt.Sprint(evaluator.calls) != "[allow-first approval-late]" {
		t.Fatalf("rules must be evaluated by ascending numeric priority: %v", evaluator.calls)
	}
}

func TestEvaluateDenyRuleWinsOverHigherPriorityAllowRule(t *testing.T) {
	p := &policy.Policy{
		ID: "deny-over-allow", Version: "1.0.0",
		Scope:   policy.Scope{Resource: "repository", Action: "write"},
		Default: policy.DecisionAllow,
		Rules: []policy.PolicyRule{
			{RuleID: "allow-first", Condition: "exists:allow_flag", Decision: policy.DecisionAllow, Priority: 1},
			{RuleID: "deny-later", Condition: "exists:deny_flag", Decision: policy.DecisionDeny, Priority: 4},
		},
	}
	c := newTestPlane(t, p)
	rec, err := c.Evaluate(Request{
		Resource: "repository", Action: "write",
		Context: map[string]any{"allow_flag": true, "deny_flag": true},
	})
	if err != nil {
		t.Fatalf("unexpected evaluation error: %v", err)
	}
	if rec.Decision != DecisionDeny || rec.RuleID != "deny-later" {
		t.Fatalf("lower-priority DENY must override ALLOW: %+v", rec)
	}
}

func TestEvaluatePolicyDefaultApprovalCannotBeDowngradedByAllowRule(t *testing.T) {
	p := &policy.Policy{
		ID: "approval-default", Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run"},
		Default: policy.DecisionRequireApproval,
		Rules: []policy.PolicyRule{
			{RuleID: "allow-rule", Condition: "exists:ready", Decision: policy.DecisionAllow, Priority: 1},
		},
	}
	c := newTestPlane(t, p)
	rec, err := c.Evaluate(Request{Resource: "deploy", Action: "run", Context: map[string]any{"ready": true}})
	if err != nil {
		t.Fatalf("unexpected evaluation error: %v", err)
	}
	if rec.Decision != DecisionRequireApproval || rec.RuleID != "" {
		t.Fatalf("rule ALLOW must not downgrade policy default REQUIRE_APPROVAL: %+v", rec)
	}
}

func TestEvaluateDenyWinsAcrossApplicablePolicies(t *testing.T) {
	allow := &policy.Policy{
		ID: "a-allow", Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run"},
		Default: policy.DecisionAllow,
	}
	deny := &policy.Policy{
		ID: "z-deny", Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run"},
		Default: policy.DecisionDeny,
	}
	c := newTestPlane(t, allow, deny)
	rec, err := c.Evaluate(Request{Resource: "deploy", Action: "run"})
	if err != nil {
		t.Fatalf("unexpected evaluation error: %v", err)
	}
	if rec.Decision != DecisionDeny || rec.PolicyID != "z-deny" {
		t.Fatalf("DENY must win across applicable policies: %+v", rec)
	}
}

func TestEvaluateEqualSeverityPolicyAttributionIsRegistrationOrderIndependent(t *testing.T) {
	alpha := &policy.Policy{
		ID: "alpha-policy", Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run"},
		Default: policy.DecisionRequireApproval,
	}
	zeta := &policy.Policy{
		ID: "zeta-policy", Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run"},
		Default: policy.DecisionRequireApproval,
	}
	for _, policies := range [][]*policy.Policy{{alpha, zeta}, {zeta, alpha}} {
		c := newTestPlane(t, policies...)
		rec, err := c.Evaluate(Request{Resource: "deploy", Action: "run"})
		if err != nil {
			t.Fatalf("unexpected evaluation error: %v", err)
		}
		if rec.Decision != DecisionRequireApproval || rec.PolicyID != "alpha-policy" {
			t.Fatalf("equal-severity attribution must be stable across registration order: %+v", rec)
		}
	}
}

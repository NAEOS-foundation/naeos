package policy

import (
	"testing"
)

func TestNewEvaluator(t *testing.T) {
	e := NewEvaluator()
	if e == nil {
		t.Fatal("expected non-nil evaluator")
	}
}

func TestEvaluateNilContext(t *testing.T) {
	e := NewEvaluator()
	err := e.Evaluate(nil)
	if err == nil {
		t.Fatal("expected error for nil context")
	}
}

func TestEvaluateValidContext(t *testing.T) {
	e := NewEvaluator()
	err := e.Evaluate(map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvaluateRulesEmpty(t *testing.T) {
	e := NewEvaluator()
	results, err := e.EvaluateRules([]Rule{}, map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestEvaluateRulesPassing(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "env:production", Priority: 1, Action: "enforce", Enabled: true},
	}
	ctx := map[string]any{"env": "production"}
	results, err := e.EvaluateRules(rules, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Passed {
		t.Fatalf("expected rule to pass, got message: %s", results[0].Message)
	}
}

func TestEvaluateRulesFailing(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "env:production", Priority: 1, Action: "block", Enabled: true},
	}
	ctx := map[string]any{"env": "staging"}
	results, err := e.EvaluateRules(rules, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Passed {
		t.Fatal("expected rule to fail")
	}
}

func TestEvaluateRulesDisabled(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "env:production", Priority: 1, Action: "enforce", Enabled: false},
	}
	results, err := e.EvaluateRules(rules, map[string]any{"env": "production"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for disabled rules, got %d", len(results))
	}
}

func TestEvaluateRulesMissingKey(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "missing_key:value", Priority: 1, Action: "block", Enabled: true},
	}
	results, err := e.EvaluateRules(rules, map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Passed {
		t.Fatal("expected rule to fail for missing key")
	}
}

func TestEvaluateRulesMultiple(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "env:production", Priority: 1, Action: "enforce", Enabled: true},
		{RuleID: "r2", Condition: "version:1.0", Priority: 2, Action: "warn", Enabled: true},
		{RuleID: "r3", Condition: "", Priority: 3, Action: "log", Enabled: true},
	}
	ctx := map[string]any{"env": "production", "version": "1.0"}
	results, err := e.EvaluateRules(rules, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Passed {
			t.Fatalf("expected rule %s to pass", r.RuleID)
		}
	}
}

func TestEvaluateRulesNilContext(t *testing.T) {
	e := NewEvaluator()
	_, err := e.EvaluateRules([]Rule{{RuleID: "r1"}}, nil)
	if err == nil {
		t.Fatal("expected error for nil context")
	}
}

func TestEvaluateExistsRule(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "exists:project", Enabled: true},
	}
	results, _ := e.EvaluateRules(rules, map[string]any{"project": "test"})
	if !results[0].Passed {
		t.Fatalf("expected exists to pass, got: %s", results[0].Message)
	}

	results, _ = e.EvaluateRules(rules, map[string]any{})
	if results[0].Passed {
		t.Fatal("expected exists to fail for missing key")
	}
}

func TestEvaluateNotEmptyRule(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "not_empty:name", Enabled: true},
	}
	results, _ := e.EvaluateRules(rules, map[string]any{"name": "test"})
	if !results[0].Passed {
		t.Fatalf("expected not_empty to pass, got: %s", results[0].Message)
	}

	results, _ = e.EvaluateRules(rules, map[string]any{"name": ""})
	if results[0].Passed {
		t.Fatal("expected not_empty to fail for empty value")
	}
}

func TestEvaluateContainsRule(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "contains:name,test", Enabled: true},
	}
	results, _ := e.EvaluateRules(rules, map[string]any{"name": "testing"})
	if !results[0].Passed {
		t.Fatalf("expected contains to pass, got: %s", results[0].Message)
	}

	results, _ = e.EvaluateRules(rules, map[string]any{"name": "hello"})
	if results[0].Passed {
		t.Fatal("expected contains to fail")
	}
}

func TestEvaluateGtRule(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "gt:port,0", Enabled: true},
	}
	results, _ := e.EvaluateRules(rules, map[string]any{"port": 8080})
	if !results[0].Passed {
		t.Fatalf("expected gt to pass, got: %s", results[0].Message)
	}

	results, _ = e.EvaluateRules(rules, map[string]any{"port": 0})
	if results[0].Passed {
		t.Fatal("expected gt to fail for equal value")
	}
}

func TestEvaluateLtRule(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "lt:count,100", Enabled: true},
	}
	results, _ := e.EvaluateRules(rules, map[string]any{"count": 50})
	if !results[0].Passed {
		t.Fatalf("expected lt to pass, got: %s", results[0].Message)
	}

	results, _ = e.EvaluateRules(rules, map[string]any{"count": 100})
	if results[0].Passed {
		t.Fatal("expected lt to fail for equal value")
	}
}

func TestEvaluateInRule(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "in:env,staging,production", Enabled: true},
	}
	results, _ := e.EvaluateRules(rules, map[string]any{"env": "production"})
	if !results[0].Passed {
		t.Fatalf("expected in to pass, got: %s", results[0].Message)
	}

	results, _ = e.EvaluateRules(rules, map[string]any{"env": "development"})
	if results[0].Passed {
		t.Fatal("expected in to fail for invalid value")
	}
}

func TestEvaluateGteRule(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "gte:version,1.3", Enabled: true},
	}
	results, _ := e.EvaluateRules(rules, map[string]any{"version": 1.3})
	if !results[0].Passed {
		t.Fatalf("expected gte to pass for equal value, got: %s", results[0].Message)
	}
	results, _ = e.EvaluateRules(rules, map[string]any{"version": "1.5"})
	if !results[0].Passed {
		t.Fatalf("expected gte to pass for greater value, got: %s", results[0].Message)
	}
	results, _ = e.EvaluateRules(rules, map[string]any{"version": "1.2"})
	if results[0].Passed {
		t.Fatalf("expected gte to fail for lower value, got: %s", results[0].Message)
	}
}

func TestEvaluateLteRule(t *testing.T) {
	e := NewEvaluator()
	rules := []Rule{
		{RuleID: "r1", Condition: "lte:count,10", Enabled: true},
	}
	results, _ := e.EvaluateRules(rules, map[string]any{"count": 10})
	if !results[0].Passed {
		t.Fatalf("expected lte to pass for equal value, got: %s", results[0].Message)
	}
	results, _ = e.EvaluateRules(rules, map[string]any{"count": 5})
	if !results[0].Passed {
		t.Fatalf("expected lte to pass for lower value, got: %s", results[0].Message)
	}
	results, _ = e.EvaluateRules(rules, map[string]any{"count": 11})
	if results[0].Passed {
		t.Fatalf("expected lte to fail for greater value, got: %s", results[0].Message)
	}
}

func TestDefaultRules(t *testing.T) {
	rules := DefaultRules()
	if len(rules) == 0 {
		t.Fatal("expected default rules to be non-empty")
	}
	for _, r := range rules {
		if r.RuleID == "" {
			t.Fatal("expected rule ID to be non-empty")
		}
		if !r.Enabled {
			t.Fatalf("expected rule %s to be enabled", r.RuleID)
		}
	}
}

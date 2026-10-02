// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
)

// TestDefaultRulesResolveAgainstPolicyContext guards the built-in rules against
// referring to context keys the policy context never produces.
//
// DefaultRules shipped three rules keyed on architecture_pattern,
// deployment_strategy and service_port while PolicyContext exposes the nested
// objects architecture, deployment and services. Every one of those rules could
// only ever fail, and the existing TestDefaultRules never evaluated them, so
// nothing noticed.
func TestDefaultRulesResolveAgainstPolicyContext(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "demo-app"},
		Architecture: &architecture.Architecture{
			Pattern: architecture.PatternHexagonal,
		},
		Deployment: &deployment.Deployment{
			Strategy: deployment.StrategyRolling,
		},
		Modules: []module.Module{
			{Name: "auth", Path: "./auth"},
		},
		Services: []service.Service{
			{Name: "gateway", Port: 8080},
		},
	}

	policyCtx, err := ContextFromNEIR(neir)
	if err != nil {
		t.Fatalf("ContextFromNEIR: %v", err)
	}

	results, err := NewEvaluator().EvaluateRules(DefaultRules(), policyCtx.Values)
	if err != nil {
		t.Fatalf("EvaluateRules: %v", err)
	}
	if len(results) != len(DefaultRules()) {
		t.Fatalf("expected %d results, got %d", len(DefaultRules()), len(results))
	}

	for _, result := range results {
		if !result.Passed {
			t.Errorf("default rule %q must pass for a valid NEIR, got: %s", result.RuleID, result.Message)
		}
	}
}

// TestDefaultRulesRejectInvalidValues checks the rules still have teeth.
func TestDefaultRulesRejectInvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		ruleID  string
		neir    *model.NEIR
		wantErr string
	}{
		{
			name:   "unknown architecture pattern is rejected",
			ruleID: "architecture-pattern-valid",
			neir: &model.NEIR{
				Architecture: &architecture.Architecture{Pattern: "not-a-pattern"},
			},
		},
		{
			name:   "unknown deployment strategy is rejected",
			ruleID: "deployment-strategy-valid",
			neir: &model.NEIR{
				Deployment: &deployment.Deployment{Strategy: "not-a-strategy"},
			},
		},
		{
			name:   "non positive service port is rejected",
			ruleID: "service-port-positive",
			neir: &model.NEIR{
				Services: []service.Service{{Name: "gateway", Port: 0}},
			},
		},
		{
			name:   "missing project is rejected",
			ruleID: "project-required",
			neir:   &model.NEIR{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policyCtx, err := ContextFromNEIR(tc.neir)
			if err != nil {
				t.Fatalf("ContextFromNEIR: %v", err)
			}

			var rule Rule
			for _, r := range DefaultRules() {
				if r.RuleID == tc.ruleID {
					rule = r
					break
				}
			}
			if rule.RuleID == "" {
				t.Fatalf("default rule %q not found", tc.ruleID)
			}

			results, err := NewEvaluator().EvaluateRules([]Rule{rule}, policyCtx.Values)
			if err != nil {
				t.Fatalf("EvaluateRules: %v", err)
			}
			if len(results) != 1 {
				t.Fatalf("expected 1 result, got %d", len(results))
			}
			if results[0].Passed {
				t.Errorf("rule %q must reject an invalid value", tc.ruleID)
			}
		})
	}
}

// TestServicePortRuleRequiresEveryService verifies the fan-out is an all, not an
// any: one bad port among good ones must fail the rule.
func TestServicePortRuleRequiresEveryService(t *testing.T) {
	policyCtx, err := ContextFromNEIR(&model.NEIR{
		Services: []service.Service{
			{Name: "gateway", Port: 8080},
			{Name: "worker", Port: -1},
		},
	})
	if err != nil {
		t.Fatalf("ContextFromNEIR: %v", err)
	}

	results, err := NewEvaluator().EvaluateRules(
		[]Rule{{RuleID: "service-port-positive", Condition: "gt:services.port,0", Enabled: true}},
		policyCtx.Values,
	)
	if err != nil {
		t.Fatalf("EvaluateRules: %v", err)
	}
	if results[0].Passed {
		t.Error("a single non positive port must fail the rule")
	}
}

// TestResolveContextPathFanOut pins the fan-out helper directly.
func TestResolveContextPathFanOut(t *testing.T) {
	ctx := map[string]any{
		"architecture": map[string]any{"pattern": "hexagonal"},
		"services": []any{
			map[string]any{"port": 8080},
			map[string]any{"port": 9090},
		},
		"empty": []any{},
	}

	t.Run("nested object", func(t *testing.T) {
		value, ok := resolveContextPath(ctx, "architecture.pattern")
		if !ok || value != "hexagonal" {
			t.Fatalf("expected hexagonal, got %v (ok=%v)", value, ok)
		}
	})

	t.Run("fan out over slice", func(t *testing.T) {
		value, ok := resolveContextPath(ctx, "services.port")
		if !ok {
			t.Fatal("expected services.port to resolve")
		}
		values := contextValues(value)
		if len(values) != 2 || fmt.Sprint(values...) != "8080 9090" {
			t.Fatalf("expected both ports, got %v", values)
		}
	})

	t.Run("flat key still resolves", func(t *testing.T) {
		if _, ok := resolveContextPath(ctx, "architecture"); !ok {
			t.Error("flat key must still resolve")
		}
	})

	t.Run("missing path fails closed", func(t *testing.T) {
		if _, ok := resolveContextPath(ctx, "architecture.missing"); ok {
			t.Error("missing nested field must not resolve")
		}
	})

	t.Run("field absent from every element fails closed", func(t *testing.T) {
		if _, ok := resolveContextPath(ctx, "services.missing"); ok {
			t.Error("fan-out matching nothing must not resolve")
		}
	})

	t.Run("path through empty slice fails closed", func(t *testing.T) {
		if _, ok := resolveContextPath(ctx, "empty.field"); ok {
			t.Error("path through empty slice must not resolve")
		}
	})
}

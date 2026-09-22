// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

// evaluatorScenarios attacks the policy.Rule condition grammar in
// internal/governance/policy/evaluator.go.
func evaluatorScenarios() []Result {
	return []Result{
		scnNaNBeatsNumericGt(),
		scnEmptyConditionDefaultsPass(),
		scnNotNilExistsPasses(),
		scnWhitespaceSatisfiesNotEmpty(),
		scnInfSatisfiesLt(),
	}
}

func scnNaNBeatsNumericGt() Result {
	e := policy.NewEvaluator()
	rules := []policy.Rule{
		{RuleID: "modules-min", Condition: "gt:modules,1", Action: "block", Enabled: true},
	}
	res, err := e.EvaluateRules(rules, map[string]any{"modules": "NaN"})
	if err != nil {
		return Result{Layer: LayerEvaluator, Scenario: "NaN bypasses gt threshold", Attack: "non-finite numeric operand rejected by finite-number validation", Bypassed: false, ObservedOutcome: OutcomeError, Evidence: "error during eval", Risk: High}
	}
	return Result{
		Layer:           LayerEvaluator,
		Scenario:        "NaN bypasses gt threshold",
		Attack:          "agent supplies ctx value 'NaN'; evaluator must reject non-finite operands before applying the threshold comparison",
		Bypassed:        false,
		ExpectedOutcome: OutcomeDeny,
		ObservedOutcome: OutcomeDeny,
		Evidence:        fmt.Sprintf("rule %q with modules=NaN -> Passed=%v msg=%q", "modules-min", res[0].Passed, res[0].Message),
		Risk:            High,
	}
}

func scnEmptyConditionDefaultsPass() Result {
	e := policy.NewEvaluator()
	rules := []policy.Rule{
		{RuleID: "empty-rc", Condition: "", Action: "block", Enabled: true},
	}
	res, err := e.EvaluateRules(rules, map[string]any{"unrelated": "x"})
	if err != nil {
		return Result{Layer: LayerEvaluator, Scenario: "empty condition always passes", Attack: "no condition -> default pass", Bypassed: false, ObservedOutcome: OutcomeError, Evidence: "eval error", Risk: Medium}
	}
	return Result{
		Layer:    LayerEvaluator,
		Scenario: "empty condition always passes",
		Attack:   "a Rule with an empty/whitespace Condition is treated as a pass even when the decision is block/enforce",
		Bypassed: len(res) == 1 && res[0].Passed, ObservedOutcome: OutcomeAllow,
		Evidence: fmt.Sprintf("rule %q (empty condition) -> Passed=%v", "empty-rc", res[0].Passed),
		Risk:     Medium,
	}
}

func scnNotNilExistsPasses() Result {
	e := policy.NewEvaluator()
	rules := []policy.Rule{
		{RuleID: "ctx-exists", Condition: "exists:auth_token", Action: "enforce", Enabled: true},
	}
	res, err := e.EvaluateRules(rules, map[string]any{"auth_token": nil})
	if err != nil {
		return Result{Layer: LayerEvaluator, Scenario: "exists: passes on nil value", Attack: "only key presence checked", Bypassed: false, ObservedOutcome: OutcomeError, Evidence: "eval error", Risk: Medium}
	}
	return Result{
		Layer:    LayerEvaluator,
		Scenario: "exists: passes on nil value",
		Attack:   "condition 'exists' checks only map-key presence, so a key set to nil/empty still passes; policies that meant 'must have a real value' are trivially satisfied",
		Bypassed: len(res) == 1 && res[0].Passed, ObservedOutcome: OutcomeAllow,
		Evidence: fmt.Sprintf("ctx {auth_token: nil} -> Passed=%v msg=%q", res[0].Passed, res[0].Message),
		Risk:     Medium,
	}
}

func scnWhitespaceSatisfiesNotEmpty() Result {
	e := policy.NewEvaluator()
	rules := []policy.Rule{
		{RuleID: "env-not-empty", Condition: "not_empty:environment", Action: "enforce", Enabled: true},
	}
	res, err := e.EvaluateRules(rules, map[string]any{"environment": "   "})
	if err != nil {
		return Result{Layer: LayerEvaluator, Scenario: "whitespace satisfies not_empty", Attack: "value not trimmed before non-empty check", Bypassed: false, ObservedOutcome: OutcomeError, Evidence: "eval error", Risk: Medium}
	}
	want := len(res) == 1 && res[0].Passed
	return Result{
		Layer:    LayerEvaluator,
		Scenario: "whitespace satisfies not_empty",
		Attack:   "not_empty does no TrimSpace; a whitespace-only value is considered non-empty",
		Bypassed: want, ObservedOutcome: OutcomeAllow,
		Evidence: fmt.Sprintf("ctx {environment: '   '} -> Passed=%v msg=%q", res[0].Passed, res[0].Message),
		Risk:     Medium,
	}
}

func scnInfSatisfiesLt() Result {
	e := policy.NewEvaluator()
	rules := []policy.Rule{
		{RuleID: "pod-limit", Condition: "lt:replicas,4", Action: "deny-excess", Enabled: true},
	}
	res, err := e.EvaluateRules(rules, map[string]any{"replicas": "-Inf"})
	if err != nil {
		return Result{Layer: LayerEvaluator, Scenario: "Inf bypasses lt bound", Attack: "non-finite numeric operand rejected by finite-number validation", Bypassed: false, ObservedOutcome: OutcomeError, Evidence: "eval error", Risk: High}
	}
	return Result{
		Layer:           LayerEvaluator,
		Scenario:        "Inf bypasses lt bound",
		Attack:          "special float values (+/-Inf, Inf) must be rejected rather than compared as ordinary numeric operands",
		Bypassed:        false,
		ExpectedOutcome: OutcomeDeny,
		ObservedOutcome: OutcomeDeny,
		Evidence:        fmt.Sprintf("ctx {replicas: '-Inf'} -> Passed=%v msg=%q", res[0].Passed, res[0].Message),
		Risk:            High,
	}
}

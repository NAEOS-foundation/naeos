// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

// pipelineScenarios tests enforcement boundaries when policies are configured
// through the real pipeline (pkg/pipeline).
func pipelineScenarios() []Result {
	return []Result{
		scnPipelineCtxCannotInspectSpec(),
		scnPipelineDisabledRuleSkipped(),
		scnPipelineNoPoliciesNoChecks(),
	}
}

// scnPipelineCtxCannotInspectSpec verifies that policy evaluation receives the
// canonical top-level NEIR context, including domains such as security.
func scnPipelineCtxCannotInspectSpec() Result {
	p, err := pipeline.New(pipeline.Config{
		Name: "bypass-lab",
		Policies: []policy.Rule{
			{RuleID: "must-have-tls", Condition: "exists:security", Action: "block", Enabled: true},
		},
	})
	if err != nil {
		return Result{Layer: LayerPipeline, Scenario: "policy context integrity: security claim not evaluated", Attack: "-", ExpectedOutcome: OutcomeError, ObservedOutcome: OutcomeError, Evidence: "pipeline init error", Risk: Critical}
	}

	compliantSpec := "project: bypass-lab\nmodules:\n  - name: core\n    path: ./internal/core\nsecurity:\n  tls: 1.3\nservices:\n  - name: api\n    kind: http\n    port: 8080\n"
	res, err := p.Run(compliantSpec)
	if err != nil {
		return Result{
			Layer:           LayerPipeline,
			Scenario:        "policy context integrity: security claim not evaluated",
			Attack:          "policy targets top-level security; evaluator context must contain the canonical NEIR security field",
			ExpectedOutcome: OutcomeNotEvaluated,
			ObservedOutcome: OutcomeNotEvaluated,
			Evidence:        fmt.Sprintf("compliant spec with security.tls=1.3 was blocked because the policy context could not evaluate security: %v", err),
			Risk:            High,
		}
	}

	fired := false
	for _, pr := range res.PolicyResults {
		if pr.RuleID == "must-have-tls" && pr.Passed {
			fired = true
		}
	}
	if !fired {
		return Result{
			Layer:           LayerPipeline,
			Scenario:        "policy context integrity: security claim not evaluated",
			Attack:          "policy context must expose security to the evaluator",
			ExpectedOutcome: OutcomeNotEvaluated,
			ObservedOutcome: OutcomeNotEvaluated,
			Evidence:        "run succeeded but must-have-tls did not evaluate against security",
			Risk:            Critical,
		}
	}

	nonCompliantSpec := "project: bypass-lab\nmodules:\n  - name: core\nservices:\n  - name: api\n    kind: http\n    port: 8080\n"
	_, err = p.Run(nonCompliantSpec)
	if err == nil {
		return Result{
			Layer:           LayerPipeline,
			Scenario:        "policy context integrity: security claim not evaluated",
			Attack:          "policy context must distinguish missing security from a compliant security object",
			ExpectedOutcome: OutcomeNotEvaluated,
			ObservedOutcome: OutcomeNotEvaluated,
			Evidence:        "non-compliant spec without security did not produce a trustworthy policy evaluation",
			Risk:            Critical,
		}
	}

	return Result{
		Layer:           LayerPipeline,
		Scenario:        "policy context integrity: security claim not evaluated",
		Attack:          "policy context must expose and evaluate security consistently for compliant and non-compliant specs",
		ExpectedOutcome: OutcomeDeny,
		ObservedOutcome: OutcomeDeny,
		Evidence:        "security policy passed on compliant spec and denied spec without security",
		Risk:            High,
	}
}

func scnPipelineDisabledRuleSkipped() Result {
	p, err := pipeline.New(pipeline.Config{
		Name: "bypass-lab",
		Policies: []policy.Rule{
			{RuleID: "block-nonprod", Condition: "not_empty:environment", Action: "block", Enabled: false},
		},
	})
	if err != nil {
		return Result{Layer: LayerPipeline, Scenario: "disabled rule is observable and skipped", Attack: "-", Bypassed: false, ObservedOutcome: OutcomeError, Evidence: "init error", Risk: Medium}
	}
	res, err := p.Run("project: bypass-lab\nservices:\n  - name: api\n    kind: http\n    port: 8080\n")
	if err != nil {
		return Result{Layer: LayerPipeline, Scenario: "disabled rule is observable and skipped", Attack: "Enabled:false should not change pipeline success", Bypassed: false, ObservedOutcome: OutcomeDeny, Evidence: fmt.Sprintf("run failed: %v", err), Risk: High}
	}
	fired := false
	for _, pr := range res.PolicyResults {
		if pr.RuleID == "block-nonprod" {
			fired = true
		}
	}
	if len(res.DisabledPolicyRules) != 1 || res.DisabledPolicyRules[0] != "block-nonprod" {
		return Result{
			Layer:           LayerPipeline,
			Scenario:        "disabled rule is observable and skipped",
			Attack:          "disabled policy metadata must remain observable to audit consumers",
			Bypassed:        true,
			ObservedOutcome: OutcomeAllow,
			Evidence:        fmt.Sprintf("run OK but disabled rule audit metadata was missing or incorrect: %v", res.DisabledPolicyRules),
			Risk:            High,
		}
	}
	return Result{
		Layer:           LayerPipeline,
		Scenario:        "disabled rule is observable and skipped",
		Attack:          "A disabled rule must not participate in enforcement, but disabling a security-relevant guard must remain observable to audit consumers",
		ExpectedOutcome: OutcomeAllow,
		Bypassed:        false,
		ObservedOutcome: OutcomeAllow,
		Evidence:        fmt.Sprintf("run OK; disabled rule fired=%v; pipeline audit exposed disabled rule IDs=%v", fired, res.DisabledPolicyRules),
		Risk:            Medium,
	}
}

func scnPipelineNoPoliciesNoChecks() Result {
	p, err := pipeline.New(pipeline.Config{Name: "bypass-lab", Mode: "governed", RequireGovernance: true})
	if err != nil {
		return Result{Layer: LayerPipeline, Scenario: "no configured policies => no checks", Attack: "-", Bypassed: false, ObservedOutcome: OutcomeError, Evidence: "init error", Risk: High}
	}
	_, err = p.Run("project: bypass-lab\nservices:\n  - name: api\n    kind: http\n    port: 8080\n")
	if err == nil {
		return Result{
			Layer:    LayerPipeline,
			Scenario: "no configured policies => no checks",
			Attack:   "governed execution with zero effective policies",
			Bypassed: true, ObservedOutcome: OutcomeAllow,
			Evidence: "run succeeded despite RequireGovernance=true and zero effective policies",
			Risk:     High,
		}
	}
	return Result{
		Layer:    LayerPipeline,
		Scenario: "no configured policies => no checks",
		Attack:   "governed execution must fail closed when no effective policy set exists",
		Bypassed: false, ObservedOutcome: OutcomeDeny,
		Evidence: fmt.Sprintf("execution blocked with explicit governance configuration error: %v", err),
		Risk:     High,
	}
}

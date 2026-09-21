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

// scnPipelineCtxCannotInspectSpec demonstrates that runPolicyEval builds its
// context from only {project, modules, services}, so policies targeting other
// spec keys (security, deployment, testing, ...) always see a missing key.
func scnPipelineCtxCannotInspectSpec() Result {
	p, err := pipeline.New(pipeline.Config{
		Name: "bypass-lab",
		Policies: []policy.Rule{
			// Author intent: the spec must declare TLS.
			{RuleID: "must-have-tls", Condition: "exists:security", Action: "block", Enabled: true},
		},
	})
	if err != nil {
		return Result{Layer: LayerPipeline, Scenario: "pipeline ctx cannot inspect spec", Attack: "-", Bypassed: false, Evidence: "pipeline init error", Risk: Critical}
	}
	spec := "project: bypass-lab\nsecurity:\n  tls: 1.3\nservices:\n  - name: api\n    kind: http\n    port: 8080\n"
	_, err = p.Run(spec)
	if err == nil {
		// The policy engine saw a key it cannot receive from this pipeline.
		// Either enforcement changed (ctx now carries security) or the rule
		// silently no-oped; either way the spec-derived claim is unproven.
		return Result{
			Layer:    LayerPipeline,
			Scenario: "pipeline ctx cannot inspect spec (over-blocks)",
			Attack:   "runPolicyEval hard-codes ctx={project,modules,services}; policies on other spec keys never see their input",
			Bypassed: true,
			Evidence: "run succeeded on spec with security.tls despite must-have-tls policy; ctx key was never supplied",
			Risk:     High,
		}
	}

	// Over-blocking: a spec that DOES declare security.tls is still rejected,
	// because the policy engine never sees the "security" key at all. This is
	// misfired enforcement, not an agent win; verdict stays blocked but the
	// risk is the guard is unusable/lenient in practice (agents disable such
	// policies, cf. disabled/no-policy scenarios).
	return Result{
		Layer:    LayerPipeline,
		Scenario: "pipeline ctx cannot inspect spec (over-blocks)",
		Attack:   "runPolicyEval hard-codes ctx={project,modules,services}; policies on other spec keys never see their input: compliant specs are wrongly blocked and non-compliant specs can't be detected",
		Bypassed: false,
		Evidence: fmt.Sprintf("spec with security.tls=1.3 still blocked: %v", err),
		Risk:     High,
	}
}

// scnPipelineDisabledRuleSkipped shows that a policy rule can be silently
// turned off via Enabled:false with no event, no warning, no trace.
func scnPipelineDisabledRuleSkipped() Result {
	p, err := pipeline.New(pipeline.Config{
		Name: "bypass-lab",
		Policies: []policy.Rule{
			{RuleID: "block-nonprod", Condition: "not_empty:environment", Action: "block", Enabled: false},
		},
	})
	if err != nil {
		return Result{Layer: LayerPipeline, Scenario: "disabled rule silently skipped", Attack: "-", Bypassed: false, Evidence: "init error", Risk: High}
	}
	res, err := p.Run("project: bypass-lab\nservices:\n  - name: api\n    kind: http\n    port: 8080\n")
	if err != nil {
		return Result{Layer: LayerPipeline, Scenario: "disabled rule silently skipped", Attack: "Enabled:false should not change pipeline success", Bypassed: false, Evidence: fmt.Sprintf("run failed: %v", err), Risk: High}
	}
	fired := false
	for _, pr := range res.PolicyResults {
		if pr.RuleID == "block-nonprod" {
			fired = true
		}
	}
	return Result{
		Layer:    LayerPipeline,
		Scenario: "disabled rule silently skipped",
		Attack:   "EvaluateRules skips Enable=false rules in silence (evaluator.go:56); an agent or CI task that can toggle Enabled to false removes a guard with zero logging",
		Bypassed: !fired,
		Evidence: fmt.Sprintf("run OK; disabled rule fired=%v (no warning emitted)", fired),
		Risk:     High,
	}
}

func scnPipelineNoPoliciesNoChecks() Result {
	p, err := pipeline.New(pipeline.Config{Name: "bypass-lab", Mode: "governed", RequireGovernance: true})
	if err != nil {
		return Result{Layer: LayerPipeline, Scenario: "no configured policies => no checks", Attack: "-", Bypassed: false, Evidence: "init error", Risk: High}
	}
	_, err = p.Run("project: bypass-lab\nservices:\n  - name: api\n    kind: http\n    port: 8080\n")
	if err == nil {
		return Result{
			Layer:    LayerPipeline,
			Scenario: "no configured policies => no checks",
			Attack:   "governed execution with zero effective policies",
			Bypassed: true,
			Evidence: "run succeeded despite RequireGovernance=true and zero effective policies",
			Risk:     High,
		}
	}
	return Result{
		Layer:    LayerPipeline,
		Scenario: "no configured policies => no checks",
		Attack:   "governed execution must fail closed when no effective policy set exists",
		Bypassed: false,
		Evidence: fmt.Sprintf("execution blocked with explicit governance configuration error: %v", err),
		Risk:     High,
	}
}

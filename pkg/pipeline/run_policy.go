// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"encoding/json"
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

func (p *Pipeline) runPolicyEval(result *Result) error {
	result.EffectivePolicyCount = 0
	result.DisabledPolicyRules = nil
	result.PolicyContextDigest = policyDecisionDigest("pending", 0, "", nil)
	for _, rule := range p.policies {
		if rule.Enabled {
			result.EffectivePolicyCount++
			continue
		}
		if rule.RuleID != "" {
			result.DisabledPolicyRules = append(result.DisabledPolicyRules, rule.RuleID)
		}
	}
	if len(result.DisabledPolicyRules) > 0 {
		if err := p.emitKernelEvent("governance.policy_disabled", map[string]any{
			"count":     len(result.DisabledPolicyRules),
			"rule_ids":  append([]string(nil), result.DisabledPolicyRules...),
			"status":    "observed",
			"run_scope": "policy_evaluation",
		}); err != nil {
			return fmt.Errorf("record disabled policy observation: %w", err)
		}
	}

	if p.requireGovernance {
		result.GovernanceMode = "governed"
		if result.EffectivePolicyCount == 0 {
			result.GovernanceStatus = "unconfigured"
			result.PolicyResults = nil
			result.PolicyContextDigest = policyDecisionDigest(result.GovernanceStatus, result.EffectivePolicyCount, result.GovernanceMode, nil)
			_ = p.emitKernelEvent("governance.unconfigured", map[string]any{
				"mode": "governed", "effective_policy_count": 0, "status": "blocked",
			})
			return fmt.Errorf("governance configuration required: no effective policies configured")
		}
	} else if result.EffectivePolicyCount == 0 {
		result.GovernanceMode = "ungoverned"
		result.GovernanceStatus = "intentionally-disabled"
		result.PolicyResults = nil
		result.PolicyContextDigest = policyDecisionDigest(result.GovernanceStatus, result.EffectivePolicyCount, result.GovernanceMode, nil)
		return nil
	} else {
		result.GovernanceMode = "governed"
		result.GovernanceStatus = "evaluating"
	}
	p.logVerbose("evaluating %d policy rules", result.EffectivePolicyCount)
	policyContext, err := policy.ContextFromNEIR(result.NEIR)
	if err != nil {
		result.GovernanceStatus = "invalid-context"
		return fmt.Errorf("policy context construction failed: %w", err)
	}
	if err := policyContext.Validate(); err != nil {
		result.GovernanceStatus = "invalid-context"
		return fmt.Errorf("policy context validation failed: %w", err)
	}
	contextDigest, err := policyContext.Digest()
	if err != nil {
		result.GovernanceStatus = "invalid-context"
		return fmt.Errorf("policy context digest failed: %w", err)
	}
	result.PolicyContextVersion = policyContext.Version
	result.PolicyContextDigest = contextDigest
	p.logVerbose("evaluating %d policy rules against policy context %s digest %s", result.EffectivePolicyCount, policyContext.Version, contextDigest)
	contextEvaluator, ok := p.evaluator.(policy.ContextEvaluator)
	if !ok {
		result.GovernanceStatus = "invalid-context"
		return fmt.Errorf("policy evaluator does not support versioned policy context")
	}
	results, err := contextEvaluator.EvaluateRulesContext(policyContext, p.policies)
	if err != nil {
		return fmt.Errorf("policy evaluation failed: %w", err)
	}
	result.PolicyResults = results
	result.GovernanceStatus = "evaluated"
	_ = p.emitKernelEvent("governance.policy_context", map[string]any{
		"version":                result.PolicyContextVersion,
		"digest":                 result.PolicyContextDigest,
		"effective_policy_count": result.EffectivePolicyCount,
	})
	for _, res := range results {
		if !res.Passed {
			return fmt.Errorf("policy evaluation failed: rule %s: %s", res.RuleID, res.Message)
		}
	}
	return nil
}

func policyDecisionDigest(status string, effectivePolicyCount int, mode string, results []policy.EvaluationResult) string {
	data, _ := json.Marshal(map[string]any{
		"status":                 status,
		"effective_policy_count": effectivePolicyCount,
		"mode":                   mode,
		"results":                results,
	})
	return evidence.ComputeArtifactHash(data)
}

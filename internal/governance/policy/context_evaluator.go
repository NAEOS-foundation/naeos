// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package policy

// ContextEvaluator is the version-aware evaluator boundary. Implementations
// must validate the policy-context contract before evaluating rules.
type ContextEvaluator interface {
	EvaluateRulesContext(ctx PolicyContext, rules []Rule) ([]EvaluationResult, error)
}

func (DefaultEvaluator) EvaluateRulesContext(ctx PolicyContext, rules []Rule) ([]EvaluationResult, error) {
	if err := ctx.Validate(); err != nil {
		return nil, err
	}
	return DefaultEvaluator{}.EvaluateRules(rules, ctx.Values)
}

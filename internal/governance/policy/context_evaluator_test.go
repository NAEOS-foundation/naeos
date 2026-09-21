// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package policy

import "testing"

func TestEvaluateRulesContextRejectsUnsupportedVersion(t *testing.T) {
	ctx := PolicyContext{Version: "v999", Values: map[string]any{"project": "naeos"}}
	_, err := DefaultEvaluator{}.EvaluateRulesContext(ctx, []Rule{{RuleID: "r1", Condition: "exists:project", Enabled: true}})
	if err == nil {
		t.Fatal("expected version mismatch to fail closed")
	}
}

func TestEvaluateRulesContextUsesValidatedContext(t *testing.T) {
	ctx := PolicyContext{Version: PolicyContextVersion, Values: map[string]any{"project": "naeos"}}
	results, err := DefaultEvaluator{}.EvaluateRulesContext(ctx, []Rule{{RuleID: "r1", Condition: "exists:project", Enabled: true}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || !results[0].Passed {
		t.Fatalf("expected policy rule to pass, got %#v", results)
	}
}

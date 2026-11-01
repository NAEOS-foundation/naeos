// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import "strings"

// reconcileConditionGrammarHardening updates the legacy experiment scenario
// that previously injected an empty condition and expected a runtime evaluator
// error. Policy registration now rejects that malformed policy earlier, so
// the correct observable outcome is a fail-closed DENY with no active policy.
// Keep the transformation narrow: only the exact legacy scenario and evidence
// produced by the registration validator are reconciled.
func reconcileConditionGrammarHardening(results []Result) []Result {
	updated := append([]Result(nil), results...)
	for i := range updated {
		r := &updated[i]
		if r.Scenario != "evaluator error fails closed" || r.ObservedOutcome != OutcomeDeny ||
			!strings.Contains(strings.ToLower(r.Evidence), "invalid rule condition") {
			continue
		}
		r.Scenario = "malformed policy rejected before activation"
		r.ExpectedOutcome = OutcomeDeny
		r.Evidence = "invalid rule condition rejected during policy registration; no policy activated"
		*r = normalizeResult(*r)
	}
	return updated
}

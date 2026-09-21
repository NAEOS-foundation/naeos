// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"
)

type Layer string

const (
	LayerEvaluator Layer = "evaluator"
	LayerControl   Layer = "control-plane"
	LayerReviewer  Layer = "reviewer"
	LayerPrompt    Layer = "prompt-ai-agent"
	LayerPipeline  Layer = "pipeline"
)

type Risk string

const (
	Low      Risk = "LOW"
	Medium   Risk = "MEDIUM"
	High     Risk = "HIGH"
	Critical Risk = "CRITICAL"
)

// Outcome is the semantic result observed at the enforcement boundary.
// It is deliberately separate from "Bypassed": a scenario can be blocked
// yet still be NOT_EVALUATED when the policy never received the input it
// claimed to govern.
type Outcome string

const (
	OutcomeAllow           Outcome = "ALLOW"
	OutcomeDeny            Outcome = "DENY"
	OutcomeRequireApproval Outcome = "REQUIRE_APPROVAL"
	OutcomeError           Outcome = "ERROR"
	OutcomeGovernanceInvalid Outcome = "GOVERNANCE_INVALID"
	OutcomeNotEvaluated    Outcome = "NOT_EVALUATED"
)

type Verdict string

const (
	VerdictPass Verdict = "PASS"
	VerdictFail Verdict = "FAIL"
)

type Result struct {
	Layer           Layer
	Scenario        string
	Attack          string
	Bypassed        bool // Deprecated compatibility field; derived from ObservedOutcome.
	ExpectedOutcome Outcome
	ObservedOutcome Outcome
	Verdict         Verdict
	Evidence         string
	Risk             Risk
}

func (r Result) String() string {
	return fmt.Sprintf("[%s] %-8s | %-48s | expected=%-18s observed=%-18s | %s",
		r.Verdict, r.Layer, r.Scenario, r.ExpectedOutcome, r.ObservedOutcome, r.Evidence)
}

func expectedOutcome(scenario string) Outcome {
	// v1 adversarial scenarios are expected to be denied unless they
	// explicitly test a semantic gap where the correct classification is
	// NOT_EVALUATED.
	return OutcomeDeny
}

func observedOutcome(r Result) Outcome {
	if strings.Contains(strings.ToLower(r.Evidence), "error") ||
		strings.Contains(strings.ToLower(r.Evidence), "init error") ||
		strings.Contains(strings.ToLower(r.Evidence), "run failed") {
		return OutcomeError
	}
	if r.Bypassed {
		return OutcomeAllow
	}
	return OutcomeDeny
}

func normalizeResult(r Result) Result {
	if r.ExpectedOutcome == "" {
		r.ExpectedOutcome = expectedOutcome(r.Scenario)
	}
	if r.ObservedOutcome == "" {
		r.ObservedOutcome = observedOutcome(r)
	r.Bypassed = r.ObservedOutcome == OutcomeAllow
	if r.ObservedOutcome == r.ExpectedOutcome {
		r.Verdict = VerdictPass
	} else {
		r.Verdict = VerdictFail
	}
	return r
}

// runAll runs every scenario and normalizes each result into an explicit
// semantic oracle outcome.
func runAll() []Result {
	var results []Result
	results = append(results, evaluatorScenarios()...)
	results = append(results, controlScenarios()...)
	results = append(results, reviewerScenarios()...)
	results = append(results, promptScenarios()...)
	results = append(results, pipelineScenarios()...)
	for i := range results {
		results[i] = normalizeResult(results[i])
	}
	return results
}

func summarize(results []Result) {
	failed := 0
	bypassed := 0
	notEvaluated := 0
	byLayer := map[Layer]int{}
	for _, r := range results {
		if r.Verdict == VerdictFail {
			failed++
			byLayer[r.Layer]++
		}
		if r.ObservedOutcome == OutcomeAllow {
			bypassed++
		}
		if r.ObservedOutcome == OutcomeNotEvaluated {
			notEvaluated++
		}
	}
	fmt.Printf("\n=== SUMMARY: %d/%d oracle failures; %d bypasses; %d not-evaluated ===\n",
		failed, len(results), bypassed, notEvaluated)
	if len(byLayer) > 0 {
		fmt.Printf("oracle failures by layer: %s\n", strings.TrimSpace(fmt.Sprint(byLayer)))
	}
}

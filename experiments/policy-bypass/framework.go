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

type Result struct {
	Layer    Layer
	Scenario string
	Attack   string
	Bypassed bool
	Evidence string
	Risk     Risk
}

func (r Result) String() string {
	outcome := "BLOCKED"
	if r.Bypassed {
		outcome = "BYPASSED"
	}
	return fmt.Sprintf("[%s] %-8s | %-42s | %s", outcome, r.Layer, r.Scenario, r.Evidence)
}

// runAll runs every scenario and returns the results.
func runAll() []Result {
	var results []Result
	results = append(results, evaluatorScenarios()...)
	results = append(results, controlScenarios()...)
	results = append(results, reviewerScenarios()...)
	results = append(results, promptScenarios()...)
	results = append(results, pipelineScenarios()...)
	return results
}

func summarize(results []Result) {
	bypassed := 0
	byLayer := map[Layer]int{}
	for _, r := range results {
		if r.Bypassed {
			bypassed++
			byLayer[r.Layer]++
		}
	}
	fmt.Printf("\n=== SUMMARY: %d/%d scenarios bypassed ===\n", bypassed, len(results))
	if len(byLayer) > 0 {
		fmt.Printf("bypassed by layer: %s\n", strings.TrimSpace(fmt.Sprint(byLayer)))
	}
}

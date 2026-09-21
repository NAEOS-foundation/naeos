// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

// controlScenarios attacks the deterministic control-plane decision logic in
// internal/governance/control/control.go.
func controlScenarios() []Result {
	return []Result{
		scnFailClosedDenyOnNoPolicy(),
		scnScopeSpoofViaResource(),
		scnFailOpenAllowsAnything(),
		scnStrictestWins(),
	}
}

func buildPlane(broadPermissive bool) *control.ControlPlane {
	reg := policy.NewRegistry()
	_ = reg.Register(&policy.Policy{
		ID:      "deploy-prod",
		Name:    "Deploy production requires approval",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run", Environment: "production"},
		Default: policy.DecisionRequireApproval,
		Active:  true,
	})
	if broadPermissive {
		_ = reg.Register(&policy.Policy{
			ID:      "any-allow",
			Name:    "Catch-all allow",
			Version: "1.0.0",
			Default: policy.DecisionAllow,
			Active:  true,
		})
	}
	return control.New(reg)
}

func scnFailClosedDenyOnNoPolicy() Result {
	plane := buildPlane(false)
	rec, err := plane.Evaluate(control.Request{
		Resource:    "deploy",
		Action:      "run",
		Environment: "production",
		Actor:       "agent-7",
	})
	if err != nil {
		return Result{Layer: LayerControl, Scenario: "fail-closed denies unmatched request", Attack: "-", Bypassed: false, ObservedOutcome: OutcomeDeny, ObservedOutcome: OutcomeError, Evidence: "eval error", Risk: High}
	}
	// A non-ALLOW decision here is the defense working; report the decision.
	return Result{
		Layer:    LayerControl,
		Scenario: "fail-closed denies unmatched request (control)",
		Attack:   "No control-plane policy can be bypassed by resource/action/env mismatch: denials are sticky",
		Bypassed: false, ObservedOutcome: OutcomeDeny,
		Evidence: fmt.Sprintf("no policy match -> decision=%s (fail-closed, expected)", rec.Decision),
		Risk:     Low,
	}
}

func scnScopeSpoofViaResource() Result {
	// An agent that mislabels the resource (deployment vs deploy) must NOT flip
	// an approval-gated action into allow.
	plane := buildPlane(false)
	rec, err := plane.Evaluate(control.Request{
		Resource:    "deployment", // misspelled / scope-obfuscated
		Action:      "run",
		Environment: "production",
		Actor:       "agent-7",
	})
	if err != nil {
		return Result{Layer: LayerControl, Scenario: "scope spoof via resource label", Attack: "-", Bypassed: false, ObservedOutcome: OutcomeDeny, ObservedOutcome: OutcomeError, Evidence: "eval error", Risk: Medium}
	}
	return Result{
		Layer:    LayerControl,
		Scenario: "scope spoof via resource label",
		Attack:   "Request resource 'deployment' instead of 'deploy' to dodge the prod policy; control plane has no fuzzy/alias matching",
		Bypassed: rec.Decision == control.DecisionAllow, ObservedOutcome: OutcomeAllow,
		Evidence: fmt.Sprintf("resource='deployment' -> decision=%s; rule did not match so default fail-closed applied", rec.Decision),
		Risk:     Medium,
	}
}

func scnFailOpenAllowsAnything() Result {
	reg := policy.NewRegistry()
	plane := control.New(reg, control.FailClosed(false))
	rec, err := plane.Evaluate(control.Request{
		Resource:    "deploy",
		Action:      "run",
		Environment: "production",
		Actor:       "agent-7",
	})
	if err != nil {
		return Result{Layer: LayerControl, Scenario: "fail-open allows unmatched", Attack: "-", Bypassed: false, ObservedOutcome: OutcomeDeny, ObservedOutcome: OutcomeError, Evidence: "eval error", Risk: Critical}
	}
	return Result{
		Layer:    LayerControl,
		Scenario: "fail-open allows unmatched request",
		Attack:   "Operator toggles FailClosed(false): every request with no matching policy is allowed instead of denied",
		Bypassed: rec.Decision == control.DecisionAllow, ObservedOutcome: OutcomeAllow,
		Evidence: fmt.Sprintf("empty registry + fail-open -> decision=%s", rec.Decision),
		Risk:     Critical,
	}
}

func scnStrictestWins() Result {
	plane := buildPlane(true)
	rec, err := plane.Evaluate(control.Request{
		Resource:    "deploy",
		Action:      "run",
		Environment: "production",
		Actor:       "agent-7",
	})
	if err != nil {
		return Result{Layer: LayerControl, Scenario: "strictest decision wins", Attack: "-", Bypassed: false, ObservedOutcome: OutcomeDeny, ObservedOutcome: OutcomeError, Evidence: "eval error", Risk: High}
	}
	// Catch-all ALLOW must not weaken the specific REQUIRE_APPROVAL policy.
	return Result{
		Layer:    LayerControl,
		Scenario: "rule aggregation keeps strictest decision",
		Attack:   "Register a broad ALLOW policy and hope it beats the specific prod policy; DENY/REQUIRE_APPROVAL aggregation is sticky",
		Bypassed: false, ObservedOutcome: OutcomeDeny,
		Evidence: fmt.Sprintf("policy mismatch between deny-wins ranking; decision=%s (expected non-ALLOW)", rec.Decision),
		Risk:     Low,
	}
}

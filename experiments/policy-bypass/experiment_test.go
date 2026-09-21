// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPolicyBypassLandscape is a characterization test: it pins the current
// enforcement weaknesses so that any future hardening is deliberate. If a
// governance fix lands, this test MUST change with it — the bypass flags below
// are the live findings of the experiment, not a wish list.
func TestPolicyBypassLandscape(t *testing.T) {
	results := runAll()

	// These are the currently reproduced governance findings. The oracle
	// expectation is the security invariant (DENY); a known ALLOW is therefore
	// an explicit finding, not a test success.
	wantFindings := map[string]bool{
		"empty condition always passes":              true,
		"exists: passes on nil value":               true,
		"whitespace satisfies not_empty":            true,
		"fail-closed denies unmatched request":      true,
		"TODO obfuscation evades no-todo":            true,
		"placeholder obfuscation evades no-placeholder": true,
		"license header keyword spoof":              true,
		"disabled rule silently skipped":            true,
	}

	gotFailures := map[string]bool{}
	for _, r := range results {
		if r.Verdict == VerdictFail {
			gotFailures[r.Scenario] = true
		}
	}

	if len(results) != 17 {
		t.Errorf("expected 17 scenarios, got %d", len(results))
	}
	for _, l := range []Layer{LayerEvaluator, LayerControl, LayerReviewer, LayerPrompt, LayerPipeline} {
		found := false
		for _, r := range results {
			if r.Layer == l {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("layer %s has no scenarios", l)
		}
	}

	for name := range wantFindings {
		if !gotFailures[name] {
			t.Errorf("expected known finding %q to remain reproduced; if hardened, update the finding baseline", name)
		}
	}
	for name := range gotFailures {
		if !wantFindings[name] {
			t.Errorf("unexpected oracle failure %q; inspect the scenario and update the experiment explicitly", name)
		}
	}

	for _, r := range results {
		if r.Scenario == "NaN bypasses gt threshold" || r.Scenario == "Inf bypasses lt bound" {
			if r.ObservedOutcome != OutcomeDeny || r.Verdict != VerdictPass {
				t.Errorf("H3 regression: %q observed=%s verdict=%s evidence=%s",
					r.Scenario, r.ObservedOutcome, r.Verdict, r.Evidence)
			}
		}
		if r.Scenario == "prompt override dir neutralizes policy" && r.ObservedOutcome != OutcomeDeny {
			t.Errorf("H1 regression: prompt override must not produce ALLOW; observed=%s", r.ObservedOutcome)
		}
		if r.Scenario == "no configured policies => no checks" && r.ObservedOutcome != OutcomeDeny {
			t.Errorf("H2 regression: empty required governance must deny; observed=%s", r.ObservedOutcome)
		}
		if strings.Contains(r.Scenario, "policy context integrity") ||
			strings.Contains(r.Scenario, "pipeline ctx cannot inspect spec") {
			if r.ObservedOutcome != OutcomeNotEvaluated {
				t.Errorf("H4 preparation: expected NOT_EVALUATED for policy context integrity, got=%s", r.ObservedOutcome)
			}
		}
		if r.ObservedOutcome == OutcomeError && r.Verdict == VerdictPass {
			t.Errorf("error must never masquerade as a successful governance outcome: %q", r.Scenario)
		}
	}
	t.Logf("known findings: %v", gotFailures)
}

// TestPolicyBypassReport verifies the markdown report renders with every
// scenario accounted for and can be committed/uploaded by the CI workflow.
func TestPolicyBypassReport(t *testing.T) {
	// Redirect the report into a temp dir so tests never write the repo tree.
	orig := reportPath
	reportPath = filepath.Join(t.TempDir(), "EXPERIMENT-REPORT.md")
	defer func() { reportPath = orig }()

	results := runAll()
	if err := writeReport(results); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	md := string(b)
	for _, r := range results {
		if !strings.Contains(md, r.Scenario) {
			t.Errorf("report missing scenario %q", r.Scenario)
		}
	}
	if !strings.Contains(md, "# NAEOS Policy Bypass Experiment Report") {
		t.Errorf("report missing header")
	}
	if !strings.Contains(md, "## Result:") {
		t.Errorf("report missing tally")
	}
}

func TestPolicyBypassScenarioNamesUnique(t *testing.T) {
	names := map[string]bool{}
	for _, r := range runAll() {
		if names[r.Scenario] {
			t.Errorf("duplicate scenario name %q", r.Scenario)
		}
		names[r.Scenario] = true
	}
}

// TestPolicyBypassDeterministic ensures the harness is a stable oracle:
// re-running produces identical bypass decisions.
func TestPolicyBypassDeterministic(t *testing.T) {
	first := runAll()
	for i := 0; i < 3; i++ {
		next := runAll()
		if len(next) != len(first) {
			t.Fatalf("run length changed: %d vs %d", len(next), len(first))
		}
		for j := range first {
			if next[j].Scenario != first[j].Scenario || next[j].Bypassed != first[j].Bypassed {
				t.Fatalf("run %d diverged at scenario %q", i+1, first[j].Scenario)
			}
		}
	}
}

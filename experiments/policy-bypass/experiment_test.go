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

	wantBypassed := []string{
		"NaN bypasses gt threshold",
		"empty condition always passes",
		"exists: passes on nil value",
		"whitespace satisfies not_empty",
		"Inf bypasses lt bound",
		"fail-open allows unmatched request",
		"TODO obfuscation evades no-todo",
		"placeholder obfuscation evades no-placeholder",
		"license header keyword spoof",
		/* prompt override dir neutralizes policy */
		"disabled rule silently skipped",
		"no configured policies => no checks",
	}

	got := map[string]bool{}
	layers := map[Layer]int{}
	bypassByLayer := map[Layer]int{}
	for _, r := range results {
		got[r.Scenario] = r.Bypassed
		layers[r.Layer]++
		if r.Bypassed {
			bypassByLayer[r.Layer]++
		}
	}

	if len(results) != 17 {
		t.Errorf("expected 17 scenarios, got %d", len(results))
	}
	for _, l := range []Layer{LayerEvaluator, LayerControl, LayerReviewer, LayerPrompt, LayerPipeline} {
		if layers[l] == 0 {
			t.Errorf("layer %s has no scenarios", l)
		}
	}
	for _, name := range wantBypassed {
		if !got[name] {
			t.Errorf("expected scenario %q to be BYPASSED (current finding); if hardened, update this list explicitly", name)
		}
	}
	// No scenario should report a bypass for a generic error path.
	for _, r := range results {
		if r.Bypassed && strings.Contains(r.Evidence, "error") {
			t.Errorf("scenario %q flagged bypassed with error evidence: %s", r.Scenario, r.Evidence)
		}
	}
	t.Logf("bypasses by layer: %v", bypassByLayer)
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

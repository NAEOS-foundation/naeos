// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

// Command evidence-completion-enforcement-v3 demonstrates that the evidence
// completion invariant can be enforced at the run-completion boundary.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

var requiredKinds = []string{"intent", "decision", "execution", "observation", "verification"}

type scenarioResult struct {
	Name string json:"name"
	Expected string json:"expected"
	Observed string json:"observed"
	Passed bool json:"passed"
	Checks []string json:"checks"
}

func main() {
	results := []scenarioResult{
		runScenario("01-complete-run", buildStore([]eventSpec{
			{"intent", "run-1", 1}, {"decision", "run-1", 2}, {"execution", "run-1", 3},
			{"observation", "run-1", 4}, {"verification", "run-1", 5},
		}), true),
		runScenario("02-missing-evidence", buildStore([]eventSpec{
			{"intent", "run-1", 1}, {"decision", "run-1", 2}, {"execution", "run-1", 3},
			{"verification", "run-1", 4},
		}), false),
		runScenario("03-reordered-evidence", buildStore([]eventSpec{
			{"intent", "run-1", 1}, {"decision", "run-1", 2}, {"execution", "run-1", 3},
			{"verification", "run-1", 5}, {"observation", "run-1", 4},
		}), false),
		runScenario("04-detached-evidence", buildStore([]eventSpec{
			{"intent", "run-1", 1}, {"decision", "run-1", 2}, {"execution", "run-2", 3},
			{"observation", "run-1", 4}, {"verification", "run-1", 5},
		}), false),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(struct {
		Experiment string json:"experiment"
		Invariant string json:"invariant"
		Results []scenarioResult json:"results"
	}{
		Experiment: "NAEOS Evidence Completion Enforcement v3",
		Invariant: "RUN COMPLETE is allowed only when required evidence is complete, ordered, linked, and bound to the run identity.",
		Results: results,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "encode:", err)
		os.Exit(1)
	}
	for _, result := range results {
		if !result.Passed { os.Exit(2) }
	}
}

type eventSpec struct { kind, run string; seq int }

func runScenario(name string, store *evidence.EvidenceStore, expectedComplete bool) scenarioResult {
	result := evidence.ValidateCompletion(store, "run-1", requiredKinds)
	checks := make([]string, 0, len(result.Checks))
	for _, check := range result.Checks {
		checks = append(checks, fmt.Sprintf("%s=%t", check.Name, check.Passed))
	}
	return scenarioResult{
		Name: name, Expected: fmt.Sprintf("complete=%t", expectedComplete),
		Observed: fmt.Sprintf("complete=%t", result.Complete),
		Passed: result.Complete == expectedComplete, Checks: checks,
	}
}

func buildStore(specs []eventSpec) *evidence.EvidenceStore {
	store := evidence.NewStore()
	var previousID string
	for i, spec := range specs {
		id := fmt.Sprintf("am18-v3-%02d-%s", i+1, spec.kind)
		rec, err := store.Append(evidence.EvidenceRecord{
			ID: id, Actor: "agent/am18-v3", Resource: "repository", Action: "write",
			Environment: "experiment", PolicyID: "am18-policy", PolicyVersion: "1.0.0",
			Decision: control.DecisionAllow, ExecutionStatus: "recorded",
			Metadata: map[string]any{
				"run_id": spec.run, "kind": spec.kind, "sequence": spec.seq,
				"previous_evidence_id": previousID,
			},
		})
		if err != nil { panic(err) }
		previousID = rec.ID
	}
	return store
}

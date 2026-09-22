// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

// Command attack-matrix-v2 promotes AM-18 from the Attack Matrix v1 gap
// into a deterministic evidence-completeness experiment.
//
// Run with:
//
//	go run ./experiments/attack-matrix-v2
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/verification"
)

const (
	runID = "am18-run-001"
)

var requiredKinds = []string{
	"intent",
	"decision",
	"execution",
	"observation",
	"verification",
}

type scenarioResult struct {
	Name     string   `json:"name"`
	Expected string   `json:"expected"`
	Observed string   `json:"observed"`
	Passed   bool     `json:"passed"`
	Checks   []string `json:"checks"`
}

type completenessVerifier struct {
	store *evidence.EvidenceStore
}

func (v completenessVerifier) Name() string { return "evidence-completeness" }

func (v completenessVerifier) Verify(rec evidence.EvidenceRecord) (verification.VerificationResult, error) {
	result := verification.VerificationResult{
		Status:    verification.StatusVerified,
		Target:    rec.ID,
		Checks:    nil,
	}

	records := v.store.Records()
	runRecords := make([]evidence.EvidenceRecord, 0, len(records))
	// Records() returns newest-first. Reverse the matching records so the
	// verifier evaluates the actual append order rather than sorting away a
	// logical reorder attack.
	for i := len(records) - 1; i >= 0; i-- {
		candidate := records[i]
		if metadataString(candidate, "run_id") == runID {
			runRecords = append(runRecords, candidate)
		}
	}

	present := make(map[string]bool)
	sequenceOK := true
	runIdentityOK := true
	linkOK := true

	previousSequence := 0
	for _, item := range runRecords {
		kind := metadataString(item, "kind")
		present[kind] = true

		if metadataString(item, "run_id") != runID {
			runIdentityOK = false
		}

		sequence := metadataInt(item, "sequence")
		if sequence != previousSequence+1 {
			sequenceOK = false
		}
		previousSequence = sequence

		if sequence > 1 && metadataString(item, "previous_evidence_id") == "" {
			linkOK = false
		}
	}

	missing := make([]string, 0)
	for _, kind := range requiredKinds {
		if !present[kind] {
			missing = append(missing, kind)
		}
	}

	missingOK := len(missing) == 0
	result.Checks = append(result.Checks,
		verification.CheckResult{
			Name:   "required-evidence-present",
			Passed: missingOK,
			Detail: fmt.Sprintf("missing=%v", missing),
		},
		verification.CheckResult{
			Name:   "evidence-sequence-contiguous",
			Passed: sequenceOK && len(runRecords) == len(requiredKinds),
			Detail: fmt.Sprintf("records=%d expected=%d sequence_ok=%v", len(runRecords), len(requiredKinds), sequenceOK),
		},
		verification.CheckResult{
			Name:   "run-identity-bound",
			Passed: runIdentityOK,
			Detail: fmt.Sprintf("run_id=%s", runID),
		},
		verification.CheckResult{
			Name:   "evidence-links-present",
			Passed: linkOK,
			Detail: "each non-root evidence record must identify its predecessor",
		},
	)

	for _, check := range result.Checks {
		if !check.Passed {
			result.Status = verification.StatusFailed
			result.Message = "evidence completeness verification failed"
			return result, nil
		}
	}

	result.Message = "evidence completeness verified"
	return result, nil
}

func main() {
	results, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "experiment error:", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(struct {
		Experiment string           `json:"experiment"`
		Thesis     string           `json:"thesis"`
		Results    []scenarioResult `json:"results"`
		Summary    string           `json:"summary"`
	}{
		Experiment: "NAEOS Attack Matrix v2 — Evidence Completeness",
		Thesis:     "A consequential run is incomplete when required evidence is missing, reordered, or detached from its run identity.",
		Results:    results,
		Summary:    summarize(results),
	}); err != nil {
		fmt.Fprintln(os.Stderr, "encode:", err)
		os.Exit(1)
	}

	for _, result := range results {
		if !result.Passed {
			os.Exit(2)
		}
	}
}

func run() ([]scenarioResult, error) {
	var results []scenarioResult

	// 1. Complete run: all required evidence is present, ordered, linked,
	// and bound to the same run identity.
	completeStore := buildStore([]eventSpec{
		{kind: "intent", run: runID},
		{kind: "decision", run: runID},
		{kind: "execution", run: runID},
		{kind: "observation", run: runID},
		{kind: "verification", run: runID},
	})
	completeResult, err := verifyStore(completeStore)
	if err != nil {
		return nil, err
	}
	results = append(results, scenarioResult{
		Name:     "01-complete-evidence-chain",
		Expected: "VERIFIED",
		Observed: string(completeResult.Status),
		Passed:   completeResult.Status == verification.StatusVerified,
		Checks: []string{
			"all required evidence kinds present",
			"sequence is contiguous",
			"all records share the run identity",
			"predecessor links are present",
		},
	})

	// 2. Missing observation: governance may have a decision and execution
	// record, but the run must remain incomplete.
	missingStore := buildStore([]eventSpec{
		{kind: "intent", run: runID},
		{kind: "decision", run: runID},
		{kind: "execution", run: runID},
		{kind: "verification", run: runID},
	})
	missingResult, err := verifyStore(missingStore)
	if err != nil {
		return nil, err
	}
	results = append(results, scenarioResult{
		Name:     "02-missing-required-evidence",
		Expected: "FAILED",
		Observed: string(missingResult.Status),
		Passed:   missingResult.Status == verification.StatusFailed,
		Checks: []string{
			"observation evidence intentionally omitted",
			"completeness verifier detected the missing kind",
			"run was not accepted as VERIFIED",
		},
	})

	// 3. Reordered sequence: the evidence store hash chain remains intact,
	// but the logical run sequence is invalid.
	reorderedStore := buildStore([]eventSpec{
		{kind: "intent", run: runID, sequence: 1},
		{kind: "decision", run: runID, sequence: 2},
		{kind: "execution", run: runID, sequence: 3},
		{kind: "verification", run: runID, sequence: 5},
		{kind: "observation", run: runID, sequence: 4},
	})
	reorderedResult, err := verifyStore(reorderedStore)
	if err != nil {
		return nil, err
	}
	results = append(results, scenarioResult{
		Name:     "03-reordered-evidence-sequence",
		Expected: "FAILED",
		Observed: string(reorderedResult.Status),
		Passed:   reorderedResult.Status == verification.StatusFailed,
		Checks: []string{
			"all evidence kinds are present",
			"evidence sequence is intentionally non-contiguous in append order",
			"completeness verifier rejected the logical ordering",
		},
	})

	// 4. Detached evidence: one record belongs to another run. The record is
	// valid evidence by itself, but it must not complete this run.
	detachedStore := buildStore([]eventSpec{
		{kind: "intent", run: runID},
		{kind: "decision", run: runID},
		{kind: "execution", run: "different-run"},
		{kind: "observation", run: runID},
		{kind: "verification", run: runID},
	})
	detachedResult, err := verifyStore(detachedStore)
	if err != nil {
		return nil, err
	}
	results = append(results, scenarioResult{
		Name:     "04-detached-run-identity",
		Expected: "FAILED",
		Observed: string(detachedResult.Status),
		Passed:   detachedResult.Status == verification.StatusFailed,
		Checks: []string{
			"execution evidence intentionally belongs to another run",
			"evidence remains hash-chain valid",
			"run completeness verifier rejected detached evidence",
		},
	})

	return results, nil
}

type eventSpec struct {
	kind     string
	run      string
	sequence int
}

func buildStore(specs []eventSpec) *evidence.EvidenceStore {
	store := evidence.NewStore()
	var previousID string
	for i, spec := range specs {
		sequence := spec.sequence
		if sequence == 0 {
			sequence = i + 1
		}
		id := fmt.Sprintf("am18-%02d-%s", i+1, spec.kind)
		rec, _ := store.Append(evidence.EvidenceRecord{
			ID:              id,
			Actor:           "agent/am18",
			Resource:        "repository",
			Action:          "write",
			Environment:     "experiment",
			PolicyID:        "am18-policy",
			PolicyVersion:   "1.0.0",
			Decision:        control.DecisionAllow,
			ExecutionStatus: "recorded",
			Metadata: map[string]any{
				"run_id":               spec.run,
				"kind":                 spec.kind,
				"sequence":             sequence,
				"previous_evidence_id": previousID,
			},
		})
		previousID = rec.ID
	}
	return store
}

func verifyStore(store *evidence.EvidenceStore) (verification.VerificationResult, error) {
	idx, err := store.Verify()
	if err != nil {
		return verification.VerificationResult{}, fmt.Errorf("evidence chain integrity failed at %d: %w", idx, err)
	}

	records := store.Records()
	if len(records) == 0 {
		return verification.VerificationResult{}, fmt.Errorf("empty evidence store")
	}

	chain := verification.NewChain(
		verification.Contract{
			Name:         "am18-evidence-completeness-v2",
			Version:      "2.0.0",
			Description:  "A consequential run requires complete, ordered, run-bound evidence.",
			Requirements: []string{"required evidence", "sequence continuity", "run identity", "evidence links"},
		},
		completenessVerifier{store: store},
	)
	return chain.Verify(records[0])
}

func metadataString(rec evidence.EvidenceRecord, key string) string {
	value, ok := rec.Metadata[key]
	if !ok {
		return ""
	}
	s, _ := value.(string)
	return s
}

func metadataInt(rec evidence.EvidenceRecord, key string) int {
	value, ok := rec.Metadata[key]
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

func summarize(results []scenarioResult) string {
	passed := 0
	for _, result := range results {
		if result.Passed {
			passed++
		}
	}
	return fmt.Sprintf("%d/%d expected scenario assertions passed", passed, len(results))
}

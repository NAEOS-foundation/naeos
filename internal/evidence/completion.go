// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"crypto/sha256"
	"fmt"
)

// CompletionResult describes whether a consequential run has satisfied its
// evidence completion contract.
type CompletionResult struct {
	Complete bool
	RunID    string
	Required []string
	Observed []string
	Missing  []string
	Checks   []CompletionCheck
}

// CompletionCheck records one enforcement invariant.
type CompletionCheck struct {
	Name   string
	Passed bool
	Detail string
}

// ProvenanceDigest returns a deterministic digest for the runtime provenance
// attached to one evidence record.
func ProvenanceDigest(stage, event string, payloadDigest string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("naeos:evidence:provenance:v1:%s:%s:%s", stage, event, payloadDigest)))
	return fmt.Sprintf("%x", h)
}

// ValidateCompletion is the controllable lifecycle boundary for consequential
// runs. Evidence is accepted only when required kinds are unique, the backing
// chain is intact, run records are contiguous, each record is bound to the
// requested run, predecessor identity is exact, and runtime provenance is valid.
func ValidateCompletion(store *EvidenceStore, runID string, requiredKinds []string) CompletionResult {
	result := CompletionResult{RunID: runID, Required: append([]string(nil), requiredKinds...)}
	if store == nil {
		result.Checks = append(result.Checks, CompletionCheck{Name: "store-present", Passed: false, Detail: "evidence store is nil"})
		result.Missing = append(result.Missing, requiredKinds...)
		return result
	}
	if runID == "" {
		result.Checks = append(result.Checks, CompletionCheck{Name: "run-identity-present", Passed: false, Detail: "run identity is required"})
		result.Missing = append(result.Missing, requiredKinds...)
		return result
	}
	if duplicates := duplicateStrings(requiredKinds); len(duplicates) > 0 {
		result.Checks = append(result.Checks, CompletionCheck{Name: "required-kinds-unique", Passed: false, Detail: fmt.Sprintf("duplicate required kinds=%v", duplicates)})
		return result
	}
	if brokenIndex, err := store.Verify(); err != nil {
		result.Checks = append(result.Checks, CompletionCheck{Name: "evidence-chain-intact", Passed: false, Detail: fmt.Sprintf("broken_index=%d error=%v", brokenIndex, err)})
		return result
	}

	records := store.Records()
	runRecords := make([]EvidenceRecord, 0, len(records))
	seenRun, endedRun, mixedRun := false, false, false
	for i := len(records) - 1; i >= 0; i-- {
		record := records[i]
		recordRunID := metadataString(record, "run_id")
		if recordRunID == runID {
			if endedRun {
				mixedRun = true
			}
			seenRun = true
			runRecords = append(runRecords, record)
			continue
		}
		if seenRun {
			endedRun = true
		}
	}
	if mixedRun {
		result.Checks = append(result.Checks, CompletionCheck{Name: "run-records-contiguous", Passed: false, Detail: "evidence for the target run is interleaved with another run"})
		return result
	}

	present := make(map[string]bool, len(runRecords))
	sequenceOK, linksOK, bindingOK, provenanceOK := true, true, true, true
	previousSequence := 0
	expectedBinding := RunBindingDigest(runID)
	expectedProvenance := map[string][2]string{
		"intent":       {"run", "pipeline.start"},
		"decision":     {"policy_eval", "pipeline.policy_decision"},
		"execution":    {"write_artifacts", "pipeline.execution"},
		"observation":  {"observation", "pipeline.observation"},
		"verification": {"completion", "pipeline.verification"},
	}
	for _, record := range runRecords {
		kind := metadataString(record, "kind")
		if kind != "" {
			present[kind] = true
		}
		if metadataString(record, "run_binding") != expectedBinding {
			bindingOK = false
		}
		sequence := metadataInt(record, "sequence")
		provenanceStage := metadataString(record, "provenance_stage")
		provenanceEvent := metadataString(record, "provenance_event")
		payloadDigest := metadataString(record, "payload_digest")
		provenanceDigest := metadataString(record, "provenance_digest")
		if expected, ok := expectedProvenance[kind]; !ok || provenanceStage != expected[0] || provenanceEvent != expected[1] || payloadDigest == "" || provenanceDigest != ProvenanceDigest(provenanceStage, provenanceEvent, payloadDigest) {
			provenanceOK = false
		}
		if sequence != previousSequence+1 {
			sequenceOK = false
		}
		previousSequence = sequence
		if sequence > 1 {
			previousID := metadataString(record, "previous_evidence_id")
			if previousID == "" {
				linksOK = false
			} else if recordIndex := findEvidenceIndex(runRecords, record.ID); recordIndex <= 0 || runRecords[recordIndex-1].ID != previousID {
				linksOK = false
			}
		}
	}
	for _, kind := range requiredKinds {
		if !present[kind] {
			result.Missing = append(result.Missing, kind)
		} else {
			result.Observed = append(result.Observed, kind)
		}
	}
	requiredOK := len(result.Missing) == 0
	countOK := len(runRecords) == len(requiredKinds)
	result.Checks = append(result.Checks,
		CompletionCheck{Name: "required-evidence-present", Passed: requiredOK, Detail: fmt.Sprintf("missing=%v", result.Missing)},
		CompletionCheck{Name: "evidence-count-exact", Passed: countOK, Detail: fmt.Sprintf("observed=%d required=%d", len(runRecords), len(requiredKinds))},
		CompletionCheck{Name: "evidence-sequence-contiguous", Passed: sequenceOK, Detail: fmt.Sprintf("sequence_ok=%v", sequenceOK)},
		CompletionCheck{Name: "evidence-links-present", Passed: linksOK, Detail: "each non-root record identifies its exact predecessor"},
		CompletionCheck{Name: "run-binding-valid", Passed: bindingOK, Detail: "each record carries the deterministic run binding digest"},
		CompletionCheck{Name: "runtime-provenance-valid", Passed: provenanceOK, Detail: "each record is bound to an expected runtime stage/event and payload digest"},
	)
	result.Complete = requiredOK && countOK && sequenceOK && linksOK && bindingOK && provenanceOK
	return result
}

func duplicateStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var duplicates []string
	for _, value := range values {
		if _, ok := seen[value]; ok {
			duplicates = append(duplicates, value)
			continue
		}
		seen[value] = struct{}{}
	}
	return duplicates
}

func metadataString(rec EvidenceRecord, key string) string {
	value, ok := rec.Metadata[key]
	if !ok {
		return ""
	}
	s, _ := value.(string)
	return s
}

func metadataInt(rec EvidenceRecord, key string) int {
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

func findEvidenceIndex(records []EvidenceRecord, id string) int {
	for i, record := range records {
		if record.ID == id {
			return i
		}
	}
	return -1
}

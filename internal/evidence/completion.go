// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import "fmt"

// CompletionResult describes whether a consequential run has satisfied its
// evidence completion contract.
type CompletionResult struct {
	Complete bool
	RunID string
	Required []string
	Observed []string
	Missing []string
	Checks []CompletionCheck
}

// CompletionCheck records one enforcement invariant.
type CompletionCheck struct {
	Name string
	Passed bool
	Detail string
}

// ValidateCompletion is the controllable lifecycle boundary for consequential
// runs. Evidence is accepted only when required kinds are unique, the backing
// chain is intact, run records are contiguous, each record is bound to the
// requested run, and predecessor identity is exact.
func ValidateCompletion(store *EvidenceStore, runID string, requiredKinds []string) CompletionResult {
	result := CompletionResult{RunID: runID, Required: append([]string(nil), requiredKinds...)}
	if store == nil {
		result.Checks = append(result.Checks, CompletionCheck{Name:"store-present", Passed:false, Detail:"evidence store is nil"})
		result.Missing = append(result.Missing, requiredKinds...)
		return result
	}
	if runID == "" {
		result.Checks = append(result.Checks, CompletionCheck{Name:"run-identity-present", Passed:false, Detail:"run identity is required"})
		result.Missing = append(result.Missing, requiredKinds...)
		return result
	}
	if duplicates := duplicateStrings(requiredKinds); len(duplicates) > 0 {
		result.Checks = append(result.Checks, CompletionCheck{Name:"required-kinds-unique", Passed:false, Detail:fmt.Sprintf("duplicate required kinds=%v", duplicates)})
		return result
	}
	if brokenIndex, err := store.Verify(); err != nil {
		result.Checks = append(result.Checks, CompletionCheck{Name:"evidence-chain-intact", Passed:false, Detail:fmt.Sprintf("broken_index=%d error=%v", brokenIndex, err)})
		return result
	}

	records := store.Records()
	runRecords := make([]EvidenceRecord, 0, len(records))
	seenRun, endedRun, mixedRun := false, false, false
	for i := len(records)-1; i >= 0; i-- {
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
		result.Checks = append(result.Checks, CompletionCheck{Name:"run-records-contiguous", Passed:false, Detail:"evidence for the target run is interleaved with another run"})
		return result
	}

	present := make(map[string]bool, len(runRecords))
	sequenceOK, linksOK, bindingOK := true, true, true
	previousSequence := 0
	expectedBinding := RunBindingDigest(runID)
	for _, record := range runRecords {
		kind := metadataString(record, "kind")
		if kind != "" { present[kind] = true }
		if metadataString(record, "run_binding") != expectedBinding { bindingOK = false }
		sequence := metadataInt(record, "sequence")
		if sequence != previousSequence+1 { sequenceOK = false }
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
		if !present[kind] { result.Missing = append(result.Missing, kind) } else { result.Observed = append(result.Observed, kind) }
	}
	requiredOK := len(result.Missing) == 0
	countOK := len(runRecords) == len(requiredKinds)
	result.Checks = append(result.Checks,
		CompletionCheck{Name:"required-evidence-present", Passed:requiredOK, Detail:fmt.Sprintf("missing=%v", result.Missing)},
		CompletionCheck{Name:"evidence-count-exact", Passed:countOK, Detail:fmt.Sprintf("observed=%d required=%d", len(runRecords), len(requiredKinds))},
		CompletionCheck{Name:"evidence-sequence-contiguous", Passed:sequenceOK, Detail:fmt.Sprintf("sequence_ok=%v", sequenceOK)},
		CompletionCheck{Name:"evidence-links-present", Passed:linksOK, Detail:"each non-root record identifies its exact predecessor"},
		CompletionCheck{Name:"run-binding-valid", Passed:bindingOK, Detail:"each record carries the deterministic run binding digest"},
	)
	result.Complete = requiredOK && countOK && sequenceOK && linksOK && bindingOK
	return result
}

func duplicateStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var duplicates []string
	for _, value := range values {
		if _, ok := seen[value]; ok { duplicates = append(duplicates, value); continue }
		seen[value] = struct{}{}
	}
	return duplicates
}

func metadataString(rec EvidenceRecord, key string) string {
	value, ok := rec.Metadata[key]
	if !ok { return "" }
	s, _ := value.(string)
	return s
}

func metadataInt(rec EvidenceRecord, key string) int {
	value, ok := rec.Metadata[key]
	if !ok { return 0 }
	switch v := value.(type) {
	case int: return v
	case float64: return int(v)
	default: return 0
	}
}

func findEvidenceIndex(records []EvidenceRecord, id string) int {
	for i, record := range records { if record.ID == id { return i } }
	return -1
}

// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import "fmt"

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

// ValidateCompletion is the controllable lifecycle boundary for consequential
// runs. A caller MUST gate run completion on Complete == true.
//
// The store may contain evidence for many runs. Only records bound to runID
// are considered part of this run; therefore evidence from another run cannot
// satisfy a required evidence kind.
func ValidateCompletion(store *EvidenceStore, runID string, requiredKinds []string) CompletionResult {
	result := CompletionResult{
		RunID:    runID,
		Required: append([]string(nil), requiredKinds...),
	}

	if store == nil {
		result.Checks = append(result.Checks, CompletionCheck{
			Name:   "store-present",
			Passed: false,
			Detail: "evidence store is nil",
		})
		result.Missing = append(result.Missing, requiredKinds...)
		return result
	}
	if runID == "" {
		result.Checks = append(result.Checks, CompletionCheck{
			Name:   "run-identity-present",
			Passed: false,
			Detail: "run identity is required",
		})
		result.Missing = append(result.Missing, requiredKinds...)
		return result
	}

	records := store.Records()
	runRecords := make([]EvidenceRecord, 0, len(records))
	for i := len(records) - 1; i >= 0; i-- {
		if metadataString(records[i], "run_id") == runID {
			runRecords = append(runRecords, records[i])
		}
	}

	present := make(map[string]bool, len(runRecords))
	sequenceOK := true
	linksOK := true
	previousSequence := 0

	for _, record := range runRecords {
		kind := metadataString(record, "kind")
		if kind != "" {
			present[kind] = true
		}

		sequence := metadataInt(record, "sequence")
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
		CompletionCheck{
			Name:   "required-evidence-present",
			Passed: requiredOK,
			Detail: fmt.Sprintf("missing=%v", result.Missing),
		},
		CompletionCheck{
			Name:   "evidence-count-exact",
			Passed: countOK,
			Detail: fmt.Sprintf("observed=%d required=%d", len(runRecords), len(requiredKinds)),
		},
		CompletionCheck{
			Name:   "evidence-sequence-contiguous",
			Passed: sequenceOK,
			Detail: fmt.Sprintf("sequence_ok=%v", sequenceOK),
		},
		CompletionCheck{
			Name:   "evidence-links-present",
			Passed: linksOK,
			Detail: "each non-root record identifies its predecessor",
		},
	)

	result.Complete = requiredOK && countOK && sequenceOK && linksOK
	return result
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

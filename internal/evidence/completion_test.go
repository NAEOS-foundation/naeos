// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

func completionStore(specs []struct {
	kind string
	run  string
	seq  int
}) *EvidenceStore {
	store := NewStore()
	var previousID string
	for i, spec := range specs {
		rec, err := store.Append(EvidenceRecord{
			ID:              spec.kind + "-" + spec.run,
			Actor:           "test",
			Resource:        "test",
			Action:          "write",
			Environment:     "test",
			PolicyID:        "test",
			PolicyVersion:   "1.0.0",
			Decision:        control.DecisionAllow,
			ExecutionStatus: "recorded",
			Metadata: map[string]any{
				"run_id":               spec.run,
				"kind":                 spec.kind,
				"sequence":             spec.seq,
				"previous_evidence_id": previousID,
			},
		})
		if err != nil {
			panic(err)
		}
		previousID = rec.ID
		if spec.seq == 0 {
			_ = i
		}
	}
	return store
}

var completionKinds = []string{"intent", "decision", "execution", "observation", "verification"}

func TestValidateCompletionAllowsCompleteRun(t *testing.T) {
	store := completionStore([]struct {
		kind string
		run  string
		seq  int
	}{
		{"intent", "run-1", 1},
		{"decision", "run-1", 2},
		{"execution", "run-1", 3},
		{"observation", "run-1", 4},
		{"verification", "run-1", 5},
	})

	result := ValidateCompletion(store, "run-1", completionKinds)
	if !result.Complete {
		t.Fatalf("expected complete run, got %+v", result)
	}
}

func TestValidateCompletionBlocksMissingEvidence(t *testing.T) {
	store := completionStore([]struct {
		kind string
		run  string
		seq  int
	}{
		{"intent", "run-1", 1},
		{"decision", "run-1", 2},
		{"execution", "run-1", 3},
		{"verification", "run-1", 4},
	})

	result := ValidateCompletion(store, "run-1", completionKinds)
	if result.Complete {
		t.Fatalf("expected incomplete run, got %+v", result)
	}
	if len(result.Missing) != 1 || result.Missing[0] != "observation" {
		t.Fatalf("expected missing observation, got %v", result.Missing)
	}
}

func TestValidateCompletionBlocksReorderedSequence(t *testing.T) {
	store := completionStore([]struct {
		kind string
		run  string
		seq  int
	}{
		{"intent", "run-1", 1},
		{"decision", "run-1", 2},
		{"execution", "run-1", 3},
		{"verification", "run-1", 5},
		{"observation", "run-1", 4},
	})

	result := ValidateCompletion(store, "run-1", completionKinds)
	if result.Complete {
		t.Fatalf("expected reordered run to be blocked, got %+v", result)
	}
}

func TestValidateCompletionBlocksDetachedEvidence(t *testing.T) {
	store := completionStore([]struct {
		kind string
		run  string
		seq  int
	}{
		{"intent", "run-1", 1},
		{"decision", "run-1", 2},
		{"execution", "run-2", 3},
		{"observation", "run-1", 4},
		{"verification", "run-1", 5},
	})

	result := ValidateCompletion(store, "run-1", completionKinds)
	if result.Complete {
		t.Fatalf("expected detached evidence to block completion, got %+v", result)
	}
}

func TestValidateCompletionRejectsNilStoreAndEmptyRunID(t *testing.T) {
	if result := ValidateCompletion(nil, "run-1", completionKinds); result.Complete {
		t.Fatal("nil store must not complete a run")
	}
	if result := ValidateCompletion(NewStore(), "", completionKinds); result.Complete {
		t.Fatal("empty run identity must not complete a run")
	}
}

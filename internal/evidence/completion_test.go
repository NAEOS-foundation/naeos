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
				"run_binding":          RunBindingDigest(spec.run),
				"kind":                 spec.kind,
				"sequence":             spec.seq,
				"provenance_stage":     testProvenance(spec.kind)[0],
				"provenance_event":     testProvenance(spec.kind)[1],
				"payload_digest":       "payload-" + spec.kind,
				"provenance_digest":    ProvenanceDigest(testProvenance(spec.kind)[0], testProvenance(spec.kind)[1], "payload-"+spec.kind),
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

func testProvenance(kind string) [2]string {
	switch kind {
	case "intent":
		return [2]string{"run", "pipeline.start"}
	case "decision":
		return [2]string{"policy_eval", "pipeline.policy_decision"}
	case "execution":
		return [2]string{"write_artifacts", "pipeline.execution"}
	case "observation":
		return [2]string{"observation", "pipeline.observation"}
	case "verification":
		return [2]string{"completion", "pipeline.verification"}
	default:
		return [2]string{"unknown", "unknown"}
	}
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

func TestValidateCompletionRejectsDuplicateRequiredKinds(t *testing.T) {
	store := completionStore([]struct {
		kind, run string
		seq       int
	}{
		{"intent", "run-1", 1},
	})
	result := ValidateCompletion(store, "run-1", []string{"intent", "intent"})
	if result.Complete {
		t.Fatal("duplicate required kinds must block completion")
	}
}

func TestValidateCompletionRejectsMixedRunEvidence(t *testing.T) {
	store := completionStore([]struct {
		kind, run string
		seq       int
	}{
		{"intent", "run-1", 1},
		{"decision", "run-1", 2},
		{"intent", "run-2", 1},
		{"execution", "run-1", 3},
		{"observation", "run-1", 4},
		{"verification", "run-1", 5},
	})
	result := ValidateCompletion(store, "run-1", completionKinds)
	if result.Complete {
		t.Fatal("interleaved evidence from another run must block completion")
	}
}

func TestValidateCompletionRejectsRunBindingTampering(t *testing.T) {
	store := completionStore([]struct {
		kind, run string
		seq       int
	}{
		{"intent", "run-1", 1},
		{"decision", "run-1", 2},
		{"execution", "run-1", 3},
		{"observation", "run-1", 4},
		{"verification", "run-1", 5},
	})
	records := store.records
	records[2].Metadata["run_binding"] = RunBindingDigest("other-run")
	result := ValidateCompletion(store, "run-1", completionKinds)
	if result.Complete {
		t.Fatal("tampered run binding must block completion")
	}
}

func TestValidateCompletionRejectsTamperedEvidenceChain(t *testing.T) {
	store := completionStore([]struct {
		kind, run string
		seq       int
	}{
		{"intent", "run-1", 1},
		{"decision", "run-1", 2},
		{"execution", "run-1", 3},
		{"observation", "run-1", 4},
		{"verification", "run-1", 5},
	})
	store.records[2].ExecutionStatus = "tampered"
	result := ValidateCompletion(store, "run-1", completionKinds)
	if result.Complete {
		t.Fatal("tampered evidence content must block completion")
	}
}

func TestValidateCompletionRejectsMissingRuntimeProvenance(t *testing.T) {
	store := completionStore([]struct {
		kind, run string
		seq       int
	}{
		{"intent", "run-1", 1}, {"decision", "run-1", 2}, {"execution", "run-1", 3},
		{"observation", "run-1", 4}, {"verification", "run-1", 5},
	})
	store.records[2].Metadata["provenance_event"] = ""
	result := ValidateCompletion(store, "run-1", completionKinds)
	if result.Complete {
		t.Fatal("missing runtime provenance must block completion")
	}
}

func TestProvenanceDigestChangesWithPayload(t *testing.T) {
	a := ProvenanceDigest("policy_eval", "pipeline.policy_decision", "digest-a")
	b := ProvenanceDigest("policy_eval", "pipeline.policy_decision", "digest-b")
	if a == b {
		t.Fatal("provenance digest must change when payload digest changes")
	}
}

func TestValidateCompletionRejectsProvenanceDigestMismatch(t *testing.T) {
	store := completionStore([]struct {
		kind, run string
		seq       int
	}{
		{"intent", "run-1", 1}, {"decision", "run-1", 2}, {"execution", "run-1", 3},
		{"observation", "run-1", 4}, {"verification", "run-1", 5},
	})
	store.records[1].Metadata["payload_digest"] = "tampered-payload"
	result := ValidateCompletion(store, "run-1", completionKinds)
	if result.Complete {
		t.Fatal("provenance digest mismatch must block completion")
	}
}

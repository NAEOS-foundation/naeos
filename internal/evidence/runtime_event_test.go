// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import "testing"

func appendTestLifecycleEvidence(t *testing.T, store *EvidenceStore, events *RuntimeEventStore, runID string) {
	t.Helper()
	spec := []struct {
		kind, stage, name, payload string
	}{
		{"intent", "run", "pipeline.start", "payload-intent"},
		{"decision", "policy_eval", "pipeline.policy_decision", "payload-decision"},
		{"execution", "write_artifacts", "pipeline.execution", "payload-execution"},
		{"observation", "observation", "pipeline.observation", "payload-observation"},
		{"verification", "completion", "pipeline.verification", "payload-verification"},
	}
	var previous string
	for i, item := range spec {
		event, err := events.Append(runID, item.name, item.payload, i+1)
		if err != nil {
			t.Fatal(err)
		}
		metadata := map[string]any{
			"run_id":               runID,
			"run_binding":          RunBindingDigest(runID),
			"kind":                 item.kind,
			"sequence":             i + 1,
			"previous_evidence_id": previous,
			"provenance_stage":     item.stage,
			"provenance_event":     item.name,
			"payload_digest":       item.payload,
			"provenance_digest":    ProvenanceDigest(item.stage, item.name, item.payload),
			"runtime_event_id":     event.ID,
		}
		record, err := store.Append(EvidenceRecord{ID: runID + "-" + item.kind, Metadata: metadata})
		if err != nil {
			t.Fatal(err)
		}
		previous = record.ID
	}
}

func TestValidateCompletionWithRuntimeEventsAllowsCompleteBinding(t *testing.T) {
	store := NewStore()
	events := NewRuntimeEventStore()
	appendTestLifecycleEvidence(t, store, events, "run-valid")
	result := ValidateCompletionWithRuntimeEvents(store, events, "run-valid", []string{"intent", "decision", "execution", "observation", "verification"})
	if !result.Complete {
		t.Fatalf("expected complete runtime binding, got %#v", result.Checks)
	}
}

func TestValidateCompletionWithRuntimeEventsBlocksMissingEvent(t *testing.T) {
	store := NewStore()
	events := NewRuntimeEventStore()
	appendTestLifecycleEvidence(t, store, events, "run-missing")
	records := events.Records()
	_ = records
	// A detached evidence reference cannot be satisfied by a missing event.
	event := store.ByID("run-missing-verification")
	event.Metadata["runtime_event_id"] = "evt-missing"
	result := ValidateCompletionWithRuntimeEvents(store, events, "run-missing", []string{"intent", "decision", "execution", "observation", "verification"})
	if result.Complete {
		t.Fatal("expected missing runtime event binding to block completion")
	}
}

func TestValidateCompletionWithRuntimeEventsBlocksEventTypeMismatch(t *testing.T) {
	store := NewStore()
	events := NewRuntimeEventStore()
	appendTestLifecycleEvidence(t, store, events, "run-type")
	record := store.ByID("run-type-execution")
	record.Metadata["provenance_event"] = "pipeline.wrong"
	result := ValidateCompletionWithRuntimeEvents(store, events, "run-type", []string{"intent", "decision", "execution", "observation", "verification"})
	if result.Complete {
		t.Fatal("expected event type mismatch to block completion")
	}
}

func TestValidateCompletionWithRuntimeEventsBlocksPayloadMismatch(t *testing.T) {
	store := NewStore()
	events := NewRuntimeEventStore()
	appendTestLifecycleEvidence(t, store, events, "run-payload")
	event := store.ByID("run-payload-execution")
	event.Metadata["payload_digest"] = "tampered"
	result := ValidateCompletionWithRuntimeEvents(store, events, "run-payload", []string{"intent", "decision", "execution", "observation", "verification"})
	if result.Complete {
		t.Fatal("expected payload mismatch to block completion")
	}
}

func TestValidateCompletionWithRuntimeEventsBlocksAnotherRun(t *testing.T) {
	store := NewStore()
	events := NewRuntimeEventStore()
	appendTestLifecycleEvidence(t, store, events, "run-one")
	other, err := events.Append("run-two", "pipeline.verification", "payload-other", 6)
	if err != nil {
		t.Fatal(err)
	}
	_ = other
	result := ValidateCompletionWithRuntimeEvents(store, events, "run-one", []string{"intent", "decision", "execution", "observation", "verification"})
	if result.Complete {
		t.Fatal("expected mixed-run runtime event to block completion")
	}
}

func TestValidateCompletionWithRuntimeEventsBlocksEventAfterCompletion(t *testing.T) {
	store := NewStore()
	events := NewRuntimeEventStore()
	appendTestLifecycleEvidence(t, store, events, "run-after")
	if _, err := events.Append("run-after", "pipeline.after_completion", "payload-after", 6); err != nil {
		t.Fatal(err)
	}
	result := ValidateCompletionWithRuntimeEvents(store, events, "run-after", []string{"intent", "decision", "execution", "observation", "verification"})
	if result.Complete {
		t.Fatal("expected event after completion boundary to block completion")
	}
}

// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import "testing"

func TestRuntimeEventLedgerSealBlocksLateEvent(t *testing.T) {
	ledger := NewRuntimeEventLedger()
	if _, err := ledger.Publish("run-1", "pipeline.start", "payload-1", 1); err != nil {
		t.Fatal(err)
	}
	ledger.Seal()
	if _, err := ledger.Publish("run-1", "pipeline.execution", "payload-2", 2); err == nil {
		t.Fatal("expected sealed ledger to reject late event")
	}
}

func TestRuntimeEvidenceBuilderRequiresExistingEvent(t *testing.T) {
	store := NewStore()
	ledger := NewRuntimeEventLedger()
	builder := NewRuntimeEvidenceBuilder(store, ledger)
	event, err := ledger.Publish("run-1", "pipeline.start", "payload-1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.Build("run-1", "intent", 1, "run", "pipeline.start", event); err != nil {
		t.Fatal(err)
	}
	if len(store.Records()) != 1 {
		t.Fatalf("expected one evidence record, got %d", len(store.Records()))
	}
}

func TestRuntimeEvidenceBuilderRejectsForeignEvent(t *testing.T) {
	store := NewStore()
	ledger := NewRuntimeEventLedger()
	builder := NewRuntimeEvidenceBuilder(store, ledger)
	event, err := ledger.Publish("run-2", "pipeline.start", "payload-1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.Build("run-1", "intent", 1, "run", "pipeline.start", event); err == nil {
		t.Fatal("expected cross-run event to be rejected")
	}
}

func TestRuntimeEvidenceBuilderRejectsEventPayloadMismatch(t *testing.T) {
	store := NewStore()
	ledger := NewRuntimeEventLedger()
	builder := NewRuntimeEvidenceBuilder(store, ledger)
	event, err := ledger.Publish("run-1", "pipeline.start", "payload-1", 1)
	if err != nil {
		t.Fatal(err)
	}
	event.PayloadDigest = "tampered"
	if err := builder.Build("run-1", "intent", 1, "run", "pipeline.start", event); err == nil {
		t.Fatal("expected mutated event reference to be rejected")
	}
}

func TestCompletionRejectsMixedRunLedger(t *testing.T) {
	store := NewStore()
	ledger := NewRuntimeEventLedger()
	builder := NewRuntimeEvidenceBuilder(store, ledger)
	for i, item := range []struct {
		kind  string
		stage string
		event string
		runID string
	}{
		{"intent", "run", "pipeline.start", "run-1"},
		{"decision", "policy_eval", "pipeline.policy_decision", "run-2"},
		{"execution", "write_artifacts", "pipeline.execution", "run-1"},
		{"observation", "observation", "pipeline.observation", "run-1"},
		{"verification", "completion", "pipeline.verification", "run-1"},
	} {
		event, err := ledger.Publish(item.runID, item.event, "payload", i+1)
		if err != nil {
			t.Fatal(err)
		}
		if item.runID == "run-1" {
			if err := builder.Build("run-1", item.kind, i+1, item.stage, item.event, event); err != nil {
				t.Fatal(err)
			}
		}
	}
	ledger.Seal()
	result := ValidateCompletionWithRuntimeLedger(store, ledger, "run-1", []string{"intent", "decision", "execution", "observation", "verification"})
	if result.Complete {
		t.Fatal("expected mixed-run runtime ledger to block completion")
	}
}

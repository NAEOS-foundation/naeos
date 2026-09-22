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

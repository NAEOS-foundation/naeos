// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import "testing"

func TestIndependentRuntimeObserverSeparatesObservationFromEvidence(t *testing.T) {
	observer := NewIndependentRuntimeObserver()
	event, err := observer.Observe("run-1", "pipeline.start", "payload-1", 1)
	if err != nil {
		t.Fatal(err)
	}

	var readOnly RuntimeEventObserver = observer
	if observed := readOnly.ByID(event.ID); observed == nil {
		t.Fatal("expected observer event to be readable")
	}

	store := NewStore()
	builder := NewRuntimeEvidenceBuilder(store, readOnly)
	if err := builder.Build("run-1", "intent", 1, "run", "pipeline.start", event); err != nil {
		t.Fatal(err)
	}
	if len(store.Records()) != 1 {
		t.Fatalf("expected one evidence record, got %d", len(store.Records()))
	}
}

func TestIndependentRuntimeObserverRejectsLateObservation(t *testing.T) {
	observer := NewIndependentRuntimeObserver()
	if _, err := observer.Observe("run-1", "pipeline.start", "payload-1", 1); err != nil {
		t.Fatal(err)
	}
	observer.Seal()
	if _, err := observer.Observe("run-1", "pipeline.execution", "payload-2", 2); err == nil {
		t.Fatal("expected sealed observer to reject late observation")
	}
}

func TestIndependentRuntimeObserverDetectsTamperedSnapshot(t *testing.T) {
	observer := NewIndependentRuntimeObserver()
	if _, err := observer.Observe("run-1", "pipeline.start", "payload-1", 1); err != nil {
		t.Fatal(err)
	}
	records := observer.Records()
	if len(records) != 1 {
		t.Fatal("expected one record")
	}
	records[0].PayloadDigest = "tampered"
	if err := observer.Verify(); err != nil {
		t.Fatalf("observer should remain intact after snapshot mutation: %v", err)
	}
}

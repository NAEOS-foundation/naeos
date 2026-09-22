// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import "fmt"

// RuntimeEventObserver is the read-only observation boundary consumed by
// evidence and completion code. It intentionally exposes no publish method.
type RuntimeEventObserver interface {
	Records() []RuntimeEvent
	ByID(id string) *RuntimeEvent
	Verify() error
	Seal()
}

// IndependentRuntimeObserver is the lifecycle-facing observer. Runtime code
// sends observations here; evidence construction cannot manufacture events.
type IndependentRuntimeObserver struct {
	ledger *RuntimeEventLedger
}

// NewIndependentRuntimeObserver creates a fresh observer with a private ledger.
func NewIndependentRuntimeObserver() *IndependentRuntimeObserver {
	return &IndependentRuntimeObserver{ledger: NewRuntimeEventLedger()}
}

// Observe records one runtime observation before it can be referenced by
// evidence. The observer owns event publication; consumers receive only the
// read-only observation interface.
func (o *IndependentRuntimeObserver) Observe(runID, name, payloadDigest string, sequence int) (RuntimeEvent, error) {
	if o == nil || o.ledger == nil {
		return RuntimeEvent{}, fmt.Errorf("runtime observer is nil")
	}
	return o.ledger.Publish(runID, name, payloadDigest, sequence)
}

// Records returns an append-order snapshot without granting append access.
func (o *IndependentRuntimeObserver) Records() []RuntimeEvent {
	if o == nil || o.ledger == nil {
		return nil
	}
	return o.ledger.Records()
}

// ByID resolves an observed event without granting append access.
func (o *IndependentRuntimeObserver) ByID(id string) *RuntimeEvent {
	if o == nil || o.ledger == nil {
		return nil
	}
	return o.ledger.ByID(id)
}

// Verify validates all observations before completion.
func (o *IndependentRuntimeObserver) Verify() error {
	if o == nil || o.ledger == nil {
		return fmt.Errorf("runtime observer is nil")
	}
	return o.ledger.Verify()
}

// Seal closes the observation window.
func (o *IndependentRuntimeObserver) Seal() {
	if o != nil && o.ledger != nil {
		o.ledger.Seal()
	}
}

// Ledger exposes the underlying ledger only to the completion boundary.
// This keeps publication capability outside evidence construction while
// preserving the existing V5.3 completion validator during the migration.
func (o *IndependentRuntimeObserver) Ledger() *RuntimeEventLedger {
	if o == nil {
		return nil
	}
	return o.ledger
}

var _ RuntimeEventObserver = (*IndependentRuntimeObserver)(nil)

// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import "fmt"

import "github.com/NAEOS-foundation/naeos/internal/governance/control"

// RuntimeEvidenceBuilder converts already-observed ledger events into evidence.
// It has no capability to publish runtime events, which keeps observation and
// evidence construction on separate trust boundaries.
type RuntimeEvidenceBuilder struct {
	store  *EvidenceStore
	ledger *RuntimeEventLedger
}

// NewRuntimeEvidenceBuilder creates a builder bound to one evidence store and
// one independent runtime event ledger.
func NewRuntimeEvidenceBuilder(store *EvidenceStore, ledger *RuntimeEventLedger) *RuntimeEvidenceBuilder {
	return &RuntimeEvidenceBuilder{store: store, ledger: ledger}
}

// Build appends evidence only for an event that already exists in the ledger.
// The event ID and payload digest are copied from the ledger, never supplied by
// the evidence caller.
func (b *RuntimeEvidenceBuilder) Build(runID, kind string, sequence int, stage, event string, runtimeEvent RuntimeEvent) error {
	if b == nil || b.store == nil || b.ledger == nil {
		return fmt.Errorf("runtime evidence builder requires evidence store and runtime event ledger")
	}
	if runID == "" || kind == "" || sequence <= 0 {
		return fmt.Errorf("runtime evidence requires run_id, kind, and positive sequence")
	}
	observed := b.ledger.ByID(runtimeEvent.ID)
	if observed == nil {
		return fmt.Errorf("runtime event %s is not present in ledger", runtimeEvent.ID)
	}
	if observed.RunID != runID || observed.Sequence != sequence || observed.Name != event || observed.PayloadDigest == "" {
		return fmt.Errorf("runtime event %s does not match requested evidence binding", runtimeEvent.ID)
	}
	if err := b.ledger.Verify(); err != nil {
		return fmt.Errorf("runtime event ledger verification failed: %w", err)
	}
	previousID := ""
	if latest := b.store.Latest(); latest != nil {
		previousID = latest.ID
	}
	_, err := b.store.Append(EvidenceRecord{
		ID:              "evidence-" + observed.ID + "-" + kind,
		Actor:           "runtime-observer",
		Resource:        "pipeline",
		Action:          "observed",
		Environment:     "runtime",
		PolicyID:        "pipeline-lifecycle",
		PolicyVersion:   "1.0.0",
		Decision:        control.DecisionAllow,
		ExecutionStatus: "observed",
		Metadata: map[string]any{
			"run_id":               runID,
			"run_binding":          RunBindingDigest(runID),
			"kind":                 kind,
			"sequence":             sequence,
			"previous_evidence_id": previousID,
			"provenance_stage":     stage,
			"provenance_event":     event,
			"payload_digest":       observed.PayloadDigest,
			"provenance_digest":    ProvenanceDigest(stage, event, observed.PayloadDigest),
			"runtime_event_id":     observed.ID,
		},
	})
	return err
}

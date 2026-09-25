// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
)

func TestValidateRunCompletionBlocksIncompleteEvidence(t *testing.T) {
	store := evidence.NewStore()
	durableLedger, err := evidence.NewDurableRuntimeEventLedger(t.TempDir() + "/runtime.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	observer := evidence.NewIndependentRuntimeObserverWithDurableLedger(durableLedger)
	builder := evidence.NewRuntimeEvidenceBuilder(store, observer)
	if err := appendRunEvidence(builder, observer, "run-test", "intent", 1, "run", "pipeline.start", "payload-intent"); err != nil {
		t.Fatal(err)
	}
	if err := validateRunCompletion(store, observer.Ledger(), nil, evidence.DurableRuntimeReceipt{}, "run-test"); err == nil {
		t.Fatal("expected incomplete evidence to block completion")
	} else if !strings.Contains(err.Error(), "run completion blocked") {
		t.Fatalf("unexpected completion error: %v", err)
	}
}

func TestValidateRunCompletionAllowsCompleteEvidence(t *testing.T) {
	store := evidence.NewStore()
	durableLedger, err := evidence.NewDurableRuntimeEventLedger(t.TempDir() + "/runtime.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	observer := evidence.NewIndependentRuntimeObserverWithDurableLedger(durableLedger)
	builder := evidence.NewRuntimeEvidenceBuilder(store, observer)
	stages := [][2]string{
		{"run", "pipeline.start"},
		{"policy_eval", "pipeline.policy_decision"},
		{"write_artifacts", "pipeline.execution"},
		{"observation", "pipeline.observation"},
		{"completion", "pipeline.verification"},
	}
	for i, kind := range requiredRunEvidenceKinds {
		if err := appendRunEvidence(builder, observer, "run-test", kind, i+1, stages[i][0], stages[i][1], "payload"); err != nil {
			t.Fatal(err)
		}
	}
	observer.Seal()
	receipt, err := evidence.CreateDurableRuntimeReceipt(durableLedger, "run-test", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRunCompletion(store, observer.Ledger(), durableLedger, receipt, "run-test"); err != nil {
		t.Fatalf("expected complete evidence to allow completion: %v", err)
	}
}

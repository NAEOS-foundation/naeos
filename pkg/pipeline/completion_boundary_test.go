// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
)

func TestValidateRunCompletionBlocksIncompleteEvidence(t *testing.T) {
	store := evidence.NewStore()
	ledger := evidence.NewRuntimeEventLedger()
	builder := evidence.NewRuntimeEvidenceBuilder(store, ledger)
	if err := appendRunEvidence(builder, ledger, "run-test", "intent", 1, "run", "pipeline.start", "payload-intent"); err != nil {
		t.Fatal(err)
	}
	if err := validateRunCompletion(store, ledger, "run-test"); err == nil {
		t.Fatal("expected incomplete evidence to block completion")
	} else if !strings.Contains(err.Error(), "run completion blocked") {
		t.Fatalf("unexpected completion error: %v", err)
	}
}

func TestValidateRunCompletionAllowsCompleteEvidence(t *testing.T) {
	store := evidence.NewStore()
	ledger := evidence.NewRuntimeEventLedger()
	builder := evidence.NewRuntimeEvidenceBuilder(store, ledger)
	stages := [][2]string{
		{"run", "pipeline.start"},
		{"policy_eval", "pipeline.policy_decision"},
		{"write_artifacts", "pipeline.execution"},
		{"observation", "pipeline.observation"},
		{"completion", "pipeline.verification"},
	}
	for i, kind := range requiredRunEvidenceKinds {
		if err := appendRunEvidence(builder, ledger, "run-test", kind, i+1, stages[i][0], stages[i][1], "payload"); err != nil {
			t.Fatal(err)
		}
	}
	if err := validateRunCompletion(store, ledger, "run-test"); err != nil {
		t.Fatalf("expected complete evidence to allow completion: %v", err)
	}
}

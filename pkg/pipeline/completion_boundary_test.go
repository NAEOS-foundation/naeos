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
	if err := appendRunEvidence(store, "run-test", "intent", 1); err != nil {
		t.Fatal(err)
	}
	if err := validateRunCompletion(store, "run-test"); err == nil {
		t.Fatal("expected incomplete evidence to block completion")
	} else if !strings.Contains(err.Error(), "run completion blocked") {
		t.Fatalf("unexpected completion error: %v", err)
	}
}

func TestValidateRunCompletionAllowsCompleteEvidence(t *testing.T) {
	store := evidence.NewStore()
	for i, kind := range requiredRunEvidenceKinds {
		if err := appendRunEvidence(store, "run-test", kind, i+1); err != nil {
			t.Fatal(err)
		}
	}
	if err := validateRunCompletion(store, "run-test"); err != nil {
		t.Fatalf("expected complete evidence to allow completion: %v", err)
	}
}

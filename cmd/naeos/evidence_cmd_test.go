// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
)

func TestEvidenceLogPersistsRequestID(t *testing.T) {
	store := evidence.NewStore()
	saved, err := store.Append(evidence.EvidenceRecord{
		RequestID: "req-p26-001",
		Actor:     "test",
		Action:    "read",
	})
	if err != nil {
		t.Fatalf("append evidence: %v", err)
	}
	if saved.RequestID != "req-p26-001" {
		t.Fatalf("expected request ID to persist, got %q", saved.RequestID)
	}
}

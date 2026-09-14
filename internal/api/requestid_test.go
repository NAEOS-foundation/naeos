// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"testing"
)

func TestGenerateRequestIDNotEmpty(t *testing.T) {
	id := GenerateRequestID()
	if id == "" {
		t.Fatal("expected non-empty request ID")
	}
	if len(id) != 36 {
		t.Fatalf("expected 36-char UUID, got %d chars: %s", len(id), id)
	}
}

func TestGenerateRequestIDUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateRequestID()
		if seen[id] {
			t.Fatalf("duplicate request ID on iteration %d: %s", i, id)
		}
		seen[id] = true
	}
}

func TestContextRoundTrip(t *testing.T) {
	id := "test-request-id-123"
	ctx := ContextWithRequestID(context.Background(), id)
	got := RequestIDFromContext(ctx)
	if got != id {
		t.Fatalf("expected %q, got %q", id, got)
	}
}

func TestRequestIDFromEmptyContext(t *testing.T) {
	got := RequestIDFromContext(context.Background())
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

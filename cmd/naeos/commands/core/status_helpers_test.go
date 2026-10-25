// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"testing"
)

func TestJoinStrings(t *testing.T) {
	if got := joinStrings(nil); got != "(none)" {
		t.Fatalf("expected (none), got %q", got)
	}
	if got := joinStrings([]string{"go", "ts"}); got != "go, ts" {
		t.Fatalf("expected 'go, ts', got %q", got)
	}
}

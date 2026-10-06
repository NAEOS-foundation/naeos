// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"testing"
)

func TestExitCodeError(t *testing.T) {
	err := newExitCodeError(ExitCodePolicyDeny, "denied")
	var typed *ExitCodeError
	if !errors.As(err, &typed) {
		t.Fatal("expected ExitCodeError")
	}
	if typed.Code != ExitCodePolicyDeny {
		t.Fatalf("got %d", typed.Code)
	}
	if err.Error() != "denied" {
		t.Fatalf("unexpected message: %v", err)
	}
}

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package clihelpers

import (
	"errors"
	"testing"
)

func TestExitCodeError(t *testing.T) {
	err := NewExitCodeError(ExitCodePolicyDeny, "denied")
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

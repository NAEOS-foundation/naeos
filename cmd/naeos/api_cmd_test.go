// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestAPICommandSecretUsesEnvironmentDefault(t *testing.T) {
	previous := apiSecret
	t.Cleanup(func() {
		apiSecret = previous
	})
	t.Setenv("NAEOS_API_JWT_SECRET", "environment-secret")

	cmd := newAPICommand()
	if apiSecret != "environment-secret" {
		t.Fatalf("expected secret from environment, got %q", apiSecret)
	}

	if err := cmd.ParseFlags([]string{"--secret", "flag-secret"}); err != nil {
		t.Fatalf("parse secret flag: %v", err)
	}
	if apiSecret != "flag-secret" {
		t.Fatalf("expected explicit flag to override environment, got %q", apiSecret)
	}
}

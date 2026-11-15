// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestValidateAPISecretSeparation(t *testing.T) {
	tests := []struct {
		name          string
		jwtSecret     string
		encryptionKey string
		wantErr       bool
	}{
		{
			name:          "same secrets are rejected",
			jwtSecret:     "shared-secret",
			encryptionKey: "shared-secret",
			wantErr:       true,
		},
		{
			name:          "different secrets are accepted",
			jwtSecret:     "jwt-secret",
			encryptionKey: "encryption-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAPISecretSeparation(tt.jwtSecret, tt.encryptionKey)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateAPISecretSeparation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAPICommandRejectsSameJWTAndEncryptionSecret(t *testing.T) {
	previousAuth, previousSecret := apiAuth, apiSecret
	t.Cleanup(func() {
		apiAuth, apiSecret = previousAuth, previousSecret
	})
	t.Setenv("NAEOS_API_JWT_SECRET", "shared-secret")
	t.Setenv("NAEOS_ENCRYPTION_KEY", "shared-secret")

	cmd := newAPICommand()
	if err := cmd.ParseFlags([]string{"--auth"}); err != nil {
		t.Fatalf("parse auth flag: %v", err)
	}
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected API startup to reject identical JWT and encryption secrets")
	}
}

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

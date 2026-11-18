// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"strings"
	"testing"
)

func TestValidateAPIAuthSecrets(t *testing.T) {
	tests := []struct {
		name          string
		jwtSecret     string
		encryptionKey string
		wantErr       bool
	}{
		{
			name:          "accepts distinct strong secrets",
			jwtSecret:     strings.Repeat("j", 32),
			encryptionKey: strings.Repeat("e", 32),
		},
		{
			name:          "rejects short jwt secret",
			jwtSecret:     "short",
			encryptionKey: strings.Repeat("e", 32),
			wantErr:       true,
		},
		{
			name:          "rejects short encryption key",
			jwtSecret:     strings.Repeat("j", 32),
			encryptionKey: "short",
			wantErr:       true,
		},
		{
			name:          "rejects identical secrets",
			jwtSecret:     strings.Repeat("x", 32),
			encryptionKey: strings.Repeat("x", 32),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAPIAuthSecrets(tt.jwtSecret, tt.encryptionKey)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateAPIAuthSecrets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveAPIJWTSecret(t *testing.T) {
	previous := apiSecret
	t.Cleanup(func() {
		apiSecret = previous
	})
	apiSecret = ""
	t.Setenv("NAEOS_API_JWT_SECRET", "environment-secret")
	t.Setenv("NAEOS_API_JWT_SECRET_FILE", "")

	cmd := newAPICommand()
	if got, err := resolveAPIJWTSecret(cmd); err != nil || got != "environment-secret" {
		t.Fatalf("expected environment secret, got %q", got)
	}

	if err := cmd.ParseFlags([]string{"--secret", "flag-secret"}); err != nil {
		t.Fatalf("parse secret flag: %v", err)
	}
	if got, err := resolveAPIJWTSecret(cmd); err != nil || got != "flag-secret" {
		t.Fatalf("expected explicit flag secret, got %q", got)
	}
}

func TestResolveAPIAuthModeFlagOverridesEnvironment(t *testing.T) {
	previous := apiAuthMode
	t.Cleanup(func() {
		apiAuthMode = previous
	})
	apiAuthMode = ""
	t.Setenv(apiAuthModeEnv, "supabase")
	cmd := newAPICommand()
	if got := resolveAPIAuthMode(cmd); got != "supabase" {
		t.Fatalf("auth mode = %q, want supabase", got)
	}
	if err := cmd.ParseFlags([]string{"--auth-mode", "hs256"}); err != nil {
		t.Fatalf("parse auth mode: %v", err)
	}
	if got := resolveAPIAuthMode(cmd); got != "hs256" {
		t.Fatalf("auth mode = %q, want hs256", got)
	}
}

func TestReadSecretEnvSupportsSecretFiles(t *testing.T) {
	secretFile := t.TempDir() + "/secret"
	if err := os.WriteFile(secretFile, []byte("file-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_SECRET", "")
	t.Setenv("TEST_SECRET_FILE", secretFile)

	got, err := readSecretEnv("TEST_SECRET", "TEST_SECRET_FILE")
	if err != nil {
		t.Fatalf("read secret file: %v", err)
	}
	if got != "file-secret" {
		t.Fatalf("secret file content = %q, want trailing newline removed", got)
	}

	t.Setenv("TEST_SECRET", "environment-secret")
	if _, err := readSecretEnv("TEST_SECRET", "TEST_SECRET_FILE"); err == nil {
		t.Fatal("expected conflicting secret sources to fail")
	}
}

func TestAPICommandRequiresSecretsWhenAuthEnabled(t *testing.T) {
	previousAuth, previousSecret := apiAuth, apiSecret
	t.Cleanup(func() {
		apiAuth, apiSecret = previousAuth, previousSecret
	})
	apiSecret = ""
	t.Setenv("NAEOS_API_JWT_SECRET", "")
	t.Setenv("NAEOS_ENCRYPTION_KEY", "")
	t.Setenv(apiAuthModeEnv, "hs256")

	cmd := newAPICommand()
	if err := cmd.ParseFlags([]string{"--auth"}); err != nil {
		t.Fatalf("parse auth flag: %v", err)
	}
	if err := cmd.RunE(cmd, nil); err == nil {
		t.Fatal("expected API auth to reject missing secrets")
	}
}

func TestAPICommandRequiresSupabaseConfiguration(t *testing.T) {
	previousAuth, previousSecret := apiAuth, apiSecret
	t.Cleanup(func() {
		apiAuth, apiSecret = previousAuth, previousSecret
	})
	apiSecret = ""
	t.Setenv(apiAuthModeEnv, "supabase")
	t.Setenv("SUPABASE_URL", "")
	t.Setenv("NAEOS_API_JWT_SECRET", "")

	cmd := newAPICommand()
	if err := cmd.ParseFlags([]string{"--auth"}); err != nil {
		t.Fatalf("parse auth flag: %v", err)
	}
	if err := cmd.RunE(cmd, nil); err == nil {
		t.Fatal("expected Supabase mode to reject missing project URL")
	}
}

func TestAPICommandRejectsHSSecretInSupabaseMode(t *testing.T) {
	previousAuth, previousSecret := apiAuth, apiSecret
	t.Cleanup(func() {
		apiAuth, apiSecret = previousAuth, previousSecret
	})
	apiSecret = ""
	t.Setenv(apiAuthModeEnv, "supabase")
	t.Setenv("SUPABASE_URL", "https://project.example")
	t.Setenv("NAEOS_API_JWT_SECRET", strings.Repeat("x", 32))

	cmd := newAPICommand()
	if err := cmd.ParseFlags([]string{"--auth"}); err != nil {
		t.Fatalf("parse auth flag: %v", err)
	}
	if err := cmd.RunE(cmd, nil); err == nil {
		t.Fatal("expected Supabase mode to reject an HS256 secret")
	}
}

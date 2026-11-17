// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestSupabaseAuthEnv(t *testing.T) {
	t.Setenv("SUPABASE_URL", "https://ref.supabase.co/")
	t.Setenv("SUPABASE_AUTH_AUDIENCE", "")
	t.Setenv("SUPABASE_JWKS_URL", "")

	issuer, audience, jwksURL, ok := supabaseAuthEnv()
	if !ok {
		t.Fatal("expected Supabase auth to be configured")
	}
	if issuer != "https://ref.supabase.co/auth/v1" {
		t.Errorf("issuer = %q", issuer)
	}
	if audience != "authenticated" {
		t.Errorf("audience = %q", audience)
	}
	if jwksURL != "https://ref.supabase.co/auth/v1/.well-known/jwks.json" {
		t.Errorf("jwksURL = %q", jwksURL)
	}
}

func TestSupabaseAuthEnvOverrides(t *testing.T) {
	t.Setenv("SUPABASE_URL", "https://ref.supabase.co")
	t.Setenv("SUPABASE_AUTH_AUDIENCE", "naeos-api")
	t.Setenv("SUPABASE_JWKS_URL", "https://keys.example.test/jwks.json")

	_, audience, jwksURL, ok := supabaseAuthEnv()
	if !ok {
		t.Fatal("expected Supabase auth to be configured")
	}
	if audience != "naeos-api" {
		t.Errorf("audience = %q", audience)
	}
	if jwksURL != "https://keys.example.test/jwks.json" {
		t.Errorf("jwksURL = %q", jwksURL)
	}
}

func TestSupabaseAuthEnvUnconfigured(t *testing.T) {
	t.Setenv("SUPABASE_URL", "  ")
	if _, _, _, ok := supabaseAuthEnv(); ok {
		t.Fatal("expected empty SUPABASE_URL to disable Supabase auth")
	}
}

func TestAllowLoopbackAuthEnabled(t *testing.T) {
	cases := map[string]bool{
		"":        false,
		"false":   false,
		"true":    true,
		"1":       true,
		"garbage": false,
	}
	for value, want := range cases {
		t.Setenv("NAEOS_SUPABASE_ALLOW_INSECURE_LOOPBACK", value)
		if got := allowLoopbackAuthEnabled(); got != want {
			t.Errorf("value %q = %v, want %v", value, got, want)
		}
	}
}

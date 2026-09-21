// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model/security"
	"github.com/NAEOS-foundation/naeos/internal/specification/resolver"
)

func TestBuilderExtractsSecurityContext(t *testing.T) {
	resolved := &resolver.ResolvedSpec{Context: map[string]any{
		"project": "security-test",
		"security": map[string]any{
			"tls": "1.3",
			"authentication": map[string]any{
				"method":   "oidc",
				"provider": "example",
			},
			"encryption": map[string]any{
				"in_transit": true,
				"algorithm":  "AES-256",
			},
			"attributes": map[string]any{"classification": "confidential"},
		},
	}}

	neir, err := DefaultBuilder{}.Build(resolved)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if neir.Security == nil {
		t.Fatal("expected security context to be preserved")
	}
	if neir.Security.Attributes["tls"] != "1.3" {
		t.Fatalf("expected legacy tls claim to be preserved, got %#v", neir.Security.Attributes)
	}
	if neir.Security.Attributes["classification"] != "confidential" {
		t.Fatalf("expected security attribute to be preserved, got %#v", neir.Security.Attributes)
	}
	if neir.Security.Authentication == nil || neir.Security.Authentication.Method != "oidc" {
		t.Fatalf("expected authentication to be extracted, got %#v", neir.Security.Authentication)
	}
	if neir.Security.Encryption == nil || !neir.Security.Encryption.InTransit {
		t.Fatalf("expected in_transit encryption to be extracted, got %#v", neir.Security.Encryption)
	}
	if neir.Security.Encryption.Algorithm != "AES-256" {
		t.Fatalf("expected encryption algorithm, got %q", neir.Security.Encryption.Algorithm)
	}

	var _ *security.Security = neir.Security
}

// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package normalizer

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

func TestNormalizerPreservesSecurityContext(t *testing.T) {
	doc, err := parser.NewParser(".").Parse(`project: security-trace
modules:
  - name: core
    path: ./internal/core
security:
  tls: "1.3"
  authentication:
    method: oidc
    provider: example
`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	normalized, err := NewNormalizer().Normalize(doc)
	if err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}

	securityContext, ok := normalized.Values["security"].(map[string]any)
	if !ok {
		t.Fatalf("expected normalized security context, got %T", normalized.Values["security"])
	}
	if securityContext["tls"] != "1.3" {
		t.Fatalf("expected security.tls=1.3, got %#v", securityContext["tls"])
	}
	auth, ok := securityContext["authentication"].(map[string]any)
	if !ok {
		t.Fatalf("expected authentication map, got %T", securityContext["authentication"])
	}
	if auth["method"] != "oidc" {
		t.Fatalf("expected authentication.method=oidc, got %#v", auth["method"])
	}
}

// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/security"
)

func TestContextFromNEIRIncludesGovernedDomains(t *testing.T) {
	neir := &model.NEIR{
		Security: &security.Security{},
	}

	ctx, err := ContextFromNEIR(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Version != PolicyContextVersion {
		t.Fatalf("expected context version %q, got %q", PolicyContextVersion, ctx.Version)
	}
	if _, ok := ctx.Values["security"]; !ok {
		t.Fatal("expected security in policy context")
	}
}

func TestContextFromNEIRPreservesTopLevelFields(t *testing.T) {
	neir := &model.NEIR{ActiveProfile: "enterprise"}
	ctx, err := ContextFromNEIR(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := ctx.Values["active_profile"]; !ok || got != "enterprise" {
		t.Fatalf("expected active_profile to be preserved, got %#v", got)
	}
}

func TestContextFromNEIRRejectsNil(t *testing.T) {
	if _, err := ContextFromNEIR(nil); err == nil {
		t.Fatal("expected nil NEIR to be rejected")
	}
}

func TestPolicyContextRejectsUnsupportedVersion(t *testing.T) {
	ctx := PolicyContext{Version: "v999", Values: map[string]any{"project": "x"}}
	if err := ctx.Validate(); err == nil {
		t.Fatal("expected unsupported context version to fail closed")
	}
}

func TestPolicyContextRejectsNilValues(t *testing.T) {
	ctx := PolicyContext{Version: PolicyContextVersion}
	if err := ctx.Validate(); err == nil {
		t.Fatal("expected nil context values to fail validation")
	}
}

func TestContextFromNEIRUsesExplicitAllowList(t *testing.T) {
	neir := &model.NEIR{Project: nil, ActiveProfile: "enterprise"}
	ctx, err := ContextFromNEIR(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := ctx.Values["active_profile"]; !ok {
		t.Fatal("expected active_profile in policy context")
	}
	// This test documents that only fields in PolicyContextFields are exposed.
	if _, ok := ctx.Values["future_uncontracted_field"]; ok {
		t.Fatal("unexpected uncontracted field exposed")
	}
}

func TestPolicyContextDigestDeterministic(t *testing.T) {
	ctx := PolicyContext{
		Version: PolicyContextVersion,
		Values: map[string]any{
			"project":  "naeos",
			"security": map[string]any{"tls": "1.3"},
		},
	}
	d1, err := ctx.Digest()
	if err != nil {
		t.Fatalf("digest failed: %v", err)
	}
	d2, err := ctx.Digest()
	if err != nil {
		t.Fatalf("second digest failed: %v", err)
	}
	if d1 == "" || d1 != d2 {
		t.Fatalf("expected deterministic digest, got %q and %q", d1, d2)
	}
}

func TestPolicyContextFieldNamesReturnsCopy(t *testing.T) {
	fields := PolicyContextFieldNames()
	if len(fields) == 0 {
		t.Fatal("expected policy context fields")
	}
	original := fields[0]
	fields[0] = "mutated"
	fieldsAgain := PolicyContextFieldNames()
	if fieldsAgain[0] != original {
		t.Fatalf("policy context field contract was mutated through accessor: got %q, want %q", fieldsAgain[0], original)
	}
}

func TestPolicyContextDigestBindsContractFields(t *testing.T) {
	values := map[string]any{"project": "naeos"}

	payload1, err := policyContextDigestPayload(PolicyContextVersion, []string{"project", "security"}, values)
	if err != nil {
		t.Fatalf("first digest payload failed: %v", err)
	}
	payload2, err := policyContextDigestPayload(PolicyContextVersion, []string{"project", "security", "testing"}, values)
	if err != nil {
		t.Fatalf("second digest payload failed: %v", err)
	}

	sum1 := sha256.Sum256(payload1)
	sum2 := sha256.Sum256(payload2)
	if sum1 == sum2 {
		t.Fatal("expected policy context digest to change when contract fields change")
	}

	ctx := PolicyContext{Version: PolicyContextVersion, Values: values}
	digest, err := ctx.Digest()
	if err != nil {
		t.Fatalf("digest failed: %v", err)
	}
	if digest == "" {
		t.Fatal("expected non-empty digest")
	}
}

func TestPolicyContextDigestChangesWithVersion(t *testing.T) {
	ctx := PolicyContext{Version: PolicyContextVersion, Values: map[string]any{"project": "naeos"}}
	d1, err := ctx.Digest()
	if err != nil {
		t.Fatalf("digest failed: %v", err)
	}
	ctx.Version = "v2"
	if _, err := ctx.Digest(); err == nil {
		t.Fatal("expected unsupported version to fail digest")
	}
	_ = d1
}

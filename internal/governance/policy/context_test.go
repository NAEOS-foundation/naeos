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

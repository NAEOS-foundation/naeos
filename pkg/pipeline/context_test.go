// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"context"
	"errors"
	"testing"
)

func TestRunContextStopsAtStageBoundaryWhenCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p, err := New(Config{
		DryRun: true,
		Hooks: &Hooks{AfterParse: []HookFunc{func(*HookContext) error {
			cancel()
			return nil
		}}},
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := p.RunContext(ctx, "project: context-check")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled after validation stage, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected no result after cancellation, got %#v", result)
	}
}

func TestValidateContextPopulatesHashes(t *testing.T) {
	p, err := New(Config{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	result, err := p.ValidateContext(context.Background(), "project: validate-hash")
	if err != nil {
		t.Fatal(err)
	}
	if result.SpecificationHash == "" || result.NEIRHash == "" {
		t.Fatalf("expected validation hashes, got specification=%q neir=%q", result.SpecificationHash, result.NEIRHash)
	}
}

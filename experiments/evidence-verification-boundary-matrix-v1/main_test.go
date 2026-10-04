// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestEvidenceVerificationBoundaryMatrix(t *testing.T) {
	results, err := run()
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]Outcome{
		"internal-reversible-runtime-record":   Confirmed,
		"external-reversible-runtime-only":     Unknown,
		"external-reversible-provider-receipt": Confirmed,
		"internal-irreversible-estimated":      Unknown,
		"external-irreversible-authoritative":  Confirmed,
		"external-stale":                       Reverify,
		"external-contradictory":               Unknown,
		"missing-policy":                       Escalate,
	}
	for _, r := range results {
		if want := expected[r.CaseName]; r.Outcome != want {
			t.Fatalf("%s: got %s, want %s (%s)", r.CaseName, r.Outcome, want, r.Reason)
		}
	}
}

func TestRuntimeRecordIsNotExternalProof(t *testing.T) {
	results, err := run()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.CaseName == "external-reversible-runtime-only" && r.Outcome != Unknown {
			t.Fatalf("runtime-only external evidence must remain UNKNOWN: %+v", r)
		}
	}
}

func TestEstimatedEvidenceCannotConfirmIrreversibleEffect(t *testing.T) {
	results, err := run()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.CaseName == "internal-irreversible-estimated" && r.Outcome != Unknown {
			t.Fatalf("estimated evidence must not confirm irreversible effect: %+v", r)
		}
	}
}

func TestMissingPolicyFailsClosed(t *testing.T) {
	results, err := run()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.CaseName == "missing-policy" && r.Outcome != Escalate {
			t.Fatalf("missing verification policy must escalate/fail closed: %+v", r)
		}
	}
}

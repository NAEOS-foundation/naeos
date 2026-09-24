// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package changerisk

import "testing"

func TestClassifyDocsIsLowRisk(t *testing.T) {
	r := Classify(Request{Paths: []string{"docs/example.md"}, EvidenceAvailable: true})
	if r.Criticality != Low || r.Risk != Risk("low") || r.Decision != Allow {
		t.Fatalf("unexpected result: %+v", r)
	}
}

func TestClassifyCriticalDomainRequiresReview(t *testing.T) {
	r := Classify(Request{
		Paths:             []string{"internal/governance/policy.go"},
		Additions:         20,
		EvidenceAvailable: true,
	})
	if r.Criticality != High || r.Risk != Risk("high") || r.Decision != RequireReview {
		t.Fatalf("unexpected result: %+v", r)
	}
}

func TestClassifyLargeChangeBecomesCritical(t *testing.T) {
	r := Classify(Request{
		Paths:             []string{"internal/foo/bar.go"},
		Additions:         1000,
		EvidenceAvailable: true,
	})
	if r.Criticality != Critical || r.Risk != Risk("critical") || r.Decision != RequireReview {
		t.Fatalf("unexpected result: %+v", r)
	}
}

func TestClassifyUnknownFailsClosed(t *testing.T) {
	r := Classify(Request{EvidenceAvailable: true})
	if r.Decision != Deny || r.Risk != Risk("unknown") {
		t.Fatalf("expected fail closed: %+v", r)
	}
}

func TestClassifyMissingEvidenceFailsClosed(t *testing.T) {
	r := Classify(Request{Paths: []string{"internal/foo/bar.go"}, EvidenceAvailable: false})
	if r.Decision != Deny || r.Risk != Risk("unknown") {
		t.Fatalf("expected deny without evidence: %+v", r)
	}
}

func TestDomainsForPaths(t *testing.T) {
	got := DomainsForPaths([]string{
		"internal/governance/policy.go",
		"internal/audit/event.go",
		"runtime/executor.go",
	})
	if len(got) != 3 || got[0] != AuditEvidence || got[1] != Policy || got[2] != Runtime {
		t.Fatalf("unexpected domains: %+v", got)
	}
}

func TestSurfaceMixed(t *testing.T) {
	got := SurfaceForPaths([]string{"docs/example.md", "internal/foo/bar.go"})
	if got != Mixed {
		t.Fatalf("expected mixed surface, got %s", got)
	}
}

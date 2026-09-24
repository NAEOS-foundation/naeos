// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package dependencyrisk

import "testing"

func TestClassifyPatchGeneral(t *testing.T) {
 r := Classify(Request{Ecosystem:"go", Name:"example", VersionChange:Patch, Paths:[]string{"internal/foo/bar.go"}, EvidenceAvailable:true})
 if r.Criticality != Low || r.Risk != Risk("low") || r.Decision != Allow { t.Fatalf("unexpected result: %+v", r) }
 if len(r.RequiredGates) != 1 || r.RequiredGates[0] != "ci" { t.Fatalf("unexpected gates: %+v", r.RequiredGates) }
}
func TestClassifyCriticalDomainRequiresReview(t *testing.T) {
 r := Classify(Request{Ecosystem:"github-actions", Name:"actions/checkout", VersionChange:Patch, Paths:[]string{".github/workflows/ci.yml"}, EvidenceAvailable:true})
 if r.Criticality != High || r.Risk != Risk("high") || r.Decision != RequireReview { t.Fatalf("unexpected result: %+v", r) }
}
func TestClassifyUnknownFailsClosed(t *testing.T) {
 r := Classify(Request{Ecosystem:"npm", Name:"unknown", VersionChange:Unknown, Paths:nil, EvidenceAvailable:false})
 if r.Decision != Deny || r.Risk != Risk("unknown") { t.Fatalf("expected fail closed: %+v", r) }
}
func TestClassifyUnmappedImpactFailsClosed(t *testing.T) {
 r := Classify(Request{Ecosystem:"go", Name:"example", VersionChange:Patch, Paths:nil, EvidenceAvailable:true})
 if r.Decision != Deny || r.Criticality != Criticality("unknown") { t.Fatalf("expected unknown impact to fail closed: %+v", r) }
}
func TestClassifyMissingEvidenceFailsClosed(t *testing.T) {
 r := Classify(Request{Ecosystem:"go", Name:"example", VersionChange:Patch, Paths:[]string{"internal/foo/bar.go"}, EvidenceAvailable:false})
 if r.Decision != Deny { t.Fatalf("expected deny without evidence: %+v", r) }
}
func TestDomainsForPaths(t *testing.T) {
 got := DomainsForPaths([]string{"internal/governance/policy.go", "internal/audit/event.go"})
 if len(got) != 2 || got[0] != AuditEvidence || got[1] != Policy { t.Fatalf("unexpected domains: %+v", got) }
}

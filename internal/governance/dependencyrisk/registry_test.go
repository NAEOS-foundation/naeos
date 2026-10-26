// Copyright 2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package dependencyrisk

import (
	"strings"
	"testing"
	"time"
)

func validTestRegistry() ReviewRegistry {
	return ReviewRegistry{
		SchemaVersion: RegistrySchemaVersion,
		PolicyVersion: "1.0.0",
		Records: []ReviewRecord{{
			Ecosystem:          "go",
			Name:               "github.com/example/module",
			Version:            "v1.2.3",
			UpstreamURL:        "https://github.com/example/module",
			SPDXLicense:        "Apache-2.0",
			LicenseEvidenceURL: "https://github.com/example/module/blob/v1.2.3/LICENSE",
			Checksum:           "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ProvenanceURL:      "https://github.com/example/module/releases/tag/v1.2.3",
			Purpose:            "Canonical JSON encoding for request digests",
			Domains:            []Domain{Runtime},
			Criticality:        High,
			Reviewer:           "security-reviewer",
			ReviewReference:    "https://github.com/NAEOS-foundation/naeos/pull/999",
			ReviewedAt:         "2026-10-20T12:00:00Z",
			ReviewExpiresAt:    "2026-11-20T12:00:00Z",
		}},
	}
}

func TestValidateReviewRegistryAcceptsCompleteRecord(t *testing.T) {
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	if err := ValidateReviewRegistry(validTestRegistry(), "1.0.0", now); err != nil {
		t.Fatalf("valid registry rejected: %v", err)
	}
}

func TestValidateReviewRegistryRejectsPolicyOrSchemaMismatch(t *testing.T) {
	reg := validTestRegistry()
	reg.SchemaVersion = "9.9.9"
	if err := ValidateReviewRegistry(reg, "1.0.0", time.Now().UTC()); err == nil {
		t.Fatal("expected schema mismatch to fail")
	}
	reg = validTestRegistry()
	if err := ValidateReviewRegistry(reg, "2.0.0", time.Now().UTC()); err == nil {
		t.Fatal("expected policy mismatch to fail")
	}
}

func TestValidateReviewRegistryRejectsMissingReviewAndInvalidEvidence(t *testing.T) {
	tests := []struct {
		name string
		edit func(*ReviewRecord)
		want string
	}{
		{"missing reviewer", func(r *ReviewRecord) { r.Reviewer = "" }, "missing reviewer"},
		{"missing review reference", func(r *ReviewRecord) { r.ReviewReference = "" }, "missing review reference"},
		{"invalid license evidence URL", func(r *ReviewRecord) { r.LicenseEvidenceURL = "http://example.com/license" }, "license_evidence_url"},
		{"invalid checksum", func(r *ReviewRecord) { r.Checksum = "sha256:bad" }, "checksum"},
		{"invalid criticality", func(r *ReviewRecord) { r.Criticality = Criticality("unknown") }, "criticality"},
		{"invalid domain", func(r *ReviewRecord) { r.Domains = []Domain{UnknownDomain} }, "unrecognized domain"},
		{"future review", func(r *ReviewRecord) { r.ReviewedAt = "2027-01-01T00:00:00Z" }, "future"},
		{"stale review", func(r *ReviewRecord) { r.ReviewExpiresAt = "2026-10-21T12:00:00Z" }, "stale"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := validTestRegistry()
			tt.edit(&reg.Records[0])
			now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
			err := ValidateReviewRegistry(reg, "1.0.0", now)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestValidateReviewRegistryRejectsDuplicateExactVersion(t *testing.T) {
	reg := validTestRegistry()
	reg.Records = append(reg.Records, reg.Records[0])
	if err := ValidateReviewRegistry(reg, "1.0.0", time.Now().UTC()); err == nil {
		t.Fatal("expected duplicate record to fail")
	}
}

func TestFindExactReviewRequiresExactVersion(t *testing.T) {
	reg := validTestRegistry()
	if _, ok := reg.FindExactReview("go", "github.com/example/module", "v1.2.3"); !ok {
		t.Fatal("expected exact record to match")
	}
	if _, ok := reg.FindExactReview("go", "github.com/example/module", "v1.2.4"); ok {
		t.Fatal("version mismatch must not match")
	}
	if _, ok := reg.FindExactReview("npm", "github.com/example/module", "v1.2.3"); ok {
		t.Fatal("ecosystem mismatch must not match")
	}
}

func TestSortedRecordKeysIsStable(t *testing.T) {
	reg := validTestRegistry()
	reg.Records = append(reg.Records, ReviewRecord{Ecosystem: "npm", Name: "z", Version: "1.0.0"})
	got := reg.SortedRecordKeys()
	if len(got) != 2 || got[0] != "go:github.com/example/module@v1.2.3" || got[1] != "npm:z@1.0.0" {
		t.Fatalf("unexpected sorted keys: %#v", got)
	}
}

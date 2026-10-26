// Copyright 2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/governance/dependencyrisk"
)

func trustedRegistryFixture(t *testing.T) []byte {
	t.Helper()
	registry := dependencyrisk.ReviewRegistry{
		SchemaVersion: dependencyrisk.RegistrySchemaVersion,
		PolicyVersion: "1.0.0",
		Records: []dependencyrisk.ReviewRecord{{
			Ecosystem:          "go",
			Name:               "github.com/example/module",
			Version:            "v1.2.3",
			UpstreamURL:        "https://github.com/example/module",
			SPDXLicense:        "Apache-2.0",
			LicenseEvidenceURL: "https://github.com/example/module/blob/v1.2.3/LICENSE",
			Checksum:           "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ProvenanceURL:      "https://github.com/example/module/releases/tag/v1.2.3",
			Purpose:            "Dependency purpose for a regression test",
			Domains:            []dependencyrisk.Domain{dependencyrisk.Runtime},
			Criticality:        dependencyrisk.High,
			Reviewer:           "security-reviewer",
			ReviewReference:    "https://github.com/NAEOS-foundation/naeos/pull/999",
			ReviewedAt:         "2026-10-20T12:00:00Z",
			ReviewExpiresAt:    "2026-11-20T12:00:00Z",
		}},
	}
	raw, err := json.Marshal(registry)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func trustedReviewRequest(version string) requestFile {
	dep := dependencyIdentity{Ecosystem: "go", Name: "github.com/example/module", Version: version}
	return requestFile{Ecosystem: dep.Ecosystem, Name: dep.Name, Version: version, Dependencies: []dependencyIdentity{dep}}
}

func TestValidateTrustedReviewRequiresExactVersion(t *testing.T) {
	now := time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)
	if _, err := validateTrustedReview(trustedRegistryFixture(t), "base-sha", "1.0.0", trustedReviewRequest("v1.2.4"), now, "pull-request-author"); err == nil || !strings.Contains(err.Error(), "no trusted review for exact dependency") {
		t.Fatalf("expected exact-version mismatch to fail closed, got %v", err)
	}
	if _, err := validateTrustedReview(trustedRegistryFixture(t), "base-sha", "1.0.0", trustedReviewRequest(""), now, "pull-request-author"); err == nil || !strings.Contains(err.Error(), "has no exact version") {
		t.Fatalf("expected unknown version to fail closed, got %v", err)
	}
}

func TestValidateTrustedReviewRejectsExpiredRegistryAndSelfApproval(t *testing.T) {
	now := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	if _, err := validateTrustedReview(trustedRegistryFixture(t), "base-sha", "1.0.0", trustedReviewRequest("v1.2.3"), now, "pull-request-author"); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("expected expired registry to fail closed, got %v", err)
	}
	now = time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)
	if _, err := validateTrustedReview(trustedRegistryFixture(t), "base-sha", "1.0.0", trustedReviewRequest("v1.2.3"), now, "SECURITY-REVIEWER"); err == nil || !strings.Contains(err.Error(), "cannot self-approve") {
		t.Fatalf("expected self-approval to fail closed, got %v", err)
	}
}

func TestValidateTrustedReviewRejectsMalformedAndMismatchedRegistry(t *testing.T) {
	now := time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)
	if _, err := validateTrustedReview([]byte("{"), "base-sha", "1.0.0", trustedReviewRequest("v1.2.3"), now, ""); err == nil {
		t.Fatal("expected malformed registry to fail")
	}
	if _, err := validateTrustedReview(trustedRegistryFixture(t), "base-sha", "2.0.0", trustedReviewRequest("v1.2.3"), now, ""); err == nil || !strings.Contains(err.Error(), "policy version mismatch") {
		t.Fatalf("expected policy mismatch to fail closed, got %v", err)
	}
	if _, err := validateTrustedReview(append(trustedRegistryFixture(t), []byte(" {}")...), "base-sha", "1.0.0", trustedReviewRequest("v1.2.3"), now, ""); err == nil || !strings.Contains(err.Error(), "trailing or malformed") {
		t.Fatalf("expected trailing data to fail closed, got %v", err)
	}
}

func TestExactPackageVersionRejectsRanges(t *testing.T) {
	for _, value := range []string{"^1.2.3", "~1.2.3", ">=1.2.3", "1.2.x", "*"} {
		if got := exactPackageVersion(value); got != "" {
			t.Errorf("exactPackageVersion(%q) = %q, want empty", value, got)
		}
	}
	if got := exactPackageVersion("v1.2.3"); got != "1.2.3" {
		t.Fatalf("exactPackageVersion(v1.2.3) = %q", got)
	}
}

func TestValidateTrustedReviewRejectsMixedRiskSetWithUnreviewedDependency(t *testing.T) {
	now := time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)
	req := trustedReviewRequest("v1.2.3")
	req.Dependencies = append(req.Dependencies, dependencyIdentity{
		Ecosystem: "go", Name: "github.com/example/unreviewed", Version: "v9.9.9",
	})
	if _, err := validateTrustedReview(trustedRegistryFixture(t), "base-sha", "1.0.0", req, now, "pull-request-author"); err == nil || !strings.Contains(err.Error(), "no trusted review for exact dependency go:github.com/example/unreviewed@v9.9.9") {
		t.Fatalf("expected an unreviewed dependency in a mixed change to fail closed, got %v", err)
	}
}

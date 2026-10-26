// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"
)

func validRegistryJSON() string {
	return `{"schema_version":"1.0.0","policy_version":"1.0.0","records":[{"ecosystem":"go","name":"example.org/lib","version":"v1.2.3","upstream_source":"https://example.org/lib","spdx_license":"Apache-2.0","checksum":"h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","provenance":"https://example.org/provenance/1.2.3","purpose":"canonicalization","domains":["runtime"],"criticality":"high","reviewer":"reviewer-id","review_reference":"https://github.com/NAEOS-foundation/naeos/pull/1","policy_version":"1.0.0","schema_version":"1.0.0"}]}`
}

func TestParseDependencyReviewRegistryAcceptsCompleteRecord(t *testing.T) {
	registry, err := parseDependencyReviewRegistry([]byte(validRegistryJSON()))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := findReviewedDependency(registry, "go", "example.org/lib", "v1.2.3"); !ok {
		t.Fatal("expected exact reviewed version to match")
	}
	if _, ok := findReviewedDependency(registry, "go", "example.org/lib", "v1.2.4"); ok {
		t.Fatal("version mismatch must not match")
	}
}

func TestParseDependencyReviewRegistryRejectsInvalidEvidence(t *testing.T) {
	cases := []struct {
		name, input string
	}{
		{"missing review reference", strings.Replace(validRegistryJSON(), `"review_reference":"https://github.com/NAEOS-foundation/naeos/pull/1"`, `"review_reference":""`, 1)},
		{"invalid checksum", strings.Replace(validRegistryJSON(), "h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", "bad-checksum", 1)},
		{"invalid license", strings.Replace(validRegistryJSON(), `"spdx_license":"Apache-2.0"`, `"spdx_license":"Apache 2.0"`, 1)},
		{"unknown field", strings.Replace(validRegistryJSON(), `"schema_version":"1.0.0"}]}`, `"schema_version":"1.0.0","approved":true}]}`, 1)},
		{"malformed json", "{not-json"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseDependencyReviewRegistry([]byte(tc.input)); err == nil {
				t.Fatal("expected registry validation failure")
			}
		})
	}
}

func TestParseDependencyReviewRegistryRejectsDuplicateRecords(t *testing.T) {
	record := strings.TrimSuffix(strings.TrimPrefix(validRegistryJSON(), `{"schema_version":"1.0.0","policy_version":"1.0.0","records":[`), `]}`)
	input := `{"schema_version":"1.0.0","policy_version":"1.0.0","records":[` + record + `,` + record + `]}`
	if _, err := parseDependencyReviewRegistry([]byte(input)); err == nil {
		t.Fatal("expected duplicate registry records to be rejected")
	}
}

func TestChecksumMustMatchTrustedGoSumEntry(t *testing.T) {
	registry, err := parseDependencyReviewRegistry([]byte(validRegistryJSON()))
	if err != nil {
		t.Fatal(err)
	}
	record := registry.Records[0]
	line := record.Name + " " + record.Version + " " + record.Checksum + "\n"
	if !checksumPresentInGoSum([]byte(line), record) {
		t.Fatal("expected exact module/version/checksum entry to match")
	}
	mismatch := record.Name + " " + record.Version + " h1:BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=\n"
	if checksumPresentInGoSum([]byte(mismatch), record) {
		t.Fatal("mismatched checksum must not match")
	}
}

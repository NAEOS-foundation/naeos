// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestChecksumPresentInNPMLock(t *testing.T) {
	checksum := "sha512-" + strings.Repeat("A", 86) + "=="
	record := dependencyReviewRecord{Ecosystem: "npm", Name: "wrangler", Version: "4.156.0", Checksum: checksum}
	lock := `{"lockfileVersion":3,"packages":{"":{"name":"worker"},"node_modules/wrangler":{"name":"wrangler","version":"4.156.0","integrity":"` + checksum + `"}}}`
	path := filepath.Join(t.TempDir(), "package-lock.json")
	if err := os.WriteFile(path, []byte(lock), 0o600); err != nil {
		t.Fatal(err)
	}
	if !checksumPresentInNPMLock(path, record) {
		t.Fatal("expected matching package-lock integrity to verify")
	}

	record.Version = "4.155.0"
	if checksumPresentInNPMLock(path, record) {
		t.Fatal("unexpected verification for mismatched version")
	}
	record.Version = "4.156.0"
	record.Checksum = "sha512-" + strings.Repeat("B", 86) + "=="
	if checksumPresentInNPMLock(path, record) {
		t.Fatal("unexpected verification for mismatched integrity")
	}
}

func TestChecksumPresentInNPMLockFailsClosedOnUnsupportedVersions(t *testing.T) {
	checksum := "sha512-" + strings.Repeat("A", 86) + "=="
	record := dependencyReviewRecord{Ecosystem: "npm", Name: "wrangler", Version: "4.156.0", Checksum: checksum}
	path := filepath.Join(t.TempDir(), "package-lock.json")

	for _, version := range []int{0, 1, 4, -1} {
		t.Run("lockfileVersion_"+strconv.Itoa(version), func(t *testing.T) {
			lock := `{"lockfileVersion":` + strconv.Itoa(version) + `, "packages":{"node_modules/wrangler":{"name":"wrangler","version":"4.156.0","integrity":"` + checksum + `"}}}`
			if err := os.WriteFile(path, []byte(lock), 0o600); err != nil {
				t.Fatal(err)
			}
			if checksumPresentInNPMLock(path, record) {
				t.Fatalf("lockfile version %d must fail closed", version)
			}
		})
	}
}

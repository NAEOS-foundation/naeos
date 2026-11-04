// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
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

func TestChecksumPresentInNPMLockFailsClosedOnV1(t *testing.T) {
	record := dependencyReviewRecord{Ecosystem: "npm", Name: "wrangler", Version: "4.156.0", Checksum: "sha512-" + strings.Repeat("A", 86) + "=="}
	path := filepath.Join(t.TempDir(), "package-lock.json")
	if err := os.WriteFile(path, []byte(`{"lockfileVersion":1,"dependencies":{"wrangler":{"version":"4.156.0","integrity":"`+record.Checksum+`"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if checksumPresentInNPMLock(path, record) {
		t.Fatal("lockfile v1 must fail closed")
	}
}

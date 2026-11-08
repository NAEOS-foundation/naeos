// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestNPMVersionFromLockUsesExactInstalledVersion(t *testing.T) {
	lock := []byte(`{"lockfileVersion":3,"packages":{"node_modules/wrangler":{"name":"wrangler","version":"4.156.0"},"node_modules/@cloudflare/workers-types":{"name":"@cloudflare/workers-types","version":"5.20260911.1"}}}`)

	for name, want := range map[string]string{
		"wrangler":                  "4.156.0",
		"@cloudflare/workers-types": "5.20260911.1",
	} {
		got, ok := npmPackageVersionInLock(lock, name)
		if !ok || got != want {
			t.Errorf("npmPackageVersionInLock(%q) = %q, %v; want %q, true", name, got, ok, want)
		}
	}
}

func TestNPMVersionFromLockFailsClosed(t *testing.T) {
	for _, lock := range [][]byte{
		[]byte(`{"lockfileVersion":1,"packages":{"node_modules/wrangler":{"version":"4.156.0"}}}`),
		[]byte(`{"lockfileVersion":3,"packages":{"node_modules/other":{"version":"4.156.0"}}}`),
		[]byte(`{"lockfileVersion":3,"packages":{"node_modules/wrangler":{"name":"other","version":"4.156.0"}}}`),
	} {
		if version, ok := npmPackageVersionInLock(lock, "wrangler"); ok {
			t.Errorf("npmPackageVersionInLock() = %q, true; want fail-closed result", version)
		}
	}
}

func TestDiffNPMLockIncludesTransitiveChangesAndRemovals(t *testing.T) {
	oldLock := []byte(`{"lockfileVersion":3,"packages":{"node_modules/wrangler":{"name":"wrangler","version":"4.151.1","integrity":"old"},"node_modules/wrangler/node_modules/miniflare":{"name":"miniflare","version":"5.20261012.1-beta","integrity":"old-miniflare"},"node_modules/removed":{"name":"removed","version":"1.0.0","integrity":"old-removed"}}}`)
	newLock := []byte(`{"lockfileVersion":3,"packages":{"node_modules/wrangler":{"name":"wrangler","version":"4.156.0","integrity":"new"},"node_modules/wrangler/node_modules/miniflare":{"name":"miniflare","version":"5.20261023.3-beta","integrity":"new-miniflare"},"node_modules/added":{"name":"added","version":"2.0.0","integrity":"new-added"}}}`)

	changes, ok := diffNPMLock(oldLock, newLock)
	if !ok {
		t.Fatal("diffNPMLock() rejected supported lockfiles")
	}
	want := []npmLockChange{
		{Name: "added", NewVersion: "2.0.0", NewIntegrity: "new-added"},
		{Name: "miniflare", OldVersion: "5.20261012.1-beta", NewVersion: "5.20261023.3-beta", OldIntegrity: "old-miniflare", NewIntegrity: "new-miniflare"},
		{Name: "removed", OldVersion: "1.0.0", OldIntegrity: "old-removed"},
		{Name: "wrangler", OldVersion: "4.151.1", NewVersion: "4.156.0", OldIntegrity: "old", NewIntegrity: "new"},
	}
	if !reflect.DeepEqual(changes, want) {
		t.Fatalf("diffNPMLock() = %#v, want %#v", changes, want)
	}
}

func TestDiffNPMLockFailsClosedOnUnsupportedLockfile(t *testing.T) {
	if changes, ok := diffNPMLock(nil, []byte(`{"lockfileVersion":1,"packages":{}}`)); ok || changes != nil {
		t.Fatalf("diffNPMLock() = %#v, %v; want nil, false", changes, ok)
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

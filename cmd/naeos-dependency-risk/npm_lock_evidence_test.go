// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNPMLockPathsAndVerifiedEvidence(t *testing.T) {
	tests := []struct {
		name       string
		changed    map[string]bool
		lockfiles  []string
		wantPaths  []string
		wantVerify bool
	}{
		{
			name:       "no changed lockfile",
			changed:    map[string]bool{"package.json": true},
			wantPaths:  nil,
			wantVerify: false,
		},
		{
			name:       "nested workspace lockfile",
			changed:    map[string]bool{"packages/app/package-lock.json": true},
			lockfiles:  []string{"packages/app/package-lock.json"},
			wantPaths:  []string{"packages/app/package-lock.json"},
			wantVerify: true,
		},
		{
			name:       "multiple lockfiles",
			changed:    map[string]bool{"package-lock.json": true, "packages/api/package-lock.json": true},
			lockfiles:  []string{"package-lock.json", "packages/api/package-lock.json"},
			wantPaths:  []string{"package-lock.json", "packages/api/package-lock.json"},
			wantVerify: true,
		},
		{
			name:       "npm shrinkwrap",
			changed:    map[string]bool{"npm-shrinkwrap.json": true},
			lockfiles:  []string{"npm-shrinkwrap.json"},
			wantPaths:  []string{"npm-shrinkwrap.json"},
			wantVerify: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := setupNPMGitFixture(t, tt.changed, tt.lockfiles)
			got, err := npmLockPaths(base)
			if err != nil {
				t.Fatalf("npmLockPaths() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.wantPaths) {
				t.Fatalf("npmLockPaths() = %#v, want %#v", got, tt.wantPaths)
			}

			checksum := "sha512-" + "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" + "=="
			record := dependencyReviewRecord{Ecosystem: "npm", Name: "wrangler", Version: "4.156.0", Checksum: checksum}
			if got := verifiedNPMDependencyLockEvidence(base, record); got != tt.wantVerify {
				t.Fatalf("verifiedNPMDependencyLockEvidence() = %v, want %v", got, tt.wantVerify)
			}
		})
	}
}

func TestNPMLockPathsFailsClosedForMissingOrInvalidBase(t *testing.T) {
	setupNPMGitFixture(t, map[string]bool{"package-lock.json": true}, []string{"package-lock.json"})

	for _, base := range []string{"", "not-a-commit-sha"} {
		t.Run("base_"+base, func(t *testing.T) {
			if _, err := npmLockPaths(base); err == nil {
				t.Fatalf("npmLockPaths(%q) expected error", base)
			}
			checksum := "sha512-" + "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" + "=="
			record := dependencyReviewRecord{Ecosystem: "npm", Name: "wrangler", Version: "4.156.0", Checksum: checksum}
			if verifiedNPMDependencyLockEvidence(base, record) {
				t.Fatalf("invalid base %q must fail closed", base)
			}
		})
	}
}

// setupNPMGitFixture creates a tiny two-commit repository to exercise git diff
// based lockfile discovery. It changes the process working directory only for
// this test and restores it during cleanup.
func setupNPMGitFixture(t *testing.T, changed map[string]bool, lockfiles []string) string {
	t.Helper()
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	runGit := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	runGit("init")
	runGit("config", "user.name", "NAEOS Test")
	runGit("config", "user.email", "naeos-test@example.invalid")

	if err := os.WriteFile("package.json", []byte("{\"name\":\"fixture\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit("add", "package.json")
	runGit("commit", "-m", "base")
	base := runGit("rev-parse", "HEAD")
	base = filepath.Clean(base)
	base = stringTrimSpace(base)

	for path := range changed {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		content := []byte("{}\n")
		if path == "package.json" {
			content = []byte("{\"name\":\"fixture\",\"dependencies\":{\"wrangler\":\"4.156.0\"}}\n")
		}
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range lockfiles {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		checksum := "sha512-" + "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" + "=="
		lock := map[string]any{
			"lockfileVersion": 3,
			"packages": map[string]any{
				"node_modules/wrangler": map[string]string{"name": "wrangler", "version": "4.156.0", "integrity": checksum},
			},
		}
		data, err := json.Marshal(lock)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runGit("add", ".")
	runGit("commit", "-m", "change dependency lockfiles")
	return base
}

func stringTrimSpace(value string) string {
	return strings.TrimSpace(value)
}

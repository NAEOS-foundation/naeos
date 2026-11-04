// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// checksumPresentInNPMLock verifies the review checksum against package-lock's
// SRI integrity for the exact package name and version. Only lockfile v2/v3
// package entries are supported; unsupported layouts fail closed.
func checksumPresentInNPMLock(lockPath string, record dependencyReviewRecord) bool {
	if !strings.EqualFold(record.Ecosystem, "npm") || !strings.HasPrefix(record.Checksum, "sha512-") {
		return false
	}
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return false
	}
	var lock struct {
		LockfileVersion int `json:"lockfileVersion"`
		Packages        map[string]struct {
			Name      string `json:"name"`
			Version   string `json:"version"`
			Integrity string `json:"integrity"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(data, &lock); err != nil || lock.LockfileVersion < 2 || lock.Packages == nil {
		return false
	}
	moduleSuffix := "node_modules/" + record.Name
	for packagePath, entry := range lock.Packages {
		if packagePath != moduleSuffix && !strings.HasSuffix(packagePath, "/"+moduleSuffix) {
			continue
		}
		if entry.Version != record.Version || (entry.Name != "" && entry.Name != record.Name) {
			continue
		}
		for _, integrity := range strings.Fields(entry.Integrity) {
			if integrity == record.Checksum {
				return true
			}
		}
	}
	return false
}

// npmLockPaths returns dependency lockfiles present in the working tree.
// It is called only after the diff's base SHA has been supplied by CI.
func npmLockPaths(base string) ([]string, error) {
	out, err := exec.CommandContext(context.Background(), "git", "diff", "--name-only", base+"...HEAD").Output() //nolint:gosec // base is supplied by trusted CI
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, path := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		switch filepath.Base(path) {
		case "package-lock.json", "npm-shrinkwrap.json":
			if path != "" {
				paths = append(paths, path)
			}
		}
	}
	return paths, nil
}

func verifiedNPMDependencyLockEvidence(base string, record dependencyReviewRecord) bool {
	paths, err := npmLockPaths(base)
	if err != nil || len(paths) == 0 {
		return false
	}
	for _, path := range paths {
		if checksumPresentInNPMLock(path, record) {
			return true
		}
	}
	return false
}

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type npmLockFile struct {
	LockfileVersion int                       `json:"lockfileVersion"`
	Packages        map[string]npmLockPackage `json:"packages"`
}

type npmLockPackage struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Integrity string `json:"integrity"`
}

type npmLockChange struct {
	Name         string
	OldVersion   string
	NewVersion   string
	OldIntegrity string
	NewIntegrity string
}

func parseNPMLock(data []byte) (npmLockFile, bool) {
	var lock npmLockFile
	if err := json.Unmarshal(data, &lock); err != nil || (lock.LockfileVersion != 2 && lock.LockfileVersion != 3) || lock.Packages == nil {
		return npmLockFile{}, false
	}
	return lock, true
}

func npmPackageVersionInLock(data []byte, name string) (string, bool) {
	lock, ok := parseNPMLock(data)
	if !ok {
		return "", false
	}
	entry, ok := lock.Packages["node_modules/"+name]
	if !ok || entry.Version == "" || (entry.Name != "" && entry.Name != name) {
		return "", false
	}
	return entry.Version, true
}

func diffNPMLock(oldBytes, newBytes []byte) ([]npmLockChange, bool) {
	newLock, ok := parseNPMLock(newBytes)
	if !ok {
		return nil, false
	}
	oldLock := npmLockFile{Packages: map[string]npmLockPackage{}}
	if len(oldBytes) != 0 {
		oldLock, ok = parseNPMLock(oldBytes)
		if !ok {
			return nil, false
		}
	}

	paths := make([]string, 0, len(oldLock.Packages)+len(newLock.Packages))
	seen := make(map[string]bool, len(oldLock.Packages)+len(newLock.Packages))
	for path := range oldLock.Packages {
		if path != "" {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	for path := range newLock.Packages {
		if path != "" && !seen[path] {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)

	changes := make([]npmLockChange, 0)
	for _, path := range paths {
		oldEntry, hadOld := oldLock.Packages[path]
		newEntry, hasNew := newLock.Packages[path]
		if hadOld && hasNew && oldEntry.Name == newEntry.Name && oldEntry.Version == newEntry.Version && oldEntry.Integrity == newEntry.Integrity {
			continue
		}
		entry := newEntry
		if !hasNew {
			entry = oldEntry
		}
		name := entry.Name
		if name == "" {
			name = npmPackageNameFromPath(path)
		}
		if name == "" {
			return nil, false
		}
		change := npmLockChange{Name: name}
		if hadOld {
			change.OldVersion = oldEntry.Version
			change.OldIntegrity = oldEntry.Integrity
		}
		if hasNew {
			change.NewVersion = newEntry.Version
			change.NewIntegrity = newEntry.Integrity
		}
		changes = append(changes, change)
	}
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Name != changes[j].Name {
			return changes[i].Name < changes[j].Name
		}
		return changes[i].NewVersion < changes[j].NewVersion
	})
	return changes, true
}

func npmPackageNameFromPath(path string) string {
	const marker = "node_modules/"
	index := strings.LastIndex(path, marker)
	if index < 0 {
		return ""
	}
	return strings.TrimPrefix(path[index+len(marker):], "/")
}

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
	lock, ok := parseNPMLock(data)
	if !ok {
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
	if strings.TrimSpace(base) == "" {
		return nil, errors.New("base SHA is required")
	}
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

func npmLockPathForManifest(base, manifestPath string) (string, error) {
	manifestPath, err := safeRelativePath(manifestPath)
	if err != nil {
		return "", fmt.Errorf("invalid npm manifest path: %w", err)
	}
	paths, err := npmLockPaths(base)
	if err != nil {
		return "", err
	}
	manifestDir := filepath.Dir(manifestPath)
	for _, path := range paths {
		lockPath, pathErr := safeRelativePath(path)
		if pathErr != nil {
			return "", fmt.Errorf("invalid npm lockfile path: %w", pathErr)
		}
		if filepath.Dir(lockPath) == manifestDir {
			return lockPath, nil
		}
	}
	return "", nil
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

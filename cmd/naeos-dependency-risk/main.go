// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/governance/dependencyrisk"
	"golang.org/x/mod/semver"
)

type requestFile struct {
	PolicyVersion   string   `json:"policy_version"`
	SchemaVersion   string   `json:"schema_version"`
	Ecosystem       string   `json:"ecosystem"`
	Name            string   `json:"name"`
	VersionChange   string   `json:"version_change"`
	Paths           []string `json:"paths"`
	Evidence        bool     `json:"evidence_available"`
	KnownDependency bool     `json:"known_dependency"`
}

type policyFile struct {
	PolicyID      string `json:"policy_id"`
	PolicyVersion string `json:"policy_version"`
	SchemaVersion string `json:"schema_version"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	policyBytes, err := os.ReadFile("governance/dependency-risk-policy.json")
	if err != nil {
		return fmt.Errorf("read policy: %w", err)
	}
	var policy policyFile
	if err := json.Unmarshal(policyBytes, &policy); err != nil {
		return fmt.Errorf("parse policy: %w", err)
	}
	if policy.PolicyID == "" || policy.PolicyVersion != "1.0.0" || policy.SchemaVersion != "1.0.0" {
		return errors.New("unsupported dependency-risk policy version")
	}

	requestPath := os.Getenv("NAEOS_DEPENDENCY_RISK_REQUEST")
	var req requestFile
	if requestPath == "" {
		req, err = deriveRequest()
		if err != nil {
			return err
		}
	} else {
		requestPath, err = safeRelativePath(requestPath)
		if err != nil {
			return fmt.Errorf("invalid request path: %w", err)
		}
		requestBytes, err := os.ReadFile(requestPath) //nolint:gosec // validated as workspace-relative above
		if err != nil {
			return fmt.Errorf("read request: %w", err)
		}
		if err := json.Unmarshal(requestBytes, &req); err != nil {
			return fmt.Errorf("parse request: %w", err)
		}
	}
	if req.PolicyVersion == "" {
		req.PolicyVersion = policy.PolicyVersion
	}
	if req.SchemaVersion == "" {
		req.SchemaVersion = policy.SchemaVersion
	}
	if req.PolicyVersion != policy.PolicyVersion || req.SchemaVersion != policy.SchemaVersion {
		return errors.New("request policy/schema version is unsupported")
	}

	result := dependencyrisk.Classify(dependencyrisk.Request{
		Ecosystem:         req.Ecosystem,
		Name:              req.Name,
		VersionChange:     dependencyrisk.VersionChange(req.VersionChange),
		Paths:             req.Paths,
		EvidenceAvailable: req.Evidence,
		KnownDependency:   req.KnownDependency,
	})
	evidence := struct {
		PolicyID      string                `json:"policy_id"`
		PolicyVersion string                `json:"policy_version"`
		Result        dependencyrisk.Result `json:"result"`
	}{policy.PolicyID, policy.PolicyVersion, result}

	out := os.Getenv("NAEOS_DEPENDENCY_RISK_OUTPUT")
	if out == "" {
		out = "dependency-risk-evidence.json"
	}
	out, err = safeRelativePath(out)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}
	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	data = append(data, '\n')
	// The output path is constrained by safeRelativePath before filesystem access.
	if err := os.WriteFile(out, data, 0o600); err != nil { //nolint:gosec // validated as workspace-relative above
		return fmt.Errorf("write evidence: %w", err)
	}
	fmt.Printf("dependency risk: decision=%s risk=%s criticality=%s output=%s\n", result.Decision, result.Risk, result.Criticality, out)
	if result.Decision == dependencyrisk.Deny {
		return errors.New("dependency risk policy denied the change")
	}
	return nil
}

func safeRelativePath(value string) (string, error) {
	clean := filepath.Clean(value)
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errors.New("path must be relative and remain within the workspace")
	}
	return clean, nil
}

var requireLine = regexp.MustCompile("^[+-]\\s*([^\\s]+)\\s+v?([^\\s]+)")

func deriveRequest() (requestFile, error) {
	base := os.Getenv("NAEOS_DEPENDENCY_RISK_BASE_SHA")
	if base == "" {
		return requestFile{}, errors.New("NAEOS_DEPENDENCY_RISK_BASE_SHA is required")
	}
	raw, err := exec.Command("git", "diff", base+"...HEAD", "--", "go.mod").Output()
	if err != nil {
		return requestFile{}, fmt.Errorf("read dependency diff: %w", err)
	}
	oldv, newv := map[string]string{}, map[string]string{}
	sc := bufio.NewScanner(strings.NewReader(string(raw)))
	for sc.Scan() {
		line := sc.Text()
		if len(line) < 2 || (line[0] != '+' && line[0] != '-') || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		}
		mm := requireLine.FindStringSubmatch(line)
		if len(mm) != 3 {
			continue
		}
		if line[0] == '-' {
			oldv[mm[1]] = mm[2]
		} else {
			newv[mm[1]] = mm[2]
		}
	}
	names := []string{}
	for n := range oldv {
		if _, ok := newv[n]; ok {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return requestFile{PolicyVersion: "1.0.0", SchemaVersion: "1.0.0", Ecosystem: "go", Name: "dependency-change", VersionChange: "unknown", Evidence: true, KnownDependency: false}, nil
	}
	name := names[0]
	return requestFile{PolicyVersion: "1.0.0", SchemaVersion: "1.0.0", Ecosystem: "go", Name: name, VersionChange: versionChange(oldv[name], newv[name]), Evidence: true, KnownDependency: true}, nil
}

func versionChange(oldv, newv string) string {
	oldv = "v" + strings.TrimPrefix(oldv, "v")
	newv = "v" + strings.TrimPrefix(newv, "v")
	if !semver.IsValid(oldv) || !semver.IsValid(newv) {
		return "unknown"
	}
	if semver.Major(oldv) != semver.Major(newv) {
		return "major"
	}
	if semver.MajorMinor(oldv) != semver.MajorMinor(newv) {
		return "minor"
	}
	if semver.Compare(oldv, newv) != 0 {
		return "patch"
	}
	return "unknown"
}

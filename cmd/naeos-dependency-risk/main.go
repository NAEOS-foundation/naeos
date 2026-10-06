// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/NAEOS-foundation/naeos/internal/governance/dependencyrisk"
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

var requireLine = regexp.MustCompile(`^[+-]\s*([^\s]+)\s+v?([^\s]+)`)

func deriveRequest() (requestFile, error) {
	base := os.Getenv("NAEOS_DEPENDENCY_RISK_BASE_SHA")
	if base == "" {
		return requestFile{}, errors.New("NAEOS_DEPENDENCY_RISK_BASE_SHA is required")
	}
	changed, err := changedDependencyManifests(base)
	if err != nil {
		return requestFile{}, err
	}
	if len(changed) == 0 {
		return requestFile{PolicyVersion: "1.0.0", SchemaVersion: "1.0.0", Ecosystem: "unknown", Name: "dependency-change", VersionChange: "unknown", Evidence: automaticEvidenceAvailable(), KnownDependency: false}, nil
	}
	for _, manifest := range changed {
		if filepath.Base(manifest) == "package.json" {
			if req, ok, err := deriveNPMRequest(base, manifest); err != nil {
				return requestFile{}, err
			} else if ok {
				return req, nil
			}
		}
	}
	raw, err := exec.CommandContext(context.Background(), "git", "diff", base+"...HEAD", "--", "go.mod").Output() //nolint:gosec // BASE_SHA is supplied by the trusted CI workflow
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

	// Evaluate every dependency touched by the diff. A single permissive
	// dependency must never mask a new, unknown, or higher-risk dependency.
	names := make([]string, 0, len(oldv)+len(newv))
	seen := make(map[string]struct{}, len(oldv)+len(newv))
	for name := range oldv {
		seen[name] = struct{}{}
		names = append(names, name)
	}
	for name := range newv {
		if _, ok := seen[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return requestFile{PolicyVersion: "1.0.0", SchemaVersion: "1.0.0", Ecosystem: "go", Name: "dependency-change", VersionChange: "unknown", Evidence: automaticEvidenceAvailable(), KnownDependency: false}, nil
	}

	return selectHighestRiskRequest(oldv, newv), nil
}

func changedDependencyManifests(base string) ([]string, error) {
	raw, err := exec.CommandContext(context.Background(), "git", "diff", "--name-only", base+"...HEAD").Output() //nolint:gosec // BASE_SHA is supplied by trusted CI
	if err != nil {
		return nil, fmt.Errorf("list changed files: %w", err)
	}
	var out []string
	for _, path := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if path == "" { continue }
		baseName := filepath.Base(path)
		switch baseName {
		case "go.mod", "go.sum", "package.json", "package-lock.json", "npm-shrinkwrap.json", "pnpm-lock.yaml", "yarn.lock", "requirements.txt", "pyproject.toml", "Cargo.toml", "Cargo.lock":
			out = append(out, path)
		}
	}
	sort.Strings(out)
	return out, nil
}

func gitShowFile(base, path string) ([]byte, error) {
	out, err := exec.CommandContext(context.Background(), "git", "show", base+":"+path).Output() //nolint:gosec // refs are supplied by trusted CI
	if err != nil { return nil, err }
	return out, nil
}

func deriveNPMRequest(base, path string) (requestFile, bool, error) {
	newBytes, err := os.ReadFile(path)
	if err != nil { return requestFile{}, false, fmt.Errorf("read npm manifest %s: %w", path, err) }
	oldBytes, oldErr := gitShowFile(base, path)
	if oldErr != nil { oldBytes = []byte("{}") }
	var oldPkg, newPkg map[string]any
	if err := json.Unmarshal(oldBytes, &oldPkg); err != nil { return requestFile{}, false, fmt.Errorf("parse base npm manifest %s: %w", path, err) }
	if err := json.Unmarshal(newBytes, &newPkg); err != nil { return requestFile{}, false, fmt.Errorf("parse npm manifest %s: %w", path, err) }

	sections := []string{"dependencies", "devDependencies", "optionalDependencies", "peerDependencies"}
	best := requestFile{PolicyVersion:"1.0.0", SchemaVersion:"1.0.0", Ecosystem:"npm", Name:"dependency-change", VersionChange:"unknown", Evidence:automaticEvidenceAvailable(), KnownDependency:false}
	bestRank := -1
	for _, section := range sections {
		oldDeps, _ := oldPkg[section].(map[string]any)
		newDeps, _ := newPkg[section].(map[string]any)
		names := map[string]bool{}
		for n := range oldDeps { names[n] = true }
		for n := range newDeps { names[n] = true }
		for name := range names {
			ov, ook := oldDeps[name].(string)
			nv, nok := newDeps[name].(string)
			if ook && nok && ov == nv { continue }
			candidate := requestFile{PolicyVersion:"1.0.0", SchemaVersion:"1.0.0", Ecosystem:"npm", Name:name, Paths:dependencyUsagePathsNPM(name), Evidence:automaticEvidenceAvailable(), KnownDependency:ook && nok}
			if candidate.KnownDependency { candidate.VersionChange = npmVersionChange(ov,nv) } else { candidate.VersionChange="unknown" }
			rank := requestRiskRank(candidate)
			if rank > bestRank { best, bestRank = candidate, rank }
		}
	}
	return best, bestRank >= 0, nil
}

func dependencyUsagePathsNPM(name string) []string {
	out, err := exec.CommandContext(context.Background(), "git", "grep", "-l", "--fixed-strings", name, "--", "*.js", "*.jsx", "*.ts", "*.tsx").Output() //nolint:gosec // name is derived from package.json
	if err != nil { return nil }
	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") { if line != "" { paths = append(paths,line) } }
	sort.Strings(paths)
	return paths
}

func npmVersionChange(oldv, newv string) string {
	oldv = strings.TrimSpace(strings.TrimLeft(oldv, "^~>=<v"))
	newv = strings.TrimSpace(strings.TrimLeft(newv, "^~>=<v"))
	if !semver.IsValid("v"+oldv) || !semver.IsValid("v"+newv) { return "unknown" }
	return versionChange(oldv,newv)
}

func selectHighestRiskRequest(oldv, newv map[string]string) requestFile {
	names := make([]string, 0, len(oldv)+len(newv))
	seen := make(map[string]struct{}, len(oldv)+len(newv))
	for name := range oldv {
		seen[name] = struct{}{}
		names = append(names, name)
	}
	for name := range newv {
		if _, ok := seen[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	base := requestFile{PolicyVersion: "1.0.0", SchemaVersion: "1.0.0", Ecosystem: "go", Name: "dependency-change", VersionChange: "unknown", Evidence: automaticEvidenceAvailable(), KnownDependency: false}
	if len(names) == 0 {
		return base
	}
	best := base
	bestRank := -1
	for _, name := range names {
		oldVersion, oldOK := oldv[name]
		newVersion, newOK := newv[name]
		candidate := requestFile{
			PolicyVersion:   "1.0.0",
			SchemaVersion:   "1.0.0",
			Ecosystem:       "go",
			Name:            name,
			Paths:           dependencyUsagePaths(name),
			Evidence:        automaticEvidenceAvailable(),
			KnownDependency: oldOK && newOK,
		}
		if oldOK && newOK {
			candidate.VersionChange = versionChange(oldVersion, newVersion)
		} else {
			candidate.VersionChange = "unknown"
		}
		rank := requestRiskRank(candidate)
		if rank > bestRank {
			best = candidate
			bestRank = rank
		}
	}
	return best
}

func automaticEvidenceAvailable() bool {
	return strings.EqualFold(os.Getenv("NAEOS_DEPENDENCY_RISK_EVIDENCE"), "true")
}

func dependencyUsagePaths(name string) []string {
	if name == "" {
		return nil
	}
	out, err := exec.CommandContext(context.Background(), "git", "grep", "-l", "--fixed-strings", name, "--", "*.go").Output() //nolint:gosec // dependency name is derived from go.mod
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	paths := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			paths = append(paths, line)
		}
	}
	sort.Strings(paths)
	return paths
}

func requestRiskRank(req requestFile) int {
	result := dependencyrisk.Classify(dependencyrisk.Request{
		Ecosystem:         req.Ecosystem,
		Name:              req.Name,
		VersionChange:     dependencyrisk.VersionChange(req.VersionChange),
		Paths:             req.Paths,
		EvidenceAvailable: req.Evidence,
		KnownDependency:   req.KnownDependency,
	})

	decisionRank := map[dependencyrisk.Decision]int{
		dependencyrisk.Allow:         0,
		dependencyrisk.RequireReview: 100,
		dependencyrisk.Deny:          200,
	}
	criticalityRank := map[dependencyrisk.Criticality]int{
		dependencyrisk.Low:      0,
		dependencyrisk.Medium:   10,
		dependencyrisk.High:     20,
		dependencyrisk.Critical: 30,
	}

	return decisionRank[result.Decision]*100 + criticalityRank[result.Criticality]
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

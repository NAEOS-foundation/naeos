// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

const dependencyRegistryPath = "governance/dependency-review-registry.json"

type dependencyReviewRegistry struct {
	SchemaVersion string                   `json:"schema_version"`
	PolicyVersion string                   `json:"policy_version"`
	Records       []dependencyReviewRecord `json:"records"`
}

type dependencyReviewRecord struct {
	Ecosystem       string   `json:"ecosystem"`
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	UpstreamSource  string   `json:"upstream_source"`
	SPDXLicense     string   `json:"spdx_license"`
	Checksum        string   `json:"checksum"`
	Provenance      string   `json:"provenance"`
	Purpose         string   `json:"purpose"`
	Domains         []string `json:"domains"`
	Criticality     string   `json:"criticality"`
	Reviewer        string   `json:"reviewer"`
	ReviewReference string   `json:"review_reference"`
	PolicyVersion   string   `json:"policy_version"`
	SchemaVersion   string   `json:"schema_version"`
}

var checksumPattern = regexp.MustCompile(`^(h1:[A-Za-z0-9+/]{43}=|sha256:[a-fA-F0-9]{64})$`)
var spdxTokens = regexp.MustCompile(`[()]|[^[:space:]()]+`)

// parseDependencyReviewRegistry rejects malformed or incomplete records. The
// caller must supply bytes loaded from the trusted merge-base, never the PR tree.
func parseDependencyReviewRegistry(data []byte) (dependencyReviewRegistry, error) {
	var registry dependencyReviewRegistry
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&registry); err != nil {
		return registry, fmt.Errorf("decode dependency review registry: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return registry, errors.New("dependency review registry contains trailing JSON or malformed trailing data")
	}
	if registry.SchemaVersion != "1.0.0" || registry.PolicyVersion != "1.0.0" || registry.Records == nil {
		return registry, errors.New("unsupported or incomplete dependency review registry")
	}
	seen := make(map[string]bool, len(registry.Records))
	for i, record := range registry.Records {
		fields := []string{record.Ecosystem, record.Name, record.Version, record.UpstreamSource, record.SPDXLicense, record.Checksum, record.Provenance, record.Purpose, record.Criticality, record.Reviewer, record.ReviewReference, record.PolicyVersion, record.SchemaVersion}
		for _, field := range fields {
			if strings.TrimSpace(field) == "" {
				return registry, fmt.Errorf("dependency review record %d has a missing required field", i)
			}
		}
		if record.PolicyVersion != registry.PolicyVersion || record.SchemaVersion != registry.SchemaVersion {
			return registry, fmt.Errorf("dependency review record %d has unsupported policy/schema version", i)
		}
		if !strings.HasPrefix(record.UpstreamSource, "https://") || !strings.HasPrefix(record.Provenance, "https://") || !strings.HasPrefix(record.ReviewReference, "https://") {
			return registry, fmt.Errorf("dependency review record %d has invalid upstream source", i)
		}
		if !validSPDXLicense(record.SPDXLicense) || !checksumPattern.MatchString(record.Checksum) || (strings.EqualFold(record.Ecosystem, "go") && !strings.HasPrefix(record.Checksum, "h1:")) {
			return registry, fmt.Errorf("dependency review record %d has invalid SPDX license or checksum", i)
		}
		switch record.Criticality {
		case "low", "medium", "high", "critical":
		default:
			return registry, fmt.Errorf("dependency review record %d has invalid criticality", i)
		}
		if len(record.Domains) == 0 {
			return registry, fmt.Errorf("dependency review record %d has no affected domains", i)
		}
		for _, domain := range record.Domains {
			if !validDependencyDomain(domain) {
				return registry, fmt.Errorf("dependency review record %d has an empty affected domain", i)
			}
		}
		sort.Strings(record.Domains)
		key := strings.ToLower(record.Ecosystem + "|" + record.Name + "|" + record.Version)
		if seen[key] {
			return registry, fmt.Errorf("duplicate dependency review record for %s", key)
		}
		seen[key] = true
	}
	return registry, nil
}

func checksumPresentInGoSum(data []byte, record dependencyReviewRecord) bool {
	if !strings.EqualFold(record.Ecosystem, "go") {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == record.Name && fields[1] == record.Version && fields[2] == record.Checksum {
			return true
		}
	}
	return false
}

// validSPDXLicense accepts a deliberately bounded SPDX expression grammar.
// Only explicitly reviewed license identifiers are accepted; LicenseRef and
// WITH exceptions remain unsupported until their provenance is modeled.
func validSPDXLicense(value string) bool {
	tokens := spdxTokens.FindAllString(strings.TrimSpace(value), -1)
	if len(tokens) == 0 {
		return false
	}
	parser := spdxExpressionParser{tokens: tokens}
	return parser.parseOr() && parser.pos == len(parser.tokens)
}

type spdxExpressionParser struct {
	tokens []string
	pos    int
}

func (p *spdxExpressionParser) parseOr() bool {
	if !p.parseAnd() {
		return false
	}
	for p.consume("OR") {
		if !p.parseAnd() {
			return false
		}
	}
	return true
}

func (p *spdxExpressionParser) parseAnd() bool {
	if !p.parsePrimary() {
		return false
	}
	for p.consume("AND") {
		if !p.parsePrimary() {
			return false
		}
	}
	return true
}

func (p *spdxExpressionParser) parsePrimary() bool {
	if p.consume("(") {
		if !p.parseOr() || !p.consume(")") {
			return false
		}
		return true
	}
	if p.pos >= len(p.tokens) {
		return false
	}
	identifier := p.tokens[p.pos]
	if identifier == ")" || identifier == "(" || identifier == "AND" || identifier == "OR" {
		return false
	}
	p.pos++
	switch identifier {
	case "Apache-2.0", "MIT", "BSD-2-Clause", "BSD-3-Clause", "BSD-0-Clause", "ISC", "MPL-2.0", "LGPL-2.1-only", "LGPL-3.0-only", "GPL-2.0-only", "GPL-3.0-only", "Unlicense", "CC0-1.0", "Zlib", "EPL-2.0", "Unicode-3.0":
		return true
	default:
		return false
	}
}

func (p *spdxExpressionParser) consume(token string) bool {
	if p.pos >= len(p.tokens) || p.tokens[p.pos] != token {
		return false
	}
	p.pos++
	return true
}

func validDependencyDomain(value string) bool {
	switch value {
	case "policy", "security", "audit_evidence", "runtime", "deployment_ci", "general":
		return true
	default:
		return false
	}
}

func findReviewedDependency(registry dependencyReviewRegistry, ecosystem, name, version string) (dependencyReviewRecord, bool) {
	for _, record := range registry.Records {
		if strings.EqualFold(record.Ecosystem, ecosystem) && record.Name == name && record.Version == version {
			return record, true
		}
	}
	return dependencyReviewRecord{}, false
}

func loadTrustedDependencyReviewRegistry(base string) (dependencyReviewRegistry, error) {
	if strings.TrimSpace(base) == "" {
		return dependencyReviewRegistry{}, errors.New("trusted dependency registry base SHA is required")
	}
	data, err := exec.CommandContext(context.Background(), "git", "show", base+":"+dependencyRegistryPath).Output() //nolint:gosec // base is a trusted SHA from the protected CI workflow; path is a constant
	if err != nil {
		return dependencyReviewRegistry{}, fmt.Errorf("read trusted registry from merge base: %w", err)
	}
	return parseDependencyReviewRegistry(data)
}

func trustedReviewRecord(base, ecosystem, name, version string) (dependencyReviewRecord, bool) {
	registry, err := loadTrustedDependencyReviewRegistry(base)
	if err != nil {
		return dependencyReviewRecord{}, false
	}
	return findReviewedDependency(registry, ecosystem, name, version)
}

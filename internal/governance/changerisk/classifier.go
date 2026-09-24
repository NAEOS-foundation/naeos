// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package changerisk

import "sort"

type ChangeSurface string

const (
	Docs          ChangeSurface = "docs"
	Tests         ChangeSurface = "tests"
	Code          ChangeSurface = "code"
	Configuration ChangeSurface = "configuration"
	Mixed         ChangeSurface = "mixed"
	Unknown       ChangeSurface = "unknown"
)

type Criticality string

const (
	Low      Criticality = "low"
	Medium   Criticality = "medium"
	High     Criticality = "high"
	Critical Criticality = "critical"
)

type Risk string
type Decision string

const Allow Decision = "allow"
const RequireReview Decision = "require_review"
const Deny Decision = "deny"

type Domain string

const Policy Domain = "policy"
const Security Domain = "security"
const AuditEvidence Domain = "audit_evidence"
const Runtime Domain = "runtime"
const DeploymentCI Domain = "deployment_ci"
const Architecture Domain = "architecture"
const General Domain = "general"
const UnknownDomain Domain = "unknown"

type Request struct {
	Paths             []string
	Additions         int
	Deletions         int
	EvidenceAvailable bool
}

type Result struct {
	SchemaVersion    string        `json:"schema_version"`
	ChangeSurface    ChangeSurface `json:"change_surface"`
	Criticality      Criticality   `json:"criticality"`
	Risk             Risk           `json:"risk"`
	Domains          []Domain      `json:"domains"`
	ChangedFiles     int           `json:"changed_files"`
	Additions        int           `json:"additions"`
	Deletions        int           `json:"deletions"`
	RequiredGates    []string      `json:"required_gates"`
	EvidenceRequired bool          `json:"evidence_required"`
	Decision         Decision       `json:"decision"`
}

func Classify(req Request) Result {
	domains := DomainsForPaths(req.Paths)
	surface := SurfaceForPaths(req.Paths)
	criticality := criticalityFor(req, surface, domains)

	gates, decision := gatesFor(criticality)
	risk := Risk(criticality)
	if !req.EvidenceAvailable || criticality == Criticality("unknown") {
		decision = Deny
		risk = Risk("unknown")
		gates = []string{"ci", "governance", "human_review", "security"}
	}

	sort.Strings(gates)
	return Result{
		SchemaVersion:    "1.0.0",
		ChangeSurface:    surface,
		Criticality:      criticality,
		Risk:             risk,
		Domains:          domains,
		ChangedFiles:     len(req.Paths),
		Additions:        req.Additions,
		Deletions:        req.Deletions,
		RequiredGates:    gates,
		EvidenceRequired: true,
		Decision:         decision,
	}
}

func SurfaceForPaths(paths []string) ChangeSurface {
	if len(paths) == 0 {
		return Unknown
	}

	hasDocs, hasTests, hasConfig, hasCode := false, false, false, false
	for _, path := range paths {
		switch {
		case isDoc(path):
			hasDocs = true
		case isTest(path):
			hasTests = true
		case isConfig(path):
			hasConfig = true
		default:
			hasCode = true
		}
	}

	count := 0
	for _, present := range []bool{hasDocs, hasTests, hasConfig, hasCode} {
		if present {
			count++
		}
	}
	if count > 1 {
		return Mixed
	}
	switch {
	case hasDocs:
		return Docs
	case hasTests:
		return Tests
	case hasConfig:
		return Configuration
	case hasCode:
		return Code
	default:
		return Unknown
	}
}

func DomainsForPaths(paths []string) []Domain {
	seen := map[Domain]bool{}
	for _, path := range paths {
		switch {
		case hasPrefix(path, "internal/governance/"), hasPrefix(path, "governance/"), hasPrefix(path, "constitution/"), hasPrefix(path, "policy/"):
			seen[Policy] = true
		case hasPrefix(path, "internal/security/"), hasPrefix(path, "security/"):
			seen[Security] = true
		case hasPrefix(path, "internal/audit/"), hasPrefix(path, "audit/"), hasPrefix(path, "internal/evidence/"):
			seen[AuditEvidence] = true
		case hasPrefix(path, "runtime/"), hasPrefix(path, "internal/runtime/"), hasPrefix(path, "internal/agent/"), hasPrefix(path, "internal/controlplane/"):
			seen[Runtime] = true
		case hasPrefix(path, ".github/"), hasPrefix(path, "scripts/"):
			seen[DeploymentCI] = true
		case hasPrefix(path, "architecture/"), hasPrefix(path, "Reference Architecture/"), hasPrefix(path, "specification/"):
			seen[Architecture] = true
		case isTest(path):
			// Test-only changes are governed by the normal CI verification boundary.
			continue
		default:
			seen[General] = true
		}
	}
	if len(seen) == 0 {
		return []Domain{UnknownDomain}
	}
	out := make([]Domain, 0, len(seen))
	for domain := range seen {
		out = append(out, domain)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func criticalityFor(req Request, surface ChangeSurface, domains []Domain) Criticality {
	if surface == Unknown || len(req.Paths) == 0 {
		return Criticality("unknown")
	}
	criticality := surfaceCriticality(surface)
	for _, domain := range domains {
		switch domain {
		case Policy, Security, AuditEvidence, Runtime, DeploymentCI, Architecture:
			criticality = maxCriticality(criticality, High)
		case UnknownDomain:
			return Criticality("unknown")
		}
	}
	total := req.Additions + req.Deletions
	if total >= 1000 {
		criticality = maxCriticality(criticality, Critical)
	} else if total >= 500 {
		criticality = maxCriticality(criticality, High)
	}
	return criticality
}

func surfaceCriticality(surface ChangeSurface) Criticality {
	switch surface {
	case Docs, Tests:
		return Low
	case Configuration, Code, Mixed:
		return Medium
	default:
		return Criticality("unknown")
	}
}

func gatesFor(criticality Criticality) ([]string, Decision) {
	switch criticality {
	case Low:
		return []string{"ci"}, Allow
	case Medium:
		return []string{"ci", "governance"}, Allow
	case High:
		return []string{"ci", "governance", "human_review", "security"}, RequireReview
	case Critical:
		return []string{"benchmark", "ci", "governance", "human_review", "security"}, RequireReview
	default:
		return []string{"ci", "governance", "human_review", "security"}, Deny
	}
}

func maxCriticality(a, b Criticality) Criticality {
	order := map[Criticality]int{Low: 1, Medium: 2, High: 3, Critical: 4}
	if order[b] > order[a] {
		return b
	}
	return a
}

func isDoc(path string) bool {
	return hasSuffix(path, ".md") || hasSuffix(path, ".mdx")
}

func isTest(path string) bool {
	return hasSuffix(path, "_test.go") || hasPrefix(path, "tests/")
}

func isConfig(path string) bool {
	return hasPrefix(path, ".github/") || hasPrefix(path, ".devcontainer/") ||
		hasSuffix(path, ".yaml") || hasSuffix(path, ".yml") ||
		hasSuffix(path, ".json") || hasSuffix(path, ".toml")
}

func hasPrefix(value, prefix string) bool {
	return len(value) >= len(prefix) && value[:len(prefix)] == prefix
}

func hasSuffix(value, suffix string) bool {
	return len(value) >= len(suffix) && value[len(value)-len(suffix):] == suffix
}

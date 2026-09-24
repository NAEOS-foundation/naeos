// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package dependencyrisk

import "sort"

type VersionChange string
const (
 Patch VersionChange = "patch"
 Minor VersionChange = "minor"
 Major VersionChange = "major"
 Unknown VersionChange = "unknown"
)

type Criticality string
const (
 Low Criticality = "low"
 Medium Criticality = "medium"
 High Criticality = "high"
 Critical Criticality = "critical"
)

type Risk string
type Decision string
const (
 Allow Decision = "allow"
 RequireReview Decision = "require_review"
 Deny Decision = "deny"
)

type Domain string
const (
 Policy Domain = "policy"
 Security Domain = "security"
 AuditEvidence Domain = "audit_evidence"
 Runtime Domain = "runtime"
 DeploymentCI Domain = "deployment_ci"
 General Domain = "general"
 UnknownDomain Domain = "unknown"
)

type Request struct {
 Ecosystem string
 Name string
 VersionChange VersionChange
 Paths []string
 EvidenceAvailable bool
}

type Result struct {
 SchemaVersion string
 Dependency Request
 Criticality Criticality
 Risk Risk
 Domains []Domain
 RequiredGates []string
 EvidenceRequired bool
 Decision Decision
}

// Classify is deterministic. Unknown impact or unavailable evidence fails closed.
func Classify(req Request) Result {
 domains := DomainsForPaths(req.Paths)
 criticality := criticalityFor(req, domains)
 gates, decision := gatesFor(criticality)
 risk := Risk(criticality)

 if !req.EvidenceAvailable || req.VersionChange == Unknown || criticality == Criticality("unknown") {
  decision = Deny
  risk = Risk("unknown")
  if !contains(gates, "human_review") { gates = append(gates, "human_review") }
 }
 sort.Strings(gates)
 return Result{"1.0.0", req, criticality, risk, domains, gates, true, decision}
}

func DomainsForPaths(paths []string) []Domain {
 seen := map[Domain]bool{}
 for _, p := range paths {
  switch {
  case hasPrefix(p, ".github/"), hasPrefix(p, "scripts/"):
   seen[DeploymentCI] = true
  case hasPrefix(p, "internal/governance/"), hasPrefix(p, "governance/"), hasPrefix(p, "constitution/"):
   seen[Policy] = true
  case hasPrefix(p, "internal/security/"), hasPrefix(p, "security/"):
   seen[Security] = true
  case hasPrefix(p, "internal/audit/"), hasPrefix(p, "audit/"):
   seen[AuditEvidence] = true
  case hasPrefix(p, "runtime/"), hasPrefix(p, "internal/runtime/"), hasPrefix(p, "internal/agent/"):
   seen[Runtime] = true
  default:
   seen[General] = true
  }
 }
 if len(seen) == 0 { return []Domain{UnknownDomain} }
 out := make([]Domain, 0, len(seen))
 for d := range seen { out = append(out, d) }
 sort.Slice(out, func(i,j int) bool { return out[i] < out[j] })
 return out
}

func criticalityFor(req Request, domains []Domain) Criticality {
 if req.VersionChange == Unknown { return Criticality("unknown") }
 for _, d := range domains {
  if d == Policy || d == Security || d == AuditEvidence || d == Runtime || d == DeploymentCI { return High }
 }
 switch req.VersionChange {
 case Patch: return Low
 case Minor: return Medium
 case Major: return High
 default: return Criticality("unknown")
 }
}

func gatesFor(c Criticality) ([]string, Decision) {
 switch c {
 case Low: return []string{"ci"}, Allow
 case Medium: return []string{"ci", "governance"}, Allow
 case High: return []string{"ci", "governance", "human_review", "security"}, RequireReview
 case Critical: return []string{"benchmark", "ci", "governance", "human_review", "security"}, RequireReview
 default: return []string{"ci", "governance", "human_review", "security"}, Deny
 }
}

func hasPrefix(v, prefix string) bool { return len(v) >= len(prefix) && v[:len(prefix)] == prefix }
func contains(values []string, target string) bool {
 for _, v := range values { if v == target { return true } }
 return false
}

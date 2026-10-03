// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type State string
const (
	Unknown State = "UNKNOWN"
	Confirmed State = "CONFIRMED"
	Rejected State = "REJECTED"
)

type Reversibility string
const (
	Reversible Reversibility = "REVERSIBLE"
	Irreversible Reversibility = "IRREVERSIBLE"
)

type EvidenceType string
const (
	ProviderReceipt EvidenceType = "PROVIDER_RECEIPT"
	ResourceState EvidenceType = "RESOURCE_STATE"
	AuditEvent EvidenceType = "AUDIT_EVENT"
)

type Evidence struct {
	Type EvidenceType
	Source string
	Outcome State
	ObservedAt time.Time
}

type VerificationPolicy struct {
	Reversibility Reversibility
	MaxAge time.Duration
	MinIndependentProof int
	AllowedEvidence []EvidenceType
}

type Scenario struct {
	Name string
	InitialState State
	Reversibility Reversibility
	Evidence []Evidence
	Expected State
	Observed State
	Passed bool
	Reason string
}

func main() {
	now := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	scenarios := []Scenario{
		{Name: "01-reversible-single-fresh-receipt", InitialState: Unknown, Reversibility: Reversible,
			Evidence: []Evidence{{ProviderReceipt, "provider-a", Confirmed, now.Add(-2*time.Minute)}}, Expected: Confirmed},
		{Name: "02-irreversible-single-signal-insufficient", InitialState: Unknown, Reversibility: Irreversible,
			Evidence: []Evidence{{ProviderReceipt, "provider-a", Confirmed, now.Add(-2*time.Minute)}}, Expected: Unknown},
		{Name: "03-irreversible-two-independent-signals", InitialState: Unknown, Reversibility: Irreversible,
			Evidence: []Evidence{{ProviderReceipt, "provider-a", Confirmed, now.Add(-2*time.Minute)}, {ResourceState, "resource-observer", Confirmed, now.Add(-1*time.Minute)}}, Expected: Confirmed},
		{Name: "04-stale-evidence-remains-unknown", InitialState: Unknown, Reversibility: Reversible,
			Evidence: []Evidence{{ResourceState, "resource-observer", Confirmed, now.Add(-30*time.Minute)}}, Expected: Unknown},
		{Name: "05-conflicting-independent-evidence-remains-unknown", InitialState: Unknown, Reversibility: Irreversible,
			Evidence: []Evidence{{ProviderReceipt, "provider-a", Confirmed, now.Add(-2*time.Minute)}, {ResourceState, "resource-observer", Rejected, now.Add(-1*time.Minute)}}, Expected: Unknown},
		{Name: "06-two-signals-from-same-source-are-not-independent", InitialState: Unknown, Reversibility: Irreversible,
			Evidence: []Evidence{{ProviderReceipt, "provider-a", Confirmed, now.Add(-2*time.Minute)}, {AuditEvent, "provider-a", Confirmed, now.Add(-1*time.Minute)}}, Expected: Unknown},
	}
	for i := range scenarios {
		scenarios[i].Observed, scenarios[i].Reason = evaluate(scenarios[i].Evidence, policyFor(scenarios[i].Reversibility), now)
		scenarios[i].Passed = scenarios[i].Observed == scenarios[i].Expected
	}
	out := struct {
		Experiment string
		Thesis string
		Results []Scenario
		Summary string
	}{
		"NAEOS Evidence Policy + Reversibility Experiment v1",
		"Evidence sufficiency must be explicit, freshness-aware, and stricter for irreversible actions.",
		scenarios,
		summarize(scenarios),
	}
	enc := json.NewEncoder(os.Stdout); enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
	for _, s := range scenarios { if !s.Passed { os.Exit(2) } }
}

func policyFor(r Reversibility) VerificationPolicy {
	if r == Irreversible {
		return VerificationPolicy{r, 10*time.Minute, 2, []EvidenceType{ProviderReceipt, ResourceState, AuditEvent}}
	}
	return VerificationPolicy{r, 10*time.Minute, 1, []EvidenceType{ProviderReceipt, ResourceState, AuditEvent}}
}

func evaluate(es []Evidence, p VerificationPolicy, now time.Time) (State, string) {
	if len(es) == 0 { return Unknown, "no evidence" }
	sources := map[string]bool{}
	var outcome State
	for _, e := range es {
		if now.Sub(e.ObservedAt) > p.MaxAge { return Unknown, "evidence is stale" }
		if !contains(p.AllowedEvidence, e.Type) { continue }
		if outcome == "" { outcome = e.Outcome } else if outcome != e.Outcome { return Unknown, "independent evidence conflicts" }
		sources[e.Source] = true
	}
	if outcome == "" { return Unknown, "no acceptable evidence" }
	if len(sources) < p.MinIndependentProof { return Unknown, "insufficient independent evidence for reversibility policy" }
	if len(sources) != len(es) { return Unknown, "evidence signals are not independent" }
	return outcome, "evidence satisfies verification policy"
}

func contains(values []EvidenceType, target EvidenceType) bool {
	for _, v := range values { if v == target { return true } }
	return false
}

func summarize(results []Scenario) string {
	passed := 0
	for _, r := range results { if r.Passed { passed++ } }
	return fmt.Sprintf("%d/%d expected transition assertions passed", passed, len(results))
}

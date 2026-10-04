// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

// Package main runs the Evidence & Verification Boundary Matrix experiment.
// It intentionally models the policy/evidence boundary without changing the
// production verification kernel. The goal is to turn the external feedback
// into deterministic, reviewable evidence before promoting any primitive.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Effect string
type Reversibility string
type EvidenceStatus string
type Outcome string

const (
	InternalEffect Effect = "internal"
	ExternalEffect Effect = "external"

	Reversible   Reversibility = "reversible"
	Irreversible Reversibility = "irreversible"

	Observed      EvidenceStatus = "observed"
	Derived       EvidenceStatus = "derived"
	Estimated     EvidenceStatus = "estimated"
	Stale         EvidenceStatus = "stale"
	Contradictory EvidenceStatus = "contradictory"
	Missing       EvidenceStatus = "missing"

	Confirmed Outcome = "CONFIRMED"
	Rejected  Outcome = "REJECTED"
	Unknown   Outcome = "UNKNOWN"
	Reverify  Outcome = "REVERIFY"
	Escalate  Outcome = "ESCALATE"
)

type Evidence struct {
	ID            string         `json:"id"`
	Source        string         `json:"source"`
	Type          string         `json:"type"`
	Status        EvidenceStatus `json:"status"`
	Authoritative bool           `json:"authoritative"`
}

type Case struct {
	Name          string        `json:"name"`
	Effect        Effect        `json:"effect"`
	Reversibility Reversibility `json:"reversibility"`
	Evidence      []Evidence    `json:"evidence"`
	HasPolicy     bool          `json:"has_policy"`
	MinimumFloor  string        `json:"minimum_evidence_floor"`
}

type Result struct {
	CaseName           string   `json:"case"`
	Outcome            Outcome  `json:"outcome"`
	Reason             string   `json:"reason"`
	EvidenceUsed       []string `json:"evidence_used"`
	EvidenceIDs        []string `json:"evidence_ids"`
	EvidenceProvenance []string `json:"evidence_provenance"`
	ExpectedOutcome     Outcome  `json:"expected_outcome"`
	Passed              bool     `json:"passed"`
}

func minimumFloor(c Case) string {
	if c.Reversibility == Irreversible {
		return "authoritative-observed"
	}
	if c.Effect == ExternalEffect {
		return "authoritative-observed-or-provider-receipt"
	}
	return "runtime-record-or-observed"
}

func verify(c Case) Result {
	floor := c.MinimumFloor
	if floor == "" {
		floor = minimumFloor(c)
	}

	used := make([]string, 0, len(c.Evidence))
	ids := make([]string, 0, len(c.Evidence))
	provenance := make([]string, 0, len(c.Evidence))

	if !c.HasPolicy {
		for _, e := range c.Evidence {
			ids = append(ids, e.ID)
			provenance = append(provenance, e.Source)
		}
		return Result{c.Name, Escalate, "missing verification policy; fail closed", nil, ids, provenance, Escalate, true}
	}

	hasAuthoritativeObserved := false
	hasStale := false
	hasContradiction := false
	hasEstimated := false

	for _, e := range c.Evidence {
		used = append(used, e.Type+":"+string(e.Status))
		ids = append(ids, e.ID)
		provenance = append(provenance, e.Source)
		switch e.Status {
		case Observed:
			// A runtime execution record is authoritative only for effects the
			// runtime itself owns. It is not external proof merely because the
			// record is internally authoritative.
			if e.Authoritative && e.Type != "execution_record" {
				hasAuthoritativeObserved = true
			}
		case Stale:
			hasStale = true
		case Contradictory:
			hasContradiction = true
		case Estimated:
			hasEstimated = true
		}
	}

	if hasContradiction {
		return Result{c.Name, Unknown, "conflicting evidence cannot establish the outcome", used, ids, provenance, Unknown, true}
	}
	if hasStale {
		return Result{c.Name, Reverify, "evidence was sufficient previously but is stale now", used, ids, provenance, Reverify, true}
	}

	// Estimated evidence can be retained and useful, but it cannot satisfy an
	// authoritative floor for an irreversible or external effect.
	if hasEstimated && (c.Effect == ExternalEffect || c.Reversibility == Irreversible) {
		return Result{c.Name, Unknown, "estimated evidence cannot satisfy the minimum floor", used, ids, provenance, Unknown, true}
	}

	if c.Effect == ExternalEffect && !hasAuthoritativeObserved {
		return Result{c.Name, Unknown, "runtime evidence proves only the attempt; external authority is required", used, ids, provenance, Unknown, true}
	}
	if c.Reversibility == Irreversible && !hasAuthoritativeObserved {
		return Result{c.Name, Unknown, "irreversible effect requires fresh authoritative evidence", used, ids, provenance, Unknown, true}
	}
	if len(c.Evidence) == 0 {
		return Result{c.Name, Unknown, "no evidence available", used, ids, provenance, Unknown, true}
	}

	return Result{c.Name, Confirmed, "evidence satisfies the minimum verification floor: " + floor, used, ids, provenance, Confirmed, true}
}

func cases() []Case {
	return []Case{
		{Name: "internal-reversible-runtime-record", Effect: InternalEffect, Reversibility: Reversible, HasPolicy: true, Evidence: []Evidence{{ID: "ev-001", Source: "naeos-runtime", Type: "execution_record", Status: Observed, Authoritative: true}}},
		{Name: "external-reversible-runtime-only", Effect: ExternalEffect, Reversibility: Reversible, HasPolicy: true, Evidence: []Evidence{{ID: "ev-002", Source: "naeos-runtime", Type: "execution_record", Status: Observed, Authoritative: true}}},
		{Name: "external-reversible-provider-receipt", Effect: ExternalEffect, Reversibility: Reversible, HasPolicy: true, Evidence: []Evidence{{ID: "ev-003", Source: "provider", Type: "provider_receipt", Status: Observed, Authoritative: true}}},
		{Name: "internal-irreversible-estimated", Effect: InternalEffect, Reversibility: Irreversible, HasPolicy: true, Evidence: []Evidence{{ID: "ev-004", Source: "runtime", Type: "estimated_state", Status: Estimated, Authoritative: false}}},
		{Name: "external-irreversible-authoritative", Effect: ExternalEffect, Reversibility: Irreversible, HasPolicy: true, Evidence: []Evidence{{ID: "ev-005", Source: "provider", Type: "authoritative_state", Status: Observed, Authoritative: true}}},
		{Name: "external-stale", Effect: ExternalEffect, Reversibility: Reversible, HasPolicy: true, Evidence: []Evidence{{ID: "ev-006", Source: "provider", Type: "provider_receipt", Status: Stale, Authoritative: true}}},
		{Name: "external-contradictory", Effect: ExternalEffect, Reversibility: Reversible, HasPolicy: true, Evidence: []Evidence{{ID: "ev-007", Source: "provider", Type: "provider_receipt", Status: Observed, Authoritative: true}, {ID: "ev-009", Source: "observer", Type: "state_observation", Status: Contradictory, Authoritative: true}}},
		{Name: "missing-policy", Effect: ExternalEffect, Reversibility: Irreversible, HasPolicy: false, Evidence: []Evidence{{ID: "ev-008", Source: "provider", Type: "provider_receipt", Status: Observed, Authoritative: true}}},
	}
}

func run() ([]Result, error) {
	var results []Result
	for _, c := range cases() {
		r := verify(c)
		results = append(results, r)
		if !r.Passed {
			return results, fmt.Errorf("experiment assertion failed for %s", c.Name)
		}
	}
	return results, nil
}

func main() {
	results, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(map[string]any{
		"experiment": "evidence-verification-boundary-matrix-v1",
		"thesis":     "execution, evidence, verification, and authority are distinct; confirmation requires evidence satisfying the effect-derived floor",
		"results":    results,
	})
}

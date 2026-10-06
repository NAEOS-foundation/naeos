// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/verification"
)

type tamperCase struct {
	Name   string
	Mutate func(evidence.EvidenceRecord) evidence.EvidenceRecord
	Reason string
}

type result struct {
	Case       string `json:"case"`
	Mutation   string `json:"mutation"`
	Detected   bool   `json:"detected"`
	Expected   string `json:"expected"`
	Observed   string `json:"observed"`
	HashBefore string `json:"hash_before"`
	HashAfter  string `json:"hash_after"`
}

func baseline() (evidence.EvidenceRecord, *evidence.EvidenceStore) {
	store := evidence.NewStore()
	rec, _ := store.Append(evidence.EvidenceRecord{
		ID: "tamper-v1-001", Actor: "ai-agent/benchmark", Resource: "filesystem",
		Action: "write", Environment: "benchmark", PolicyID: "ai-agent-boundary",
		PolicyVersion: "1.0.0", Decision: control.DecisionAllow,
		DecisionReasons: []string{"policy explicitly allows bounded filesystem write"},
		ArtifactName:    "agent-action.txt",
		ArtifactHash:    evidence.ComputeArtifactHash([]byte("authorized agent change\n")),
		ArtifactSize:    len([]byte("authorized agent change\n")),
		ExecutionStatus: "completed", ExecutionOutput: "wrote agent-action.txt",
	})
	return rec, store
}

func verifyRecord(rec evidence.EvidenceRecord, store *evidence.EvidenceStore) (verification.VerificationResult, error) {
	return verification.NewEvidenceChainVerifier(store).Verify(rec)
}

func run() ([]result, error) {
	original, store := baseline()
	if index, err := store.Verify(); err != nil || index != -1 {
		return nil, fmt.Errorf("baseline evidence chain is not intact: index=%d err=%w", index, err)
	}
	initial, err := verifyRecord(original, store)
	if err != nil {
		return nil, err
	}
	if initial.Status != verification.StatusVerified {
		return nil, fmt.Errorf("baseline evidence unexpectedly failed verification: %s", initial.Status)
	}

	cases := []tamperCase{
		{Name: "execution-output", Reason: "mutating a semantic evidence field must invalidate the stored hash",
			Mutate: func(r evidence.EvidenceRecord) evidence.EvidenceRecord {
				r.ExecutionOutput = "wrote attacker-controlled-output.txt"
				return r
			}},
		{Name: "authorization-decision", Reason: "mutating the recorded policy decision must invalidate the stored hash",
			Mutate: func(r evidence.EvidenceRecord) evidence.EvidenceRecord {
				r.Decision = control.DecisionDeny
				return r
			}},
		{Name: "stored-hash", Reason: "mutating the stored hash must fail recomputation",
			Mutate: func(r evidence.EvidenceRecord) evidence.EvidenceRecord {
				r.Hash = "0000000000000000000000000000000000000000000000000000000000000000"
				return r
			}},
	}

	results := make([]result, 0, len(cases))
	for _, tc := range cases {
		tampered := tc.Mutate(original)
		verified, err := verifyRecord(tampered, store)
		if err != nil {
			return nil, fmt.Errorf("%s verifier error: %w", tc.Name, err)
		}
		detected := verified.Status == verification.StatusFailed
		results = append(results, result{
			Case: tc.Name, Mutation: tc.Reason, Detected: detected,
			Expected: string(verification.StatusFailed), Observed: string(verified.Status),
			HashBefore: original.Hash, HashAfter: evidence.RecomputeHash(tampered),
		})
		if !detected {
			return results, fmt.Errorf("tampering case %s was accepted", tc.Name)
		}
	}

	if index, err := store.Verify(); err != nil || index != -1 {
		return results, fmt.Errorf("authoritative store changed during tampering checks: index=%d err=%w", index, err)
	}
	return results, nil
}

func main() {
	results, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "EVIDENCE TAMPERING BENCHMARK: FAIL")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(map[string]any{
		"benchmark": "NAEOS Evidence Tampering Benchmark v1",
		"thesis":    "an evidence record modified after append must not verify against its original integrity hash",
		"baseline":  "untampered evidence verifies and the authoritative chain remains intact",
		"results":   results,
		"summary":   fmt.Sprintf("%d/%d tampering assertions detected", len(results), len(results)),
	})
}

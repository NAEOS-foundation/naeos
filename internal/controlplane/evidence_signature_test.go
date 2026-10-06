// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import "testing"

func TestVerifyEvidenceRejectsRehashedForgery(t *testing.T) {
	bundle := EvidenceBundle{
		SchemaVersion: "1.1",
		RequestID:     "req-1",
		DecisionID:    "dec-1",
		AgentID:       "agent-1",
		DecisionEvent: LedgerEvent{
			RequestID:  "req-1",
			DecisionID: "dec-1",
			AgentID:    "agent-1",
		},
		Verification: EvidenceVerification{
			Result:              "PASS",
			DecisionConsistent:  true,
			ExecutionConsistent: true,
			LedgerIntegrity:     true,
		},
	}
	digest, err := evidenceDigest(bundle)
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	bundle.EvidenceDigest = digest
	if err := signEvidenceBundle(&bundle); err != nil {
		t.Fatalf("sign: %v", err)
	}

	if result := VerifyEvidence(bundle); result.Result != "PASS" {
		t.Fatalf("expected signed bundle to verify, got %s: %v", result.Result, result.Issues)
	}

	// An attacker can recompute the public SHA-256 digest after changing a
	// field. The signature must still fail because the attacker cannot produce
	// a valid Ed25519 signature for the new digest.
	bundle.PolicyID = "forged-policy"
	forgedDigest, err := evidenceDigest(bundle)
	if err != nil {
		t.Fatalf("forged digest: %v", err)
	}
	bundle.EvidenceDigest = forgedDigest

	result := VerifyEvidence(bundle)
	if result.Result != "FAIL" {
		t.Fatal("expected rehashed forgery to fail verification")
	}
	found := false
	for _, issue := range result.Issues {
		if issue == "evidence signature invalid or missing" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected signature failure, got issues: %v", result.Issues)
	}
}

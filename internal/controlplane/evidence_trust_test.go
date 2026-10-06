// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"
)

func TestVerifyEvidenceWithTrustedKeyRejectsUnknownSigner(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := NewLedgerWithEvidenceSigner(private)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_760_000_000, 0).UTC()
	ledger.Append(LedgerEvent{Timestamp: now, RequestID: "REQ-TRUST", DecisionID: "DEC-TRUST", AgentID: "agent", Capability: "repository.write", ArtifactHash: "sha256:artifact", EventType: "AUTHORIZATION_DECISION", Decision: DecisionAllow, Reason: ReasonAllowed, Metadata: map[string]string{"policy_id": "p", "policy_version": "1", "grant_id": "g"}})
	ledger.Append(LedgerEvent{Timestamp: now, RequestID: "REQ-TRUST", DecisionID: "DEC-TRUST", ExecutionID: "EXEC-TRUST", AgentID: "agent", Capability: "repository.write", ArtifactHash: "sha256:artifact", EventType: "EXECUTION_ALLOWED", Decision: DecisionAllow, Reason: ReasonAllowed})
	ledger.Append(LedgerEvent{Timestamp: now, RequestID: "REQ-TRUST", DecisionID: "DEC-TRUST", ExecutionID: "EXEC-TRUST", AgentID: "agent", Capability: "repository.write", ArtifactHash: "sha256:artifact", EventType: "SIDE_EFFECT_OBSERVED", Decision: DecisionAllow, Reason: ReasonAllowed})
	bundle, err := ledger.BuildEvidence("DEC-TRUST")
	if err != nil {
		t.Fatal(err)
	}
	if verification := VerifyEvidenceWithTrustedKey(bundle, public); verification.Result != "PASS" {
		t.Fatalf("trusted key should pass: %+v", verification)
	}
	attackerPublic, _, _ := ed25519.GenerateKey(rand.Reader)
	if verification := VerifyEvidenceWithTrustedKey(bundle, attackerPublic); verification.Result != "FAIL" {
		t.Fatalf("unknown signer must fail: %+v", verification)
	}
}

func TestEvidenceSignerRequiresValidKey(t *testing.T) {
	if _, err := NewLedgerWithEvidenceSigner(make(ed25519.PrivateKey, ed25519.PrivateKeySize-1)); err == nil {
		t.Fatal("expected invalid signing key to be rejected")
	}
}

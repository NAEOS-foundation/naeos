// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

func TestEvidenceVerifyBundleCommandAcceptsValidBundle(t *testing.T) {
	bundle := testEvidenceBundle()
	bundle.EvidenceDigest = testEvidenceDigest(t, bundle)
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signature := ed25519.Sign(private, []byte(bundle.EvidenceDigest))
	bundle.EvidenceSignature = base64.RawStdEncoding.EncodeToString(signature)
	bundle.EvidencePublicKey = base64.RawStdEncoding.EncodeToString(public)
	bundle.EvidenceSignatureAlgorithm = "Ed25519"

	path := filepath.Join(t.TempDir(), "evidence.json")
	writeTestEvidenceBundle(t, path, bundle)

	cmd := newEvidenceVerifyBundleCommand()
	cmd.SetArgs([]string{"--input-file", path})
	var out strings.Builder
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected valid bundle, got %v", err)
	}
	if !strings.Contains(out.String(), "Evidence verification: PASS") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestEvidenceVerifyBundleCommandRejectsTampering(t *testing.T) {
	bundle := testEvidenceBundle()
	bundle.EvidenceDigest = testEvidenceDigest(t, bundle)
	bundle.ArtifactHash = "sha256:tampered"

	path := filepath.Join(t.TempDir(), "evidence.json")
	writeTestEvidenceBundle(t, path, bundle)

	cmd := newEvidenceVerifyBundleCommand()
	cmd.SetArgs([]string{"--input-file", path})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected tampered bundle to fail verification")
	}
}

func testEvidenceBundle() controlplane.EvidenceBundle {
	return controlplane.EvidenceBundle{
		SchemaVersion: "1.0",
		RequestID:     "request-1",
		DecisionID:    "decision-1",
		ExecutionID:   "execution-1",
		AgentID:       "agent-1",
		Capability:    controlplane.Capability("repository.read"),
		ArtifactHash:  "sha256:artifact",
		Decision:      controlplane.DecisionAllow,
		Reason:        controlplane.DecisionReason("policy_allow"),
		DecisionEvent: controlplane.LedgerEvent{
			ID:           "event-decision-1",
			RequestID:    "request-1",
			DecisionID:   "decision-1",
			AgentID:      "agent-1",
			Capability:   controlplane.Capability("repository.read"),
			ArtifactHash: "sha256:artifact",
		},
		ExecutionEvent: &controlplane.LedgerEvent{
			ID:           "event-execution-1",
			RequestID:    "request-1",
			DecisionID:   "decision-1",
			ExecutionID:  "execution-1",
			AgentID:      "agent-1",
			Capability:   controlplane.Capability("repository.read"),
			ArtifactHash: "sha256:artifact",
			EventType:    "EXECUTION_ALLOWED",
		},
		Verification: controlplane.EvidenceVerification{
			Result:              "PASS",
			DecisionConsistent:  true,
			ExecutionConsistent: true,
			LedgerIntegrity:     true,
		},
	}
}

func writeTestEvidenceBundle(t *testing.T, path string, bundle controlplane.EvidenceBundle) {
	t.Helper()
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func testEvidenceDigest(t *testing.T, bundle controlplane.EvidenceBundle) string {
	t.Helper()
	bundle.EvidenceDigest = ""
	bundle.Verification.Issues = nil
	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

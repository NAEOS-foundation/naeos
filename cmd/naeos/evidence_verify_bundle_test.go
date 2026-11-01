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

	dir := t.TempDir()
	path := filepath.Join(dir, "evidence.json")
	trustedPath := filepath.Join(dir, "trusted.pub")
	writeTestEvidenceBundle(t, path, bundle)
	writeTrustedEvidenceKey(t, trustedPath, public)

	cmd := newEvidenceVerifyBundleCommand()
	cmd.SetArgs([]string{"--input-file", path, "--trusted-public-key", trustedPath})
	var out strings.Builder
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected valid bundle, got %v", err)
	}
	for _, expected := range []string{"Evidence verification: PASS", "Digest integrity:      true", "Signature valid:       true", "Signer trusted:        true"} {
		if !strings.Contains(out.String(), expected) {
			t.Fatalf("output missing %q: %s", expected, out.String())
		}
	}
}

func TestEvidenceVerifyBundleCommandRejectsTampering(t *testing.T) {
	bundle, public := signedTestEvidenceBundle(t)
	bundle.ArtifactHash = "sha256:tampered"

	dir := t.TempDir()
	path := filepath.Join(dir, "evidence.json")
	trustedPath := filepath.Join(dir, "trusted.pub")
	writeTestEvidenceBundle(t, path, bundle)
	writeTrustedEvidenceKey(t, trustedPath, public)

	cmd := newEvidenceVerifyBundleCommand()
	cmd.SetArgs([]string{"--input-file", path, "--trusted-public-key", trustedPath})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected tampered bundle to fail verification")
	}
}

func TestEvidenceVerifyBundleCommandRequiresTrustedKey(t *testing.T) {
	bundle, _ := signedTestEvidenceBundle(t)
	path := filepath.Join(t.TempDir(), "evidence.json")
	writeTestEvidenceBundle(t, path, bundle)

	cmd := newEvidenceVerifyBundleCommand()
	cmd.SetArgs([]string{"--input-file", path, "--output", "json"})
	var out strings.Builder
	cmd.SetOut(&out)
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected missing trust anchor to fail verification")
	}
	if !strings.Contains(out.String(), "trusted evidence signing key is missing or invalid") {
		t.Fatalf("missing distinct trust-anchor failure: %s", out.String())
	}
}

func TestEvidenceVerifyBundleCommandRejectsUnknownTrustedKey(t *testing.T) {
	bundle, _ := signedTestEvidenceBundle(t)
	unknown, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "evidence.json")
	trustedPath := filepath.Join(dir, "unknown.pub")
	writeTestEvidenceBundle(t, path, bundle)
	writeTrustedEvidenceKey(t, trustedPath, unknown)

	cmd := newEvidenceVerifyBundleCommand()
	cmd.SetArgs([]string{"--input-file", path, "--trusted-public-key", trustedPath})
	var out strings.Builder
	cmd.SetOut(&out)
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected unknown trust anchor to fail verification")
	}
	if !strings.Contains(out.String(), "evidence signer is not the configured trust anchor") {
		t.Fatalf("missing signer-trust failure: %s", out.String())
	}
}

func TestEvidenceVerifyBundleCommandRejectsMalformedTrustedKey(t *testing.T) {
	bundle, _ := signedTestEvidenceBundle(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "evidence.json")
	trustedPath := filepath.Join(dir, "malformed.pub")
	writeTestEvidenceBundle(t, path, bundle)
	if err := os.WriteFile(trustedPath, []byte("not-a-public-key"), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := newEvidenceVerifyBundleCommand()
	cmd.SetArgs([]string{"--input-file", path, "--trusted-public-key", trustedPath})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "parse trusted evidence public key") {
		t.Fatalf("expected malformed trust key to fail parsing, got %v", err)
	}
}

func signedTestEvidenceBundle(t *testing.T) (controlplane.EvidenceBundle, ed25519.PublicKey) {
	t.Helper()
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
	return bundle, public
}

func writeTrustedEvidenceKey(t *testing.T, path string, key ed25519.PublicKey) {
	t.Helper()
	if err := os.WriteFile(path, key, 0o600); err != nil {
		t.Fatal(err)
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
		ObservationEvent: &controlplane.LedgerEvent{
			ID:           "event-observation-1",
			RequestID:    "request-1",
			DecisionID:   "decision-1",
			ExecutionID:  "execution-1",
			AgentID:      "agent-1",
			Capability:   controlplane.Capability("repository.read"),
			ArtifactHash: "sha256:artifact",
			EventType:    "SIDE_EFFECT_OBSERVED",
		},
		Verification: controlplane.EvidenceVerification{
			Result:                "PASS",
			DecisionConsistent:    true,
			ExecutionConsistent:   true,
			ObservationConsistent: true,
			LedgerIntegrity:       true,
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

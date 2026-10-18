// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

func TestRuntimeEvidenceBindsExactInvocation(t *testing.T) {
	req := ToolRequest{
		RequestID: "req-evidence-01", InvocationID: "inv-evidence-01",
		Capability: "filesystem.write", Tool: "filesystem", Action: "write",
		Resource: "filesystem", Actor: "manus",
		Payload: map[string]any{"path": "a.txt", "content": "one"},
		Context: map[string]any{"policy_context": "v1"},
	}
	digest := InvocationDigest(req)
	candidateDigest, err := CandidateDigestV1(req)
	if err != nil {
		t.Fatalf("candidate digest: %v", err)
	}
	mutated := req
	mutated.Payload = map[string]any{"path": "a.txt", "content": "two"}
	if digest == InvocationDigest(mutated) {
		t.Fatal("payload mutation must change invocation digest")
	}

	obs := Observation{
		RequestID: req.RequestID, InvocationID: req.InvocationID,
		InvocationDigest: digest, Status: "observed", Observed: true,
		ArtifactHash: "sha256:artifact",
	}
	result := ExecutionResult{
		RequestID: req.RequestID, InvocationID: req.InvocationID, Request: req,
		Decision: control.DecisionAllow, PolicyID: "policy", PolicyVersion: "1.0",
		CandidateDigest: candidateDigest, CandidateDigestProfile: CandidateDigestProfileV1,
		Status: "completed", Hash: "sha256:output", Observation: &obs,
	}
	evidence, err := BuildRuntimeEvidence(result)
	if err != nil {
		t.Fatalf("build evidence: %v", err)
	}
	if evidence.InvocationDigest != digest {
		t.Fatalf("invocation digest = %q, want %q", evidence.InvocationDigest, digest)
	}
	if err := VerifyRuntimeEvidence(evidence); err != nil {
		t.Fatalf("verify evidence: %v", err)
	}

	evidence.Observation.ArtifactHash = "sha256:tampered"
	if err := VerifyRuntimeEvidence(evidence); err == nil {
		t.Fatal("tampered observation must invalidate canonical evidence")
	}
}

func TestRuntimeEvidenceRejectsUnobservedAllow(t *testing.T) {
	req := ToolRequest{RequestID: "req-evidence-02", InvocationID: "inv-evidence-02", Tool: "filesystem", Action: "write"}
	candidateDigest, err := CandidateDigestV1(req)
	if err != nil {
		t.Fatalf("candidate digest: %v", err)
	}
	result := ExecutionResult{
		RequestID: req.RequestID, InvocationID: req.InvocationID, Request: req,
		Decision: control.DecisionAllow, CandidateDigest: candidateDigest,
		CandidateDigestProfile: CandidateDigestProfileV1, Status: "completed",
	}
	if _, err := BuildRuntimeEvidence(result); err == nil {
		t.Fatal("ALLOW completion without independent observation must not become evidence")
	}
}

func TestRuntimeEvidenceSignatureRejectsDigestRecomputationAttack(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	request := ToolRequest{
		RequestID: "req-sign-01", InvocationID: "inv-sign-01",
		Tool: "filesystem", Action: "write",
		Payload: map[string]any{"path": "signed.txt", "content": "one"},
	}
	digest := InvocationDigest(request)
	result := ExecutionResult{
		RequestID: request.RequestID, InvocationID: request.InvocationID, Request: request,
		Decision: control.DecisionAllow, PolicyID: "policy", PolicyVersion: "1.0",
		Status: "completed", Observation: &Observation{
			RequestID: request.RequestID, InvocationID: request.InvocationID,
			InvocationDigest: digest, Status: "observed", Observed: true,
		},
	}
	evidence, err := BuildRuntimeEvidence(result)
	if err != nil {
		t.Fatalf("build evidence: %v", err)
	}
	if err := SignRuntimeEvidence(&evidence, privateKey, "naeos-runtime", "key-1"); err != nil {
		t.Fatalf("sign evidence: %v", err)
	}
	publicKey := privateKey.Public().(ed25519.PublicKey)
	if err := VerifyRuntimeEvidenceWithPublicKey(evidence, publicKey, "naeos-runtime", "key-1"); err != nil {
		t.Fatalf("verify signed evidence: %v", err)
	}

	evidence.Observation.Timestamp = time.Now()
	evidence = sealEvidenceDigest(evidence)
	if err := VerifyRuntimeEvidenceWithPublicKey(evidence, publicKey, "naeos-runtime", "key-1"); err == nil {
		t.Fatal("recomputed digest after tampering must fail signature verification")
	}
}

func TestRuntimeEvidenceRejectsUnsignedReceiptWithTrustedKey(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	evidence := RuntimeEvidence{
		SchemaVersion:    "naeos.runtime-evidence.v1",
		RequestID:        "req-sign-02",
		InvocationID:     "inv-sign-02",
		InvocationDigest: "sha256:inv", Tool: "filesystem", Action: "write",
		Decision: "DENY", ExecutionStatus: "denied",
	}
	evidence = sealEvidenceDigest(evidence)
	if err := VerifyRuntimeEvidenceWithPublicKey(evidence, publicKey, "naeos-runtime", "key-1"); err == nil {
		t.Fatal("unsigned evidence must fail trusted verification")
	}
}

func TestParseEd25519PrivateKeyPreservesRawSeedWhitespaceBytes(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	seed[0] = '\n'
	seed[len(seed)-1] = '\t'

	got, err := ParseEd25519PrivateKey(seed)
	if err != nil {
		t.Fatalf("parse raw seed: %v", err)
	}
	want := ed25519.NewKeyFromSeed(seed)
	if !bytes.Equal(got, want) {
		t.Fatal("parsed raw seed does not match original key material")
	}
}

func TestRuntimeEvidenceRejectsCandidateDigestMismatch(t *testing.T) {
	req := ToolRequest{
		RequestID: "req-evidence-03", InvocationID: "inv-evidence-03",
		Tool: "filesystem", Action: "write", Payload: map[string]any{"path": "a.txt"},
	}
	result := ExecutionResult{
		RequestID: req.RequestID, InvocationID: req.InvocationID, Request: req,
		Decision: control.DecisionDeny, CandidateDigest: strings.Repeat("0", 64),
		CandidateDigestProfile: CandidateDigestProfileV1, Status: "denied",
	}
	if _, err := BuildRuntimeEvidence(result); err == nil {
		t.Fatal("evidence must reject a digest that does not match the request")
	}
}

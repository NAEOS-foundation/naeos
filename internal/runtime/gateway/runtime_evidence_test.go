// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"testing"

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
	result := ExecutionResult{
		RequestID: req.RequestID, InvocationID: req.InvocationID, Request: req,
		Decision: control.DecisionAllow, Status: "completed",
	}
	if _, err := BuildRuntimeEvidence(result); err == nil {
		t.Fatal("ALLOW completion without independent observation must not become evidence")
	}
}

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/runtime/gateway"
)

type bridgeAllowControlPlane struct{}

func (bridgeAllowControlPlane) Evaluate(req control.Request) (control.DecisionRecord, error) {
	return control.DecisionRecord{Request: req, Decision: control.DecisionAllow, PolicyID: "bridge-test", RuleID: "allow"}, nil
}

func (bridgeAllowControlPlane) ValidateDecision(_ control.Request, issued control.DecisionRecord) (control.DecisionRecord, error) {
	return issued, nil
}

func TestRuntimeBridgeGatewayRejectsReplayBeforeSecondWrite(t *testing.T) {
	root := t.TempDir()
	gw := newRuntimeBridgeGateway(bridgeAllowControlPlane{}, root, gateway.NewInMemoryInvocationStore())

	request := map[string]any{
		"request_id":    "req-bridge-replay",
		"invocation_id": "inv-bridge-replay",
		"actor":         "manus",
		"tool":          "filesystem",
		"action":        "write",
		"resource":      "filesystem",
		"payload": map[string]any{
			"path":    "replay.txt",
			"content": "one write",
		},
	}

	first, err := gw.AuthorizeFromAdapter("json", request)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	if first.Status != "completed" {
		t.Fatalf("first request status = %q, want completed: %+v", first.Status, first)
	}
	if first.Observation == nil {
		t.Fatal("first request missing observation")
	}
	if first.Observation.RequestID != request["request_id"] || first.Observation.InvocationID != request["invocation_id"] || first.Observation.InvocationDigest != gateway.InvocationDigest(first.Request) {
		t.Fatalf("observation identity = (%q, %q), want (%q, %q)",
			first.Observation.RequestID, first.Observation.InvocationID,
			request["request_id"], request["invocation_id"])
	}

	second, err := gw.AuthorizeFromAdapter("json", request)
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	if second.Status != "denied" || second.Output != "invocation already consumed" {
		t.Fatalf("second request = %+v, want replay denial", second)
	}

	data, err := os.ReadFile(filepath.Join(root, "replay.txt"))
	if err != nil {
		t.Fatalf("read governed artifact: %v", err)
	}
	if string(data) != "one write" {
		t.Fatalf("artifact content = %q, want one write", string(data))
	}
}

func TestRuntimeBridgeEvidenceRejectsMutatedInvocation(t *testing.T) {
	root := t.TempDir()
	gw := newRuntimeBridgeGateway(bridgeAllowControlPlane{}, root, gateway.NewInMemoryInvocationStore())
	request := map[string]any{
		"request_id":    "req-bridge-mutation",
		"invocation_id": "inv-bridge-mutation",
		"actor":         "manus",
		"tool":          "filesystem",
		"action":        "write",
		"resource":      "filesystem",
		"payload":       map[string]any{"path": "mutation.txt", "content": "authorized"},
	}
	result, err := gw.AuthorizeFromAdapter("json", request)
	if err != nil || result.Status != "completed" {
		t.Fatalf("initial execution failed: result=%+v err=%v", result, err)
	}
	evidence, err := gateway.BuildRuntimeEvidence(result)
	if err != nil {
		t.Fatalf("build evidence: %v", err)
	}
	mutated := result
	mutated.Request.Payload = map[string]any{"path": "mutation.txt", "content": "tampered"}
	if gateway.InvocationDigest(mutated.Request) == evidence.InvocationDigest {
		t.Fatal("mutated invocation must not retain the authorization digest")
	}
}

func TestRuntimeBridgeRejectsUnsupportedToolBeforeExecution(t *testing.T) {
	root := t.TempDir()
	gw := newRuntimeBridgeGateway(bridgeAllowControlPlane{}, root, gateway.NewInMemoryInvocationStore())
	request := map[string]any{
		"request_id":    "req-bridge-unsupported",
		"invocation_id": "inv-bridge-unsupported",
		"actor":         "manus",
		"tool":          "shell",
		"action":        "run",
		"resource":      "shell",
		"payload":       map[string]any{"command": "echo should-not-run"},
	}
	result, err := gw.AuthorizeFromAdapter("json", request)
	if err != nil {
		t.Fatalf("unsupported request should return a governed adapter denial: %v", err)
	}
	if result.Status != "denied" || result.Output != "adapter privilege boundary denied request" {
		t.Fatalf("unsupported tool was not denied at adapter boundary: %+v", result)
	}
	if result.Observation != nil {
		t.Fatalf("unsupported tool unexpectedly produced runtime observation: %+v", result.Observation)
	}
}

func TestFilesystemObserverDoesNotClaimUnsupportedActionObserved(t *testing.T) {
	observer := filesystemObserver{root: t.TempDir()}
	req := gateway.ToolRequest{
		RequestID:    "req-observer-unsupported",
		InvocationID: "inv-observer-unsupported",
		Tool:         "shell",
		Action:       "run",
	}
	obs, err := observer.Observe(req, gateway.ExecutionResult{Timestamp: time.Now().UTC()})
	if err != nil {
		t.Fatalf("unsupported observation should be reported as unavailable, not error: %v", err)
	}
	if obs.Status != "unavailable" || obs.Observed {
		t.Fatalf("unsupported action must not be represented as independently observed: %+v", obs)
	}
	if obs.RequestID != req.RequestID || obs.InvocationID != req.InvocationID || obs.InvocationDigest != gateway.InvocationDigest(req) {
		t.Fatalf("unsupported observation must preserve invocation identity: %+v", obs)
	}
}

func TestAppendRuntimeEvidenceSyncsJSONLRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "runtime-evidence.jsonl")
	if err := appendRuntimeEvidence(path, []byte(`{"evidence_digest":"sha256:test"}`)); err != nil {
		t.Fatalf("append runtime evidence: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read runtime evidence: %v", err)
	}
	if string(data) != "{\"evidence_digest\":\"sha256:test\"}\n" {
		t.Fatalf("runtime evidence record = %q", string(data))
	}
}

func TestWriteAtomicRuntimeReceiptPersistsCompleteRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "runtime-receipt.json")
	want := []byte("{\"receipt_id\":\"receipt-test\"}\n")
	if err := writeAtomicRuntimeReceipt(path, want); err != nil {
		t.Fatalf("write runtime receipt: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read runtime receipt: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("runtime receipt = %q, want %q", got, want)
	}
}

func TestRuntimeBridgeSideEffectPersistsWhenEvidencePersistenceFails(t *testing.T) {
	root := t.TempDir()
	gw := newRuntimeBridgeGateway(bridgeAllowControlPlane{}, root, gateway.NewInMemoryInvocationStore())
	request := map[string]any{
		"request_id":    "req-evidence-persist-failure",
		"invocation_id": "inv-evidence-persist-failure",
		"actor":         "test",
		"tool":          "filesystem",
		"action":        "write",
		"resource":      "filesystem",
		"payload": map[string]any{
			"path":    "persisted-before-evidence.txt",
			"content": "side effect completed",
		},
	}

	result, err := gw.AuthorizeFromAdapter("json", request)
	if err != nil {
		t.Fatalf("governed side effect: %v", err)
	}
	if result.Status != "completed" {
		t.Fatalf("side effect status = %q, want completed: %+v", result.Status, result)
	}

	runtimeEvidence, err := gateway.BuildRuntimeEvidence(result)
	if err != nil {
		t.Fatalf("build runtime evidence: %v", err)
	}
	evidenceData, err := json.Marshal(runtimeEvidence)
	if err != nil {
		t.Fatalf("marshal runtime evidence: %v", err)
	}

	// An existing directory at the evidence-file path deterministically makes
	// the post-side-effect append fail without relying on permissions or disk state.
	evidencePath := filepath.Join(t.TempDir(), "evidence-is-a-directory")
	if err := os.Mkdir(evidencePath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := appendRuntimeEvidence(evidencePath, evidenceData); err == nil {
		t.Fatal("expected evidence persistence to fail")
	}

	artifactPath := filepath.Join(root, "persisted-before-evidence.txt")
	artifact, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("side effect must remain after evidence persistence failure: %v", err)
	}
	if string(artifact) != "side effect completed" {
		t.Fatalf("persisted side effect = %q", artifact)
	}

	// Replay protection prevents a blind retry from executing the same
	// invocation again. This does not roll back the first side effect.
	retry, err := gw.AuthorizeFromAdapter("json", request)
	if err != nil {
		t.Fatalf("retry should return a governed denial: %v", err)
	}
	if retry.Status != "denied" || retry.Output != "invocation already consumed" {
		t.Fatalf("retry result = %+v, want replay denial", retry)
	}
}

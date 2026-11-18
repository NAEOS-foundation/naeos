// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/runtime/gateway"
)

func TestRuntimeBridgeDurableAllowLifecycleWithRealSideEffectAndRecovery(t *testing.T) {
	root := t.TempDir()
	dir := t.TempDir()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	request := map[string]any{
		"request_id":    "req-ws3-durable-allow",
		"invocation_id": "inv-ws3-durable-allow",
		"actor":         "test",
		"tool":          "filesystem",
		"action":        "write",
		"resource":      "filesystem",
		"payload":       map[string]any{"path": "allow-artifact.txt", "content": "authorized durable side effect"},
	}
	gw := newRuntimeBridgeGateway(bridgeAllowControlPlane{}, root, gateway.NewInMemoryInvocationStore())
	result, err := gw.AuthorizeFromAdapter("json", request)
	if err != nil || result.Status != "completed" {
		t.Fatalf("governed ALLOW execution: result=%+v err=%v", result, err)
	}
	artifactPath := filepath.Join(root, "allow-artifact.txt")
	artifact, err := os.ReadFile(artifactPath)
	if err != nil || string(artifact) != "authorized durable side effect" {
		t.Fatalf("independent side-effect readback: data=%q err=%v", artifact, err)
	}
	if result.Observation == nil || !result.Observation.Observed || result.Observation.Status != "observed" {
		t.Fatalf("ALLOW result lacks positive independent observation: %+v", result.Observation)
	}

	runtimeEvidence, err := gateway.BuildRuntimeEvidence(result)
	if err != nil {
		t.Fatal(err)
	}
	if err := gateway.SignRuntimeEvidence(&runtimeEvidence, privateKey, "naeos-runtime", "ws3-test-key"); err != nil {
		t.Fatal(err)
	}
	evidencePath := filepath.Join(dir, "runtime.jsonl")
	evidenceBytes, err := json.Marshal(runtimeEvidence)
	if err != nil {
		t.Fatal(err)
	}
	if err := appendRuntimeEvidence(evidencePath, evidenceBytes); err != nil {
		t.Fatal(err)
	}

	ledgerPath := filepath.Join(dir, "events.jsonl")
	ledger, err := evidence.NewDurableRuntimeEventLedger(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ledger.Publish("run-ws3-allow", "runtime.execution_evidence", runtimeEvidence.EvidenceDigest, 1); err != nil {
		t.Fatal(err)
	}
	ledger.Seal()
	receipt, err := evidence.CreateDurableRuntimeReceipt(ledger, "run-ws3-allow", time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	signedReceipt, err := evidence.SignDurableRuntimeReceipt(receipt, privateKey, "naeos-runtime", "ws3-test-key")
	if err != nil {
		t.Fatal(err)
	}
	if err := evidence.VerifySignedDurableRuntimeReceipt(signedReceipt, ledger, "run-ws3-allow", publicKey, "naeos-runtime", "ws3-test-key"); err != nil {
		t.Fatalf("verify receipt with separately trusted key: %v", err)
	}
	receiptPath := filepath.Join(dir, "receipt.json")
	receiptBytes, err := json.Marshal(signedReceipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeAtomicRuntimeReceipt(receiptPath, append(receiptBytes, '\n')); err != nil {
		t.Fatal(err)
	}

	// Recovery is performed from fresh ledger and receipt objects, as after process restart.
	recoveredLedger, err := evidence.NewDurableRuntimeEventLedger(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	recoveredLedger.Seal()
	recoveredReceiptBytes, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	var recoveredReceipt evidence.SignedDurableRuntimeReceipt
	if err := json.Unmarshal(recoveredReceiptBytes, &recoveredReceipt); err != nil {
		t.Fatal(err)
	}
	if err := verifyDurableRuntimeEvidenceFile(evidencePath, recoveredLedger, "run-ws3-allow", publicKey, "naeos-runtime", "ws3-test-key"); err != nil {
		t.Fatalf("trusted evidence recovery: %v", err)
	}
	if err := evidence.VerifySignedDurableRuntimeReceipt(recoveredReceipt, recoveredLedger, "run-ws3-allow", publicKey, "naeos-runtime", "ws3-test-key"); err != nil {
		t.Fatalf("trusted receipt recovery: %v", err)
	}

	replay, err := gw.AuthorizeFromAdapter("json", request)
	if err != nil || replay.Status != "denied" || replay.Output != "invocation already consumed" {
		t.Fatalf("replay must be denied without second side effect: result=%+v err=%v", replay, err)
	}
	artifactAfterReplay, err := os.ReadFile(artifactPath)
	if err != nil || string(artifactAfterReplay) != "authorized durable side effect" {
		t.Fatalf("replay changed side effect: data=%q err=%v", artifactAfterReplay, err)
	}

	t.Run("tampered evidence", func(t *testing.T) {
		mutated := runtimeEvidence
		mutated.ExecutionStatus = "tampered"
		data, err := json.Marshal(mutated)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "runtime.jsonl")
		if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := verifyDurableRuntimeEvidenceFile(path, recoveredLedger, "run-ws3-allow", publicKey, "naeos-runtime", "ws3-test-key"); err == nil {
			t.Fatal("tampered signed evidence was accepted")
		}
	})
	t.Run("missing evidence", func(t *testing.T) {
		if err := verifyDurableRuntimeEvidenceFile(filepath.Join(t.TempDir(), "missing.jsonl"), recoveredLedger, "run-ws3-allow", publicKey, "naeos-runtime", "ws3-test-key"); err == nil {
			t.Fatal("missing evidence was accepted")
		}
	})
	t.Run("orphan evidence", func(t *testing.T) {
		data, err := os.ReadFile(evidencePath)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "runtime.jsonl")
		if err := os.WriteFile(path, append(data, data...), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := verifyDurableRuntimeEvidenceFile(path, recoveredLedger, "run-ws3-allow", publicKey, "naeos-runtime", "ws3-test-key"); err == nil {
			t.Fatal("orphan evidence record was accepted")
		}
	})
	t.Run("truncated evidence", func(t *testing.T) {
		data, err := os.ReadFile(evidencePath)
		if err != nil {
			t.Fatal(err)
		}
		data = bytes.TrimSuffix(data, []byte("\n"))
		path := filepath.Join(t.TempDir(), "runtime.jsonl")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := verifyDurableRuntimeEvidenceFile(path, recoveredLedger, "run-ws3-allow", publicKey, "naeos-runtime", "ws3-test-key"); err == nil {
			t.Fatal("truncated evidence was accepted")
		}
	})
	t.Run("invalid receipt signature", func(t *testing.T) {
		bad := recoveredReceipt
		bad.Signature = "invalid"
		if err := evidence.VerifySignedDurableRuntimeReceipt(bad, recoveredLedger, "run-ws3-allow", publicKey, "naeos-runtime", "ws3-test-key"); err == nil {
			t.Fatal("invalid receipt signature was accepted")
		}
	})
}

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
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
)

func TestRuntimeBridgeAuthenticatedDurableEvidenceLifecycle(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	privatePath := filepath.Join(home, "runtime-private.key")
	publicPath := filepath.Join(home, "runtime-public.key")
	if err := os.WriteFile(privatePath, privateKey, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(publicPath, publicKey, 0o600); err != nil {
		t.Fatal(err)
	}

	evidenceFile := filepath.Join(home, "evidence", "runtime.jsonl")
	ledgerFile := filepath.Join(home, "evidence", "events.jsonl")
	receiptFile := filepath.Join(home, "evidence", "receipt.json")
	request := `{"request_id":"req-durable-test","invocation_id":"inv-durable-test","actor":"test","tool":"shell","action":"run","resource":"shell","payload":{"command":"must-not-run"}}` + "\n"

	cmd := newRuntimeBridgeCommand()
	cmd.SetIn(strings.NewReader(request))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"--evidence-file", evidenceFile,
		"--evidence-ledger", ledgerFile,
		"--evidence-receipt", receiptFile,
		"--evidence-run-id", "run-durable-test",
		"--evidence-signing-key", privatePath,
		"--evidence-trusted-public-key", publicPath,
		"--evidence-issuer", "naeos-runtime",
		"--evidence-key-id", "test-key-1",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run authenticated durable bridge: %v", err)
	}

	evidenceData, err := os.ReadFile(evidenceFile)
	if err != nil {
		t.Fatalf("read evidence file: %v", err)
	}
	if !bytes.HasSuffix(evidenceData, []byte("\n")) {
		t.Fatal("evidence file must contain a complete synced JSONL record")
	}
	var runtimeEvidence map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(evidenceData), &runtimeEvidence); err != nil {
		t.Fatalf("decode runtime evidence: %v", err)
	}
	if runtimeEvidence["signature"] == nil {
		t.Fatal("authenticated durable mode must sign each runtime evidence record")
	}

	ledger, err := evidence.NewDurableRuntimeEventLedger(ledgerFile)
	if err != nil {
		t.Fatalf("reload durable ledger: %v", err)
	}
	ledger.Seal()
	var signed evidence.SignedDurableRuntimeReceipt
	receiptData, err := os.ReadFile(receiptFile)
	if err != nil {
		t.Fatalf("read signed receipt: %v", err)
	}
	if err := json.Unmarshal(receiptData, &signed); err != nil {
		t.Fatalf("decode signed receipt: %v", err)
	}
	if err := evidence.VerifySignedDurableRuntimeReceipt(signed, ledger, "run-durable-test", publicKey, "naeos-runtime", "test-key-1"); err != nil {
		t.Fatalf("verify durable receipt after restart: %v", err)
	}
	if len(ledger.Records()) != 1 {
		t.Fatalf("durable ledger event count = %d, want 1", len(ledger.Records()))
	}
}

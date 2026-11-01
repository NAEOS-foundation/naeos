// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"crypto/ed25519"
	"crypto/rand"
	"path/filepath"
	"testing"
	"time"
)

func TestSignedDurableRuntimeReceiptRequiresTrustedSigner(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := NewDurableRuntimeEventLedger(filepath.Join(t.TempDir(), "events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ledger.Publish("run-signed", "runtime.evidence", "sha256:evidence", 1); err != nil {
		t.Fatal(err)
	}
	ledger.Seal()
	receipt, err := CreateDurableRuntimeReceipt(ledger, "run-signed", time.Unix(10, 0))
	if err != nil {
		t.Fatal(err)
	}
	signed, err := SignDurableRuntimeReceipt(receipt, privateKey, "naeos-runtime", "key-2026-01")
	if err != nil {
		t.Fatal(err)
	}
	publicKey := privateKey.Public().(ed25519.PublicKey)
	if err := VerifySignedDurableRuntimeReceipt(signed, ledger, "run-signed", publicKey, "naeos-runtime", "key-2026-01"); err != nil {
		t.Fatalf("trusted receipt should verify: %v", err)
	}
	if err := VerifySignedDurableRuntimeReceipt(signed, ledger, "run-signed", publicKey, "other-issuer", "key-2026-01"); err == nil {
		t.Fatal("unexpected issuer must be rejected")
	}
	attackerPublic, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifySignedDurableRuntimeReceipt(signed, ledger, "run-signed", attackerPublic, "naeos-runtime", "key-2026-01"); err == nil {
		t.Fatal("untrusted key must be rejected")
	}
	signed.Receipt.EventCount++
	if err := VerifySignedDurableRuntimeReceipt(signed, ledger, "run-signed", publicKey, "naeos-runtime", "key-2026-01"); err == nil {
		t.Fatal("mutated receipt must be rejected")
	}
}

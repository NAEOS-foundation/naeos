// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDurableRuntimeReceiptRoundTripAndRecovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	receiptPath := filepath.Join(t.TempDir(), "runtime-receipt.json")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil { t.Fatal(err) }
	if _, err := ledger.Publish("run-1", "pipeline.start", "payload-1", 1); err != nil { t.Fatal(err) }
	if _, err := ledger.Publish("run-1", "pipeline.policy_decision", "payload-2", 2); err != nil { t.Fatal(err) }
	ledger.Seal()

	receipt, err := CreateDurableRuntimeReceipt(ledger, "run-1", time.Unix(1, 0))
	if err != nil { t.Fatal(err) }
	if err := WriteDurableRuntimeReceipt(receiptPath, receipt); err != nil { t.Fatal(err) }

	reloaded, err := NewDurableRuntimeEventLedger(path)
	if err != nil { t.Fatal(err) }
	loadedReceipt, err := LoadDurableRuntimeReceipt(receiptPath)
	if err != nil { t.Fatal(err) }
	if err := VerifyDurableRuntimeReceipt(loadedReceipt, reloaded, "run-1"); err != nil { t.Fatal(err) }
}

func TestDurableRuntimeReceiptRejectsLedgerMutation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil { t.Fatal(err) }
	if _, err := ledger.Publish("run-1", "pipeline.start", "payload-1", 1); err != nil { t.Fatal(err) }
	ledger.Seal()
	receipt, err := CreateDurableRuntimeReceipt(ledger, "run-1", time.Unix(1, 0))
	if err != nil { t.Fatal(err) }
	if err := ledger.Publish("run-1", "pipeline.policy_decision", "payload-2", 2); err == nil {
		t.Fatal("expected sealed ledger to reject mutation")
	}
	if err := VerifyDurableRuntimeReceipt(receipt, ledger, "run-1"); err != nil { t.Fatal(err) }
}

func TestDurableRuntimeReceiptRejectsForeignRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil { t.Fatal(err) }
	if _, err := ledger.Publish("run-1", "pipeline.start", "payload-1", 1); err != nil { t.Fatal(err) }
	ledger.Seal()
	receipt, err := CreateDurableRuntimeReceipt(ledger, "run-1", time.Unix(1, 0))
	if err != nil { t.Fatal(err) }
	if err := VerifyDurableRuntimeReceipt(receipt, ledger, "run-2"); err == nil {
		t.Fatal("expected foreign run to be rejected")
	}
}

func TestDurableRuntimeReceiptRejectsTamperedReceipt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil { t.Fatal(err) }
	if _, err := ledger.Publish("run-1", "pipeline.start", "payload-1", 1); err != nil { t.Fatal(err) }
	ledger.Seal()
	receipt, err := CreateDurableRuntimeReceipt(ledger, "run-1", time.Unix(1, 0))
	if err != nil { t.Fatal(err) }
	receipt.EventCount++
	if err := VerifyDurableRuntimeReceipt(receipt, ledger, "run-1"); err == nil {
		t.Fatal("expected tampered receipt to be rejected")
	}
}

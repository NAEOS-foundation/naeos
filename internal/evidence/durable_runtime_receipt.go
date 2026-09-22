// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const DurableRuntimeReceiptVersion = "naeos:durable-runtime-receipt:v1"

// DurableRuntimeReceipt is the durable completion receipt for one sealed
// runtime ledger. It binds the run identity, event count, terminal event,
// and canonical ledger digest into a restart-verifiable record.
type DurableRuntimeReceipt struct {
	Version      string    `json:"version"`
	RunID        string    `json:"run_id"`
	EventCount   int       `json:"event_count"`
	FirstEventID string    `json:"first_event_id"`
	LastEventID  string    `json:"last_event_id"`
	LedgerDigest string    `json:"ledger_digest"`
	SealedAt     time.Time `json:"sealed_at"`
	ReceiptID    string    `json:"receipt_id"`
}

// CreateDurableRuntimeReceipt creates a receipt only from a ledger that is
// already sealed. The receipt is deterministic except for the seal timestamp.
func CreateDurableRuntimeReceipt(ledger *DurableRuntimeEventLedger, runID string, sealedAt time.Time) (DurableRuntimeReceipt, error) {
	if ledger == nil {
		return DurableRuntimeReceipt{}, fmt.Errorf("durable runtime ledger is nil")
	}
	if runID == "" {
		return DurableRuntimeReceipt{}, fmt.Errorf("run identity is required")
	}
	events := ledger.Records()
	if len(events) == 0 {
		return DurableRuntimeReceipt{}, fmt.Errorf("cannot receipt an empty runtime ledger")
	}
	if err := ledger.Verify(); err != nil {
		return DurableRuntimeReceipt{}, fmt.Errorf("verify ledger before receipt: %w", err)
	}
	if events[0].RunID != runID {
		return DurableRuntimeReceipt{}, fmt.Errorf("ledger run mismatch: expected %s got %s", runID, events[0].RunID)
	}
	for _, event := range events {
		if event.RunID != runID {
			return DurableRuntimeReceipt{}, fmt.Errorf("ledger contains foreign run %s", event.RunID)
		}
	}
	if sealedAt.IsZero() {
		sealedAt = time.Now().UTC()
	} else {
		sealedAt = sealedAt.UTC()
	}
	digest, err := durableLedgerDigest(events)
	if err != nil {
		return DurableRuntimeReceipt{}, err
	}
	receipt := DurableRuntimeReceipt{
		Version: DurableRuntimeReceiptVersion, RunID: runID, EventCount: len(events),
		FirstEventID: events[0].ID, LastEventID: events[len(events)-1].ID,
		LedgerDigest: digest, SealedAt: sealedAt,
	}
	receipt.ReceiptID = durableReceiptID(receipt)
	return receipt, nil
}

// VerifyDurableRuntimeReceipt verifies a receipt against the current sealed
// ledger. This is the recovery boundary used after process restart.
func VerifyDurableRuntimeReceipt(receipt DurableRuntimeReceipt, ledger *DurableRuntimeEventLedger, runID string) error {
	if ledger == nil {
		return fmt.Errorf("durable runtime ledger is nil")
	}
	if receipt.Version != DurableRuntimeReceiptVersion {
		return fmt.Errorf("unsupported durable runtime receipt version %q", receipt.Version)
	}
	if receipt.RunID != runID || runID == "" {
		return fmt.Errorf("durable runtime receipt run mismatch")
	}
	if receipt.ReceiptID != durableReceiptID(receiptWithoutID(receipt)) {
		return fmt.Errorf("durable runtime receipt identity mismatch")
	}
	events := ledger.Records()
	if len(events) != receipt.EventCount || len(events) == 0 {
		return fmt.Errorf("durable runtime receipt event count mismatch")
	}
	if events[0].ID != receipt.FirstEventID || events[len(events)-1].ID != receipt.LastEventID {
		return fmt.Errorf("durable runtime receipt event boundary mismatch")
	}
	if err := ledger.Verify(); err != nil {
		return fmt.Errorf("durable runtime ledger failed verification: %w", err)
	}
	digest, err := durableLedgerDigest(events)
	if err != nil {
		return err
	}
	if digest != receipt.LedgerDigest {
		return fmt.Errorf("durable runtime receipt ledger digest mismatch")
	}
	for _, event := range events {
		if event.RunID != runID {
			return fmt.Errorf("durable runtime receipt contains foreign run")
		}
	}
	return nil
}

// WriteDurableRuntimeReceipt atomically writes a receipt to a JSON file.
func WriteDurableRuntimeReceipt(path string, receipt DurableRuntimeReceipt) error {
	if path == "" {
		return fmt.Errorf("durable runtime receipt path is required")
	}
	if receipt.ReceiptID != durableReceiptID(receiptWithoutID(receipt)) {
		return fmt.Errorf("durable runtime receipt identity mismatch")
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return fmt.Errorf("encode durable runtime receipt: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Clean(path)), 0o750); err != nil {
		return fmt.Errorf("create durable runtime receipt directory: %w", err)
	}
	tmp := filepath.Clean(path) + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write durable runtime receipt: %w", err)
	}
	if err := os.Rename(tmp, filepath.Clean(path)); err != nil {
		return fmt.Errorf("commit durable runtime receipt: %w", err)
	}
	return nil
}

// LoadDurableRuntimeReceipt loads a receipt and verifies its identity.
func LoadDurableRuntimeReceipt(path string) (DurableRuntimeReceipt, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DurableRuntimeReceipt{}, fmt.Errorf("read durable runtime receipt: %w", err)
	}
	var receipt DurableRuntimeReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return DurableRuntimeReceipt{}, fmt.Errorf("decode durable runtime receipt: %w", err)
	}
	if receipt.ReceiptID != durableReceiptID(receiptWithoutID(receipt)) {
		return DurableRuntimeReceipt{}, fmt.Errorf("durable runtime receipt identity mismatch")
	}
	return receipt, nil
}

func durableLedgerDigest(events []RuntimeEvent) (string, error) {
	data, err := json.Marshal(events)
	if err != nil {
		return "", fmt.Errorf("encode durable runtime ledger digest: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func durableReceiptID(receipt DurableRuntimeReceipt) string {
	data, _ := json.Marshal(receipt)
	sum := sha256.Sum256(data)
	return "receipt-" + hex.EncodeToString(sum[:])
}

func receiptWithoutID(receipt DurableRuntimeReceipt) DurableRuntimeReceipt {
	receipt.ReceiptID = ""
	return receipt
}

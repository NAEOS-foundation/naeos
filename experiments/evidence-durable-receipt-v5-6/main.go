// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
)

type scenario struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

func main() {
	var results []scenario

	results = append(results, validScenario(), mutationScenario(), foreignRunScenario(), tamperedReceiptScenario())

	encoded, _ := json.MarshalIndent(map[string]any{
		"experiment": "evidence-durable-receipt-v5-6",
		"scenarios":  results,
	}, "", "  ")
	fmt.Println(string(encoded))
}

func validScenario() scenario {
	dir := filepath.Join(tempDir(), "valid")
	ledger, err := evidence.NewDurableRuntimeEventLedger(filepath.Join(dir, "events.jsonl"))
	if err != nil { return scenario{"valid-recovery", false, err.Error()} }
	if _, err = ledger.Publish("run-1", "pipeline.start", "p1", 1); err != nil { return scenario{"valid-recovery", false, err.Error()} }
	if _, err = ledger.Publish("run-1", "pipeline.verification", "p2", 2); err != nil { return scenario{"valid-recovery", false, err.Error()} }
	ledger.Seal()
	receipt, err := evidence.CreateDurableRuntimeReceipt(ledger, "run-1", time.Unix(1, 0))
	if err != nil { return scenario{"valid-recovery", false, err.Error()} }
	receiptPath := filepath.Join(dir, "receipt.json")
	if err := evidence.WriteDurableRuntimeReceipt(receiptPath, receipt); err != nil { return scenario{"valid-recovery", false, err.Error()} }
	reloaded, err := evidence.NewDurableRuntimeEventLedger(filepath.Join(dir, "events.jsonl"))
	if err != nil { return scenario{"valid-recovery", false, err.Error()} }
	loaded, err := evidence.LoadDurableRuntimeReceipt(receiptPath)
	if err != nil { return scenario{"valid-recovery", false, err.Error()} }
	err = evidence.VerifyDurableRuntimeReceipt(loaded, reloaded, "run-1")
	return scenario{"valid-recovery", err == nil, detail(err)}
}

func mutationScenario() scenario {
	ledger, err := evidence.NewDurableRuntimeEventLedger(filepath.Join(tempDir(), "mutation", "events.jsonl"))
	if err != nil { return scenario{"sealed-mutation", false, err.Error()} }
	if _, err = ledger.Publish("run-1", "pipeline.start", "p1", 1); err != nil { return scenario{"sealed-mutation", false, err.Error()} }
	ledger.Seal()
	_, err = ledger.Publish("run-1", "pipeline.verification", "p2", 2)
	return scenario{"sealed-mutation", err != nil, detail(err)}
}

func foreignRunScenario() scenario {
	ledger, err := evidence.NewDurableRuntimeEventLedger(filepath.Join(tempDir(), "foreign", "events.jsonl"))
	if err != nil { return scenario{"foreign-run", false, err.Error()} }
	if _, err = ledger.Publish("run-1", "pipeline.start", "p1", 1); err != nil { return scenario{"foreign-run", false, err.Error()} }
	ledger.Seal()
	receipt, err := evidence.CreateDurableRuntimeReceipt(ledger, "run-1", time.Unix(1, 0))
	if err != nil { return scenario{"foreign-run", false, err.Error()} }
	err = evidence.VerifyDurableRuntimeReceipt(receipt, ledger, "run-2")
	return scenario{"foreign-run", err != nil, detail(err)}
}

func tamperedReceiptScenario() scenario {
	ledger, err := evidence.NewDurableRuntimeEventLedger(filepath.Join(tempDir(), "tamper", "events.jsonl"))
	if err != nil { return scenario{"receipt-tamper", false, err.Error()} }
	if _, err = ledger.Publish("run-1", "pipeline.start", "p1", 1); err != nil { return scenario{"receipt-tamper", false, err.Error()} }
	ledger.Seal()
	receipt, err := evidence.CreateDurableRuntimeReceipt(ledger, "run-1", time.Unix(1, 0))
	if err != nil { return scenario{"receipt-tamper", false, err.Error()} }
	receipt.EventCount++
	err = evidence.VerifyDurableRuntimeReceipt(receipt, ledger, "run-1")
	return scenario{"receipt-tamper", err != nil, detail(err)}
}

func detail(err error) string {
	if err == nil { return "accepted" }
	return err.Error()
}

func tempDir() string {
	return filepath.Join("/tmp", fmt.Sprintf("naeos-v5-6-%d", time.Now().UnixNano()))
}

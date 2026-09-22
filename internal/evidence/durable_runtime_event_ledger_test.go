// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDurableRuntimeEventLedgerPersistsAndReloads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil { t.Fatal(err) }
	if _, err := ledger.Publish("run-1", "pipeline.start", "payload-1", 1); err != nil { t.Fatal(err) }
	ledger2, err := NewDurableRuntimeEventLedger(path)
	if err != nil { t.Fatal(err) }
	if err := ledger2.Verify(); err != nil { t.Fatal(err) }
	if len(ledger2.Records()) != 1 { t.Fatalf("expected one durable event") }
}

func TestDurableRuntimeEventLedgerRejectsMutation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil { t.Fatal(err) }
	if _, err := ledger.Publish("run-1", "pipeline.start", "payload-1", 1); err != nil { t.Fatal(err) }
	data, err := os.ReadFile(path)
	if err != nil { t.Fatal(err) }
	for i := range data {
		if data[i] == '1' { data[i] = '9'; break }
	}
	if err := os.WriteFile(path, data, 0o640); err != nil { t.Fatal(err) }
	if _, err := NewDurableRuntimeEventLedger(path); err == nil { t.Fatal("expected mutated ledger to fail integrity verification") }
}

func TestDurableRuntimeEventLedgerRejectsTruncationAndReordering(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil { t.Fatal(err) }
	for i := 1; i <= 2; i++ {
		if _, err := ledger.Publish("run-1", "pipeline.event", "payload", i); err != nil { t.Fatal(err) }
	}
	data, err := os.ReadFile(path)
	if err != nil { t.Fatal(err) }
	first, second := splitLines(data)[0], splitLines(data)[1]
	reordered := append(append([]byte{}, second...), '\n')
	reordered = append(reordered, first...)
	reordered = append(reordered, '\n')
	if err := os.WriteFile(path, reordered, 0o640); err != nil { t.Fatal(err) }
	if _, err := NewDurableRuntimeEventLedger(path); err == nil { t.Fatal("expected reordered ledger to fail") }
}

func TestDurableRuntimeEventLedgerRejectsLateWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil { t.Fatal(err) }
	ledger.Seal()
	if _, err := ledger.Publish("run-1", "pipeline.start", "payload", 1); err == nil { t.Fatal("expected sealed ledger to reject write") }
}

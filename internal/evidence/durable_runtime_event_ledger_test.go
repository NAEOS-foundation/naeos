// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDurableRuntimeEventLedgerHashChain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	first, err := ledger.Publish("run-1", "started", "sha256:a", 1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ledger.Publish("run-1", "completed", "sha256:b", 2)
	if err != nil {
		t.Fatal(err)
	}
	if first.PreviousHash != "" || first.EventHash == "" {
		t.Fatalf("invalid first hash state: %+v", first)
	}
	if second.PreviousHash != first.EventHash || second.EventHash == "" {
		t.Fatalf("invalid second hash link: %+v", second)
	}
	if err := ledger.Verify(); err != nil {
		t.Fatalf("Verify: %v", err)
	}

	reloaded, err := NewDurableRuntimeEventLedger(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if err := reloaded.Verify(); err != nil {
		t.Fatalf("reloaded Verify: %v", err)
	}
	third, err := reloaded.Publish("run-1", "observed", "sha256:c", 3)
	if err != nil {
		t.Fatal(err)
	}
	if third.PreviousHash != second.EventHash {
		t.Fatalf("restart broke hash chain: previous=%q want=%q", third.PreviousHash, second.EventHash)
	}
	if err := reloaded.Verify(); err != nil {
		t.Fatalf("post-restart Verify: %v", err)
	}
}

func TestDurableRuntimeEventLedgerPersistsAndReloads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ledger.Publish("run-1", "pipeline.start", "payload-1", 1); err != nil {
		t.Fatal(err)
	}
	ledger2, err := NewDurableRuntimeEventLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := ledger2.Verify(); err != nil {
		t.Fatal(err)
	}
	if len(ledger2.Records()) != 1 {
		t.Fatalf("expected one durable event")
	}
}

func TestDurableRuntimeEventLedgerRejectsMutation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ledger.Publish("run-1", "pipeline.start", "payload-1", 1); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for i := range data {
		if data[i] == '1' {
			data[i] = '9'
			break
		}
	}
	if err := os.WriteFile(path, data, 0o640); err != nil {
		t.Fatal(err)
	}
	if _, err := NewDurableRuntimeEventLedger(path); err == nil {
		t.Fatal("expected mutated ledger to fail integrity verification")
	}
}

func TestDurableRuntimeEventLedgerRejectsTruncationAndReordering(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2; i++ {
		if _, err := ledger.Publish("run-1", "pipeline.event", "payload", i); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	first, second := splitLines(data)[0], splitLines(data)[1]
	reordered := append(append([]byte{}, second...), '\n')
	reordered = append(reordered, first...)
	reordered = append(reordered, '\n')
	if err := os.WriteFile(path, reordered, 0o640); err != nil {
		t.Fatal(err)
	}
	if _, err := NewDurableRuntimeEventLedger(path); err == nil {
		t.Fatal("expected reordered ledger to fail")
	}
}

func TestDurableRuntimeEventLedgerRejectsLateWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime-events.jsonl")
	ledger, err := NewDurableRuntimeEventLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	ledger.Seal()
	if _, err := ledger.Publish("run-1", "pipeline.start", "payload", 1); err == nil {
		t.Fatal("expected sealed ledger to reject write")
	}
}

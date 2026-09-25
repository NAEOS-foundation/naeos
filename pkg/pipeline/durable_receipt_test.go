// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
)

func TestValidateRunCompletionRequiresDurableReceipt(t *testing.T) {
	runID := "run-durable-receipt"
	store := evidence.NewStore()
	observer := evidence.NewIndependentRuntimeObserver()
	builder := evidence.NewRuntimeEvidenceBuilder(store, observer)

	lifecycle := []struct {
		kind, stage, event string
	}{
		{"intent", "run", "pipeline.start"},
		{"decision", "policy_eval", "pipeline.policy_decision"},
		{"execution", "write_artifacts", "pipeline.execution"},
		{"observation", "observation", "pipeline.observation"},
		{"verification", "completion", "pipeline.verification"},
	}
	for i, item := range lifecycle {
		payload := evidence.ComputeArtifactHash([]byte(item.event))
		event, err := observer.Observe(runID, item.event, payload, i+1)
		if err != nil {
			t.Fatalf("observe %s: %v", item.kind, err)
		}
		if err := builder.Build(runID, item.kind, i+1, item.stage, item.event, event); err != nil {
			t.Fatalf("build %s evidence: %v", item.kind, err)
		}
	}
	observer.Seal()

	if err := validateRunCompletion(store, observer.Ledger(), nil, evidence.DurableRuntimeReceipt{}, runID); err == nil {
		t.Fatal("expected completion to reject missing durable receipt")
	}
}

func TestValidateRunCompletionAcceptsVerifiedDurableReceipt(t *testing.T) {
	runID := "run-durable-receipt-valid"
	ledgerPath := filepath.Join(t.TempDir(), "runtime.jsonl")
	durableLedger, err := evidence.NewDurableRuntimeEventLedger(ledgerPath)
	if err != nil {
		t.Fatalf("new durable ledger: %v", err)
	}
	observer := evidence.NewIndependentRuntimeObserverWithDurableLedger(durableLedger)
	store := evidence.NewStore()
	builder := evidence.NewRuntimeEvidenceBuilder(store, observer)

	lifecycle := []struct {
		kind, stage, event string
	}{
		{"intent", "run", "pipeline.start"},
		{"decision", "policy_eval", "pipeline.policy_decision"},
		{"execution", "write_artifacts", "pipeline.execution"},
		{"observation", "observation", "pipeline.observation"},
		{"verification", "completion", "pipeline.verification"},
	}
	for i, item := range lifecycle {
		payload := evidence.ComputeArtifactHash([]byte(item.event))
		event, err := observer.Observe(runID, item.event, payload, i+1)
		if err != nil {
			t.Fatalf("observe %s: %v", item.kind, err)
		}
		if err := builder.Build(runID, item.kind, i+1, item.stage, item.event, event); err != nil {
			t.Fatalf("build %s evidence: %v", item.kind, err)
		}
	}
	observer.Seal()

	receipt, err := evidence.CreateDurableRuntimeReceipt(durableLedger, runID, time.Now().UTC())
	if err != nil {
		t.Fatalf("create receipt: %v", err)
	}
	if err := validateRunCompletion(store, observer.Ledger(), durableLedger, receipt, runID); err != nil {
		t.Fatalf("expected verified durable receipt to satisfy completion: %v", err)
	}
}

// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
)

var lifecycle = []struct {
	kind  string
	stage string
	event string
}{
	{"intent", "run", "pipeline.start"},
	{"decision", "policy_eval", "pipeline.policy_decision"},
	{"execution", "write_artifacts", "pipeline.execution"},
	{"observation", "observation", "pipeline.observation"},
	{"verification", "completion", "pipeline.verification"},
}

func main() {
	fmt.Println("NAEOS Independent Runtime Observer v5.4")
	fmt.Println("Runtime → Independent Observer → Runtime Event Ledger → Evidence Builder → Verification → Completion")

	if err := validScenario(); err != nil {
		fmt.Printf("VALID: FAIL (%v)\n", err)
		os.Exit(1)
	}
	fmt.Println("VALID: PASS")

	if err := observerBypassScenario(); err == nil {
		fmt.Println("OBSERVER_BYPASS: FAIL (evidence escaped observer boundary)")
		os.Exit(1)
	}
	fmt.Println("OBSERVER_BYPASS: BLOCKED")

	if err := lateObservationScenario(); err == nil {
		fmt.Println("LATE_OBSERVATION: FAIL (observer accepted event after seal)")
		os.Exit(1)
	}
	fmt.Println("LATE_OBSERVATION: BLOCKED")

	if err := foreignObservationScenario(); err == nil {
		fmt.Println("FOREIGN_RUN: FAIL (foreign event bound to target run)")
		os.Exit(1)
	}
	fmt.Println("FOREIGN_RUN: BLOCKED")
}

func validScenario() error {
	store := evidence.NewStore()
	observer := evidence.NewIndependentRuntimeObserver()
	builder := evidence.NewRuntimeEvidenceBuilder(store, observer)

	for i, item := range lifecycle {
		event, err := observer.Observe("run-valid", item.event, fmt.Sprintf("payload-%d", i+1), i+1)
		if err != nil {
			return err
		}
		if err := builder.Build("run-valid", item.kind, i+1, item.stage, item.event, event); err != nil {
			return err
		}
	}

	observer.Seal()
	result := evidence.ValidateCompletionWithRuntimeLedger(store, observer.Ledger(), "run-valid", requiredKinds())
	if !result.Complete {
		return fmt.Errorf("completion blocked: %v", result.Missing)
	}
	return nil
}

func observerBypassScenario() error {
	store := evidence.NewStore()
	observer := evidence.NewIndependentRuntimeObserver()
	builder := evidence.NewRuntimeEvidenceBuilder(store, observer)
	fake := evidence.RuntimeEvent{ID: "evt-forged", RunID: "run-target", Name: "pipeline.start", PayloadDigest: "payload", Sequence: 1}
	if err := builder.Build("run-target", "intent", 1, "run", "pipeline.start", fake); err == nil {
		return nil
	}
	return fmt.Errorf("forged event rejected")
}

func lateObservationScenario() error {
	observer := evidence.NewIndependentRuntimeObserver()
	if _, err := observer.Observe("run-late", "pipeline.start", "payload-1", 1); err != nil {
		return err
	}
	observer.Seal()
	_, err := observer.Observe("run-late", "pipeline.verification", "payload-2", 2)
	return err
}

func foreignObservationScenario() error {
	store := evidence.NewStore()
	observer := evidence.NewIndependentRuntimeObserver()
	builder := evidence.NewRuntimeEvidenceBuilder(store, observer)
	event, err := observer.Observe("run-other", "pipeline.start", "payload-1", 1)
	if err != nil {
		return err
	}
	if err := builder.Build("run-target", "intent", 1, "run", "pipeline.start", event); err == nil {
		return nil
	}
	return fmt.Errorf("foreign event rejected")
}

func requiredKinds() []string {
	kinds := make([]string, len(lifecycle))
	for i, item := range lifecycle {
		kinds[i] = item.kind
	}
	return kinds
}

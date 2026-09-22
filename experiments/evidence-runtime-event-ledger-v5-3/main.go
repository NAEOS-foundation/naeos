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
	fmt.Println("NAEOS Evidence Runtime Event Binding v5.3")
	fmt.Println("Event Producer → Runtime Event Ledger → Evidence Builder → Verification → Completion")

	if err := validScenario(); err != nil {
		fmt.Printf("VALID: FAIL (%v)\n", err)
		os.Exit(1)
	}
	fmt.Println("VALID: PASS")

	if err := missingEventScenario(); err == nil {
		fmt.Println("MISSING_EVENT: FAIL (unexpected completion)")
		os.Exit(1)
	}
	fmt.Println("MISSING_EVENT: BLOCKED")

	if err := foreignEventScenario(); err == nil {
		fmt.Println("FOREIGN_EVENT: FAIL (unexpected evidence binding)")
		os.Exit(1)
	}
	fmt.Println("FOREIGN_EVENT: BLOCKED")

	if err := lateEventScenario(); err == nil {
		fmt.Println("LATE_EVENT: FAIL (unexpected ledger append)")
		os.Exit(1)
	}
	fmt.Println("LATE_EVENT: BLOCKED")
}

func validScenario() error {
	store := evidence.NewStore()
	ledger := evidence.NewRuntimeEventLedger()
	builder := evidence.NewRuntimeEvidenceBuilder(store, ledger)
	for i, item := range lifecycle {
		event, err := ledger.Publish("run-valid", item.event, fmt.Sprintf("payload-%d", i+1), i+1)
		if err != nil {
			return err
		}
		if err := builder.Build("run-valid", item.kind, i+1, item.stage, item.event, event); err != nil {
			return err
		}
	}
	ledger.Seal()
	result := evidence.ValidateCompletionWithRuntimeLedger(store, ledger, "run-valid", requiredKinds())
	if !result.Complete {
		return fmt.Errorf("completion blocked: %v", result.Missing)
	}
	return nil
}

func missingEventScenario() error {
	store := evidence.NewStore()
	ledger := evidence.NewRuntimeEventLedger()
	builder := evidence.NewRuntimeEvidenceBuilder(store, ledger)
	for i, item := range lifecycle[:4] {
		event, err := ledger.Publish("run-missing", item.event, fmt.Sprintf("payload-%d", i+1), i+1)
		if err != nil {
			return err
		}
		if err := builder.Build("run-missing", item.kind, i+1, item.stage, item.event, event); err != nil {
			return err
		}
	}
	ledger.Seal()
	result := evidence.ValidateCompletionWithRuntimeLedger(store, ledger, "run-missing", requiredKinds())
	if result.Complete {
		return nil
	}
	return fmt.Errorf("expected incomplete run")
}

func foreignEventScenario() error {
	store := evidence.NewStore()
	ledger := evidence.NewRuntimeEventLedger()
	builder := evidence.NewRuntimeEvidenceBuilder(store, ledger)
	event, err := ledger.Publish("run-other", "pipeline.start", "payload-1", 1)
	if err != nil {
		return err
	}
	if err := builder.Build("run-target", "intent", 1, "run", "pipeline.start", event); err == nil {
		return nil
	}
	return fmt.Errorf("foreign event rejected")
}

func lateEventScenario() error {
	ledger := evidence.NewRuntimeEventLedger()
	if _, err := ledger.Publish("run-late", "pipeline.start", "payload-1", 1); err != nil {
		return err
	}
	ledger.Seal()
	_, err := ledger.Publish("run-late", "pipeline.verification", "payload-2", 2)
	return err
}

func requiredKinds() []string {
	kinds := make([]string, len(lifecycle))
	for i, item := range lifecycle {
		kinds[i] = item.kind
	}
	return kinds
}

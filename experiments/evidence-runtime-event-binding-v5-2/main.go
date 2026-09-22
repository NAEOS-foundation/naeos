// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

type scenario struct {
	Name string `json:"name"`
	Pass bool   `json:"pass"`
}

var required = []string{"intent", "decision", "execution", "observation", "verification"}

func main() {
	results := []scenario{
		runValid(),
		runMissingEvent(),
		runPayloadMismatch(),
		runWrongRun(),
		runEventAfterCompletion(),
	}
	data, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(data))
	for _, result := range results {
		if !result.Pass {
			panic("runtime event binding scenario failed: " + result.Name)
		}
	}
}

func appendEvidence(store *evidence.EvidenceStore, events *evidence.RuntimeEventStore, runID, kind string, sequence int, name, payload string) {
	event, err := events.Append(runID, name, payload, sequence)
	if err != nil {
		panic(err)
	}
	previous := ""
	if latest := store.Latest(); latest != nil {
		previous = latest.ID
	}
	stages := map[string]string{
		"intent": "run", "decision": "policy_eval", "execution": "write_artifacts",
		"observation": "observation", "verification": "completion",
	}
	_, err = store.Append(evidence.EvidenceRecord{
		ID: fmt.Sprintf("%s-%s", runID, kind), Actor: "experiment", Resource: "pipeline", Action: "run",
		Environment: "runtime", PolicyID: "pipeline-lifecycle", PolicyVersion: "1.0.0",
		Decision: control.DecisionAllow, ExecutionStatus: "recorded",
		Metadata: map[string]any{
			"run_id": runID, "run_binding": evidence.RunBindingDigest(runID), "kind": kind, "sequence": sequence,
			"previous_evidence_id": previous, "provenance_stage": stages[kind], "provenance_event": name,
			"payload_digest": payload, "provenance_digest": evidence.ProvenanceDigest(stages[kind], name, payload),
			"runtime_event_id": event.ID,
		},
	})
	if err != nil {
		panic(err)
	}
}

func complete(runID string) (*evidence.EvidenceStore, *evidence.RuntimeEventStore) {
	store, events := evidence.NewStore(), evidence.NewRuntimeEventStore()
	names := []string{"pipeline.start", "pipeline.policy_decision", "pipeline.execution", "pipeline.observation", "pipeline.verification"}
	for i, kind := range required {
		appendEvidence(store, events, runID, kind, i+1, names[i], fmt.Sprintf("payload-%d", i+1))
	}
	return store, events
}

func runValid() scenario {
	store, events := complete("run-valid")
	return scenario{"valid-event-binding", evidence.ValidateCompletionWithRuntimeEvents(store, events, "run-valid", required).Complete}
}

func runMissingEvent() scenario {
	store, events := complete("run-missing")
	records := events.Records()
	_ = records
	record := store.ByID("run-missing-verification")
	record.Metadata["runtime_event_id"] = "evt-missing"
	return scenario{"missing-event-blocked", !evidence.ValidateCompletionWithRuntimeEvents(store, events, "run-missing", required).Complete}
}

func runPayloadMismatch() scenario {
	store, events := complete("run-payload")
	record := store.ByID("run-payload-execution")
	record.Metadata["payload_digest"] = "tampered"
	return scenario{"payload-mismatch-blocked", !evidence.ValidateCompletionWithRuntimeEvents(store, events, "run-payload", required).Complete}
}

func runWrongRun() scenario {
	store, events := complete("run-one")
	_, _ = events.Append("run-two", "pipeline.verification", "payload-other", 6)
	return scenario{"event-from-another-run-blocked", !evidence.ValidateCompletionWithRuntimeEvents(store, events, "run-one", required).Complete}
}

func runEventAfterCompletion() scenario {
	store, events := complete("run-after")
	_, _ = events.Append("run-after", "pipeline.after_completion", "payload-after", 6)
	return scenario{"event-after-completion-blocked", !evidence.ValidateCompletionWithRuntimeEvents(store, events, "run-after", required).Complete}
}

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
	results := []scenario{runComplete(), runMissingProvenance(), runMismatchedProvenance()}
	data, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(data))
	for _, result := range results {
		if !result.Pass {
			panic("runtime provenance scenario failed: " + result.Name)
		}
	}
}

func appendRecord(store *evidence.EvidenceStore, runID, kind string, seq int, stage, event, payload string) {
	previous := ""
	if latest := store.Latest(); latest != nil {
		previous = latest.ID
	}
	_, err := store.Append(evidence.EvidenceRecord{
		ID: fmt.Sprintf("%s-%s", runID, kind), Actor: "experiment", Resource: "pipeline", Action: "run",
		Environment: "runtime", PolicyID: "pipeline-lifecycle", PolicyVersion: "1.0.0",
		Decision: control.DecisionAllow, ExecutionStatus: "recorded",
		Metadata: map[string]any{
			"run_id": runID, "run_binding": evidence.RunBindingDigest(runID), "kind": kind, "sequence": seq,
			"previous_evidence_id": previous, "provenance_stage": stage, "provenance_event": event,
			"payload_digest": payload, "provenance_digest": evidence.ProvenanceDigest(stage, event, payload),
		},
	})
	if err != nil {
		panic(err)
	}
}

func runComplete() scenario {
	store := evidence.NewStore()
	stages := [][2]string{{"run", "pipeline.start"}, {"policy_eval", "pipeline.policy_decision"}, {"write_artifacts", "pipeline.execution"}, {"observation", "pipeline.observation"}, {"completion", "pipeline.verification"}}
	for i, kind := range required {
		appendRecord(store, "run-provenance", kind, i+1, stages[i][0], stages[i][1], fmt.Sprintf("payload-%d", i+1))
	}
	return scenario{"complete-provenance", evidence.ValidateCompletion(store, "run-provenance", required).Complete}
}

func runMissingProvenance() scenario {
	store := evidence.NewStore()
	stages := [][2]string{{"run", "pipeline.start"}, {"policy_eval", "pipeline.policy_decision"}, {"write_artifacts", "pipeline.execution"}, {"observation", "pipeline.observation"}, {"completion", "pipeline.verification"}}
	for i, kind := range required {
		stage, event := stages[i][0], stages[i][1]
		if kind == "observation" {
			event = ""
		}
		appendRecord(store, "run-missing", kind, i+1, stage, event, fmt.Sprintf("payload-%d", i+1))
	}
	return scenario{"missing-provenance-blocked", !evidence.ValidateCompletion(store, "run-missing", required).Complete}
}

func runMismatchedProvenance() scenario {
	store := evidence.NewStore()
	stages := [][2]string{{"run", "pipeline.start"}, {"policy_eval", "pipeline.policy_decision"}, {"write_artifacts", "pipeline.execution"}, {"observation", "pipeline.observation"}, {"completion", "pipeline.verification"}}
	for i, kind := range required {
		stage, event := stages[i][0], stages[i][1]
		if kind == "decision" {
			stage, event = "run", "pipeline.start"
		}
		appendRecord(store, "run-mismatch", kind, i+1, stage, event, fmt.Sprintf("payload-%d", i+1))
	}
	return scenario{"mismatched-provenance-blocked", !evidence.ValidateCompletion(store, "run-mismatch", required).Complete}
}

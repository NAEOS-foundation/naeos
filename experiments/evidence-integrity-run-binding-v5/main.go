package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

var required = []string{"intent", "decision", "execution", "observation", "verification"}

func appendEvidence(store *evidence.EvidenceStore, runID, kind string, sequence int, previousID string) error {
	_, err := store.Append(evidence.EvidenceRecord{
		ID:    fmt.Sprintf("%s-%s", runID, kind),
		Actor: "experiment", Resource: "run", Action: kind,
		Environment: "test", PolicyID: "v5", PolicyVersion: "1.0.0",
		Decision: control.DecisionAllow, ExecutionStatus: "recorded",
		Metadata: map[string]any{
			"run_id": runID, "run_binding": evidence.RunBindingDigest(runID),
			"kind": kind, "sequence": sequence, "previous_evidence_id": previousID,
		},
	})
	return err
}

func completeRun(store *evidence.EvidenceStore, runID string) error {
	var previousID string
	for i, kind := range required {
		if err := appendEvidence(store, runID, kind, i+1, previousID); err != nil {
			return err
		}
		previousID = fmt.Sprintf("%s-%s", runID, kind)
	}
	return nil
}

func main() {
	type scenario struct {
		Name     string
		Complete bool
		Expected bool
	}
	results := []scenario{}

	store := evidence.NewStore()
	if err := completeRun(store, "run-complete"); err != nil {
		panic(err)
	}
	results = append(results, scenario{"complete-run", evidence.ValidateCompletion(store, "run-complete", required).Complete, true})

	interleaved := evidence.NewStore()
	_ = appendEvidence(interleaved, "run-a", "intent", 1, "")
	_ = appendEvidence(interleaved, "run-a", "decision", 2, "run-a-intent")
	_ = appendEvidence(interleaved, "run-b", "intent", 1, "")
	_ = appendEvidence(interleaved, "run-a", "execution", 3, "run-a-decision")
	_ = appendEvidence(interleaved, "run-a", "observation", 4, "run-a-execution")
	_ = appendEvidence(interleaved, "run-a", "verification", 5, "run-a-observation")
	results = append(results, scenario{"mixed-run-evidence", evidence.ValidateCompletion(interleaved, "run-a", required).Complete, false})

	results = append(results, scenario{"duplicate-required-kind", evidence.ValidateCompletion(store, "run-complete", []string{"intent", "intent"}).Complete, false})
	results = append(results, scenario{"run-binding", evidence.ValidateCompletion(store, "run-complete", required).Complete, true})

	tampered := evidence.NewStore()
	_ = completeRun(tampered, "run-tampered")
	if record := tampered.ByID("run-tampered-execution"); record != nil {
		record.ExecutionStatus = "tampered"
	}
	results = append(results, scenario{"tampered-evidence", evidence.ValidateCompletion(tampered, "run-tampered", required).Complete, false})

	encoded, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(encoded))
	for _, result := range results {
		if result.Complete != result.Expected {
			os.Exit(1)
		}
	}
}

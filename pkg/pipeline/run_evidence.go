// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"encoding/json"
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/governance/review"
	"github.com/NAEOS-foundation/naeos/internal/planner/scheduler"
)

var requiredRunEvidenceKinds = []string{"intent", "decision", "execution", "observation", "verification"}

func appendRunEvidence(builder *evidence.RuntimeEvidenceBuilder, observer *evidence.IndependentRuntimeObserver, runID, kind string, sequence int, stage, event, payloadDigest string) error {
	if builder == nil || observer == nil {
		return fmt.Errorf("runtime evidence builder and producer are required")
	}
	runtimeEvent, err := observer.Observe(runID, event, payloadDigest, sequence)
	if err != nil {
		return err
	}
	return builder.Build(runID, kind, sequence, stage, event, runtimeEvent)
}

func artifactDigest(artifacts []engine.Artifact) string {
	data, _ := json.Marshal(artifacts)
	return evidence.ComputeArtifactHash(data)
}

func observationDigest(tasks []scheduler.Task, artifacts []engine.Artifact, reviews []*review.ReviewResult) string {
	data, _ := json.Marshal(map[string]int{"tasks": len(tasks), "artifacts": len(artifacts), "reviews": len(reviews)})
	return evidence.ComputeArtifactHash(data)
}

func verificationDigest(tasks []scheduler.Task, artifacts []engine.Artifact, reviews []*review.ReviewResult) string {
	data, _ := json.Marshal(map[string]any{
		"tasks": len(tasks), "artifacts": len(artifacts), "reviews": len(reviews), "verification": "completion-boundary",
	})
	return evidence.ComputeArtifactHash(data)
}

func validateRunCompletion(store *evidence.EvidenceStore, runtimeLedger *evidence.RuntimeEventLedger, runID string) error {
	completion := evidence.ValidateCompletionWithRuntimeLedger(store, runtimeLedger, runID, requiredRunEvidenceKinds)
	if !completion.Complete {
		return fmt.Errorf("run completion blocked: incomplete evidence contract: %v", completion.Missing)
	}
	return nil
}

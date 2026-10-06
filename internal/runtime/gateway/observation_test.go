// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

type stubObserver struct {
	observation Observation
	err         error
	calls       int
}

func (s *stubObserver) Observe(req ToolRequest, result ExecutionResult) (Observation, error) {
	s.calls++
	return s.observation, s.err
}

func TestGatewayObservationIsFirstClass(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "sandbox claims completion"}
	obs := &stubObserver{
		observation: Observation{
			Status:       "observed",
			Observed:     true,
			ArtifactHash: "sha256:artifact",
			ArtifactSize: 42,
			Metadata:     map[string]string{"path": "/workspace/effect.txt"},
			Timestamp:    time.Now().UTC(),
		},
	}
	gw := New(cp, sb, WithObserver(obs))

	result, err := gw.Authorize(ToolRequest{
		Tool: "filesystem",
		Action: "write",
		Resource: "workspace",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.calls != 1 {
		t.Fatalf("expected one observation, got %d", obs.calls)
	}
	if result.Status != "completed" {
		t.Fatalf("expected completed after observed side effect, got %s", result.Status)
	}
	if result.Observation == nil || !result.Observation.Observed {
		t.Fatalf("expected first-class observation, got %#v", result.Observation)
	}
	if result.Observation.ArtifactHash != "sha256:artifact" {
		t.Fatalf("unexpected artifact hash: %s", result.Observation.ArtifactHash)
	}
	if len(gw.History()) != 1 || gw.History()[0].Observation == nil {
		t.Fatal("expected observation to be retained in runtime history")
	}
}

func TestGatewayObservationFailureBlocksCompletion(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "sandbox claims completion"}
	obs := &stubObserver{
		observation: Observation{
			Status:   "absent",
			Observed: false,
		},
	}
	gw := New(cp, sb, WithObserver(obs))

	result, err := gw.Authorize(ToolRequest{
		Tool: "filesystem",
		Action: "write",
		Resource: "workspace",
	})
	if err == nil {
		t.Fatal("expected observation failure to fail closed")
	}
	if result.Status != "failed" {
		t.Fatalf("expected failed, got %s", result.Status)
	}
	if result.Observation == nil || result.Observation.Observed {
		t.Fatalf("expected negative observation to be retained, got %#v", result.Observation)
	}
	if len(gw.History()) != 1 {
		t.Fatalf("expected failed result in history, got %d entries", len(gw.History()))
	}
}

func TestGatewayObserverErrorFailsClosed(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "sandbox claims completion"}
	obs := &stubObserver{err: errObservationUnavailable{}}
	gw := New(cp, sb, WithObserver(obs))

	result, err := gw.Authorize(ToolRequest{Tool: "filesystem", Action: "write"})
	if err == nil {
		t.Fatal("expected observer error to fail closed")
	}
	if result.Status != "failed" {
		t.Fatalf("expected failed, got %s", result.Status)
	}
}

type errObservationUnavailable struct{}

func (errObservationUnavailable) Error() string { return "observer unavailable" }

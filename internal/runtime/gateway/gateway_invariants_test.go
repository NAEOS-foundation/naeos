// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"fmt"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

type invariantRevalidator struct {
	decision control.Decision
	err      error
}

func (r invariantRevalidator) ValidateDecision(req control.Request, issued control.DecisionRecord) (control.DecisionRecord, error) {
	if r.err != nil {
		return control.DecisionRecord{}, r.err
	}
	issued.Decision = r.decision
	issued.Reasons = []string{"decision changed during revalidation"}
	return issued, nil
}

func TestGatewayRevalidationRequireApprovalBlocksSandbox(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &countingSandbox{}
	gw := New(cp, sb)
	gw.controlPlane = struct {
		*stubControlPlane
		invariantRevalidator
	}{cp, invariantRevalidator{decision: control.DecisionRequireApproval}}

	result, err := gw.Authorize(ToolRequest{Tool: "deploy", Action: "run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "denied" || result.Decision != control.DecisionRequireApproval {
		t.Fatalf("expected revalidated REQUIRE_APPROVAL denial, got decision=%s status=%s", result.Decision, result.Status)
	}
	if sb.Count() != 0 {
		t.Fatalf("revalidated REQUIRE_APPROVAL reached sandbox: %d calls", sb.Count())
	}
}

func TestGatewayRevalidationErrorFailsClosed(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &countingSandbox{}
	gw := New(cp, sb)
	gw.controlPlane = struct {
		*stubControlPlane
		invariantRevalidator
	}{cp, invariantRevalidator{err: fmt.Errorf("policy state unavailable")}}

	result, err := gw.Authorize(ToolRequest{Tool: "deploy", Action: "run"})
	if err != nil {
		t.Fatalf("revalidation errors are represented as a denied result, got error: %v", err)
	}
	if result.Status != "denied" {
		t.Fatalf("expected denied result, got %s", result.Status)
	}
	if result.Decision != control.DecisionAllow {
		t.Fatalf("expected original decision to remain auditable, got %s", result.Decision)
	}
	if sb.Count() != 0 {
		t.Fatalf("revalidation error reached sandbox: %d calls", sb.Count())
	}
}

type invariantObserver struct {
	observation Observation
	err         error
}

func (o invariantObserver) Observe(req ToolRequest, result ExecutionResult) (Observation, error) {
	return o.observation, o.err
}

func TestGatewayEvidenceObserverFalseFailsClosed(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "claimed side effect"}
	gw := New(cp, sb, WithObserver(invariantObserver{
		observation: Observation{Status: "absent", Observed: false},
	}))

	result, err := gw.Authorize(ToolRequest{RequestID: "req-evidence-absent", InvocationID: "inv-evidence-absent", Tool: "filesystem", Action: "write"})
	if err == nil {
		t.Fatal("expected missing observed side effect to fail closed")
	}
	if result.Status != "failed" {
		t.Fatalf("expected failed result, got %s", result.Status)
	}
	if result.Observation == nil || result.Observation.Observed {
		t.Fatalf("expected explicit negative observation, got %+v", result.Observation)
	}
}

func TestGatewayEvidenceObserverErrorFailsClosed(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "claimed side effect"}
	gw := New(cp, sb, WithObserver(invariantObserver{
		err: fmt.Errorf("target unavailable"),
	}))

	result, err := gw.Authorize(ToolRequest{RequestID: "req-evidence-error", InvocationID: "inv-evidence-error", Tool: "filesystem", Action: "write"})
	if err == nil {
		t.Fatal("expected observer error to fail closed")
	}
	if result.Status != "failed" {
		t.Fatalf("expected failed result, got %s", result.Status)
	}
	if result.Observation == nil {
		t.Fatal("expected observer result to be retained in execution evidence")
	}
}

func TestGatewayEvidenceSuccessRequiresObservedSideEffect(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &stubSandbox{output: "claimed side effect"}
	gw := New(cp, sb, WithObserver(invariantObserver{
		observation: Observation{
			Status:       "observed",
			Observed:     true,
			ArtifactHash: "artifact-sha",
			ArtifactSize: 42,
		},
	}))

	result, err := gw.Authorize(ToolRequest{RequestID: "req-evidence-ok", InvocationID: "inv-evidence-ok", Tool: "filesystem", Action: "write"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Fatalf("expected completed result, got %s", result.Status)
	}
	if result.Observation == nil || !result.Observation.Observed {
		t.Fatalf("expected independently observed evidence, got %+v", result.Observation)
	}
	if result.Observation.ArtifactHash != "artifact-sha" {
		t.Fatalf("expected evidence artifact hash to be retained, got %q", result.Observation.ArtifactHash)
	}
}


func TestGatewayInitialRequireApprovalBlocksSandbox(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionRequireApproval, policyID: "p-approval"}
	sb := &countingSandbox{}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{Tool: "deploy", Action: "run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "denied" || result.Decision != control.DecisionRequireApproval {
		t.Fatalf("expected REQUIRE_APPROVAL denial, got decision=%s status=%s", result.Decision, result.Status)
	}
	if sb.Count() != 0 {
		t.Fatalf("REQUIRE_APPROVAL reached sandbox: %d calls", sb.Count())
	}
}

func TestGatewayUnknownDecisionFailsClosed(t *testing.T) {
	cp := &stubControlPlane{decision: control.Decision("UNKNOWN"), policyID: "p-unknown"}
	sb := &countingSandbox{}
	gw := New(cp, sb)

	result, err := gw.Authorize(ToolRequest{Tool: "deploy", Action: "run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "denied" {
		t.Fatalf("expected unknown decision to be denied, got %s", result.Status)
	}
	if sb.Count() != 0 {
		t.Fatalf("unknown decision reached sandbox: %d calls", sb.Count())
	}
}

func TestGatewayEvidenceMismatchAndUnavailableAreNotSuccess(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{name: "mismatch", status: "mismatch"},
		{name: "unavailable", status: "unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
			sb := &stubSandbox{output: "claimed side effect"}
			gw := New(cp, sb, WithObserver(invariantObserver{
				observation: Observation{Status: tt.status, Observed: false},
			}))

			result, err := gw.Authorize(ToolRequest{
				RequestID:    "req-evidence-" + tt.name,
				InvocationID: "inv-evidence-" + tt.name,
				Tool:         "filesystem",
				Action:       "write",
			})
			if err == nil {
				t.Fatalf("expected %s observation to fail closed", tt.status)
			}
			if result.Status != "failed" {
				t.Fatalf("expected failed result, got %s", result.Status)
			}
			if result.Observation == nil || result.Observation.Status != tt.status || result.Observation.Observed {
				t.Fatalf("expected %s negative observation, got %+v", tt.status, result.Observation)
			}
		})
	}
}

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"strings"
	"testing"
	"time"
)

// TestDecisionGateway_IndeterminateEvaluationIsAuditedAndCannotExecute verifies
// both evaluator-unavailable and unresolved-policy-version failures remain
// internal indeterminate states, are not assigned decision IDs, and cannot
// reach the atomic side-effect callback.
func TestDecisionGateway_IndeterminateEvaluationIsAuditedAndCannotExecute(t *testing.T) {
	now := time.Now().UTC()
	policy := &Policy{
		ID: "POLICY-711", Version: 7, Status: "active",
		AllowedCapabilities: []Capability{"repository.write"},
	}
	grant := &Grant{
		GrantID: "GRANT-711", AgentID: "agent-711",
		PolicyID: policy.ID, PolicyVersion: policy.Version,
		Capabilities: []Capability{"repository.write"},
		Status:       "active", CreatedAt: now, ExpiresAt: now.Add(time.Hour),
	}

	tests := []struct {
		name      string
		evaluator *Evaluator
	}{
		{name: "evaluator unavailable", evaluator: nil},
		{name: "policy version unresolved", evaluator: NewEvaluator(NewPolicyStore())},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ledger := NewLedger()
			gateway := NewDecisionGateway(tt.evaluator, ledger)
			req := AuthorizeRequest{
				RequestID: "REQ-711-" + strings.ReplaceAll(tt.name, " ", "-"),
				AgentID:   grant.AgentID,
				Action: Action{
					AgentID: grant.AgentID, Capability: "repository.write",
					ArtifactHash: "sha256:issue-711",
					Payload:      map[string]string{"private": "must-not-appear-in-failure-metadata"},
				},
				Grant: grant, Policy: policy, Timestamp: now,
			}

			result := gateway.Authorize(req)
			if result.Status != DecisionIndeterminate || result.Reason != ReasonEvaluationFailed {
				t.Fatalf("expected internal indeterminate evaluation failure, got %+v", result)
			}
			if result.DecisionID != "" {
				t.Fatalf("evaluation failure must not mint a decision ID, got %q", result.DecisionID)
			}

			called := false
			executed, event := gateway.ExecuteAtomic(req, result, func() error {
				called = true
				return nil
			})
			if called {
				t.Fatal("indeterminate evaluation must block the side-effect callback")
			}
			if executed.Status != DecisionIndeterminate || executed.DecisionID != "" {
				t.Fatalf("execution boundary must preserve indeterminate state without decision ID, got %+v", executed)
			}
			if event.ID != "" || event.EventType != "" {
				t.Fatalf("indeterminate execution must not emit an execution receipt, got %+v", event)
			}

			events := ledger.Query(map[string]string{"request_id": req.RequestID})
			if len(events) != 1 {
				t.Fatalf("expected one failure-evidence event and no execution events, got %d: %+v", len(events), events)
			}
			failure := events[0]
			if failure.EventType != "POLICY_EVALUATION_FAILED" || failure.DecisionID != "" || failure.Decision != "" {
				t.Fatalf("failure evidence must remain separate from authorization decisions, got %+v", failure)
			}
			if failure.RequestID != req.RequestID || failure.AgentID != req.AgentID || failure.Capability != req.Action.Capability {
				t.Fatalf("failure evidence must retain request correlation, got %+v", failure)
			}
			if failure.Metadata["failure_message"] == "" {
				t.Fatal("expected failure message metadata")
			}
			if failure.Metadata["failure_code"] != "policy_evaluation_failed" ||
				failure.Metadata["failure_stage"] != "policy_evaluation" {
				t.Fatalf("expected structured failure classification, got %+v", failure.Metadata)
			}
			if failure.Metadata["policy_id"] != policy.ID ||
				failure.Metadata["policy_version"] != "7" {
				t.Fatalf("expected policy attribution in failure metadata, got %+v", failure.Metadata)
			}
			if strings.Contains(failure.Metadata["failure_message"], "must-not-appear-in-failure-metadata") {
				t.Fatal("failure metadata leaked private action payload")
			}
		})
	}
}

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"fmt"
	"time"
)

// SessionVerifier independently checks that an agent's actions were valid and policy-compliant.
type SessionVerifier struct {
	Ledger    *Ledger
	Evaluator *Evaluator
}

// NewSessionVerifier creates a verifier bound to the provided ledger and evaluator.
func NewSessionVerifier(l *Ledger, e *Evaluator) *SessionVerifier {
	if l == nil {
		l = NewLedger()
	}
	if e == nil {
		e = NewEvaluator()
	}
	return &SessionVerifier{Ledger: l, Evaluator: e}
}

// Verify checks one capability request against the session evidence ledger.
func (v *SessionVerifier) Verify(agentID string, action Action, grant *Grant, policy *Policy) DecisionResult {
	if v == nil || v.Evaluator == nil {
		return DecisionResult{Status: DecisionDeny, Reason: ReasonNoPolicyFound, Requested: action.Capability, Message: "verifier unavailable"}
	}
	if grant == nil {
		return DecisionResult{Status: DecisionDeny, Reason: ReasonDeniedByGrant, Requested: action.Capability, Message: "grant missing"}
	}
	if policy == nil {
		return DecisionResult{Status: DecisionDeny, Reason: ReasonNoPolicyFound, Requested: action.Capability, Message: "policy missing"}
	}
	decision := v.Evaluator.EvaluateAction(action, grant, policy, time.Now().UTC())
	if decision.Status != DecisionAllow && decision.Status != DecisionPending {
		return decision
	}
	if v.Ledger != nil {
		events := v.Ledger.EventsForAgent(agentID)
		for _, event := range events {
			if event.Capability == action.Capability && event.ArtifactHash == action.ArtifactHash && event.EventType == "EXECUTION_UNKNOWN" && !executionHasValidReconciliation(events, event.ExecutionID) {
				return DecisionResult{Status: DecisionUnknown, Reason: ReasonExecutionOutcomeUnknown, Requested: action.Capability, Message: fmt.Sprintf("execution outcome for %s is unconfirmed and requires reconciliation", action.Capability)}
			}
		}
		for _, event := range unresolvedExecutionStarts(events) {
			if event.Capability == action.Capability && event.ArtifactHash == action.ArtifactHash {
				return DecisionResult{Status: DecisionUnknown, Reason: ReasonExecutionOutcomeUnknown, Requested: action.Capability, Message: fmt.Sprintf("execution intent for %s has no durable outcome receipt", action.Capability)}
			}
		}
		for _, event := range events {
			if event.Capability == action.Capability && event.Decision == DecisionDeny {
				return DecisionResult{Status: DecisionDeny, Reason: ReasonDeniedByPolicy, Requested: action.Capability, Message: fmt.Sprintf("evidence ledger shows denied execution for %s", action.Capability)}
			}
		}
	}
	return decision
}

// VerifySession reviews all recorded events for an agent and reports whether they are policy compliant.
func (v *SessionVerifier) VerifySession(agentID string) VerificationSummary {
	if v == nil || v.Ledger == nil {
		return VerificationSummary{AgentID: agentID, Result: "FAIL", PolicyCompliant: false, Issues: []string{"evidence ledger unavailable"}}
	}
	summary := VerificationSummary{AgentID: agentID, Result: "PASS", PolicyCompliant: true}
	events := v.Ledger.EventsForAgent(agentID)
	unresolvedStarts := unresolvedExecutionStarts(events)
	for _, event := range events {
		if event.EventType == "EXECUTION_UNKNOWN" && !executionHasValidReconciliation(events, event.ExecutionID) {
			summary.PolicyCompliant = false
			if summary.Result == "PASS" {
				summary.Result = "UNKNOWN"
			}
			summary.Issues = append(summary.Issues, fmt.Sprintf("unconfirmed execution %s for capability %s requires reconciliation", event.ID, event.Capability))
		}
		if event.Decision == DecisionDeny || event.EventType == "EXECUTION_BLOCKED" {
			summary.BlockedAttempts++
			summary.Issues = append(summary.Issues, fmt.Sprintf("blocked attempt %s for capability %s", event.ID, event.Capability))
		}
		if event.EventType == "EXECUTION_ALLOWED" && event.Decision != DecisionAllow {
			summary.Unauthorized++
			summary.UnauthorizedExecutions++
			summary.PolicyCompliant = false
			summary.Result = "FAIL"
			summary.Issues = append(summary.Issues, fmt.Sprintf("unauthorized execution %s for capability %s", event.ID, event.Capability))
		}
	}
	for _, event := range unresolvedStarts {
		summary.PolicyCompliant = false
		if summary.Result == "PASS" {
			summary.Result = "UNKNOWN"
		}
		summary.Issues = append(summary.Issues, fmt.Sprintf("execution %s was started but has no durable outcome receipt; reconciliation required", event.ExecutionID))
	}
	if summary.PolicyCompliant && len(summary.Issues) == 0 {
		summary.Issues = nil
	}
	return summary
}

type VerificationSummary struct {
	AgentID                string   `json:"agent_id"`
	Result                 string   `json:"result"`
	PolicyCompliant        bool     `json:"policy_compliant"`
	Unauthorized           int      `json:"unauthorized"`
	BlockedAttempts        int      `json:"blocked_attempts"`
	UnauthorizedExecutions int      `json:"unauthorized_executions"`
	Issues                 []string `json:"issues,omitempty"`
}

// VerifySession inspects ledger events for unauthorized or blocked actions.
func (l *Ledger) VerifySession(agentID string) VerificationSummary {
	summary := VerificationSummary{
		AgentID:         agentID,
		Result:          "PASS",
		PolicyCompliant: true,
		Issues:          []string{},
	}
	events := l.EventsForAgent(agentID)
	unresolvedStarts := unresolvedExecutionStarts(events)
	for _, event := range events {
		if event.EventType == "EXECUTION_UNKNOWN" && !executionHasValidReconciliation(events, event.ExecutionID) {
			summary.PolicyCompliant = false
			if summary.Result == "PASS" {
				summary.Result = "UNKNOWN"
			}
			summary.Issues = append(summary.Issues, fmt.Sprintf("unconfirmed execution %s for %s requires reconciliation", event.ID, event.Capability))
		}
		if event.Decision == DecisionDeny || event.EventType == "EXECUTION_BLOCKED" {
			summary.BlockedAttempts++
			summary.Issues = append(summary.Issues, fmt.Sprintf("blocked attempt %s for %s: %s", event.ID, event.Capability, event.Reason))
		}
		if event.EventType == "EXECUTION_ALLOWED" && event.Decision != DecisionAllow {
			summary.Unauthorized++
			summary.UnauthorizedExecutions++
			summary.PolicyCompliant = false
			summary.Result = "FAIL"
			summary.Issues = append(summary.Issues, fmt.Sprintf("unauthorized execution evidence %s for %s", event.ID, event.Capability))
		}
	}
	for _, event := range unresolvedStarts {
		summary.PolicyCompliant = false
		if summary.Result == "PASS" {
			summary.Result = "UNKNOWN"
		}
		summary.Issues = append(summary.Issues, fmt.Sprintf("execution %s was started but has no durable outcome receipt; reconciliation required", event.ExecutionID))
	}
	if summary.PolicyCompliant && len(summary.Issues) == 0 {
		summary.Issues = nil
	}
	return summary
}

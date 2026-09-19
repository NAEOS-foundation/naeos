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
		for _, event := range v.Ledger.EventsForAgent(agentID) {
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
	for _, event := range v.Ledger.EventsForAgent(agentID) {
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
	if summary.PolicyCompliant && len(summary.Issues) == 0 {
		summary.Issues = nil
	}
	return summary
}

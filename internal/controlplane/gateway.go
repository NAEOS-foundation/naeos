package controlplane

import (
	"fmt"
	"time"
)

// AuthorizeRequest models an authorization decision request at the execution boundary.
type AuthorizeRequest struct {
	RequestID  string               `json:"request_id"`
	DecisionID string               `json:"decision_id,omitempty"`
	AgentID    string               `json:"agent_id"`
	Action     Action               `json:"action"`
	Grant      *Grant               `json:"grant,omitempty"`
	Policy     *Policy              `json:"policy,omitempty"`
	Context    AuthorizationContext `json:"context,omitempty"`
	ApprovalID string               `json:"approval_id,omitempty"`
	Timestamp  time.Time            `json:"timestamp"`
}

// DecisionGateway is the fail-closed gate between proposed actions and execution.
type DecisionGateway struct {
	Evaluator *Evaluator
	Ledger    *Ledger
	Approvals *ApprovalStore
}

func controlPlaneID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}

// NewDecisionGateway creates a new execution gateway using the provided evaluator and ledger.
func NewDecisionGateway(e *Evaluator, l *Ledger) *DecisionGateway {
	if e == nil {
		e = NewEvaluator()
	}
	if l == nil {
		l = NewLedger()
	}
	return &DecisionGateway{Evaluator: e, Ledger: l, Approvals: NewApprovalStore()}
}

// Authorize decides whether a requested action may proceed.
func (g *DecisionGateway) Authorize(req AuthorizeRequest) DecisionResult {
	if g == nil || g.Evaluator == nil {
		return DecisionResult{
			Status: DecisionDeny,
			Reason: DecisionReason("gateway unavailable"),
		}
	}

	if req.Policy == nil {
		return DecisionResult{
			Status: DecisionDeny,
			Reason: DecisionReason("policy missing"),
		}
	}
	if req.Grant == nil {
		return DecisionResult{
			Status: DecisionDeny,
			Reason: DecisionReason("grant missing"),
		}
	}
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}
	if req.RequestID == "" {
		req.RequestID = controlPlaneID("REQ")
	}
	effectiveNow := req.Context.Now
	if effectiveNow.IsZero() {
		effectiveNow = req.Timestamp
	}

	result := g.Evaluator.EvaluateAction(req.Action, req.Grant, req.Policy, effectiveNow)
	result.RequestID = req.RequestID
	result.AgentID = req.AgentID
	result.ArtifactHash = req.Action.ArtifactHash
	if req.DecisionID == "" {
		result.DecisionID = controlPlaneID("DEC")
	} else {
		result.DecisionID = req.DecisionID
	}
	if result.Status == DecisionPending && req.ApprovalID != "" && g.Approvals != nil {
		if approval, err := g.Approvals.Get(req.ApprovalID); err == nil &&
			approval.DecisionID == result.DecisionID &&
			approval.AgentID == result.AgentID &&
			approval.Capability == result.Requested &&
			approval.PolicyID == result.PolicyID &&
			approval.PolicyVersion == result.PolicyVersion &&
			approval.ArtifactHash == req.Action.ArtifactHash &&
			approval.IsValid(effectiveNow) {
			result.Status = DecisionAllow
			result.Reason = ReasonAllowed
			result.NeedsApproval = false
			result.Message = "approval granted for pending decision"
		}
		if result.Status == DecisionPending {
			result.Status = DecisionDeny
			result.Reason = ReasonDeniedByPolicy
			result.Message = "approval is missing, expired, consumed, or not bound to this artifact"
		}
	}
	if g.Ledger != nil {
		policyID := "unknown"
		grantID := "unknown"
		if req.Policy != nil {
			policyID = req.Policy.ID
		}
		if req.Grant != nil {
			grantID = req.Grant.GrantID
		}
		g.Ledger.Append(LedgerEvent{
			Timestamp:    req.Timestamp,
			RequestID:    req.RequestID,
			DecisionID:   result.DecisionID,
			AgentID:      req.AgentID,
			Capability:   req.Action.Capability,
			ArtifactHash: req.Action.ArtifactHash,
			EventType:    "AUTHORIZATION_DECISION",
			Decision:     result.Status,
			Reason:       result.Reason,
			Metadata: map[string]string{
				"policy_id": policyID,
				"grant_id":  grantID,
			},
		})
	}
	return result
}

// Execute enforces the decision at the execution boundary. It returns a decision and the resulting ledger event.
func (g *DecisionGateway) Execute(req AuthorizeRequest) (DecisionResult, LedgerEvent) {
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}
	if req.RequestID == "" {
		req.RequestID = controlPlaneID("REQ")
	}
	result := g.Authorize(req)
	return g.executeDecision(req, result)
}

// ExecuteDecision records an already-authorized action at the execution boundary.
// Callers must obtain the decision from Authorize immediately before execution.
func (g *DecisionGateway) ExecuteDecision(req AuthorizeRequest, result DecisionResult) (DecisionResult, LedgerEvent) {
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}
	if req.RequestID == "" {
		req.RequestID = result.RequestID
	}
	return g.executeDecision(req, result)
}

func (g *DecisionGateway) executeDecision(req AuthorizeRequest, result DecisionResult) (DecisionResult, LedgerEvent) {
	if g.Ledger == nil {
		return result, LedgerEvent{}
	}

	policyID := "unknown"
	grantID := "unknown"
	if req.Policy != nil {
		policyID = req.Policy.ID
	}
	if req.Grant != nil {
		grantID = req.Grant.GrantID
	}

	if result.Status == DecisionAllow {
		if req.ApprovalID != "" && g.Approvals != nil {
			if _, err := g.Approvals.Consume(req.ApprovalID, req.Timestamp); err != nil {
				result.Status = DecisionDeny
				result.Reason = ReasonDeniedByPolicy
				result.Message = fmt.Sprintf("approval could not be consumed: %v", err)
			}
		}
	}

	if result.Status == DecisionAllow {
		return result, g.Ledger.Append(LedgerEvent{
			Timestamp:    req.Timestamp,
			RequestID:    req.RequestID,
			DecisionID:   result.DecisionID,
			ExecutionID:  controlPlaneID("EXEC"),
			AgentID:      req.AgentID,
			Capability:   req.Action.Capability,
			ArtifactHash: req.Action.ArtifactHash,
			EventType:    "EXECUTION_ALLOWED",
			Decision:     DecisionAllow,
			Reason:       result.Reason,
			Metadata: map[string]string{
				"policy_id": policyID,
				"grant_id":  grantID,
			},
		})
	}

	return result, g.Ledger.Append(LedgerEvent{
		Timestamp:    req.Timestamp,
		RequestID:    req.RequestID,
		DecisionID:   result.DecisionID,
		ExecutionID:  controlPlaneID("EXEC"),
		AgentID:      req.AgentID,
		Capability:   req.Action.Capability,
		ArtifactHash: req.Action.ArtifactHash,
		EventType:    "EXECUTION_BLOCKED",
		Decision:     DecisionDeny,
		Reason:       result.Reason,
		Metadata: map[string]string{
			"policy_id": policyID,
			"grant_id":  grantID,
		},
	})
}

// String is a concise human-readable summary for the decision gateway.
func (g *DecisionGateway) String() string {
	if g == nil {
		return "DecisionGateway(nil)"
	}
	return fmt.Sprintf("DecisionGateway{evaluator=%T ledger=%T}", g.Evaluator, g.Ledger)
}

// RequestApproval creates an approval record for a REQUIRE_APPROVAL decision.
func (g *DecisionGateway) RequestApproval(result DecisionResult, approver string, expiresAt time.Time) (Approval, error) {
	if g == nil || g.Approvals == nil {
		return Approval{}, fmt.Errorf("approval store unavailable")
	}
	if result.Status != DecisionPending {
		return Approval{}, fmt.Errorf("decision %s does not require approval", result.DecisionID)
	}
	if g.Ledger == nil {
		return Approval{}, fmt.Errorf("control plane ledger unavailable")
	}
	event, ok := g.Ledger.Decision(result.DecisionID)
	if !ok {
		return Approval{}, fmt.Errorf("decision %s not found in control plane ledger", result.DecisionID)
	}
	if event.Decision != DecisionPending {
		return Approval{}, fmt.Errorf("decision %s is not pending approval", result.DecisionID)
	}
	if result.AgentID != "" && event.AgentID != result.AgentID {
		return Approval{}, fmt.Errorf("decision %s agent binding mismatch", result.DecisionID)
	}
	if result.Requested != "" && event.Capability != result.Requested {
		return Approval{}, fmt.Errorf("decision %s capability binding mismatch", result.DecisionID)
	}
	if result.PolicyID != "" && event.Metadata["policy_id"] != result.PolicyID {
		return Approval{}, fmt.Errorf("decision %s policy binding mismatch", result.DecisionID)
	}
	if result.ArtifactHash != "" && event.ArtifactHash != result.ArtifactHash {
		return Approval{}, fmt.Errorf("decision %s artifact binding mismatch", result.DecisionID)
	}
	result.AgentID = event.AgentID
	result.Requested = event.Capability
	result.PolicyID = event.Metadata["policy_id"]
	if result.ArtifactHash == "" {
		result.ArtifactHash = event.ArtifactHash
	}
	if version := event.Metadata["policy_version"]; version != "" {
		_, _ = fmt.Sscanf(version, "%d", &result.PolicyVersion)
	}
	return g.Approvals.Create(result, approver, expiresAt)
}

// Approve marks a pending approval as approved.
func (g *DecisionGateway) Approve(approvalID, reason, artifactHash string, now time.Time) (Approval, error) {
	if g == nil || g.Approvals == nil {
		return Approval{}, fmt.Errorf("approval store unavailable")
	}
	return g.Approvals.Approve(approvalID, reason, artifactHash, now)
}

// Reject records an explicit rejection for a pending approval.
func (g *DecisionGateway) Reject(approvalID, reason string, now time.Time) (Approval, error) {
	if g == nil || g.Approvals == nil {
		return Approval{}, fmt.Errorf("approval store unavailable")
	}
	return g.Approvals.Reject(approvalID, reason, now)
}

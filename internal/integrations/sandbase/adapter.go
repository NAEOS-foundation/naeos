// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package sandbase

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

// AuthorizationRequest is the protocol-neutral boundary proposed for SandBase Harness.
// It deliberately contains no SandBase-specific implementation types.
type AuthorizationRequest struct {
	SessionID    string    `json:"session_id"`
	RequestID    string    `json:"request_id"`
	AgentID      string    `json:"agent_id"`
	Capability   string    `json:"capability"`
	Target       string    `json:"target"`
	ArtifactHash string    `json:"artifact_hash,omitempty"`
	Context      any       `json:"context,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// AuthorizationDecision is the minimal decision envelope an execution runtime needs.
type AuthorizationDecision struct {
	Decision      string `json:"decision"`
	PolicyVersion int    `json:"policy_version"`
	Capability    string `json:"capability"`
	Target        string `json:"target"`
	ExpiresAt     string `json:"expires_at"`
	ContextDigest string `json:"context_digest"`
	RequestID     string `json:"request_id"`
	DecisionID    string `json:"decision_id,omitempty"`
	Reason        string `json:"reason"`
}

// Adapter translates the external authorization contract into the NAEOS control plane.
// It resolves the currently active policy on every request so a long-running session
// cannot continue receiving ALLOW from a stale policy pointer.
type Adapter struct {
	Gateway *controlplane.DecisionGateway
	Policy  *controlplane.Policy
	Grant   *controlplane.Grant
}

func (a *Adapter) Authorize(req AuthorizationRequest) (AuthorizationDecision, error) {
	if a == nil || a.Gateway == nil || a.Gateway.Evaluator == nil || a.Policy == nil || a.Grant == nil {
		return AuthorizationDecision{}, fmt.Errorf("sandbase authorization adapter is not configured")
	}
	if req.AgentID == "" || req.Capability == "" || req.Target == "" {
		return AuthorizationDecision{}, fmt.Errorf("agent_id, capability, and target are required")
	}
	if req.RequestID == "" {
		return AuthorizationDecision{}, fmt.Errorf("request_id is required")
	}
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}

	activePolicy, err := a.Gateway.Evaluator.ActivePolicy(a.Policy.ID)
	if err != nil {
		return AuthorizationDecision{}, fmt.Errorf("resolve active authorization policy: %w", err)
	}

	payload := map[string]string{"target": req.Target, "session_id": req.SessionID}
	if contextFields, ok := req.Context.(map[string]string); ok {
		for key, value := range contextFields {
			payload["authorization_context."+key] = value
		}
	}
	action := controlplane.Action{
		AgentID:      req.AgentID,
		Capability:   controlplane.Capability(req.Capability),
		ArtifactHash: req.ArtifactHash,
		Payload:      payload,
	}
	decision := a.Gateway.Authorize(controlplane.AuthorizeRequest{
		RequestID: req.RequestID,
		AgentID:   req.AgentID,
		Action:    action,
		Grant:     a.Grant,
		Policy:    activePolicy,
		Timestamp: req.Timestamp,
	})
	if err := a.Gateway.Ledger.PersistenceError(); err != nil {
		return AuthorizationDecision{}, fmt.Errorf("persist authorization decision: %w", err)
	}

	contextDigest, err := digest(req.Context)
	if err != nil {
		return AuthorizationDecision{}, err
	}

	return AuthorizationDecision{
		Decision:      string(decision.Status),
		PolicyVersion: decision.PolicyVersion,
		Capability:    req.Capability,
		Target:        req.Target,
		ExpiresAt:     a.Grant.ExpiresAt.UTC().Format(time.RFC3339),
		ContextDigest: contextDigest,
		RequestID:     decision.RequestID,
		DecisionID:    decision.DecisionID,
		Reason:        string(decision.Reason),
	}, nil
}

func digest(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal authorization context: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

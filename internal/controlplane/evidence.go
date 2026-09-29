// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// EvidenceBundle is a canonical, verifier-facing record for one authorization lifecycle.
type EvidenceBundle struct {
	SchemaVersion  string               `json:"schema_version"`
	RequestID      string               `json:"request_id"`
	DecisionID     string               `json:"decision_id"`
	ExecutionID    string               `json:"execution_id,omitempty"`
	AgentID        string               `json:"agent_id"`
	Capability     Capability           `json:"capability"`
	ArtifactHash   string               `json:"artifact_hash,omitempty"`
	Decision       DecisionStatus       `json:"decision"`
	Reason         DecisionReason       `json:"reason"`
	PolicyID       string               `json:"policy_id,omitempty"`
	PolicyVersion  string               `json:"policy_version,omitempty"`
	GrantID        string               `json:"grant_id,omitempty"`
	DecisionEvent  LedgerEvent          `json:"decision_event"`
	ExecutionEvent *LedgerEvent         `json:"execution_event,omitempty"`
	Verification   EvidenceVerification `json:"verification"`
	EvidenceDigest string               `json:"evidence_digest"`
}

// EvidenceVerification describes deterministic checks over the evidence lifecycle.
type EvidenceVerification struct {
	Result              string   `json:"result"`
	DecisionConsistent  bool     `json:"decision_consistent"`
	ExecutionConsistent bool     `json:"execution_consistent"`
	LedgerIntegrity     bool     `json:"ledger_integrity"`
	Issues              []string `json:"issues,omitempty"`
}

// BuildEvidence materializes a canonical evidence bundle for one decision.
func (l *Ledger) BuildEvidence(decisionID string) (EvidenceBundle, error) {
	if l == nil {
		return EvidenceBundle{}, fmt.Errorf("ledger unavailable")
	}
	decision, ok := l.Decision(decisionID)
	if !ok {
		return EvidenceBundle{}, fmt.Errorf("decision %q not found", decisionID)
	}

	bundle := EvidenceBundle{
		SchemaVersion: "1.0",
		RequestID:     decision.RequestID,
		DecisionID:    decision.DecisionID,
		AgentID:       decision.AgentID,
		Capability:    decision.Capability,
		ArtifactHash:  decision.ArtifactHash,
		Decision:      decision.Decision,
		Reason:        decision.Reason,
		PolicyID:      decision.Metadata["policy_id"],
		PolicyVersion: decision.Metadata["policy_version"],
		GrantID:       decision.Metadata["grant_id"],
		DecisionEvent: decision,
	}

	for _, event := range l.Events() {
		if event.DecisionID != decisionID || (event.EventType != "EXECUTION_ALLOWED" && event.EventType != "EXECUTION_BLOCKED") {
			continue
		}
		copyEvent := event
		bundle.ExecutionID = event.ExecutionID
		bundle.ExecutionEvent = &copyEvent
		break
	}

	bundle.Verification = l.verifyEvidenceBundle(bundle)
	digest, err := evidenceDigest(bundle)
	if err != nil {
		return EvidenceBundle{}, err
	}
	bundle.EvidenceDigest = digest
	return bundle, nil
}

// VerifyEvidence independently validates a previously materialized bundle.
func VerifyEvidence(bundle EvidenceBundle) EvidenceVerification {
	verification := EvidenceVerification{
		Result:              "PASS",
		DecisionConsistent:  true,
		ExecutionConsistent: true,
		LedgerIntegrity:     true,
	}

	if bundle.DecisionEvent.DecisionID != bundle.DecisionID ||
		bundle.DecisionEvent.RequestID != bundle.RequestID ||
		bundle.DecisionEvent.AgentID != bundle.AgentID ||
		bundle.DecisionEvent.Capability != bundle.Capability ||
		bundle.DecisionEvent.ArtifactHash != bundle.ArtifactHash {
		verification.DecisionConsistent = false
		verification.Issues = append(verification.Issues, "decision evidence does not match bundle identity")
	}

	if bundle.ExecutionEvent != nil {
		exec := bundle.ExecutionEvent
		if exec.DecisionID != bundle.DecisionID || exec.RequestID != bundle.RequestID ||
			exec.AgentID != bundle.AgentID || exec.Capability != bundle.Capability ||
			exec.ArtifactHash != bundle.ArtifactHash || exec.ExecutionID != bundle.ExecutionID {
			verification.ExecutionConsistent = false
			verification.Issues = append(verification.Issues, "execution evidence does not match bundle identity")
		}
		if bundle.Decision == DecisionAllow && exec.EventType != "EXECUTION_ALLOWED" {
			verification.ExecutionConsistent = false
			verification.Issues = append(verification.Issues, "allowed decision lacks allowed execution evidence")
		}
	}
	if bundle.Decision == DecisionAllow && bundle.ExecutionEvent == nil {
		verification.ExecutionConsistent = false
		verification.Issues = append(verification.Issues, "allowed decision has no execution evidence")
	}
	if bundle.Decision == DecisionDeny && bundle.ExecutionEvent != nil && bundle.ExecutionEvent.EventType == "EXECUTION_ALLOWED" {
		verification.ExecutionConsistent = false
		verification.Issues = append(verification.Issues, "denied decision has allowed execution evidence")
	}

	expected, err := evidenceDigest(bundle)
	if err != nil || expected != bundle.EvidenceDigest {
		verification.LedgerIntegrity = false
		verification.Issues = append(verification.Issues, "evidence digest mismatch")
	}
	if len(verification.Issues) > 0 {
		verification.Result = "FAIL"
	}
	return verification
}

func (l *Ledger) verifyEvidenceBundle(bundle EvidenceBundle) EvidenceVerification {
	v := EvidenceVerification{
		Result:              "PASS",
		DecisionConsistent:  true,
		ExecutionConsistent: true,
		LedgerIntegrity:     true,
	}

	if bundle.DecisionEvent.DecisionID != bundle.DecisionID ||
		bundle.DecisionEvent.RequestID != bundle.RequestID ||
		bundle.DecisionEvent.AgentID != bundle.AgentID ||
		bundle.DecisionEvent.Capability != bundle.Capability ||
		bundle.DecisionEvent.ArtifactHash != bundle.ArtifactHash {
		v.DecisionConsistent = false
		v.Issues = append(v.Issues, "decision evidence does not match bundle identity")
	}

	if bundle.ExecutionEvent != nil {
		exec := bundle.ExecutionEvent
		if exec.DecisionID != bundle.DecisionID || exec.RequestID != bundle.RequestID ||
			exec.AgentID != bundle.AgentID || exec.Capability != bundle.Capability ||
			exec.ArtifactHash != bundle.ArtifactHash || exec.ExecutionID != bundle.ExecutionID {
			v.ExecutionConsistent = false
			v.Issues = append(v.Issues, "execution evidence does not match bundle identity")
		}
		if bundle.Decision == DecisionAllow && exec.EventType != "EXECUTION_ALLOWED" {
			v.ExecutionConsistent = false
			v.Issues = append(v.Issues, "allowed decision lacks allowed execution evidence")
		}
	}
	if bundle.Decision == DecisionAllow && bundle.ExecutionEvent == nil {
		v.ExecutionConsistent = false
		v.Issues = append(v.Issues, "allowed decision has no execution evidence")
	}
	if bundle.Decision == DecisionDeny && bundle.ExecutionEvent != nil && bundle.ExecutionEvent.EventType == "EXECUTION_ALLOWED" {
		v.ExecutionConsistent = false
		v.Issues = append(v.Issues, "denied decision has allowed execution evidence")
	}

	var previous string
	for _, event := range l.Events() {
		if event.PreviousHash != previous || event.EventHash != hashLedgerEvent(event) {
			v.LedgerIntegrity = false
			v.Issues = append(v.Issues, fmt.Sprintf("ledger hash chain invalid at %s", event.ID))
			break
		}
		previous = event.EventHash
	}

	if !v.DecisionConsistent || !v.ExecutionConsistent || !v.LedgerIntegrity {
		v.Result = "FAIL"
	}
	return v
}

func evidenceDigest(bundle EvidenceBundle) (string, error) {
	bundle.EvidenceDigest = ""
	bundle.Verification.Issues = nil
	data, err := json.Marshal(bundle)
	if err != nil {
		return "", fmt.Errorf("marshal evidence bundle: %w", err)
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

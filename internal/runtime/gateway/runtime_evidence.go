// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

// RuntimeEvidence is the canonical, protocol-neutral receipt emitted after a
// governed execution. It binds authorization, the exact invocation, and the
// independently observed outcome without trusting agent/sandbox claims alone.
type RuntimeEvidence struct {
	SchemaVersion    string       `json:"schema_version"`
	RequestID        string       `json:"request_id"`
	InvocationID     string       `json:"invocation_id"`
	InvocationDigest string       `json:"invocation_digest"`
	Tool             string       `json:"tool"`
	Action           string       `json:"action"`
	Resource         string       `json:"resource,omitempty"`
	Environment      string       `json:"environment,omitempty"`
	Actor            string       `json:"actor,omitempty"`
	Capability       string       `json:"capability,omitempty"`
	PolicyID         string       `json:"policy_id,omitempty"`
	PolicyVersion    string       `json:"policy_version,omitempty"`
	RuleID           string       `json:"rule_id,omitempty"`
	Decision         string       `json:"decision"`
	ExecutionStatus  string       `json:"execution_status"`
	ExecutionHash    string       `json:"execution_hash,omitempty"`
	Observation      *Observation `json:"observation,omitempty"`
	EvidenceDigest   string       `json:"evidence_digest"`
}

// InvocationDigest returns a deterministic SHA-256 digest over the complete
// governed invocation identity, including canonicalized payload/context.
func InvocationDigest(req ToolRequest) string {
	type canonicalRequest struct {
		RequestID    string         `json:"request_id"`
		InvocationID string         `json:"invocation_id"`
		Capability   string         `json:"capability"`
		Tool         string         `json:"tool"`
		Action       string         `json:"action"`
		Resource     string         `json:"resource"`
		Environment  string         `json:"environment"`
		Actor        string         `json:"actor"`
		Payload      map[string]any `json:"payload"`
		Context      map[string]any `json:"context"`
	}
	data, _ := json.Marshal(canonicalRequest(req))
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// BuildRuntimeEvidence creates the canonical receipt for an execution result.
// A completed ALLOW is evidence-backed only when the observation is present,
// observed, and identity-bound.
func BuildRuntimeEvidence(result ExecutionResult) (RuntimeEvidence, error) {
	if result.Request.InvocationID == "" {
		return RuntimeEvidence{}, fmt.Errorf("invocation_id is required")
	}
	e := RuntimeEvidence{
		SchemaVersion:    "naeos.runtime-evidence.v1",
		RequestID:        result.RequestID,
		InvocationID:     result.InvocationID,
		InvocationDigest: InvocationDigest(result.Request),
		Tool:             result.Request.Tool,
		Action:           result.Request.Action,
		Resource:         result.Request.Resource,
		Environment:      result.Request.Environment,
		Actor:            result.Request.Actor,
		Capability:       result.Request.Capability,
		PolicyID:         result.PolicyID,
		PolicyVersion:    result.PolicyVersion,
		RuleID:           result.RuleID,
		Decision:         string(result.Decision),
		ExecutionStatus:  result.Status,
		ExecutionHash:    result.Hash,
		Observation:      result.Observation,
	}
	if result.Status == "completed" && result.Decision == control.DecisionAllow {
		if result.Observation == nil || !result.Observation.Observed ||
			result.Observation.RequestID != result.RequestID ||
			result.Observation.InvocationID != result.InvocationID ||
			result.Observation.InvocationDigest != e.InvocationDigest {
			return RuntimeEvidence{}, fmt.Errorf("completed execution lacks identity-bound observation")
		}
	}
	data, _ := json.Marshal(e)
	sum := sha256.Sum256(data)
	e.EvidenceDigest = "sha256:" + hex.EncodeToString(sum[:])
	return e, nil
}

// VerifyRuntimeEvidence independently recomputes the invocation and evidence
// digests and checks the authorization/execution/observation binding.
func VerifyRuntimeEvidence(e RuntimeEvidence) error {
	if e.SchemaVersion != "naeos.runtime-evidence.v1" {
		return fmt.Errorf("unsupported evidence schema %q", e.SchemaVersion)
	}
	if e.RequestID == "" || e.InvocationID == "" || e.InvocationDigest == "" {
		return fmt.Errorf("evidence identity is incomplete")
	}
	if e.Decision == string(control.DecisionAllow) && e.ExecutionStatus == "completed" {
		if e.Observation == nil || !e.Observation.Observed ||
			e.Observation.RequestID != e.RequestID ||
			e.Observation.InvocationID != e.InvocationID ||
			e.Observation.InvocationDigest != e.InvocationDigest {
			return fmt.Errorf("successful evidence is not observation-bound")
		}
	}
	got := e.EvidenceDigest
	e.EvidenceDigest = ""
	data, _ := json.Marshal(e)
	sum := sha256.Sum256(data)
	e.EvidenceDigest = got
	expected := "sha256:" + hex.EncodeToString(sum[:])
	if got != expected {
		return fmt.Errorf("evidence digest mismatch: expected %s got %s", expected, got)
	}
	return nil
}

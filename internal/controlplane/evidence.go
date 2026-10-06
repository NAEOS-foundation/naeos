// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
)

// EvidenceBundle is a canonical, verifier-facing record for one authorization lifecycle.
type EvidenceBundle struct {
	SchemaVersion              string               `json:"schema_version"`
	RequestID                  string               `json:"request_id"`
	DecisionID                 string               `json:"decision_id"`
	ExecutionID                string               `json:"execution_id,omitempty"`
	AgentID                    string               `json:"agent_id"`
	Capability                 Capability           `json:"capability"`
	ArtifactHash               string               `json:"artifact_hash,omitempty"`
	Decision                   DecisionStatus       `json:"decision"`
	Reason                     DecisionReason       `json:"reason"`
	PolicyID                   string               `json:"policy_id,omitempty"`
	PolicyVersion              string               `json:"policy_version,omitempty"`
	GrantID                    string               `json:"grant_id,omitempty"`
	DecisionEvent              LedgerEvent          `json:"decision_event"`
	ExecutionEvent             *LedgerEvent         `json:"execution_event,omitempty"`
	ObservationEvent           *LedgerEvent         `json:"observation_event,omitempty"`
	Verification               EvidenceVerification `json:"verification"`
	EvidenceDigest             string               `json:"evidence_digest"`
	EvidenceSignature          string               `json:"evidence_signature,omitempty"`
	EvidencePublicKey          string               `json:"evidence_public_key,omitempty"`
	EvidenceSignatureAlgorithm string               `json:"evidence_signature_algorithm,omitempty"`
}

// EvidenceVerification describes deterministic checks over the evidence lifecycle.
type EvidenceVerification struct {
	Result              string   `json:"result"`
	DecisionConsistent  bool     `json:"decision_consistent"`
	ExecutionConsistent bool     `json:"execution_consistent"`
	ObservationConsistent bool`json:"observation_consistent"`
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
		SchemaVersion: "1.2",
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
		if event.DecisionID != decisionID {
			continue
		}
		if (event.EventType == "EXECUTION_ALLOWED" || event.EventType == "EXECUTION_BLOCKED") && bundle.ExecutionEvent == nil {
			copyEvent := event
			bundle.ExecutionID = event.ExecutionID
			bundle.ExecutionEvent = &copyEvent
		}
		if event.EventType == "SIDE_EFFECT_OBSERVED" && bundle.ObservationEvent == nil {
			copyEvent := event
			bundle.ObservationEvent = &copyEvent
		}
	}

	bundle.Verification = l.verifyEvidenceBundle(bundle)
	digest, err := evidenceDigest(bundle)
	if err != nil {
		return EvidenceBundle{}, err
	}
	bundle.EvidenceDigest = digest
	if err := signEvidenceBundle(&bundle); err != nil {
		return EvidenceBundle{}, err
	}
	return bundle, nil
}

// VerifyEvidence independently validates a previously materialized bundle.
func VerifyEvidence(bundle EvidenceBundle) EvidenceVerification {
	verification := EvidenceVerification{
		Result:              "PASS",
		DecisionConsistent:    true,
		ExecutionConsistent:   true,
		ObservationConsistent: true,
		LedgerIntegrity:       true,
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

	if bundle.Decision == DecisionAllow {
		if bundle.ObservationEvent == nil {
			verification.ObservationConsistent = false
			verification.Issues = append(verification.Issues, "allowed decision has no observed side-effect evidence")
		} else {
			obs := bundle.ObservationEvent
			if obs.DecisionID != bundle.DecisionID || obs.RequestID != bundle.RequestID ||
				obs.AgentID != bundle.AgentID || obs.Capability != bundle.Capability ||
				obs.ArtifactHash != bundle.ArtifactHash || obs.EventType != "SIDE_EFFECT_OBSERVED" {
				verification.ObservationConsistent = false
				verification.Issues = append(verification.Issues, "observation evidence does not match bundle identity")
			}
		}
	}

	expected, err := evidenceDigest(bundle)
	if err != nil || expected != bundle.EvidenceDigest {
		verification.LedgerIntegrity = false
		verification.Issues = append(verification.Issues, "evidence digest mismatch")
	}
	if !verifyEvidenceSignature(bundle) {
		verification.LedgerIntegrity = false
		verification.Issues = append(verification.Issues, "evidence signature invalid or missing")
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

	if bundle.Decision == DecisionAllow {
		if bundle.ObservationEvent == nil {
			v.ObservationConsistent = false
			v.Issues = append(v.Issues, "allowed decision has no observed side-effect evidence")
		} else if bundle.ObservationEvent.DecisionID != bundle.DecisionID ||
			bundle.ObservationEvent.RequestID != bundle.RequestID ||
			bundle.ObservationEvent.AgentID != bundle.AgentID ||
			bundle.ObservationEvent.Capability != bundle.Capability ||
			bundle.ObservationEvent.ArtifactHash != bundle.ArtifactHash ||
			bundle.ObservationEvent.EventType != "SIDE_EFFECT_OBSERVED" {
			v.ObservationConsistent = false
			v.Issues = append(v.Issues, "observation evidence does not match bundle identity")
		}
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

	if !v.DecisionConsistent || !v.ExecutionConsistent || !v.ObservationConsistent || !v.LedgerIntegrity {
		v.Result = "FAIL"
	}
	return v
}

func evidenceDigest(bundle EvidenceBundle) (string, error) {
	bundle.EvidenceDigest = ""
	bundle.EvidenceSignature = ""
	bundle.EvidencePublicKey = ""
	bundle.EvidenceSignatureAlgorithm = ""
	bundle.Verification.Issues = nil
	data, err := json.Marshal(bundle)
	if err != nil {
		return "", fmt.Errorf("marshal evidence bundle: %w", err)
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

var (
	evidenceSigningKey    ed25519.PrivateKey
	evidenceSigningPublic ed25519.PublicKey
	evidenceSigningOnce   sync.Once
	evidenceSigningErr    error
)

func evidenceSigningKeys() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	evidenceSigningOnce.Do(func() {
		evidenceSigningPublic, evidenceSigningKey, evidenceSigningErr = ed25519.GenerateKey(rand.Reader)
	})
	if evidenceSigningErr != nil {
		return nil, nil, fmt.Errorf("generate evidence signing key: %w", evidenceSigningErr)
	}
	return evidenceSigningKey, evidenceSigningPublic, nil
}

func signEvidenceBundle(bundle *EvidenceBundle) error {
	if bundle == nil {
		return fmt.Errorf("evidence bundle is required")
	}
	private, public, err := evidenceSigningKeys()
	if err != nil {
		return err
	}
	sig := ed25519.Sign(private, []byte(bundle.EvidenceDigest))
	bundle.EvidenceSignature = base64.RawStdEncoding.EncodeToString(sig)
	bundle.EvidencePublicKey = base64.RawStdEncoding.EncodeToString(public)
	bundle.EvidenceSignatureAlgorithm = "Ed25519"
	return nil
}

func verifyEvidenceSignature(bundle EvidenceBundle) bool {
	if bundle.EvidenceSignatureAlgorithm != "Ed25519" || bundle.EvidenceDigest == "" || bundle.EvidenceSignature == "" || bundle.EvidencePublicKey == "" {
		return false
	}
	public, err := base64.RawStdEncoding.DecodeString(bundle.EvidencePublicKey)
	if err != nil || len(public) != ed25519.PublicKeySize {
		return false
	}
	sig, err := base64.RawStdEncoding.DecodeString(bundle.EvidenceSignature)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(ed25519.PublicKey(public), []byte(bundle.EvidenceDigest), sig)
}

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

// EvidenceSignature authenticates a canonical runtime evidence receipt.
// The signature covers the canonical evidence payload including EvidenceDigest,
// but excludes this signature field itself.
type EvidenceSignature struct {
	Algorithm string `json:"algorithm"`
	Issuer    string `json:"issuer"`
	KeyID     string `json:"key_id"`
	Value     string `json:"value"`
}

// RuntimeEvidence is the canonical, protocol-neutral receipt emitted after a
// governed execution. It binds authorization, the exact invocation, and the
// independently observed outcome without trusting agent/sandbox claims alone.
type RuntimeEvidence struct {
	SchemaVersion    string             `json:"schema_version"`
	RequestID        string             `json:"request_id"`
	InvocationID     string             `json:"invocation_id"`
	InvocationDigest string             `json:"invocation_digest"`
	Tool             string             `json:"tool"`
	Action           string             `json:"action"`
	Resource         string             `json:"resource,omitempty"`
	Environment      string             `json:"environment,omitempty"`
	Actor            string             `json:"actor,omitempty"`
	Capability       string             `json:"capability,omitempty"`
	PolicyID         string             `json:"policy_id,omitempty"`
	PolicyVersion    string             `json:"policy_version,omitempty"`
	RuleID           string             `json:"rule_id,omitempty"`
	Decision         string             `json:"decision"`
	ExecutionStatus  string             `json:"execution_status"`
	ExecutionHash    string             `json:"execution_hash,omitempty"`
	Observation      *Observation       `json:"observation,omitempty"`
	EvidenceDigest   string             `json:"evidence_digest"`
	Signature        *EvidenceSignature `json:"signature,omitempty"`
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
	return sealEvidenceDigest(e), nil
}

func sealEvidenceDigest(e RuntimeEvidence) RuntimeEvidence {
	e.EvidenceDigest = ""
	e.Signature = nil
	data, _ := json.Marshal(e)
	sum := sha256.Sum256(data)
	e.EvidenceDigest = "sha256:" + hex.EncodeToString(sum[:])
	return e
}

func canonicalEvidenceForSignature(e RuntimeEvidence) ([]byte, error) {
	if e.EvidenceDigest == "" {
		return nil, fmt.Errorf("evidence digest is required")
	}
	e.Signature = nil
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SignRuntimeEvidence authenticates a receipt using Ed25519. The signature is
// over the canonical receipt including its digest, so digest recomputation
// after tampering cannot forge a valid receipt without the signing key.
func SignRuntimeEvidence(e *RuntimeEvidence, privateKey ed25519.PrivateKey, issuer, keyID string) error {
	if e == nil {
		return fmt.Errorf("evidence is nil")
	}
	if len(privateKey) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid Ed25519 private key length")
	}
	if strings.TrimSpace(issuer) == "" || strings.TrimSpace(keyID) == "" {
		return fmt.Errorf("issuer and key_id are required")
	}
	payload, err := canonicalEvidenceForSignature(*e)
	if err != nil {
		return err
	}
	signature := ed25519.Sign(privateKey, payload)
	e.Signature = &EvidenceSignature{
		Algorithm: "Ed25519",
		Issuer:    issuer,
		KeyID:     keyID,
		Value:     base64.StdEncoding.EncodeToString(signature),
	}
	return nil
}

// VerifyRuntimeEvidence independently verifies receipt integrity. When a
// trusted public key is supplied, it also verifies issuer/key identity and
// cryptographic authenticity.
func VerifyRuntimeEvidence(e RuntimeEvidence) error {
	return verifyRuntimeEvidence(e, nil, "", "")
}

// VerifyRuntimeEvidenceWithPublicKey verifies the receipt against the supplied
// trusted Ed25519 public key and expected issuer/key ID.
func VerifyRuntimeEvidenceWithPublicKey(e RuntimeEvidence, publicKey ed25519.PublicKey, issuer, keyID string) error {
	return verifyRuntimeEvidence(e, publicKey, issuer, keyID)
}

func verifyRuntimeEvidence(e RuntimeEvidence, publicKey ed25519.PublicKey, issuer, keyID string) error {
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
	e.Signature = nil
	data, _ := json.Marshal(e)
	sum := sha256.Sum256(data)
	expected := "sha256:" + hex.EncodeToString(sum[:])
	e.EvidenceDigest = got
	if got != expected {
		return fmt.Errorf("evidence digest mismatch: expected %s got %s", expected, got)
	}
	if publicKey == nil {
		return nil
	}
	if e.Signature == nil {
		return fmt.Errorf("evidence signature is required")
	}
	if e.Signature.Algorithm != "Ed25519" || e.Signature.Issuer != issuer || e.Signature.KeyID != keyID {
		return fmt.Errorf("evidence signature trust metadata mismatch")
	}
	sig, err := base64.StdEncoding.DecodeString(e.Signature.Value)
	if err != nil {
		return fmt.Errorf("decode evidence signature: %w", err)
	}
	payload, err := canonicalEvidenceForSignature(e)
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, payload, sig) {
		return fmt.Errorf("evidence signature verification failed")
	}
	return nil
}

// ParseEd25519PrivateKey accepts raw, base64, or hexadecimal Ed25519 private
// key material. Raw 32-byte seeds are expanded using ed25519.NewKeyFromSeed.
func ParseEd25519PrivateKey(data []byte) (ed25519.PrivateKey, error) {
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == ed25519.SeedSize {
		return ed25519.NewKeyFromSeed(data), nil
	}
	if len(data) == ed25519.PrivateKeySize {
		return ed25519.PrivateKey(data), nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(string(data)); err == nil {
		if len(decoded) == ed25519.SeedSize {
			return ed25519.NewKeyFromSeed(decoded), nil
		}
		if len(decoded) == ed25519.PrivateKeySize {
			return ed25519.PrivateKey(decoded), nil
		}
	}
	if decoded, err := hex.DecodeString(string(data)); err == nil {
		if len(decoded) == ed25519.SeedSize {
			return ed25519.NewKeyFromSeed(decoded), nil
		}
		if len(decoded) == ed25519.PrivateKeySize {
			return ed25519.PrivateKey(decoded), nil
		}
	}
	return nil, fmt.Errorf("unsupported Ed25519 private key encoding")
}

// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package investordemo

import (
	"testing"
	"time"
)

func TestHandoffRejectsDownstreamAuthorizedCapabilityWidening(t *testing.T) {
	setup := SetupDemoEnvironment()
	parent := &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              "agent-payment-01",
		Recipient:              "agent-secondary-02",
		RequestedCapability:    "repository.write",
		AuthorizedCapabilities: []Capability{"repository.read", "repository.write", "test.execute"},
		Payload:                map[string]interface{}{},
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		Provenance:             map[string]interface{}{"source": "agent-payment-01", "destination": "agent-secondary-02"},
		CreatedAt:              time.Now().UTC(),
		ExpiresAt:              time.Now().UTC().Add(time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: "abv1-10-downstream-widening", Timestamp: time.Now().UTC()},
	}

	parent.PayloadDigest = calculatePayloadDigest(parent.Payload)
	parent.DownstreamHandoff = &HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              "agent-secondary-02",
		Recipient:              "agent-tertiary-03",
		RequestedCapability:    "repository.write",
		AuthorizedCapabilities: []Capability{"repository.write", "credential.rotate"},
		Payload:                map[string]interface{}{},
		PolicyID:               "POLICY-017",
		PolicyVersion:          17,
		Provenance:             map[string]interface{}{"source": "agent-secondary-02"},
		CreatedAt:              time.Now().UTC(),
		ExpiresAt:              time.Now().UTC().Add(time.Hour),
		ReplayProtection:       ReplayProtection{Nonce: "abv1-10-downstream-child", Timestamp: time.Now().UTC()},
	}
	parent.DownstreamHandoff.PayloadDigest = calculatePayloadDigest(parent.DownstreamHandoff.Payload)
	setup.HandoffValidator.SignContract(parent)

	validation := setup.HandoffValidator.ValidateHandoff(parent)
	if validation.Valid {
		t.Fatal("expected downstream capability widening to be rejected")
	}
	if !validation.CapabilityWideningDetected {
		t.Fatal("expected capability widening to be detected")
	}
}


func TestHandoffRejectsDownstreamMutationAfterParentSigning(t *testing.T) {
	setup := SetupDemoEnvironment()
	parent := &HandoffContract{
		ContractVersion: "1.0",
		CanonicalVersion: "1",
		Initiator: "agent-payment-01",
		Recipient: "agent-secondary-02",
		RequestedCapability: "repository.write",
		AuthorizedCapabilities: []Capability{"repository.write"},
		Payload: map[string]interface{}{},
		PolicyID: "POLICY-017",
		PolicyVersion: 17,
		Provenance: map[string]interface{}{"source": "agent-payment-01", "destination": "agent-secondary-02"},
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
		ReplayProtection: ReplayProtection{Nonce: "abv1-10-nested-signature-binding", Timestamp: time.Now().UTC()},
	}
	parent.PayloadDigest = calculatePayloadDigest(parent.Payload)
	parent.DownstreamHandoff = &HandoffContract{
		ContractVersion: "1.0",
		CanonicalVersion: "1",
		Initiator: "agent-secondary-02",
		Recipient: "agent-tertiary-03",
		RequestedCapability: "repository.write",
		AuthorizedCapabilities: []Capability{"repository.write"},
		Payload: map[string]interface{}{},
		PolicyID: "POLICY-017",
		PolicyVersion: 17,
		Provenance: map[string]interface{}{"source": "agent-secondary-02", "destination": "agent-tertiary-03"},
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
		ReplayProtection: ReplayProtection{Nonce: "abv1-10-nested-child", Timestamp: time.Now().UTC()},
	}
	parent.DownstreamHandoff.PayloadDigest = calculatePayloadDigest(parent.DownstreamHandoff.Payload)
	parent.Signature = setup.HandoffValidator.SignContract(parent)

	parent.DownstreamHandoff.Recipient = "untrusted-agent"
	if setup.HandoffValidator.VerifyContractSignature(parent) {
		t.Fatal("expected parent signature to fail after downstream recipient mutation")
	}
}

func TestHandoffRejectsRecipientMismatch(t *testing.T) {
	setup := SetupDemoEnvironment()
	parent := &HandoffContract{
		ContractVersion: "1.0",
		CanonicalVersion: "1",
		Initiator: "agent-payment-01",
		Recipient: "agent-secondary-02",
		RequestedCapability: "repository.write",
		AuthorizedCapabilities: []Capability{"repository.write"},
		Payload: map[string]interface{}{},
		PolicyID: "POLICY-017",
		PolicyVersion: 17,
		Provenance: map[string]interface{}{"source": "agent-payment-01", "destination": "agent-other-99"},
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
		ReplayProtection: ReplayProtection{Nonce: "abv1-10-recipient-mismatch", Timestamp: time.Now().UTC()},
	}
	parent.PayloadDigest = calculatePayloadDigest(parent.Payload)
	parent.Signature = setup.HandoffValidator.SignContract(parent)
	validation := setup.HandoffValidator.ValidateHandoff(parent)
	if validation.Valid || !validation.ProvenanceMismatch {
		t.Fatal("expected recipient mismatch to be rejected")
	}
}

// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package investordemo

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func TestLoadTrustedEvidenceKey(t *testing.T) {
	public, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("NAEOS_EVIDENCE_TRUSTED_PUBLIC_KEY", base64.RawStdEncoding.EncodeToString(public))

	got := loadTrustedEvidenceKey()
	if len(got) != ed25519.PublicKeySize || string(got) != string(public) {
		t.Fatalf("trusted evidence key mismatch")
	}
}

func TestLoadTrustedEvidenceKeyRejectsInvalidInput(t *testing.T) {
	t.Setenv("NAEOS_EVIDENCE_TRUSTED_PUBLIC_KEY", "not-a-key")
	if got := loadTrustedEvidenceKey(); got != nil {
		t.Fatalf("expected invalid trust anchor to be rejected")
	}
}

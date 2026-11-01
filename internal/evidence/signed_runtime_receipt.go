// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// SignedDurableRuntimeReceipt authenticates a durable receipt with an
// operator-managed Ed25519 key. The public key is deliberately not embedded:
// verifiers must supply an independently trusted key.
type SignedDurableRuntimeReceipt struct {
	Receipt   DurableRuntimeReceipt `json:"receipt"`
	Algorithm string                `json:"algorithm"`
	Issuer    string                `json:"issuer"`
	KeyID     string                `json:"key_id"`
	Signature string                `json:"signature"`
}

// SignDurableRuntimeReceipt signs the canonical receipt identity and metadata.
func SignDurableRuntimeReceipt(receipt DurableRuntimeReceipt, privateKey ed25519.PrivateKey, issuer, keyID string) (SignedDurableRuntimeReceipt, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return SignedDurableRuntimeReceipt{}, fmt.Errorf("invalid Ed25519 private key length")
	}
	if strings.TrimSpace(issuer) == "" || strings.TrimSpace(keyID) == "" {
		return SignedDurableRuntimeReceipt{}, fmt.Errorf("issuer and key_id are required")
	}
	if receipt.ReceiptID != durableReceiptID(receiptWithoutID(receipt)) {
		return SignedDurableRuntimeReceipt{}, fmt.Errorf("durable runtime receipt identity mismatch")
	}
	signed := SignedDurableRuntimeReceipt{Receipt: receipt, Algorithm: "Ed25519", Issuer: issuer, KeyID: keyID}
	payload, err := canonicalSignedReceipt(signed)
	if err != nil {
		return SignedDurableRuntimeReceipt{}, err
	}
	signed.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, payload))
	return signed, nil
}

// VerifySignedDurableRuntimeReceipt verifies both receipt/ledger integrity and
// authenticity against an independently configured trust anchor.
func VerifySignedDurableRuntimeReceipt(signed SignedDurableRuntimeReceipt, ledger *DurableRuntimeEventLedger, runID string, trusted ed25519.PublicKey, issuer, keyID string) error {
	if len(trusted) != ed25519.PublicKeySize {
		return fmt.Errorf("trusted Ed25519 public key is required")
	}
	if signed.Algorithm != "Ed25519" || signed.Issuer != issuer || signed.KeyID != keyID {
		return fmt.Errorf("durable runtime receipt trust metadata mismatch")
	}
	if err := VerifyDurableRuntimeReceipt(signed.Receipt, ledger, runID); err != nil {
		return err
	}
	sig, err := base64.StdEncoding.DecodeString(signed.Signature)
	if err != nil {
		return fmt.Errorf("decode durable runtime receipt signature: %w", err)
	}
	payload, err := canonicalSignedReceipt(signed)
	if err != nil {
		return err
	}
	if !ed25519.Verify(trusted, payload, sig) {
		return fmt.Errorf("durable runtime receipt signature verification failed")
	}
	return nil
}

func canonicalSignedReceipt(signed SignedDurableRuntimeReceipt) ([]byte, error) {
	signed.Signature = ""
	data, err := json.Marshal(signed)
	if err != nil {
		return nil, fmt.Errorf("encode signed durable runtime receipt: %w", err)
	}
	// Domain separation prevents this signature from being reused for another
	// NAEOS artifact type with an identical JSON representation.
	digest := sha256.Sum256(append([]byte("naeos:signed-durable-runtime-receipt:v1:"), data...))
	return digest[:], nil
}

// ParseEd25519PublicKey accepts raw, base64, or hexadecimal public key material.
func ParseEd25519PublicKey(data []byte) (ed25519.PublicKey, error) {
	if len(data) == ed25519.PublicKeySize {
		return ed25519.PublicKey(append([]byte(nil), data...)), nil
	}
	text := strings.TrimSpace(string(data))
	if decoded, err := base64.StdEncoding.DecodeString(text); err == nil && len(decoded) == ed25519.PublicKeySize {
		return ed25519.PublicKey(decoded), nil
	}
	if decoded, err := hex.DecodeString(text); err == nil && len(decoded) == ed25519.PublicKeySize {
		return ed25519.PublicKey(decoded), nil
	}
	return nil, fmt.Errorf("unsupported Ed25519 public key encoding")
}

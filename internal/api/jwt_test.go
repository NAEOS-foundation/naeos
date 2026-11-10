// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func TestJWTGenerateAndValidate(t *testing.T) {
	validator := NewJWTValidator("test-secret-key")

	claims := &JWTClaims{
		Sub:  "user-123",
		Name: "Test User",
	}

	token, err := validator.Generate(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("expected token to be generated")
	}

	validated, err := validator.Validate(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if validated.Sub != "user-123" {
		t.Errorf("expected sub 'user-123', got %s", validated.Sub)
	}

	if validated.Name != "Test User" {
		t.Errorf("expected name 'Test User', got %s", validated.Name)
	}
}

func TestJWTRejectsUnsupportedHeaderAlgorithm(t *testing.T) {
	validator := NewJWTValidator("test-secret-key")
	claims := &JWTClaims{
		Sub:       "user-123",
		Name:      "Test User",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		IssuedAt:  time.Now().Unix(),
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

	for _, algorithm := range []string{"none", "RS256", "HS384"} {
		t.Run(algorithm, func(t *testing.T) {
			headerJSON, err := json.Marshal(map[string]string{"alg": algorithm, "typ": "JWT"})
			if err != nil {
				t.Fatalf("marshal header: %v", err)
			}
			header := base64.RawURLEncoding.EncodeToString(headerJSON)
			signingInput := header + "." + payload
			token := signingInput + "." + validator.sign(signingInput)
			if _, err := validator.Validate(token); err == nil {
				t.Fatalf("Validate() accepted header algorithm %q", algorithm)
			}
		})
	}
}

func TestJWTExpiredToken(t *testing.T) {
	validator := NewJWTValidator("test-secret-key")

	claims := &JWTClaims{
		Sub:       "user-123",
		Name:      "Test User",
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
	}

	token, err := validator.Generate(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = validator.Validate(token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestJWTInvalidSignature(t *testing.T) {
	validator1 := NewJWTValidator("secret-1")
	validator2 := NewJWTValidator("secret-2")

	claims := &JWTClaims{
		Sub:  "user-123",
		Name: "Test User",
	}

	token, err := validator1.Generate(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = validator2.Validate(token)
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestJWTInvalidFormat(t *testing.T) {
	validator := NewJWTValidator("test-secret-key")

	_, err := validator.Validate("invalid.token")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

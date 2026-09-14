// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package api

import (
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
